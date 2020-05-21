package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func GetAsgIPs(ec2Svc *ec2.EC2, asgSvc *autoscaling.AutoScaling, asgName string) ([]string, error) {
	ips := []string{}
	asgInpput := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{&asgName},
	}

	asgOutput, err := asgSvc.DescribeAutoScalingGroups(asgInpput)
	if err != nil {
		return ips, fmt.Errorf("Unable to describe autosclaing group. %s", err.Error())
	}

	instanceIds := []*string{}
	for _, group := range asgOutput.AutoScalingGroups {
		for _, inst := range group.Instances {
			instanceIds = append(instanceIds, inst.InstanceId)
		}
	}

	ec2Reservations, err := GetEC2Reservations(ec2Svc, instanceIds)
	if err != nil {
		return ips, fmt.Errorf("Unable to ec2s instance from asg. %s", err.Error())
	}

	for _, reservation := range ec2Reservations {
		for _, ec2Instance := range reservation.Instances {
			for _, network := range ec2Instance.NetworkInterfaces {
				ips = append(ips, *network.PrivateIpAddress)
			}
		}
	}
	return ips, nil
}
