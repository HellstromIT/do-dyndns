package dyndns

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/HellstromIT/do-dyndns/app/cmd/do-dyndns/internal/config"
	"github.com/digitalocean/godo"
)

type PublicIP struct {
	IP string `json:"ip"`
}

type Domains struct {
	domains []Domain
}

type Domain struct {
	ID     int
	Name   string
	Type   string
	Data   string
	Update bool
}

func createDoClient(c config.Config) *godo.Client {
	return godo.NewFromToken(c.DigitalOcean.Token)
}

func getPublicIP(c config.Config) (*PublicIP, error) {
	var p *PublicIP
	url := c.Ifconfig.Host + c.Ifconfig.Uri
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func getParentDomain(d string) string {
	split := strings.Split(d, ".")
	domain, tld := split[len(split)-2], split[len(split)-1]

	return domain + "." + tld
}

func (d *Domains) checkRecords(client *godo.Client, ip PublicIP, c config.Config) error {
	ctx := context.TODO()
	for i, domain := range c.Domains {

		var currDomain Domain
		currDomain.Name = domain
		currDomain.Type = "A"

		d.domains = append(d.domains, currDomain)

		parent := getParentDomain(currDomain.Name)

		opt := &godo.ListOptions{
			Page:    1,
			PerPage: 1,
		}

		record, _, err := client.Domains.RecordsByTypeAndName(ctx, parent, currDomain.Type, currDomain.Name, opt)
		if err != nil {
			return err
		}
		for _, r := range record {
			d.domains[i].ID = r.ID
			if r.Data != ip.IP {
				d.domains[i].Update = true
			} else {
				d.domains[i].Update = false
			}
		}
	}
	return nil
}

func (d *Domains) updateRecords(client *godo.Client, ip PublicIP) error {
	ctx := context.TODO()
	for _, domain := range d.domains {
		if domain.Update {
			log.Printf("Starting update of %s\n", domain.Name)
			parent := getParentDomain(domain.Name)
			editRequest := &godo.DomainRecordEditRequest{
				Type: domain.Type,
				Name: strings.TrimSuffix(domain.Name, parent),
				Data: ip.IP,
			}

			_, _, err := client.Domains.EditRecord(ctx, parent, domain.ID, editRequest)
			if err != nil {
				log.Println("Update failed!")
				return err
			}
			log.Printf("Update of %s completed!\n", domain.Name)
		}
	}
	return nil
}

func run(c config.Config) {
	log.Println("Starting Check!")
	nextRun := time.Now().Truncate(time.Minute)
	nextRun = nextRun.Add(time.Minute * time.Duration(c.Interval))

	publicIP, err := getPublicIP(c)
	if err != nil {
		log.Printf("Error getting public IP\n %v", err)
	}

	client := createDoClient(c)

	var domain Domains
	err = domain.checkRecords(client, *publicIP, c)
	if err != nil {
		log.Println(err)
	}

	err = domain.updateRecords(client, *publicIP)
	if err != nil {
		log.Println(err)
	}

	time.Sleep(time.Until(nextRun))

	run(c)
}

func App() {
	var config config.Config
	config.Read("/data/config.yml")

	run(config)
}
