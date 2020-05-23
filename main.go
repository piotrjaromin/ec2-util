package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	ec2aws "github.com/piotrjaromin/ec2-util/aws"
	"github.com/piotrjaromin/ec2-util/mongo"
)

const clusterNameTag = "Cluster_Name"
const sdServiceIdTag = "Sd_Service_Id"
const mongoPort = "27017"

func main() {
	currentHost, allHosts, err := handleServiceDiscovery()
	if err != nil {
		panic(err)
	}

	if err = mongo.InitReplicaSet(currentHost, allHosts); err != nil {
		panic(err)
	}
}

func handleServiceDiscovery() (string, []string, error) {
	emptyIPs := []string{}
	currentHost := ""

	sessNoRegion := ec2aws.NewSessionWithoutRegion()
	metaSvc := ec2metadata.New(sessNoRegion)

	metaDoc, err := metaSvc.GetInstanceIdentityDocument()
	if err != nil {
		return currentHost, emptyIPs, fmt.Errorf("Unable to read metadata of current ec2. %s", err.Error())
	}

	sess := ec2aws.NewSession(metaDoc.Region)

	ec2Svc := ec2.New(sess)
	sdSvc := servicediscovery.New(sess)

	currentHost = metaDoc.PrivateIP
	instanceID := metaDoc.InstanceID
	tags, err := ec2aws.GetTags(ec2Svc, instanceID)
	if err != nil {
		return currentHost, emptyIPs, fmt.Errorf("Unable to read tags from ec2 instance. %s", err.Error())
	}

	sdServiceID, ok := tags[sdServiceIdTag]
	if !ok {
		return currentHost, emptyIPs, fmt.Errorf("Missing service id from instance tags")
	}

	clusterName, ok := tags[clusterNameTag]
	if !ok {
		return currentHost, emptyIPs, fmt.Errorf("Missing clusterName name from instance tags")
	}

	clusterIps, err := ec2aws.FindEc2IpsByTag(ec2Svc, clusterNameTag, clusterName)
	if err != nil {
		return currentHost, emptyIPs, fmt.Errorf("Unable to cluster get ips. %s", err.Error())
	}

	dnsName, err := ec2aws.GetServiceDNS(sdSvc, sdServiceID)
	if err != nil {
		return currentHost, emptyIPs, fmt.Errorf("Unable to get service discovery dns name. %s", err.Error())
	}

	// each mongo node has single domain in sd group, so 1 domain = 1 ip
	newIPs, err := ec2aws.RegisterInstacesWithIps(sdSvc, dnsName, sdServiceID, []string{currentHost})
	if err != nil {
		return currentHost, clusterIps, fmt.Errorf("Unable to register instance in sd group. %s", err.Error())
	}

	if len(newIPs) > 0 {
		log.Printf("Found new ips for service discovery group: %s\n", strings.Join(newIPs, ","))
	}

	return currentHost, clusterIps, nil
}
