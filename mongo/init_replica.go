package mongo

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/juju/replicaset"
)

const replicaName = "rs0"
const localhost = "localhost:27017"

type mongoSession interface {
	Run(cmd interface{}, result interface{}) error
}

// Not using juju implementation for everything because it was failing for some reason..

// InitReplicaSet initializes replicaSet, if replica set is working then removes unhealthy nodes, and adds new ones
func InitReplicaSet(currentHost string, hosts map[string]time.Time) error {
	info := &mgo.DialInfo{
		Addrs:   []string{localhost},
		Timeout: 25 * time.Second,
		Direct:  true,
	}

	session, err := mgo.DialWithInfo(info)
	if err != nil {
		return fmt.Errorf("Unable to connect to mongo. %s", err.Error())
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	if !isReplicaSetActive(session) {
		log.Printf("Init first time replicaset\n")
		if err = firstTimeInit(session, currentHost, hosts); err != nil {
			return fmt.Errorf("Unable to firstTimeInit mongo. %s", err.Error())
		}

		// wait a moment, if this is first time, then replica needs to be initalized
		time.Sleep(10 * time.Second)
	}

	isMasterRes, err := replicaset.IsMaster(session)
	if err != nil {
		return fmt.Errorf("Unable check if current node is Master. %s", err.Error())
	}

	if !isMasterRes.IsMaster {
		log.Printf("Not on master, skipping removal/adding of members")
		return nil
	}

	unhealthyIds, err := getUnhealthyMemberIds(session)
	if err != nil {
		return fmt.Errorf("Unable get unhealthy ids. %s", err.Error())
	}

	if len(unhealthyIds) > 0 {
		log.Printf("Found unhealthy Replicaset nodes. Ids to removal %+v\n", unhealthyIds)
		err = removeUnhealthy(session, unhealthyIds)
		if err != nil {
			log.Printf("Unable to remove unhealthy members. %s\n", err)
		}
	}

	return addNewMembers(session, hosts)
}

func firstTimeInit(session mongoSession, currentHost string, hosts map[string]time.Time) error {
	oldestHost, oldestLaunchTime := getOldestHost(hosts)

	// if instances started more than 15 minutes ago it might be that replica is already running
	_15MinAgo := time.Now().Add(time.Duration(-15) * time.Minute)
	isChanceThatReplicaSetIsRunning := oldestLaunchTime.Before(_15MinAgo)
	if isChanceThatReplicaSetIsRunning {
		return nil
	}

	// when we init new replica set, just do it for single node
	if currentHost != oldestHost {
		return nil
	}

	members := []bson.M{{"_id": 0, "host": fmt.Sprintf("%s:27017", currentHost)}}
	config := bson.M{
		"_id":     replicaName,
		"members": members,
	}

	result := bson.M{}
	if err := session.Run(bson.M{"replSetInitiate": config}, &result); err != nil {
		return err
	}

	return nil
}

func getOldestHost(hosts map[string]time.Time) (string, time.Time) {
	oldestTime := time.Now()
	oldestHost := ""

	for host, startTime := range hosts {
		if startTime.Before(oldestTime) {
			oldestTime = startTime
			oldestHost = host
		}
	}

	return oldestHost, oldestTime
}

func getHosts(hosts map[string]time.Time) []string {
	allHosts := []string{}

	for host := range hosts {
		allHosts = append(allHosts, host)
	}

	return allHosts
}

func isReplicaSetActive(session *mgo.Session) bool {
	result := bson.M{}
	if err := session.Run(bson.M{"replSetGetStatus": bson.M{}}, &result); err != nil {
		return false
	}

	return true
}

func getUnhealthyMemberIds(session *mgo.Session) ([]int, error) {
	status := ReplSetStatus{}
	err := session.Run(bson.M{"replSetGetStatus": bson.M{}}, &status)

	unhealthyIds := []int{}
	for _, member := range status.Members {
		if member.Health != healthyState {
			unhealthyIds = append(unhealthyIds, member.ID)
		}
	}

	return unhealthyIds, err
}

func removeUnhealthy(session *mgo.Session, unhealthyIds []int) error {
	errString := ""

	currentMembers, err := replicaset.CurrentMembers(session)
	if err != nil {
		return err
	}

	for _, member := range currentMembers {
		if hasID(member.Id, unhealthyIds) {
			err := replicaset.Remove(session, member.Address)
			if err != nil {
				errString += err.Error()
			}
		}
	}

	if len(errString) > 0 {
		return fmt.Errorf("Unable to remove unhealthy members. %s", errString)
	}

	return nil
}

func addNewMembers(session *mgo.Session, newHosts map[string]time.Time) error {
	errString := ""

	currentMembers, err := replicaset.CurrentMembers(session)
	if err != nil {
		return err
	}

	for host := range newHosts {
		if !hasHost(host, currentMembers) {
			log.Printf("Adding new member: %s\n", host)
			member := replicaset.Member{
				Address: host,
			}

			err := replicaset.Add(session, member)
			if err != nil {
				errString += err.Error()
			}
		}
	}

	if len(errString) > 0 {
		return fmt.Errorf("Unable to add new members. %s", errString)
	}

	return nil
}

func hasHost(host string, members []replicaset.Member) bool {
	for _, member := range members {
		// just in case host can contains port or smth
		if strings.Contains(member.Address, host) {
			return true
		}
	}

	return false
}

func hasID(id int, members []int) bool {
	for _, member := range members {
		if id == member {
			return true
		}
	}

	return false
}
