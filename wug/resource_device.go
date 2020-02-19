package wug

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/tidwall/gjson"
)

func resourceDevice() *schema.Resource {
	return &schema.Resource{
		Create: resourceDeviceCreate,
		Read:   resourceDeviceRead,
		Update: resourceDeviceUpdate,
		Delete: resourceDeviceDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Display name of the device.",
				Required:    true,
			},
			"options": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Set of options for applying the template (either l2 or basic).",
				Required:    true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					switch v {
					case
						"l2",
						"basic":
						return
					}
					errs = append(errs, fmt.Errorf("%q must be either 'basic' or 'l2', got: %s", key, v))
					return
				},
			},
			"groups": &schema.Schema{
				Type:        schema.TypeList,
				Description: "List of groups that devie will be added to.",
				Required:    true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"parents": &schema.Schema{
						Type:        schema.TypeList,
						Description: "List of parent nodes.",
						Optional:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"name": &schema.Schema{
						Type:        schema.TypeString,
						Description: "Name of the leaf group the device will be added to.",
						Required:    true,
					}},
				},
			},
			"interface": &schema.Schema{
				Type:        schema.TypeSet,
				Required:    true,
				Description: "Interfaces.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"default": &schema.Schema{
						Type:        schema.TypeBool,
						Default:     false,
						Optional:    true,
						Description: "Whether the interface is the default one.",
					},
					"poll_using_network_name": &schema.Schema{
						Type:        schema.TypeBool,
						Default:     false,
						Optional:    true,
						Description: "Poll using network name.",
					},
					"network_address": &schema.Schema{
						Type:        schema.TypeString,
						Required:    true,
						Description: "Network address of the interface.",
					},
					"network_name": &schema.Schema{
						Type:        schema.TypeString,
						Required:    true,
						Description: "Network name of the interface.",
					},
				}},
			},
		},
	}
}

func resourceDeviceCreate(d *schema.ResourceData, m interface{}) error {
	wugResty := m.(*WUGClient).Resty
	token := m.(*WUGClient).Token
	config := m.(*WUGClient).Config

	/* Reformat the interfaces array since the field names change... */
	interfaces := make([]map[string]interface{}, 0)
	for _, iface := range d.Get("interface").(*schema.Set).List() {
		interfaces = append(interfaces, map[string]interface{}{
			"defaultInterface":     iface.(map[string]interface{})["default"].(bool),
			"pollUsingNetworkName": iface.(map[string]interface{})["poll_using_network_name"].(bool),
			"networkAddress":       iface.(map[string]interface{})["network_address"].(string),
			"networkName":          iface.(map[string]interface{})["network_name"].(string),
		})
	}

	params := map[string]interface{}{
		"options": []string{
			d.Get("options").(string),
		},
		"templates": []map[string]interface{}{{
			"displayName": d.Get("name").(string),
			"interfaces":  interfaces,
			"groups":      d.Get("groups").([]interface{}),
		}},
	}

	resp, err := wugResty.SetDebug(true).R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token).
		SetBody(params).
		Patch(config.URL + "/devices/-/config/template")

	if err != nil {
		return err
	} else if resp.StatusCode() != 200 {
		return errors.New(string(resp.Body()))
	}

	d.SetId(gjson.GetBytes(resp.Body(), "data.idMap.0.resultId").String())

	log.Printf("[WUG] Created device with ID: %s\n", d.Id())

	return resourceDeviceRead(d, m)
}

func resourceDeviceRead(d *schema.ResourceData, m interface{}) error {
	resty := m.(*WUGClient).Resty
	token := m.(*WUGClient).Token
	config := m.(*WUGClient).Config

	id := d.Id()

	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token).
		Get(config.URL + "/devices/" + id + "/config/template")

	if err != nil {
		return err
	} else if resp.StatusCode() != 200 {
		return errors.New(string(resp.Body()))
	} else if resp.StatusCode() == 404 {
		/* The device does not exist anymore. */
		d.SetId("")
		return errors.New(string(resp.Body()))
	}

	deviceCount := gjson.GetBytes(resp.Body(), "data.deviceCount").Int()

	if deviceCount != 1 {
		return fmt.Errorf("Found invalid device count for %s: %d", id, deviceCount)
	}

	d.Set("name", gjson.GetBytes(resp.Body(), "data.templates.0.displayName").String())
	d.Set("groups", gjson.GetBytes(resp.Body(), "data.templates.0.groups").Array())

	/* Reformat the interfaces array since the field names change... */
	interfaces := make([]map[string]interface{}, 0)
	for _, iface := range gjson.GetBytes(resp.Body(), "data.templates.0.interfaces").Array() {
		interfaces = append(interfaces, map[string]interface{}{
			"default":                 iface.Get("defaultInterface").Bool(),
			"poll_using_network_name": iface.Get("pollUsingNetworkName").Bool(),
			"network_address":         iface.Get("networkAddress").String(),
			"network_name":            iface.Get("networkName").String(),
		})
	}

	d.Set("interface", interfaces)

	return nil
}

func resourceDeviceUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceDeviceRead(d, m)
}

func resourceDeviceDelete(d *schema.ResourceData, m interface{}) error {
	resty := m.(*WUGClient).Resty
	token := m.(*WUGClient).Token
	config := m.(*WUGClient).Config

	id := d.Id()

	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token).
		Delete(config.URL + "/devices/" + id)

	if err != nil {
		return err
	} else if resp.StatusCode() != 200 {
		return errors.New(string(resp.Body()))
	}

	d.SetId("")

	return nil
}
