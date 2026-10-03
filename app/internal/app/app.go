package dyndns

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/HellstromIT/do-dyndns/app/cmd/do-dyndns/internal/config"
	"github.com/digitalocean/godo"
)

// maxIPResponseSize caps how much of the IP lookup response is read.
const maxIPResponseSize = 4096

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
	Create bool
}

func createDoClient(c config.Config) *godo.Client {
	return godo.NewFromToken(c.DigitalOcean.Token)
}

func getPublicIP(c config.Config) (*PublicIP, error) {
	var p *PublicIP
	url := c.Ifconfig.Host + c.Ifconfig.Uri
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status from %s: %s", url, resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxIPResponseSize))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &p)
	if err != nil {
		return nil, err
	}

	if p == nil {
		return nil, fmt.Errorf("empty response from %s", url)
	}

	// Only accept a valid IPv4 address since it is written to an A record.
	ip := net.ParseIP(p.IP)
	if ip == nil || ip.To4() == nil {
		return nil, fmt.Errorf("invalid IPv4 address %q from %s", p.IP, url)
	}
	p.IP = ip.To4().String()

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

		if len(record) == 0 {
			d.domains[i].Create = true
		} else {
			for _, r := range record {
				d.domains[i].ID = r.ID
				if r.Data != ip.IP {
					d.domains[i].Update = true
				} else {
					d.domains[i].Update = false
				}
			}
		}
	}
	return nil
}

func (d *Domains) updateRecords(client *godo.Client, ip PublicIP) error {
	ctx := context.TODO()
	for _, domain := range d.domains {
		if domain.Update {
			log.Printf("Starting update of record %s\n", domain.Name)
			parent := getParentDomain(domain.Name)
			editRequest := &godo.DomainRecordEditRequest{
				Type: domain.Type,
				Name: strings.TrimSuffix(domain.Name, parent),
				Data: ip.IP,
			}

			_, _, err := client.Domains.EditRecord(ctx, parent, domain.ID, editRequest)
			if err != nil {
				log.Printf("Update of record %s failed!", domain.Name)
				return err
			}
			log.Printf("Update of record %s completed!\n", domain.Name)
		} else if domain.Create {
			log.Printf("Creating %s record \n", domain.Name)
			parent := getParentDomain(domain.Name)
			editRequest := &godo.DomainRecordEditRequest{
				Type: domain.Type,
				Name: strings.TrimSuffix(domain.Name, parent),
				Data: ip.IP,
			}

			_, _, err := client.Domains.CreateRecord(ctx, parent, editRequest)
			if err != nil {
				log.Printf("Creation of record %s failed!", domain.Name)
				return err
			}
			log.Printf("Creation of record %s completed!\n", domain.Name)
		}
	}
	return nil
}

func run(c config.Config) {
	for {
		log.Println("Starting Check!")
		nextRun := time.Now().Truncate(time.Minute)
		nextRun = nextRun.Add(time.Minute * time.Duration(c.Interval))

		check(c)

		time.Sleep(time.Until(nextRun))
	}
}

func check(c config.Config) {
	publicIP, err := getPublicIP(c)
	if err != nil {
		log.Printf("Error getting public IP, skipping update\n %v", err)
		return
	}

	client := createDoClient(c)

	var domain Domains
	err = domain.checkRecords(client, *publicIP, c)
	if err != nil {
		log.Println(err)
		return
	}

	err = domain.updateRecords(client, *publicIP)
	if err != nil {
		log.Println(err)
	}
}

func App() {
	var config config.Config
	config.Read("/data/config.yml")
	config.ReadEnv()

	run(config)
}
