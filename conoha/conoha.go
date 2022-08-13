package conoha

import (
	"log"
	"os"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/startstop"
)

var conohaClient *gophercloud.ProviderClient

func Init() {
	opts := gophercloud.AuthOptions{
		IdentityEndpoint: "https://identity." + os.Getenv("CONOHA_ENDPOINT") + ".conoha.io/v2.0",
		Username:         os.Getenv("CONOHA_USERNAME"),
		TenantName:       os.Getenv("CONOHA_TENANTNAME"),
		Password:         os.Getenv("CONOHA_PASSWORD"),
	}

	client, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		log.Fatalf("Authentication Failed: %v", err)
		return
	}

	conohaClient = client

	//DEBUG
	log.Println(conohaClient)

	return
}

func CloseServer() error {
	eo := gophercloud.EndpointOpts{
		Type:   "compute",
		Region: os.Getenv("CONOHA_ENDPOINT"),
	}
	computeClient, err := openstack.NewComputeV2(conohaClient, eo)
	if err != nil {
		log.Fatalf("Compute Client Failed: %v", err)
		return err
	}

	err = startstop.Stop(computeClient, os.Getenv("CONOHA_SERVERID")).ExtractErr()
	if err != nil {
		log.Fatalf("Stop a Server Failed: %v", err)
		return err
	}
	return nil
}
