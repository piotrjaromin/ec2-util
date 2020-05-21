package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/servicediscovery"
)

// GetServiceDNS based on serviceID returns full service DNS name
func GetServiceDNS(sdSvc *servicediscovery.ServiceDiscovery, serviceID string) (string, error) {
	getServiceInput := &servicediscovery.GetServiceInput{
		Id: &serviceID,
	}

	log.Printf("Searching for %s\n", serviceID)
	getServiceOutput, err := sdSvc.GetService(getServiceInput)
	if err != nil {
		return "", fmt.Errorf("Unable to get service for service discovery data %s", err)
	}

	svc := getServiceOutput.Service

	getNamespaceInput := &servicediscovery.GetNamespaceInput{
		Id: svc.NamespaceId,
	}

	getNamespaceOutput, err := sdSvc.GetNamespace(getNamespaceInput)
	if err != nil {
		return "", fmt.Errorf("Unable to get service discovery data %s", err)
	}

	namespace := getNamespaceOutput.Namespace
	return fmt.Sprintf("%s.%s", *svc.Name, *namespace.Name), nil
}
