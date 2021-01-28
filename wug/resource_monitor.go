package wug

import (
//	"encoding/json"
	"errors"
	"fmt"
	"log"
//	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/tidwall/gjson"
)


// MonitorActiveParameters is WUG's internal object.
type MonitorActiveParameters struct {
	CriticalOrder		int  `json:"criticalOrder,omitempty"`
	ActionPolicyName	string `json:"actionPolicyName,omitempty"`
	ActionPolicyId		string `json:"actionPolicyId,omitempty"`
	Comment			string `json:"comment,omitempty"`
	Argument		string `json:"argument,omitempty"`
	PollingIntervalSeconds  int  `json:"pollingIntervalSeconds,omitempty"`
	InterfaceId		string `json:"interfaceId,string,omitempty"`
}

// MonitorPerformanceParameters is WUG's internal object.
type MonitorPerformanceParameters struct {
	PollingIntervalMinutes int `json:"polingIntervalMinutes,omitempty"`
}

// MonitorParameters is WUG's internal object.
type MonitorTemplate struct {
	Type                string                             `json:"type,omitempty"`
	IsGlobal            bool                               `json:"isGlobal,omitempty"`
	Enabled             bool                               `json:"enabled,omitempty"`
	MonitorTypeClassId  string                             `json:"monitorTypeClassId,omitempty"`
	MonitorTypeId       string                             `json:"monitorType,omitempty"`
	MonitorTypeName     string                             `json:"monitorTypeName,omitempty"`
	Active              MonitorActiveParameters            `json:"active,omitempty"`
	Performance         MonitorPerformanceParameters       `json:"performance,omitempty"`
}

func resourceMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceMonitorCreate,
		Read:   resourceMonitorRead,
		/* No Update path since all fields are ForceNew. */
		/* Update: resourceMonitorUpdate, */
		Delete: resourceMonitorDelete,

		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:        schema.TypeString,
				Description: "ID of the device to assign the monitor.",
				Required:    true,
				ForceNew:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "Type of the monitor.",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"active",
					"performance",
				}, true),
			},
			"is_global": {
				Type:        schema.TypeBool,
				Description: "If the monitor is from the global monitor library.",
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "If the monitor assignment get enabled.",
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
			"monitor_type_class_id": {
				Type:        schema.TypeString,
				Description: "ID of the monitor type class.",
				Optional:    true,
				ForceNew:    true,
			},
			"monitor_type_id": {
				Type:        schema.TypeString,
				Description: "ID of the monitor type.",
				Optional:    true,
				ForceNew:    true,
			},
			"monitor_type_name": {
				Type:        schema.TypeString,
				Description: "Name of the monitor type.",
				Optional:    true,
				ForceNew:    true,
			},
			"active": {
				Type:        schema.TypeList,
				Description: "Parameters of an active monitor.",
				Optional:    true,
				ForceNew:    true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"critical_order": {
							Type:        schema.TypeInt,
							Description: "Critical order of the monitor.",
							Optional:    true,
						},
						"action_policy_name": {
							Type:        schema.TypeString,
							Description: "Name of the action policy to apply on the monitor.",
							Optional:    true,
						},
						"action_policy_id": {
							Type:        schema.TypeString,
							Description: "ID of the action policy to apply on the monitor.",
							Optional:    true,
						},
						"comment": {
							Type:        schema.TypeString,
							Description: "Monitor comment.",
							Optional:    true,
						},
						"argument": {
							Type:        schema.TypeString,
							Description: "Monitor argument.",
							Optional:    true,
						},
						"polling_interval_seconds": {
							Type:        schema.TypeInt,
							Description: "Polling interval of the monitor.",
							Optional:    true,
						},
						"interface_id": {
							Type:        schema.TypeString,
							Description: "Network interface ID.",
							Optional:    true,
						},
					},
				},
			},
			"performance": {
				Type:        schema.TypeList,
				Description: "Parameters of a performance monitor.",
				Optional:    true,
				ForceNew:    true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"polling_interval_minutes": {
							Type:        schema.TypeInt,
							Description: "Polling interval of the monitor.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func resourceMonitorCreate(d *schema.ResourceData, m interface{}) error {
	wugResty := m.(*Client).Resty
	token := m.(*Client).Token
	config := m.(*Client).Config

	var monitor MonitorTemplate

	/* Build our object. */

	monitor.Type = d.Get("type").(string)
	monitor.IsGlobal = d.Get("is_global").(bool)
	monitor.Enabled = d.Get("enabled").(bool)
	monitor.MonitorTypeClassId = d.Get("monitor_type_class_id").(string)
	monitor.MonitorTypeId = d.Get("monitor_type_id").(string)
	monitor.MonitorTypeName = d.Get("monitor_type_name").(string)

	if len(d.Get("active").([]interface{})) > 0 {
		activeData := d.Get("active").([]interface{})[0].(map[string]interface{})
		monitor.Active.CriticalOrder = activeData["critical_order"].(int)
		monitor.Active.ActionPolicyName = activeData["action_policy_name"].(string)
		monitor.Active.ActionPolicyId = activeData["action_policy_id"].(string)
		monitor.Active.Comment = activeData["comment"].(string)
		monitor.Active.Argument = activeData["argument"].(string)
		monitor.Active.PollingIntervalSeconds = activeData["polling_interval_seconds"].(int)
		monitor.Active.InterfaceId = activeData["interface_id"].(string)
	}

	if len(d.Get("performance").([]interface{})) > 0 {
		performanceData := d.Get("performance").([]interface{})[0].(map[string]interface{})
		monitor.Performance.PollingIntervalMinutes = performanceData["polling_interval_minutes"].(int)
	}

	var deviceId = d.Get("device_id").(string)

//	params := map[string]interface{}{
//		"type": monitor.Type,
//		"is_global": monitor.IsGlobal,
//		"enabled": monitor.Enabled,
//		"monitor_type_class_id": monitor.MonitorTypeClassId,
//		"monitor_type_id": monitor.MonitorTypeId,
//		"monitor_type_name": monitor.MonitorTypeName,
//		"active": monitor.Active,
//		"performance": monitor.Performance,
//	}

	params := monitor

	resp, err := wugResty.SetDebug(true).R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token).
		SetBody(params).
		Patch(config.URL + "/devices/" + deviceId + "/monitors/-")

	if err != nil {
		return err
	} else if resp.StatusCode() != 200 {
		return errors.New(string(resp.Body()))
	}

	monitorID := gjson.GetBytes(resp.Body(), "data.idMap.0.resultId").String()

	if len(monitorID) == 0 {
		return errors.New(string(resp.Body()))
	}

	d.SetId(monitorID)

	log.Printf("[WUG] Created monitor with ID: %s\n", d.Id())

	return resourceMonitorRead(d, m)
}

