package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
)

func GetTags(ec2Svc *ec2.EC2, instanceId string) (map[string]string, error) {
	tags := map[string]string{}

	reservations, err := GetEC2Reservations(ec2Svc, []*string{&instanceId})
	if err != nil {
		return tags, fmt.Errorf("Unable to get ec2 data. %s", err.Error())
	}

	if len(reservations) != 1 {
		return tags, fmt.Errorf("Invalid amount of reservations found in getTag, got %d", len(reservations))
	}

	res := reservations[0]
	if len(res.Instances) != 1 {
		return tags, fmt.Errorf("Invalid amount of instances found in getTag, got %d", len(res.Instances))
	}

	instance := res.Instances[0]
	for _, tag := range instance.Tags {
		tags[*tag.Key] = *tag.Value
	}

	return tags, nil
}

func GetEC2Reservations(ec2Svc *ec2.EC2, instanceIds []*string) ([]*ec2.Reservation, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: instanceIds,
	}

	instanceOutput, err := ec2Svc.DescribeInstances(input)
	if err != nil {
		return []*ec2.Reservation{}, err
	}

	return instanceOutput.Reservations, nil
}
