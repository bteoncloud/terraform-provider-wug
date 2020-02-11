package wug

import (
	"encoding/json"
	"log"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type WUGClient struct {
	Resty *resty.Client
	Token string
}

type Config struct {
	InsecureFlag bool
	User         string
	Password     string
	URL          string
}

func NewConfig(d *schema.ResourceData) (*Config, error) {
	c := &Config{
		User:         d.Get("user").(string),
		Password:     d.Get("password").(string),
		InsecureFlag: d.Get("allow_unverified_ssl").(bool),
		URL:          d.Get("url").(string),
	}

	return c, nil
}

/* Returns a REST client for WUG. */
func (c *Config) Client() (*WUGClient, error) {
	client := new(WUGClient)

	client.Resty = resty.New()

	resp, err := client.Resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{"grant_type": "password", "username": c.User, "password": c.Password}).
		Post(c.URL + "/token")

	if err != nil {
		return nil, err
	} else if resp.StatusCode() != 200 {
		return nil, err
	}

	var i map[string]interface{}
	jsonErr := json.Unmarshal(resp.Body(), &i)
	if jsonErr != nil {
		return nil, jsonErr
	}

	client.Token = i["access_token"].(string)

	log.Printf("[WUG] Access token: %s", client.Token)

	return client, nil
}
