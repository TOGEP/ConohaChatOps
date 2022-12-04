package conoha

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/startstop"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/images"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/pagination"
)

type Bot struct {
	Session        *discordgo.Session
	providerClient *gophercloud.ProviderClient
	computeClient  *gophercloud.ServiceClient
	imageClient    *gophercloud.ServiceClient
}

func NewBot(s *discordgo.Session) (*Bot, error) {
	bot := &Bot{
		Session: s,
	}
	opts := gophercloud.AuthOptions{
		IdentityEndpoint: "https://identity." + os.Getenv("CONOHA_ENDPOINT") + ".conoha.io/v2.0",
		Username:         os.Getenv("CONOHA_USERNAME"),
		TenantName:       os.Getenv("CONOHA_TENANTNAME"),
		Password:         os.Getenv("CONOHA_PASSWORD"),
		AllowReauth:      true,
	}

	pc, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		return nil, err
	}
	bot.providerClient = pc

	eo := gophercloud.EndpointOpts{
		Type:   "compute",
		Region: os.Getenv("CONOHA_ENDPOINT"),
	}
	cc, err := openstack.NewComputeV2(bot.providerClient, eo)
	if err != nil {
		return nil, err
	}
	bot.computeClient = cc

	eo = gophercloud.EndpointOpts{
		Type:   "image",
		Region: os.Getenv("CONOHA_ENDPOINT"),
	}
	ic, err := openstack.NewImageServiceV2(bot.providerClient, eo)
	if err != nil {
		log.Fatalf("Compute Client Failed: %v", err)
		return nil, err
	}
	bot.imageClient = ic

	log.Println("Start!")
	return bot, nil
}

func (bot *Bot) GetImageRef() (string, error) {
	var imageRef string
	pager := images.ListDetail(bot.computeClient, nil)
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

func (bot *Bot) GetFlavorRef(memSize int64) (string, error) {
	var flavorRef string
	pager := flavors.ListDetail(bot.computeClient, nil)
	pager.EachPage(func(page pagination.Page) (bool, error) {
		flavorList, err := flavors.ExtractFlavors(page)
		if err != nil {
			fmt.Println(err)
			return false, err
		}
		for _, f := range flavorList {
			// Conohaの旧プラン(Disk:50GB)のIDも残っている為DISKとRAMで判定
			if f.Disk == 100 && f.RAM == int(memSize)*1024 {
				flavorRef = f.ID
			}
		}
		return true, nil
	})
	return flavorRef, nil
}

func (bot *Bot) GetIPaddr() (string, error) {
	var ip string
	pager := servers.List(bot.computeClient, nil)
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

// IsServerRunは"instance_name_tag":"ConohaChatOps"のVMが存在するか確認する関数
// 戻り値がtrueの場合はVMが存在し，falseの場合はVMは存在しない．
func (bot *Bot) IsServerRun() bool {
	var uuid string
	pager := servers.List(bot.computeClient, nil)
	pager.EachPage(func(page pagination.Page) (bool, error) {
		serverList, err := servers.ExtractServers(page)
		if err != nil {
			fmt.Println(err)
			return true, nil
		}
		for _, s := range serverList {
			//TODO 好きな名前を設定できるように
			if s.Metadata["instance_name_tag"] == "ConohaChatOps" {
				uuid = s.ID
			}
		}
		return true, nil
	})
	if uuid == "" {
		return false
	}
	return true
}

func (bot *Bot) OpenServer(imageRef string, flavorRef string) error {
	co := servers.CreateOpts{
		Name:      "ConohaChatOps",
		ImageRef:  imageRef,
		FlavorRef: flavorRef,
		AdminPass: os.Getenv("CONOHA_PASSWORD"),
		SecurityGroups: []string{
			"default",
			"gncs-ipv4-all",
		},
		Metadata: map[string]string{
			"instance_name_tag": "ConohaChatOps",
		},
	}

	server, err := servers.Create(bot.computeClient, co).Extract()
	if err != nil {
		log.Fatalf("Create a Server Failed: %v", err)
		return err
	}

	err = servers.WaitForStatus(bot.computeClient, server.ID, "ACTIVE", 1000)
	if err != nil {
		log.Fatalf("Unable to create for server: %v", err)
		return err
	}

	return nil
}

func (bot *Bot) CloseServer() error {
	var uuid string
	pager := servers.List(bot.computeClient, nil)
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

	err := startstop.Stop(bot.computeClient, uuid).ExtractErr()
	if err != nil {
		log.Fatalf("Stop a Server Failed: %v", err)
		return err
	}

	err = servers.WaitForStatus(bot.computeClient, uuid, "SHUTOFF", 1000)
	if err != nil {
		log.Fatalf("Unable to stop for server: %v", err)
		return err
	}

	return nil
}

func (bot *Bot) DeleteServer() error {
	var uuid string
	pager := servers.List(bot.computeClient, nil)
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

	err := servers.Delete(bot.computeClient, uuid).ExtractErr()
	if err != nil {
		log.Fatalf("Delete a Server Failed: %v", err)
		return err
	}

	err = servers.WaitForStatus(bot.computeClient, uuid, "DELETED", 1000)
	if err != nil {
		if _, ok := err.(gophercloud.ErrDefault404); !ok {
			log.Fatalf("Deleting server %q failed: %v", uuid, err)
			return nil
		}
		log.Printf("Cannot find server: %q. It's probably already been deleted.\n", uuid)
	}

	return nil
}

func (bot *Bot) CreateImage() error {
	var uuid string
	pager := servers.List(bot.computeClient, nil)
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

	imageID, err := servers.CreateImage(bot.computeClient, uuid, snapshotOpts).ExtractImageID()
	if err != nil {
		log.Fatalf("Create Image Failed: %v", err)
		return err
	}

	err = waitForImage(bot.computeClient, imageID, "ACTIVE")
	if err != nil {
		log.Fatalf("Unable to save for server image: %v", err)
		return err
	}

	return nil
}

func (bot *Bot) DeleteImage() error {
	var uuid string
	pager := images.ListDetail(bot.computeClient, nil)
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

	err := images.Delete(bot.imageClient, uuid).ExtractErr()
	if err != nil {
		log.Fatalf("Delete Image Failed: %v", err)
		return err
	}

	//TODO Createと違ってDeleteの場合はGetしてStatusを見ることができない
	//時間のかかる処理でもないので、とりあえず決め打ち待機で...
	time.Sleep(time.Second * 30)

	return nil
}

func waitForImage(client *gophercloud.ServiceClient, imageID string, target string) error {
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
