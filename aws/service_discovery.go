package aws

import (
	"log"
	"net"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
)

func RegisterInstacesWithIps(sdSvc *servicediscovery.ServiceDiscovery, port, serviceDns, serviceID string, asgIPs []string) ([]string, error) {
	newInstanceIPs := []string{}
	// dns name contains service discovery ips
	sdIPs, err := net.LookupHost(serviceDns)
	if err != nil {
		log.Printf("Unable to resolve service dns name to ips, using empty list as default. err: %s\n", err)
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
			if err := registerInstance(sdSvc, port, asgIP, serviceID); err != nil {
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

func registerInstance(sdSvc *servicediscovery.ServiceDiscovery, port, ipV4, serviceID string) error {
	instanceID := toInstanceIDFromIP(ipV4)
	input := &servicediscovery.RegisterInstanceInput{
		Attributes: map[string]*string{
			"AWS_INSTANCE_IPV4": &ipV4,
			"AWS_INSTANCE_PORT": &port,
		},
		CreatorRequestId: aws.String("ecs-utils"),
		InstanceId:       &instanceID,
		ServiceId:        &serviceID,
	}

	_, err := sdSvc.RegisterInstance(input)
	if err != nil {
		return err
	}

	return nil
}

func removeInstance(sdSvc *servicediscovery.ServiceDiscovery, ipV4, serviceID string) error {
	instanceID := toInstanceIDFromIP(ipV4)
	input := &servicediscovery.DeregisterInstanceInput{
		InstanceId: &instanceID,
		ServiceId:  &serviceID,
	}

	_, err := sdSvc.DeregisterInstance(input)
	if err != nil {
		return err
	}

	return nil
}

func toInstanceIDFromIP(ip string) string {
	return strings.Replace(ip, ".", "-", -1)
}
