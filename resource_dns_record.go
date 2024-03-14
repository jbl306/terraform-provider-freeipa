package main

import (
	"context"
	"log"
	"strings"
	"fmt"

	ipa "github.com/RomanButsiy/go-freeipa/freeipa"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFreeIPADNSRecord() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFreeIPADNSDNSRecordCreate,
		ReadContext:   resourceFreeIPADNSDNSRecordRead,
		UpdateContext: resourceFreeIPADNSDNSRecordUpdate,
		DeleteContext: resourceFreeIPADNSDNSRecordDelete,
		Importer: &schema.ResourceImporter{
			State: resourceFreeIPADNSDNSRecordImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Record name",
			},
			"zone_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Zone name (FQDN)",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The record type",
			},
			"records": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
				Description: "A string list of records",
			},
			"ttl": {
				Type:        schema.TypeInt,
				Computed:	 true,
				Optional:    true,
				Description: "Time to live",
			},
			"set_identifier": {
				Type:        schema.TypeString,
				Computed:	 true,
				Optional:    true,
				Description: "Unique identifier to differentiate records with routing policies from one another",
			},
		},
	}
}

func resourceFreeIPADNSDNSRecordCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Creating freeipa dns record")

	client, err := meta.(*Config).Client()
	if err != nil {
		return diag.Errorf("Error creating freeipa identity client: %s", err)
	}

	name := d.Get("name").(string)
	zone_name := d.Get("zone_name").(string)

	args := ipa.DnsrecordAddArgs{
		Idnsname: name,
	}

	optArgs := ipa.DnsrecordAddOptionalArgs{
		Dnszoneidnsname: &zone_name,
	}

	_type := d.Get("type").(string)
	_records := d.Get("records").(*schema.Set).List()
	records := make([]string, len(_records))
	for i, d := range _records {
		records[i] = d.(string)
	}
	switch _type {
	case "A":
		optArgs.Arecord = &records
	case "AAAA":
		optArgs.Aaaarecord = &records
	case "CNAME":
		optArgs.Cnamerecord = &records
	case "MX":
		optArgs.Mxrecord = &records
	case "NS":
		optArgs.Nsrecord = &records
	case "PTR":
		optArgs.Ptrrecord = &records
	case "SRV":
		optArgs.Srvrecord = &records
	case "TXT":
		optArgs.Txtrecord = &records
	case "SSHFP":
		optArgs.Sshfprecord = &records
	}

	if _v, ok := d.GetOkExists("ttl"); ok {
		v := _v.(int)
		optArgs.Dnsttl = &v
	}

	_, err = client.DnsrecordAdd(&args, &optArgs)
	if err != nil {
		if strings.Contains(err.Error(), "EmptyModlist") {
			log.Printf("[DEBUG] EmptyModlist (4202): no modifications to be performed")
		} else {
			return diag.Errorf("Error creating freeipa dns record: %s", err)
		}
	}

	// Generate an ID
	vars := []string{
		zone_name,
		strings.ToLower(name),
		_type,
	}
	if v, ok := d.GetOk("set_identifier"); ok {
		vars = append(vars, v.(string))
	}

	d.SetId(strings.Join(vars, "_"))

	return resourceFreeIPADNSDNSRecordRead(ctx, d, meta)
}

func resourceFreeIPADNSDNSRecordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Read freeipa dns record")

	client, err := meta.(*Config).Client()
	if err != nil {
		return diag.Errorf("Error creating freeipa identity client: %s", err)
	}

	args := ipa.DnsrecordShowArgs{
		Idnsname: d.Get("name").(string),
	}

	zone_name := d.Get("zone_name").(string)
	all := true
	optArgs := ipa.DnsrecordShowOptionalArgs{
		Dnszoneidnsname: &zone_name,
		All:             &all,
	}

	res, err := client.DnsrecordShow(&args, &optArgs)
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			d.SetId("")
			log.Printf("[DEBUG] DNS record not found")
			return nil
		} else {
			return diag.Errorf("Error reading freeipa DNS record: %s", err)
		}
	}

	if res.Result.Idnsname != "" {
		d.Set("name", res.Result.Idnsname)
	}
	if zone_name != "" {
		d.Set("zone_name", zone_name)
	}

	_type := d.Get("type")

	switch _type {
	case "A":
		if res.Result.Arecord != nil {
			d.Set("records", res.Result.Arecord)
		}
	case "AAAA":
		if res.Result.Aaaarecord != nil {
			d.Set("records", res.Result.Aaaarecord)
		}
	case "MX":
		if res.Result.Mxrecord != nil {
			d.Set("records", res.Result.Mxrecord)
		}
	case "NS":
		if res.Result.Nsrecord != nil {
			d.Set("records", res.Result.Nsrecord)
		}
	case "PTR":
		if res.Result.Ptrrecord != nil {
			d.Set("records", res.Result.Ptrrecord)
		}
	case "SRV":
		if res.Result.Srvrecord != nil {
			d.Set("records", res.Result.Srvrecord)
		}
	case "TXT":
		if res.Result.Txtrecord != nil {
			d.Set("records", res.Result.Txtrecord)
		}
	case "SSHFP":
		if res.Result.Sshfprecord != nil {
			d.Set("records", res.Result.Sshfprecord)
		}
	}

	//d.Set("set_identifier", res.Result.Idnsrecordsetidentifier)
	if res.Result.Dnsttl != nil {
		d.Set("ttl", res.Result.Dnsttl)
	}

	return nil
}

func resourceFreeIPADNSDNSRecordUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Update freeipa dns record")

	client, err := meta.(*Config).Client()
	if err != nil {
		return diag.Errorf("Error creating freeipa identity client: %s", err)
	}

	args := ipa.DnsrecordModArgs{
		Idnsname: d.Get("name").(string),
	}

	zone_name := d.Get("zone_name").(string)
	optArgs := ipa.DnsrecordModOptionalArgs{
		Dnszoneidnsname: &zone_name,
	}

	_type := d.Get("type")
	_records := d.Get("records").(*schema.Set).List()
	records := make([]string, len(_records))
	for i, d := range _records {
		records[i] = d.(string)
	}
	switch _type {
	case "A":
		optArgs.Arecord = &records
	case "AAAA":
		optArgs.Aaaarecord = &records
	case "CNAME":
		optArgs.Cnamerecord = &records
	case "MX":
		optArgs.Mxrecord = &records
	case "NS":
		optArgs.Nsrecord = &records
	case "PTR":
		optArgs.Ptrrecord = &records
	case "SRV":
		optArgs.Srvrecord = &records
	case "TXT":
		optArgs.Txtrecord = &records
	case "SSHFP":
		optArgs.Sshfprecord = &records
	}

	if _v, ok := d.GetOkExists("ttl"); ok {
		v := _v.(int)
		optArgs.Dnsttl = &v
	}

	_, err = client.DnsrecordMod(&args, &optArgs)
	if err != nil {
		if strings.Contains(err.Error(), "EmptyModlist") {
			log.Printf("[DEBUG] EmptyModlist (4202): no modifications to be performed")
		} else {
			return diag.Errorf("Error update freeipa dns record: %s", err)
		}
	}
	return resourceFreeIPADNSDNSRecordRead(ctx, d, meta)
}

func resourceFreeIPADNSDNSRecordDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Delete freeipa dns record")

	client, err := meta.(*Config).Client()
	if err != nil {
		return diag.Errorf("Error creating freeipa identity client: %s", err)
	}
	args := ipa.DnsrecordDelArgs{
		Idnsname: d.Get("name").(string),
	}

	zone_name := d.Get("zone_name").(string)
	optArgs := ipa.DnsrecordDelOptionalArgs{
		Dnszoneidnsname: &zone_name,
	}

	_type := d.Get("type")
	_records := d.Get("records").(*schema.Set).List()
	records := make([]string, len(_records))
	for i, d := range _records {
		records[i] = d.(string)
	}
	switch _type {
	case "A":
		optArgs.Arecord = &records
	case "AAAA":
		optArgs.Aaaarecord = &records
	case "CNAME":
		optArgs.Cnamerecord = &records
	case "MX":
		optArgs.Mxrecord = &records
	case "NS":
		optArgs.Nsrecord = &records
	case "PTR":
		optArgs.Ptrrecord = &records
	case "SRV":
		optArgs.Srvrecord = &records
	case "TXT":
		optArgs.Txtrecord = &records
	case "SSHFP":
		optArgs.Sshfprecord = &records
	}

	_, err = client.DnsrecordDel(&args, &optArgs)
	if err != nil {
		return diag.Errorf("Error delete freeipa dns record: %s", err)
	}

	d.SetId("")
	return nil
}

func resourceFreeIPADNSDNSRecordImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error){
	parts := strings.Split(d.Id(), "/")
    if len(parts) != 3 {
        return []*schema.ResourceData{}, fmt.Errorf("Invalid ID: The expected format is <record name>/<zone name>/<record type>")
    }

    d.Set("name", parts[0])
    d.Set("zone_name", parts[1])
	d.Set("type", parts[2])

	record_name := parts[0]
	zone_name := parts[1]
	_type := parts[2]

	// Generate an ID
	vars := []string{
		zone_name,
		strings.ToLower(record_name),
		_type,
	}

	d.SetId(strings.Join(vars, "_"))


    return []*schema.ResourceData{d}, nil
}

