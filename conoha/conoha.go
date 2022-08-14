package conoha

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/startstop"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/images"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/pagination"
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

	log.Println("Start!")

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

	var uuid string
	pager := servers.List(computeClient, nil)
	pager.EachPage(func(page pagination.Page) (bool, error) {
		serverList, err := servers.ExtractServers(page)
		if err != nil {
			fmt.Println(err)
			return false, err
		}
		for _, s := range serverList {
			//TODO 好きな名前を設定できるように
			if s.Metadata["instance_name_tag"] == "ConohaChatOps" {
				uuid = s.ID
			}
		}
		return true, nil
	})

	err = startstop.Stop(computeClient, uuid).ExtractErr()
	if err != nil {
		log.Fatalf("Stop a Server Failed: %v", err)
		return err
	}

	err = servers.WaitForStatus(computeClient, uuid, "SHUTOFF", 300)
	if err != nil {
		log.Fatalf("Unable to create for server: %v", err)
		return err
	}

	return nil
}

func DeleteServer() error {
	eo := gophercloud.EndpointOpts{
		Type:   "compute",
		Region: os.Getenv("CONOHA_ENDPOINT"),
	}
	computeClient, err := openstack.NewComputeV2(conohaClient, eo)
	if err != nil {
		log.Fatalf("Compute Client Failed: %v", err)
		return err
	}

	var uuid string
	pager := servers.List(computeClient, nil)
	pager.EachPage(func(page pagination.Page) (bool, error) {
		serverList, err := servers.ExtractServers(page)
		if err != nil {
			fmt.Println(err)
			return false, err
		}
		for _, s := range serverList {
			//TODO 好きな名前を設定できるように
			if s.Metadata["instance_name_tag"] == "ConohaChatOps" {
				uuid = s.ID
			}
		}
		return true, nil
	})

	err = servers.Delete(computeClient, uuid).ExtractErr()
	if err != nil {
		log.Fatalf("Delete a Server Failed: %v", err)
		return err
	}

	err = servers.WaitForStatus(computeClient, uuid, "DELETED", 300)
	if err != nil {
		log.Fatalf("Unable to create for server: %v", err)
		return err
	}

	return nil
}

func CleateImage() error {
	eo := gophercloud.EndpointOpts{
		Type:   "compute",
		Region: os.Getenv("CONOHA_ENDPOINT"),
	}

	computeClient, err := openstack.NewComputeV2(conohaClient, eo)
	if err != nil {
		log.Fatalf("Compute Client Failed: %v", err)
		return err
	}

	var uuid string
	pager := servers.List(computeClient, nil)
	pager.EachPage(func(page pagination.Page) (bool, error) {
		serverList, err := servers.ExtractServers(page)
		if err != nil {
			fmt.Println(err)
			return false, err
		}
		for _, s := range serverList {
			//TODO 好きな名前を設定できるように
			if s.Metadata["instance_name_tag"] == "ConohaChatOps" {
				uuid = s.ID
			}
		}
		return true, nil
	})

	snapshotOpts := servers.CreateImageOpts{
		Name: "ConohaChatOps-snapshot",
	}

	imageID, err := servers.CreateImage(computeClient, uuid, snapshotOpts).ExtractImageID()
	if err != nil {
		log.Fatalf("Create Image Failed: %v", err)
		return err
	}

	err = WaitForImage(computeClient, imageID)
	if err != nil {
		log.Fatalf("Unable to save for server image: %v", err)
		return err
	}

	return nil
}

func WaitForImage(client *gophercloud.ServiceClient, imageID string) error {
	for {
		image, err := images.Get(client, imageID).Extract()
		if err != nil {
			return err
		}
		if image.Status == "ACTIVE" {
			break
		}
		time.Sleep(time.Second * 10)
	}
	return nil
}
