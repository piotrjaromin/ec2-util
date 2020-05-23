package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var ec2RunningState = "running"

// FindEc2IpsByTag searches for private ips with given tagName and tagValue
func FindEc2IpsByTag(ec2Svc *ec2.EC2, tagName, tagValue string) ([]string, error) {
	ips := []string{}

	instances, err := findRunningEc2ByTag(ec2Svc, tagName, tagValue)
	if err != nil {
		return ips, fmt.Errorf("Unable to get ec2 ips. %s", err.Error())
	}

	for _, instance := range instances {
		ips = append(ips, *instance.PrivateIpAddress)
	}
	return ips, nil
}

func findRunningEc2ByTag(ec2Svc *ec2.EC2, tagName, tagValue string) ([]*ec2.Instance, error) {
	instances := []*ec2.Instance{}

	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", tagName)),
				Values: []*string{&tagValue},
			},
		},
	}

	res, err := ec2Svc.DescribeInstances(params)
	if err != nil {
		return instances, err
	}

	for _, reserv := range res.Reservations {
		for _, instance := range reserv.Instances {
			if *instance.State.Name == ec2RunningState {
				instances = append(instances, instance)
			}
		}
	}

	return instances, nil
}
