package wug

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/tidwall/gjson"
)


// MonitorSearch is WUG's internal object.
type MonitorInfo struct {
	Type                string                             `json:"type,omitempty"`
	Search	            string                             `json:"search,omitempty"`
	ClassId	            string                             `json:"classId,omitempty"`
	MonitorName         string                             `json:"monitorName,omitempty"`
}

// MonitorTypeInfo is WUG's internal object.
type MonitorTypeInfo struct {
	ClassId                 string                             `json:"classId,omitempty"`
	BaseType	        string                             `json:"baseType,omitempty"`
}

// MonitorTemplate is WUG's internal object.
type MonitorSearchTemplate struct {
	MonitorId               string                             `json:"monitorId,omitempty"`
	Name			string                             `json:"name,omitempty"`
	Description             string                             `json:"description,omitempty"`
	Id			string                             `json:"id,omitempty"`
	MonitorTypeInfo		MonitorTypeInfo                    `json:"monitorTypeInfo,omitempty"`
}

func dataSourceMonitor() *schema.Resource {
	return &schema.Resource{
		Read:   dataSourceMonitorRead,
		Schema: map[string]*schema.Schema{
			"type": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Type of the device",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"active",
					"performance",
//					"passive",
				}, true),
			},
			"search": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Name of the monitor to look for",
				Required:    true,
				ForceNew:    true,
			},
			"class_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "ID of the monitor type class",
				Computed:    true,
				ForceNew:    true,
			},
			"monitor_name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Name of the monitor",
				Computed:    true,
				ForceNew:    true,
			},
		},
	}
}

func dataSourceMonitorRead(d *schema.ResourceData, m interface{}) error {
	resty := m.(*Client).Resty
	token := m.(*Client).Token
	config := m.(*Client).Config

	params := map[string]string{
		"type": d.Get("type").(string),
		"search": d.Get("search").(string),
		"includeDeviceMonitors": "true",
		"includeSystemMonitors": "true",
		"includeCoreMonitors": "true",
	}

	resp, err := resty.R().
		SetQueryParams(params).
		SetHeader("Accept", "application/json").
		SetAuthToken(token).
		Get(config.URL + "/monitors/-")

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
	if monitorCount == 0 {
		return fmt.Errorf("Found no monitor for " + d.Get("search").(string))
	}

	var data MonitorSearchTemplate
	var mode string
	if d.Get("type").(string) == "active" {
		mode = "data.activeMonitors.0"
	}
	if d.Get("type").(string) == "performance" {
		mode = "data.performanceMonitors.0"
	}
	err = json.Unmarshal([]byte(gjson.GetBytes(resp.Body(), mode).Raw), &data)
	if err != nil {
		return err
	}

	d.Set("class_id", data.MonitorTypeInfo.ClassId)
	d.Set("monitor_name", data.Name)
	d.SetId(data.MonitorId)

	return nil
}

