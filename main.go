package main

import (
	"jasonbronson/cloudflare-dynamic-dns/cloudflare"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/joho/godotenv"
	cron "github.com/robfig/cron/v3"
)

var interval string
var client *http.Client

func main() {
	c := cron.New()

	if interval == "" {
		interval = "*/59 * * * *"
	}
	//Run once on startup
	updateIP()
	//Run cronjob moving forward
	c.AddFunc(interval, updateIP)
	c.Start()
	log.Println("=====cron system started======")

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}

func updateIP() {

	err := godotenv.Load()
	if err != nil {
		log.Print("no env file found ")
	}

	domain := os.Getenv("DOMAIN")
	apiKey := os.Getenv("API_KEY")
	emailKey := os.Getenv("EMAIL_KEY")
	log.Printf("APIKEY: %s DOMAIN: %s EMAILKEY: %s", apiKey, domain, emailKey)
	if apiKey == "" || domain == "" || emailKey == "" {
		log.Println("api, domain or email env variable is missing")
		return
	}
	// Get current IP
	cloudflareClient := cloudflare.NewClient(apiKey, emailKey)
	ip := cloudflareClient.GetIP()
	if ip == "" {
		log.Println("external IP address could not be found ")
		return
	}
	domainList := getDomainList(domain)
	if len(domainList) == 0 {
		log.Println("domain format is incorrect skipping update")
		return
	}

	for _, d := range domainList {
		dns := cloudflareClient.GetDomainDNS(d)
		for _, record := range dns.Result {
			if d.DNSRecordID == record.ID && record.ZoneID == d.ZoneID && record.Type == "A" {
				if record.Content == ip {
					log.Printf("no need to update ip it's already up to date %s\n", ip)
					continue
				}
				log.Printf("attempting to update ip for domain %s from %s to %s\n", d.Name, record.Content, ip)
				if cloudflareClient.UpdateDNSRecord(record, ip) {
					log.Println("update ip completed")
				}
			}

		}

	}

}

func getDomainList(domain string) []cloudflare.Domain {
	var domainList []cloudflare.Domain
	if strings.Contains(domain, "|") {
		domains := strings.Split(domain, "|")
		for _, data := range domains {
			item := strings.Split(data, ";")
			if len(item) != 3 {
				log.Println("invalid domain format skipping update")
				return nil
			}
			domainList = append(domainList, cloudflare.Domain{
				Name:        item[0],
				ZoneID:      item[1],
				DNSRecordID: item[2],
			})
		}
	} else {
		item := strings.Split(domain, ";")
		if len(item) != 3 {
			log.Println("invalid domain format skipping update")
			return nil
		}
		domainList = append(domainList, cloudflare.Domain{
			Name:        item[0],
			ZoneID:      item[1],
			DNSRecordID: item[2],
		})
	}
	return domainList
}
