package pkg

import (
	"fmt"
	"net"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
)

func RegisterInstacesWithIps(sdSvc *servicediscovery.ServiceDiscovery, serviceDns, serviceID string, asgIPs []string) ([]string, error) {
	newInstanceIPs := []string{}
	// dns name contains service discovery ips
	sdIPs, err := net.LookupHost(serviceDns)
	if err != nil {
		return newInstanceIPs, fmt.Errorf("Unable to resolve service dns name to ips. %s", err)
	}

	for _, sdIP := range sdIPs {
		if !contains(asgIPs, sdIP) {
			if err := removeInstance(sdSvc, sdIP, serviceID); err != nil {
				return newInstanceIPs, err
			}
		}
	}

	for _, asgIP := range asgIPs {
		if !contains(sdIPs, asgIP) {
			newInstanceIPs = append(newInstanceIPs, asgIP)
			if err := registerInstance(sdSvc, asgIP, serviceID); err != nil {
				return newInstanceIPs, err
			}

		}
	}

	return newInstanceIPs, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func registerInstance(sdSvc *servicediscovery.ServiceDiscovery, ipV4, serviceID string) error {
	input := &servicediscovery.RegisterInstanceInput{
		Attributes: map[string]*string{
			"AWS_INSTANCE_IPV4": &ipV4,
		},
		CreatorRequestId: aws.String("ecs-utils"),
		InstanceId:       &ipV4,
		ServiceId:        &serviceID,
	}

	_, err := sdSvc.RegisterInstance(input)
	if err != nil {
		return err
	}

	return nil
}

func removeInstance(sdSvc *servicediscovery.ServiceDiscovery, ipV4, serviceID string) error {
	input := &servicediscovery.DeregisterInstanceInput{
		InstanceId: &ipV4,
		ServiceId:  &serviceID,
	}

	_, err := sdSvc.DeregisterInstance(input)
	if err != nil {
		return err
	}

	return nil
}
