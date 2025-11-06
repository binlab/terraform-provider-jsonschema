package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"schema_version": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "draft/2020-12",
					Description: "Default JSON Schema version to use when not specified in schema document. Supported values: `draft-04`, `draft-06`, `draft-07`, `draft/2019-09`, `draft/2020-12`",
				},
				"base_url": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Default base URL for resolving relative `$ref` URIs in schemas. Can be overridden per data source.",
				},
				"error_message_template": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "JSON Schema validation failed: {error}",
					Description: "Default error message template for validation failures. Can be overridden per data source. Available variables: {{.Error}}, {{.Schema}}, {{.Document}}, {{.Path}}",
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"jsonschema_validator": dataSourceJsonschemaValidator(),
			},
			ConfigureContextFunc: providerConfigure,
		}

		return p
	}
}

func Provider() *schema.Provider {
	return New("dev")()
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	schemaVersion := d.Get("schema_version").(string)
	baseURL := d.Get("base_url").(string)
	errorTemplate := d.Get("error_message_template").(string)

	config, err := NewProviderConfig(schemaVersion, baseURL, errorTemplate)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return config, diags
}
