package wug

import (
	"encoding/json"
	"errors"
//	"fmt"
//	"log"
//	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/tidwall/gjson"
)


// MonitorSearch is WUG's internal object.
type MonitorInfo struct {
	Type                string                             `json:"type,omitempty"`
	Search	            string                             `json:"search,omitempty"`
	ClassId	            string                             `json:"class_id,omitempty"`
	MonitorId           string                             `json:"monitor_id,omitempty"`
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
			"monitor_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "ID of the monitor",
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

	var data []MonitorInfo
	err = json.Unmarshal([]byte(gjson.GetBytes(resp.Body(), "data.activeMonitors").Raw), &data)
	if err != nil {
		return err
	}

	class_id := data[0].ClassId
	monitor_id := data[0].MonitorId
	d.Set("class_id", class_id)
	d.Set("monitor_id", monitor_id)

	return nil
}

