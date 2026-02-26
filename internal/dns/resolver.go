package dns

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"time"
)

type Resolver interface {
	Resolve(domain string) (string, error)
}

type CloudflareDoH struct {
	client *http.Client
}

func NewCloudflareDoH() *CloudflareDoH {
	return &CloudflareDoH{
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

type doHResponse struct {
	Answer []struct {
		Data string `json:"data"`
	} `json:"Answer"`
}

func (c *CloudflareDoH) Resolve(domain string) (string, error) {
	url := "https://cloudflare-dns.com/dns-query?name=" + domain + "&type=A"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("accept", "application/dns-json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result doHResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	for _, answer := range result.Answer {
		if ip := net.ParseIP(answer.Data); ip != nil && ip.To4() != nil {
			return answer.Data, nil
		}
	}

	return "", nil
}
