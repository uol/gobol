package embassy

import (
	"fmt"
	"strings"

	"github.com/hashicorp/consul/api"
)

var (
	// ErrNoServiceMatchesQuery means the query did not find any matches in Consul
	ErrNoServiceMatchesQuery = fmt.Errorf("No services match the query")

	// ErrServicePortNotUnique means there are diferent ports for the same service
	ErrServicePortNotUnique = fmt.Errorf("Service port is not unique!")
)

// GetServicePort returns the port associated with a given service
func GetServicePort(service, tag string) (int, error) {
	services, _, err := consulClient.Catalog().Service(service, tag, &api.QueryOptions{})
	if err != nil {
		// Consul should not error
		return 0, err
	}
	if len(services) == 0 {
		// There should be someone answering for the service
		return 0, ErrNoServiceMatchesQuery
	}
	port := services[0].ServicePort
	for _, serv := range services {
		// There should be a unique port for the service
		if serv.ServicePort != port {
			return 0, ErrServicePortNotUnique
		}
	}
	return port, nil
}

// GetServiceHost returns the hostname of a given service
func GetServiceHost(service, tag string) string {
	return strings.TrimPrefix(
		fmt.Sprintf("%s.%s.service.%s", tag, service, consulDomain),
		".",
	)
}

func GetSRV(srvName, tag string) (string, int, error) {

	datacenters, err := consulClient.Catalog().Datacenters()
	if err != nil {
		return "", -1, err
	}

	serviceEntries, _, err := consulClient.Health().Service(srvName, tag, false, nil)
	if err != nil {
		return "", -1, err
	}

	var datacenter string
	if serviceEntries == nil || len(serviceEntries) < 1 {
		for _, dc := range datacenters {
			if len(serviceEntries) > 0 {
				break
			}
			query := &api.QueryOptions{Datacenter: dc}
			serviceEntries, _, err = consulClient.Health().Service(srvName, tag, false, query)
			if err != nil {
				return "", -1, err
			}
			if len(serviceEntries) > 0 {
				datacenter = dc
			}
		}
	}

	var dns string
	var port int
	for _, serviceEntry := range serviceEntries {
		if serviceEntry != nil {
			if serviceEntry.Service != nil {
				dns = strings.TrimPrefix(
					fmt.Sprintf("%s.%s.service.%s.%s", tag, serviceEntry.Service.Service, datacenter, consulDomain),
					".",
				)
				dns = strings.Replace(dns, "..", ".", -1)
				port = serviceEntry.Service.Port
			}
		}

	}
	if dns == "" || port < 1 {
		return "", -1, ErrNoServiceMatchesQuery
	}

	return dns, port, nil

}
