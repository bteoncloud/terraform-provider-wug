package wug

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/tidwall/gjson"
)

// DeviceTemplateInterface is WUG's internal object.
type DeviceTemplateInterface struct {
	IsDefault            bool   `json:"defaultInterface,omitempty"`
	PollUsingNetworkName bool   `json:"pollUsingNetworkName,omitempty"`
	NetworkAddress       string `json:"networkAddress,omitempty"`
	NetworkName          string `json:"networkName,omitempty"`
}

// DeviceTemplateReferenceName is WUG's internal object.
type DeviceTemplateReferenceName struct {
	Name    string   `json:"name,omitempty"`
	Parents []string `json:"parents,omitempty"`
}

// DeviceTemplateCredentials is WUG's internal object.
type DeviceTemplateCredentials struct {
	CredentialType string `json:"credentialType,omitempty"`
	Name           string `json:"credential,omitempty"`
}

// DeviceTemplateActiveMonitor is WUG's internal object.
type DeviceTemplateActiveMonitor struct {
	Name         string `json:"name,omitempty"`
	Argument     string `json:"argument,omitempty"`
	Comment      string `json:"comment,omitempty"`
	IsCritical   string `json:"isCritical,omitempty"`
	PollingOrder int    `json:"pollingOrder,string,omitempty"`
}

// DeviceTemplatePerformanceMonitor is WUG's internal object.
type DeviceTemplatePerformanceMonitor struct {
	Name string `json:"name,omitempty"`
}

// DeviceTemplate is WUG's internal object.
type DeviceTemplate struct {
	Name                string                             `json:"displayName,omitempty"`
	Interfaces          []DeviceTemplateInterface          `json:"interfaces,omitempty"`
	Groups              []DeviceTemplateReferenceName      `json:"groups,omitempty"`
	Credentials         []DeviceTemplateCredentials        `json:"credentials,omitempty"`
	ActiveMonitors      []DeviceTemplateActiveMonitor      `json:"activeMonitors,omitempty"`
	PerformanceMonitors []DeviceTemplatePerformanceMonitor `json:"performanceMonitors,omitempty"`
	DeviceType          string                             `json:"deviceType,omitempty"`
	SnmpOid             string                             `json:"snmpOid,omitempty"`
	PrimaryRole         string                             `json:"primaryRole,omitempty"`
	SubRoles            []string                           `json:"subRoles,omitempty"`
	Os                  string                             `json:"os,omitempty"`
	Brand               string                             `json:"brand,omitempty"`
	ActionPolicy        string                             `json:"actionPolicy,omitempty"`
}

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
						Default:     false,
					},
					"polling_order": &schema.Schema{
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Monitor polling order.",
						Default:     0,
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
			"action_policy": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Policy how to get notified.",
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceDeviceCreate(d *schema.ResourceData, m interface{}) error {
	wugResty := m.(*Client).Resty
	token := m.(*Client).Token
	config := m.(*Client).Config

	var template DeviceTemplate

	/* Build our template object. */

	template.Name = d.Get("name").(string)
	template.DeviceType = d.Get("device_type").(string)
	template.SnmpOid = d.Get("snmp_oid").(string)
	template.PrimaryRole = d.Get("primary_role").(string)
	template.Os = d.Get("os").(string)
	template.Brand = d.Get("brand").(string)
	template.ActionPolicy = d.Get("action_policy").(string)

	groupList := d.Get("groups").([]interface{})
	template.Groups = make([]DeviceTemplateReferenceName, 0)
	for _, group := range groupList {
		var refName DeviceTemplateReferenceName
		refName.Name = group.(map[string]interface{})["name"].(string)

		parents := group.(map[string]interface{})["parents"].([]interface{})
		refName.Parents = make([]string, 0)
		for _, parent := range parents {
			refName.Parents = append(refName.Parents, parent.(string))
		}

		template.Groups = append(template.Groups, refName)
	}

	subRoles := d.Get("subroles").([]interface{})
	template.SubRoles = make([]string, 0)
	for _, subrole := range subRoles {
		template.SubRoles = append(template.SubRoles, subrole.(string))
	}

	interfaceList := d.Get("interface").(*schema.Set).List()
	template.Interfaces = make([]DeviceTemplateInterface, 0)
	for _, iface := range interfaceList {
		template.Interfaces = append(template.Interfaces, DeviceTemplateInterface{
			IsDefault:            iface.(map[string]interface{})["default"].(bool),
			PollUsingNetworkName: iface.(map[string]interface{})["poll_using_network_name"].(bool),
			NetworkAddress:       iface.(map[string]interface{})["network_address"].(string),
			NetworkName:          iface.(map[string]interface{})["network_name"].(string),
		})
	}

	credentialList := d.Get("credential").(*schema.Set).List()
	template.Credentials = make([]DeviceTemplateCredentials, 0)
	for _, cred := range credentialList {
		template.Credentials = append(template.Credentials, DeviceTemplateCredentials{
			CredentialType: cred.(map[string]interface{})["type"].(string),
			Name:           cred.(map[string]interface{})["name"].(string),
		})
	}

	activeMonitorsList := d.Get("active_monitor").(*schema.Set).List()
	template.ActiveMonitors = make([]DeviceTemplateActiveMonitor, 0)
	for _, mon := range activeMonitorsList {
		template.ActiveMonitors = append(template.ActiveMonitors, DeviceTemplateActiveMonitor{
			Name:         mon.(map[string]interface{})["name"].(string),
			Argument:     mon.(map[string]interface{})["argument"].(string),
			Comment:      mon.(map[string]interface{})["comment"].(string),
			IsCritical:   strconv.FormatBool(mon.(map[string]interface{})["critical"].(bool)),
			PollingOrder: mon.(map[string]interface{})["polling_order"].(int),
		})
	}

	performanceMonitorsList := d.Get("performance_monitor").(*schema.Set).List()
	template.PerformanceMonitors = make([]DeviceTemplatePerformanceMonitor, 0)
	for _, mon := range performanceMonitorsList {
		template.PerformanceMonitors = append(template.PerformanceMonitors, DeviceTemplatePerformanceMonitor{
			Name: mon.(map[string]interface{})["name"].(string),
		})
	}

	params := map[string]interface{}{
		"options": []string{
			d.Get("options").(string),
		},
		"templates": []DeviceTemplate{
			template,
		},
	}

	resp, err := wugResty.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token).
		SetBody(params).
		Patch(config.URL + "/devices/-/config/template")

	if err != nil {
		return err
	} else if resp.StatusCode() != 200 {
		return errors.New(string(resp.Body()))
	}

	deviceID := gjson.GetBytes(resp.Body(), "data.idMap.0.resultId").String()

	if len(deviceID) == 0 {
		return errors.New(string(resp.Body()))
	}

	d.SetId(deviceID)

	log.Printf("[WUG] Created device with ID: %s\n", d.Id())

	return resourceDeviceRead(d, m)
}

