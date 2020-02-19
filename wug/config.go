package wug

import (
	"errors"
	"log"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/tidwall/gjson"
)

type WUGClient struct {
	Resty  *resty.Client
	Config *Config
	Token  string
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

	params := url.Values{}
	params.Add("grant_type", "password")
	params.Add("username", c.User)
	params.Add("password", c.Password)

	resp, err := client.Resty.R().
		SetHeader("Content-Type", "application/json").
		//SetBody(map[string]interface{}{"grant_type": "password", "username": c.User, "password": c.Password}).
		SetBody(params.Encode()).
		Post(c.URL + "/token")

	if err != nil {
		return nil, err
	} else if resp.StatusCode() != 200 {
		return nil, errors.New(string(resp.Body()))
	}

	client.Token = gjson.GetBytes(resp.Body(), "access_token").String()

	log.Printf("[WUG] Access token: %s", client.Token)

	client.Config = c

	return client, nil
}
