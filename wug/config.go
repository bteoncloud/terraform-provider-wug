package wug

import (
	"errors"
	"log"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tidwall/gjson"
)

// Client holds the Resty instance and API configuration.
type Client struct {
	Resty  *resty.Client
	Config *Config
	Token  string
}

// Config holds API configuration parameters.
type Config struct {
	InsecureFlag bool
	User         string
	Password     string
	URL          string
}

// NewConfig instanciates a Config object.
func NewConfig(d *schema.ResourceData) (*Config, error) {
	c := &Config{
		User:         d.Get("user").(string),
		Password:     d.Get("password").(string),
		InsecureFlag: d.Get("allow_unverified_ssl").(bool),
		URL:          d.Get("url").(string),
	}

	return c, nil
}

// Client returns a REST client for WUG.
func (c *Config) Client() (*Client, error) {
	client := new(Client)

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
