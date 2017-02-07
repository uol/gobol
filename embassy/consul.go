package embassy

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/consul/api"
)

var consulClient *api.Client
var consulDomain string

type ConsulSettings struct {
	Address  string
	Scheme   string
	Token    string
	CertFile string
	KeyFile  string
	CAFile   string
	Domain   string
}

func NewConnection(settings ConsulSettings) error {

	cp := x509.NewCertPool()

	data, err := ioutil.ReadFile(settings.CAFile)

	if err != nil {
		return fmt.Errorf("Failed to read CA file: %v", err)
	}

	if !cp.AppendCertsFromPEM(data) {
		return fmt.Errorf("Failed to parse any CA certificates")
	}

	cert, err := tls.LoadX509KeyPair(settings.CertFile, settings.KeyFile)

	if err != nil {
		return err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      cp,
		},
	}

	config := &api.Config{
		Address:    settings.Address,
		Scheme:     settings.Scheme,
		Token:      settings.Token,
		HttpClient: &http.Client{Transport: tr},
	}

	client, err := api.NewClient(config)

	if err != nil {
		return err
	}

	consulClient = client

	if settings.Domain != "" {
		consulDomain = settings.Domain
	} else {
		consulMap, err := consulClient.Agent().Self()
		if err != nil {
			return err
		}
		domain := consulMap["Config"]["Domain"]
		consulDomain = domain.(string)
	}

	return err

}
