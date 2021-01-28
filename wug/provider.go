package wug

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider exports WUG terraform provider schemas.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"user": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("WUG_USER", nil),
				Description: "The user name for WUG API operations.",
			},

			"password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("WUG_PASSWORD", nil),
				Description: "The user password for WUG API operations.",
			},

			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WUG_URL", nil),
				Description: "The WUG endpoint for WUG API operations.",
			},
			"allow_unverified_ssl": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WUG_ALLOW_UNVERIFIED_SSL", false),
				Description: "If set, WUG client will permit unverifiable SSL certificates.",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"wug_monitor": dataSourceMonitor(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"wug_device": resourceDevice(),
			"wug_monitor": resourceMonitor(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	c, err := NewConfig(d)
	if err != nil {
		return nil, err
	}
	return c.Client()
}