func resourceMonitorRead(d *schema.ResourceData, m interface{}) error {
	resty := m.(*Client).Resty
	token := m.(*Client).Token
	config := m.(*Client).Config

	id := d.Id()
	var deviceId = d.Get("device_id").(string)

	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token).
		Get(config.URL + "/devices/" + deviceId + "/monitors/" + id )

	if err != nil {
		return err
	} else if resp.StatusCode() != 200 {
		return errors.New(string(resp.Body()))
	} else if resp.StatusCode() == 404 {
		/* The monitor does not exist anymore. */
		d.SetId("")
		return errors.New(string(resp.Body()))
	}

	monitorCount := gjson.GetBytes(resp.Body(), "paging.size").Int()

	if monitorCount != 1 {
		return fmt.Errorf("Found invalid monitor count for %s: %d", id, monitorCount)
	}

	var monitor MonitorTemplate

	d.Set("type", monitor.Type)
	d.Set("is_global", monitor.IsGlobal)
	d.Set("enabled", monitor.Enabled)
	d.Set("monitor_type_class_id", monitor.MonitorTypeClassId)
	d.Set("monitor_type_id", monitor.MonitorTypeId)
	d.Set("monitor_type_name", monitor.MonitorTypeName)
	d.Set("active", monitor.Active)
	d.Set("performance", monitor.Performance)

	return nil
}

func resourceMonitorUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceMonitorRead(d, m)
}

func resourceMonitorDelete(d *schema.ResourceData, m interface{}) error {
	resty := m.(*Client).Resty
	token := m.(*Client).Token
	config := m.(*Client).Config

	id := d.Id()
	var deviceId = d.Get("device_id").(string)

	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token).
		Delete(config.URL + "/devices/" + deviceId + "/monitors/" + id)

	if err != nil {
		return err
	} else if resp.StatusCode() != 200 {
		return errors.New(string(resp.Body()))
	}

	d.SetId("")

	return nil
}
