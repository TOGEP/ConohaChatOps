package conoha

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/startstop"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
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

func GetImageRef() (string, error) {
	eo := gophercloud.EndpointOpts{
		Type:   "compute",
		Region: os.Getenv("CONOHA_ENDPOINT"),
	}
	computeClient, err := openstack.NewComputeV2(conohaClient, eo)
	if err != nil {
		log.Fatalf("Compute Client Failed: %v", err)
		return "", err
	}

	var imageRef string
	pager := images.ListDetail(computeClient, nil)
	pager.EachPage(func(page pagination.Page) (bool, error) {
		imageList, err := images.ExtractImages(page)
		if err != nil {
			fmt.Println(err)
			return false, err
		}
		for _, i := range imageList {
			if i.Name == "ConohaChatOps-snapshot" {
				imageRef = i.ID
			}
		}
		return true, nil
	})
	return imageRef, nil
}

func GetFlavorRef(memSize int64) (string, error) {
	eo := gophercloud.EndpointOpts{
		Type:   "compute",
		Region: os.Getenv("CONOHA_ENDPOINT"),
	}
	computeClient, err := openstack.NewComputeV2(conohaClient, eo)
	if err != nil {
		log.Fatalf("Compute Client Failed: %v", err)
		return "", err
	}

	var flavorRef string
	pager := flavors.ListDetail(computeClient, nil)
	pager.EachPage(func(page pagination.Page) (bool, error) {
		flavorList, err := flavors.ExtractFlavors(page)
		if err != nil {
			fmt.Println(err)
			return false, err
		}
		for _, f := range flavorList {
			//MEMO Conohaの旧プラン(Disk:50GB)のIDも残っている為DISKとRAMで判定
			if f.Disk == 100 && f.RAM == int(memSize)*1024 {
				flavorRef = f.ID
			}
		}
		return true, nil
	})
	return flavorRef, nil
}

func GetIPaddr() (string, error) {
	eo := gophercloud.EndpointOpts{
		Type:   "compute",
		Region: os.Getenv("CONOHA_ENDPOINT"),
	}
	computeClient, err := openstack.NewComputeV2(conohaClient, eo)
	if err != nil {
		log.Fatalf("Compute Client Failed: %v", err)
		return "", err
	}

	var ip string
	pager := servers.List(computeClient, nil)
	pager.EachPage(func(page pagination.Page) (bool, error) {
		serverList, err := servers.ExtractServers(page)
		if err != nil {
			fmt.Println(err)
			return false, err
		}
		for _, s := range serverList {
			for _, networkAddresses := range s.Addresses {
				for _, element := range networkAddresses.([]interface{}) {
					address := element.(map[string]interface{})
					if address["version"].(float64) == 4 {
						ip = address["addr"].(string)
					}
				}
			}
		}
		return true, nil
	})
	return ip, nil
}

func OpenServer(imageRef string, flavorRef string) error {
	eo := gophercloud.EndpointOpts{
		Type:   "compute",
		Region: os.Getenv("CONOHA_ENDPOINT"),
	}
	computeClient, err := openstack.NewComputeV2(conohaClient, eo)
	if err != nil {
		log.Fatalf("Compute Client Failed: %v", err)
		return err
	}

	co := servers.CreateOpts{
		Name:      "ConohaChatOps",
		ImageRef:  imageRef,
		FlavorRef: flavorRef,
		AdminPass: os.Getenv("CONOHA_PASSWORD"),
		Metadata: map[string]string{
			"instance_name_tag": "ConohaChatOps",
		},
	}

	server, err := servers.Create(computeClient, co).Extract()
	if err != nil {
		log.Fatalf("Create a Server Failed: %v", err)
		return err
	}

	err = servers.WaitForStatus(computeClient, server.ID, "ACTIVE", 300)
	if err != nil {
		log.Fatalf("Unable to create for server: %v", err)
		return err
	}

	return nil
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
		log.Fatalf("Unable to stop for server: %v", err)
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
		log.Fatalf("Unable to delete for server: %v", err)
		return err
	}

	return nil
}

func CreateImage() error {
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

	err = WaitForImage(computeClient, imageID, "ACTIVE")
	if err != nil {
		log.Fatalf("Unable to save for server image: %v", err)
		return err
	}

	return nil
}

func DeleteImage() error {
	eo := gophercloud.EndpointOpts{
		Type:   "compute",
		Region: os.Getenv("CONOHA_ENDPOINT"),
	}
	computeClient, err := openstack.NewComputeV2(conohaClient, eo)
	if err != nil {
		log.Fatalf("Compute Client Failed: %v", err)
		return err
	}
	eo = gophercloud.EndpointOpts{
		Type:   "image",
		Region: os.Getenv("CONOHA_ENDPOINT"),
	}
	imageClient, err := openstack.NewImageServiceV2(conohaClient, eo)
	if err != nil {
		log.Fatalf("Compute Client Failed: %v", err)
		return err
	}

	var uuid string
	pager := images.ListDetail(computeClient, nil)
	pager.EachPage(func(page pagination.Page) (bool, error) {
		imageList, err := images.ExtractImages(page)
		if err != nil {
			fmt.Println(err)
			return false, err
		}
		for _, i := range imageList {
			if i.Name == "ConohaChatOps-snapshot" {
				uuid = i.ID
			}
		}
		return true, nil
	})

	err = images.Delete(imageClient, uuid).ExtractErr()
	if err != nil {
		log.Fatalf("Delete Image Failed: %v", err)
		return err
	}

	//TODO Createと違ってDeleteの場合はGetしてStatusを見ることができない
	//時間のかかる処理でもないので、とりあえず決め打ち待機で...
	time.Sleep(time.Second * 30)

	return nil
}

func WaitForImage(client *gophercloud.ServiceClient, imageID string, target string) error {
	for {
		image, err := images.Get(client, imageID).Extract()
		if err != nil {
			return err
		}
		if image.Status == target {
			// conflicting防止
			time.Sleep(time.Second * 10)
			break
		}
		time.Sleep(time.Second * 10)
	}
	return nil
}
