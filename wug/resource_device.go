package wug

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/tidwall/gjson"
)

func resourceDevice() *schema.Resource {
	return &schema.Resource{
		Create: resourceDeviceCreate,
		Read:   resourceDeviceRead,
		/* No Update path since all fields are ForceNew. */
		/* Update: resourceDeviceUpdate, */
		Delete: resourceDeviceDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Display name of the device.",
				Required:    true,
				ForceNew:    true,
			},
			"options": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Set of options for applying the template (either l2 or basic).",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"l2",
					"basic",
				}, true),
			},
			"groups": &schema.Schema{
				Type:        schema.TypeList,
				Description: "List of groups that device will be added to.",
				Required:    true,
				ForceNew:    true,
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
				ForceNew:    true,
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
			"credential": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				Description: "Credentials.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"type": &schema.Schema{
						Type:        schema.TypeString,
						Required:    true,
						Description: "Credential type (SNMP, Windows, etc).",
					},
					"name": &schema.Schema{
						Type:        schema.TypeString,
						Required:    true,
						Description: "Credential name.",
					},
				}},
			},
			"active_monitor": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				Description: "Active monitors.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"name": &schema.Schema{
						Type:        schema.TypeString,
						Required:    true,
						Description: "Monitor name.",
					},
					"argument": &schema.Schema{
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Monitor argument.",
					},
					"comment": &schema.Schema{
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Monitor comment.",
					},
					"critical": &schema.Schema{
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "Is monitor critical.",
					},
					"polling_order": &schema.Schema{
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Monitor polling order.",
					},
				}},
			},
			"performance_monitor": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				Description: "Performance monitors.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"name": &schema.Schema{
						Type:        schema.TypeString,
						Required:    true,
						Description: "Monitor name.",
					},
				}},
			},
			"device_type": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Type of the device.",
				Optional:    true,
				ForceNew:    true,
			},
			"snmp_oid": &schema.Schema{
				Type:        schema.TypeString,
				Description: "SNMP OID of the device.",
				Optional:    true,
				ForceNew:    true,
			},
			"primary_role": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Primary role of the device.",
				Optional:    true,
				ForceNew:    true,
			},
			"subroles": &schema.Schema{
				Type:        schema.TypeList,
				Description: "Subroles of the device.",
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"os": &schema.Schema{
				Type:        schema.TypeString,
				Description: "OS of the device.",
				Optional:    true,
				ForceNew:    true,
			},
			"brand": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Brand of the device.",
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceDeviceCreate(d *schema.ResourceData, m interface{}) error {
	wugResty := m.(*WUGClient).Resty
	token := m.(*WUGClient).Token
	config := m.(*WUGClient).Config

	/* Reformat arrays since the field names may change... */
	interfaces := make([]map[string]interface{}, 0)
	for _, iface := range d.Get("interface").(*schema.Set).List() {
		interfaces = append(interfaces, map[string]interface{}{
			"defaultInterface":     iface.(map[string]interface{})["default"].(bool),
			"pollUsingNetworkName": iface.(map[string]interface{})["poll_using_network_name"].(bool),
			"networkAddress":       iface.(map[string]interface{})["network_address"].(string),
			"networkName":          iface.(map[string]interface{})["network_name"].(string),
		})
	}

	credentials := make([]map[string]interface{}, 0)
	for _, cred := range d.Get("credential").(*schema.Set).List() {
		credentials = append(credentials, map[string]interface{}{
			"credentialType": cred.(map[string]interface{})["type"].(string),
			"credential":     cred.(map[string]interface{})["name"].(string),
		})
	}

	activeMonitors := make([]map[string]interface{}, 0)
	for _, mon := range d.Get("active_monitor").(*schema.Set).List() {
		activeMonitors = append(activeMonitors, map[string]interface{}{
			"name":         mon.(map[string]interface{})["name"].(string),
			"argument":     mon.(map[string]interface{})["argument"].(string),
			"comment":      mon.(map[string]interface{})["comment"].(string),
			"isCritical":   mon.(map[string]interface{})["critical"].(bool),
			"pollingOrder": mon.(map[string]interface{})["polling_order"].(int),
		})
	}

	params := map[string]interface{}{
		"options": []string{
			d.Get("options").(string),
		},
		"templates": []map[string]interface{}{{
			"displayName":         d.Get("name").(string),
			"interfaces":          interfaces,
			"groups":              d.Get("groups").([]interface{}),
			"credentials":         credentials,
			"activeMonitors":      activeMonitors,
			"performanceMonitors": d.Get("performance_monitor").(*schema.Set).List(),
			"deviceType":          d.Get("device_type").(string),
			"snmpOid":             d.Get("snmp_oid").(string),
			"primaryRole":         d.Get("primary_role").(string),
			"subRoles":            d.Get("subroles").([]interface{}),
			"os":                  d.Get("os").(string),
			"brand":               d.Get("brand").(string),
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

	/* Reformat arrays since the field names may change... */
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

	credentials := make([]map[string]interface{}, 0)
	for _, iface := range gjson.GetBytes(resp.Body(), "data.templates.0.credentials").Array() {
		credentials = append(credentials, map[string]interface{}{
			"type": iface.Get("credentialType").String(),
			"name": iface.Get("credential").String(),
		})
	}

	d.Set("credential", credentials)

	activeMonitors := make([]map[string]interface{}, 0)
	for _, iface := range gjson.GetBytes(resp.Body(), "data.templates.0.activeMonitors").Array() {
		activeMonitors = append(activeMonitors, map[string]interface{}{
			"name":          iface.Get("name").String(),
			"argument":      iface.Get("argument").String(),
			"comment":       iface.Get("comment").String(),
			"critical":      iface.Get("isCritical").Bool(),
			"polling_order": iface.Get("pollingOrder").Int(),
		})
	}

	d.Set("active_monitor", activeMonitors)

	d.Set("performance_monitor", gjson.GetBytes(resp.Body(), "data.templates.0.performanceMonitors").Array())

	d.Set("device_type", gjson.GetBytes(resp.Body(), "data.templates.0.deviceType").String())
	d.Set("snmp_oid", gjson.GetBytes(resp.Body(), "data.templates.0.snmpOid").String())
	d.Set("primary_role", gjson.GetBytes(resp.Body(), "data.templates.0.primaryRole").String())
	d.Set("subroles", gjson.GetBytes(resp.Body(), "data.templates.0.subRoles").Array())
	d.Set("os", gjson.GetBytes(resp.Body(), "data.templates.0.os").String())
	d.Set("brand", gjson.GetBytes(resp.Body(), "data.templates.0.brand").String())

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
