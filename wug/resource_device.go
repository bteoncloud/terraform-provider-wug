package wug

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDevice() *schema.Resource {
	return &schema.Resource{
		Create: resourceDeviceCreate,
		Read:   resourceDeviceRead,
		Update: resourceDeviceUpdate,
		Delete: resourceDeviceDelete,

		Schema: map[string]*schema.Schema{
			"address": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceDeviceCreate(d *schema.ResourceData, m interface{}) error {
	return resourceDeviceRead(d, m)
}

func resourceDeviceRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceDeviceUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceDeviceRead(d, m)
}

func resourceDeviceDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