func resourceDeviceRead(d *schema.ResourceData, m interface{}) error {
	resty := m.(*Client).Resty
	token := m.(*Client).Token
	config := m.(*Client).Config

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

	var template DeviceTemplate
	err = json.Unmarshal([]byte(gjson.GetBytes(resp.Body(), "data.templates.0").Raw), &template)
	if err != nil {
		return err
	}

	d.Set("name", template.Name)
	d.Set("groups", template.Groups)

	/* Reformat arrays since the field names may change... */
	interfaces := make([]map[string]interface{}, 0)
	for _, iface := range template.Interfaces {
		interfaces = append(interfaces, map[string]interface{}{
			"default":                 iface.IsDefault,
			"poll_using_network_name": iface.PollUsingNetworkName,
			"network_address":         iface.NetworkAddress,
			"network_name":            iface.NetworkName,
		})
	}

	d.Set("interface", interfaces)

	credentials := make([]map[string]interface{}, 0)
	for _, cred := range template.Credentials {
		credentials = append(credentials, map[string]interface{}{
			"type": cred.CredentialType,
			"name": cred.Name,
		})
	}

	d.Set("credential", credentials)

	activeMonitors := make([]map[string]interface{}, 0)
	for _, mon := range template.ActiveMonitors {
		activeMonitors = append(activeMonitors, map[string]interface{}{
			"name":          mon.Name,
			"argument":      mon.Argument,
			"comment":       mon.Comment,
			"critical":      mon.IsCritical,
			"polling_order": mon.PollingOrder,
		})
	}

	d.Set("active_monitor", activeMonitors)

	d.Set("performance_monitor", template.PerformanceMonitors)

	d.Set("device_type", template.DeviceType)
	d.Set("snmp_oid", template.SnmpOid)
	d.Set("primary_role", template.PrimaryRole)
	d.Set("subroles", template.SubRoles)
	d.Set("os", template.Os)
	d.Set("brand", template.Brand)
	d.Set("action_policy", template.ActionPolicy)

	return nil
}

func resourceDeviceUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceDeviceRead(d, m)
}

func resourceDeviceDelete(d *schema.ResourceData, m interface{}) error {
	resty := m.(*Client).Resty
	token := m.(*Client).Token
	config := m.(*Client).Config

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
