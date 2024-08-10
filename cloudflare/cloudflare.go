package cloudflare

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type cloudflareClient struct {
	apiKey   string
	emailKey string
	client   *http.Client
}

func NewClient(apiKey, emailKey string) *cloudflareClient {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	return &cloudflareClient{
		apiKey:   apiKey,
		emailKey: emailKey,
		client:   client,
	}
}

func (c *cloudflareClient) GetDomainDNS(d Domain) DNSRecordsResponse {
	var dnsRecordsResponse DNSRecordsResponse
	URL := url.URL{
		Scheme: "https",
		Host:   "api.cloudflare.com",
		Path:   fmt.Sprintf("/client/v4/zones/%s/dns_records", d.ZoneID),
	}
	log.Printf("Getting DNS record from Cloudflare %s", d.Name)
	req, err := http.NewRequest(http.MethodGet, URL.String(), nil)
	if err != nil {
		log.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Key", c.apiKey)
	req.Header.Set("X-Auth-Email", c.emailKey)

	res, err := c.client.Do(req)
	if err != nil {
		log.Println(err)
		return dnsRecordsResponse
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			log.Printf("Error reading body response from Cloudflare %v \n", res.StatusCode)
		} else {
			err := json.Unmarshal(b, &dnsRecordsResponse)
			if err != nil {
				log.Printf("Error unmarshalling body response from Cloudflare %v \n", err)
				return dnsRecordsResponse
			}
		}
	} else {
		log.Printf("Error getting DNS record from Cloudflare StatusCode:%v \n", res.StatusCode)
	}
	return dnsRecordsResponse
}

func (c *cloudflareClient) GetIP() string {
	req, err := http.NewRequest(http.MethodGet, "http://checkip.amazonaws.com/", nil)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	res, err := c.client.Do(req)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer res.Body.Close()

	ip, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	ipAddress := strings.TrimSpace(string(ip))
	ipAddress = strings.Replace(string(ipAddress), "\n", "", 1)
	return ipAddress
}

func (c *cloudflareClient) UpdateDNSRecord(dnsRecord DNSRecord, ip string) bool {
	updated := false
	URL := url.URL{
		Scheme: "https",
		Host:   "api.cloudflare.com",
		Path:   fmt.Sprintf("/client/v4/zones/%s/dns_records/%s", dnsRecord.ZoneID, dnsRecord.ID),
	}
	log.Printf("Updating DNS record in Cloudflare %s", dnsRecord.Name)
	dnsRecordUpdate := DNSRecordUpdate{
		Content: ip,
		Type:    dnsRecord.Type,
		Name:    dnsRecord.Name,
		TTL:     dnsRecord.TTL,
		ID:      dnsRecord.ID,
	}
	b, err := json.Marshal(dnsRecordUpdate)
	if err != nil {
		log.Println(err)
		return updated
	}
	req, err := http.NewRequest(http.MethodPatch, URL.String(), strings.NewReader(string(b)))
	if err != nil {
		log.Println(err)
		return updated
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Key", c.apiKey)
	req.Header.Set("X-Auth-Email", c.emailKey)
	res, err := c.client.Do(req)
	if err != nil {
		log.Println(err)
		return updated
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		return true
	} else {
		b, _ := io.ReadAll(res.Body)
		log.Printf("Error updating DNS record for Cloudflare StatusCode:%v Response:%s\n", res.StatusCode, string(b))
	}
	return updated
}

type Domain struct {
	Name        string
	ZoneID      string
	DNSRecordID string
}

type DNSRecord struct {
	ID         string   `json:"id"`
	ZoneID     string   `json:"zone_id"`
	ZoneName   string   `json:"zone_name"`
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Content    string   `json:"content"`
	Proxiable  bool     `json:"proxiable"`
	Proxied    bool     `json:"proxied"`
	TTL        int      `json:"ttl"`
	Meta       Meta     `json:"meta"`
	Comment    *string  `json:"comment"`
	Tags       []string `json:"tags"`
	CreatedOn  string   `json:"created_on"`
	ModifiedOn string   `json:"modified_on"`
	Priority   *int     `json:"priority,omitempty"` // Only applicable for MX records
}

type Meta struct {
	AutoAdded           bool `json:"auto_added"`
	ManagedByApps       bool `json:"managed_by_apps"`
	ManagedByArgoTunnel bool `json:"managed_by_argo_tunnel"`
}

type ResultInfo struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Count      int `json:"count"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

type DNSRecordsResponse struct {
	Result     []DNSRecord `json:"result"`
	Success    bool        `json:"success"`
	Errors     []string    `json:"errors"`
	Messages   []string    `json:"messages"`
	ResultInfo ResultInfo  `json:"result_info"`
}

type DNSRecordUpdate struct {
	Content string   `json:"content"`
	Name    string   `json:"name"`
	Proxied bool     `json:"proxied"`
	Type    string   `json:"type"`
	Comment string   `json:"comment"`
	ID      string   `json:"id"`
	Tags    []string `json:"tags"`
	TTL     int      `json:"ttl"`
}
