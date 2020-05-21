package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	ec2aws "github.com/piotrjaromin/ec2-util/aws"
	"github.com/piotrjaromin/ec2-util/mongo"
)

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
	// 1. Get instance tags
	// 2. Get from tags: serviceId, asgId
	// 3. Get IPS from asg
	// 4. Resolve serviceDiscovery dns name based on serviceId
	// 5. Add missing ips to serviceDiscovery
	// 6. Add missing ips to mongodb replicate set - TODO

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
	asgSvc := autoscaling.New(sess)

	currentHost = metaDoc.PrivateIP
	instanceID := metaDoc.InstanceID
	tags, err := ec2aws.GetTags(ec2Svc, instanceID)
	if err != nil {
		return currentHost, emptyIPs, fmt.Errorf("Unable to read tags from ec2 instance. %s", err.Error())
	}

	sdServiceID, ok := tags["Sd_Service_Id"]
	if !ok {
		return currentHost, emptyIPs, fmt.Errorf("Missing service id from instance tags")
	}

	asgName, ok := tags["aws:autoscaling:groupName"]
	if !ok {
		return currentHost, emptyIPs, fmt.Errorf("Missing ASG name from instance tags")
	}

	asgIPs, err := ec2aws.GetAsgIPs(ec2Svc, asgSvc, asgName)
	if err != nil {
		return currentHost, emptyIPs, fmt.Errorf("Unable to get asg ips. %s", err.Error())
	}

	dnsName, err := ec2aws.GetServiceDNS(sdSvc, sdServiceID)
	if err != nil {
		return currentHost, emptyIPs, fmt.Errorf("Unable to get service discovery dns name. %s", err.Error())
	}

	newIPs, err := ec2aws.RegisterInstacesWithIps(sdSvc, mongoPort, dnsName, sdServiceID, asgIPs)
	if err != nil {
		return currentHost, asgIPs, fmt.Errorf("Unable to register instance in sd group. %s", err.Error())
	}

	if len(newIPs) > 0 {
		log.Printf("Found new ips for service discovery group: %s\n", strings.Join(newIPs, ","))
	}

	return currentHost, asgIPs, nil
}
