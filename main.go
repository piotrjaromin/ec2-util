package main

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/piotrjaromin/ec2-util/pkg"
)

func main() {
	// 1. Get instance tags
	// 2. Get from tags: serviceId, asgId
	// 3. Get IPS from asg
	// 4. Resolve serviceDiscovery dns name based on serviceId
	// 5. Add missing ips to serviceDiscovery
	// 6. Add missing ips to mongodb replicate set - TODO

	sess := pkg.NewSession()

	metaSvc := ec2metadata.New(sess)
	ec2Svc := ec2.New(sess)
	sdSvc := servicediscovery.New(sess)
	asgSvc := autoscaling.New(sess)

	metaDoc, err := metaSvc.GetInstanceIdentityDocument()
	if err != nil {
		panic(err)
	}

	instanceID := metaDoc.InstanceID
	tags, err := pkg.GetTags(ec2Svc, instanceID)
	if err != nil {
		panic(fmt.Errorf("Unable to read tags from ec2 instance. %s", err.Error()))
	}

	sdServiceID, ok := tags["Sd_Service_Id"]
	if !ok {
		panic(fmt.Errorf("Missing service id from instance tags"))
	}

	asgName, ok := tags["aws:autoscaling:groupName"]
	if !ok {
		panic(fmt.Errorf("Missing ASG name from instance tags"))
	}

	ips, err := pkg.GetAsgIPs(ec2Svc, asgSvc, asgName)
	if err != nil {
		panic(fmt.Errorf("Unable to get asg ips. %s", err.Error()))
	}

	dnsName, err := pkg.GetServiceDNS(sdSvc, sdServiceID)
	if err != nil {
		panic(fmt.Errorf("Unable to get service discovery dns name. %s", err.Error()))
	}

	newIPs, err := pkg.RegisterInstacesWithIps(sdSvc, sdServiceID, dnsName, ips)
	if err != nil {
		panic(fmt.Errorf("Unable to register instance in sd group. %s", err.Error()))
	}

	fmt.Printf("New ips found: %s", strings.Join(newIPs, ","))
}
