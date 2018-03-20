package azurerm

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/scheduler/mgmt/2016-03-01/scheduler"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/response"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmSchedulerJobCollection() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmSchedulerJobCollectionCreateUpdate,
		Read:   resourceArmSchedulerJobCollectionRead,
		Update: resourceArmSchedulerJobCollectionCreateUpdate,
		Delete: resourceArmSchedulerJobCollectionDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"location": locationSchema(),

			"resource_group_name": resourceGroupNameSchema(),

			"tags": tagsSchema(),

			"sku": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
				ValidateFunc: validation.StringInSlice([]string{
					string(scheduler.Free),
					string(scheduler.Standard),
					string(scheduler.P10Premium),
					string(scheduler.P20Premium),
				}, true),
			},

			//optional
			"state": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          string(scheduler.Enabled),
				DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
				ValidateFunc: validation.StringInSlice([]string{
					string(scheduler.Enabled),
					string(scheduler.Suspended),
					string(scheduler.Disabled),
				}, true),
			},

			"quota": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{

						//max_job_occurrence doesn't seem to do anything and always remains empty

						"max_job_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},

						"max_recurrence_frequency": {
							Type:             schema.TypeString,
							Required:         true,
							DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
							ValidateFunc: validation.StringInSlice([]string{
								string(scheduler.Minute),
								string(scheduler.Hour),
								string(scheduler.Day),
								string(scheduler.Week),
								string(scheduler.Month),
							}, true),
						},

						//this sets MaxRecurrance.Interval, and the documentation in the api states:
						//  Gets or sets the interval between retries.
						"max_retry_interval": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1), //changes depending on the frequency, unknown maximums
						},
					},
				},
			},
		},
	}
}

func resourceArmSchedulerJobCollectionCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).schedulerJobCollectionsClient
	ctx := meta.(*ArmClient).StopContext

	name := d.Get("name").(string)
	location := d.Get("location").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	tags := d.Get("tags").(map[string]interface{})

	log.Printf("[DEBUG] Creating/updating Scheduler Job Collection %q (resource group %q)", name, resourceGroup)

	collection := scheduler.JobCollectionDefinition{
		Location: utils.String(location),
		Tags:     expandTags(tags),
		Properties: &scheduler.JobCollectionProperties{
			Sku: &scheduler.Sku{
				Name: scheduler.SkuDefinition(d.Get("sku").(string)),
			},
		},
	}

	if state, ok := d.Get("state").(string); ok {
		collection.Properties.State = scheduler.JobCollectionState(state)
	}
	collection.Properties.Quota = expandAzureArmSchedulerJobCollectionQuota(d)

	//create job collection
	collection, err := client.CreateOrUpdate(ctx, resourceGroup, name, collection)
	if err != nil {
		return fmt.Errorf("Error creating/updating Scheduler Job Collection %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	//ensure collection actually exists and we have the correct ID
	collection, err = client.Get(ctx, resourceGroup, name)
	if err != nil {
		return fmt.Errorf("Error reading Scheduler Job Collection %q after create/update (Resource Group %q): %+v", name, resourceGroup, err)
	}

	d.SetId(*collection.ID)

	return resourceArmSchedulerJobCollectionPopulate(d, resourceGroup, &collection)
}

func resourceArmSchedulerJobCollectionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).schedulerJobCollectionsClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	name := id.Path["jobCollections"]
	resourceGroup := id.ResourceGroup

	log.Printf("[DEBUG] Reading Scheduler Job Collection %q (resource group %q)", name, resourceGroup)

	collection, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(collection.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error making Read request on Scheduler Job Collection %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	return resourceArmSchedulerJobCollectionPopulate(d, resourceGroup, &collection)
}

func resourceArmSchedulerJobCollectionPopulate(d *schema.ResourceData, resourceGroup string, resp *scheduler.JobCollectionDefinition) error {

	//standard properties
	d.Set("name", resp.Name)
	d.Set("location", azureRMNormalizeLocation(*resp.Location))
	d.Set("resource_group_name", resourceGroup)

	//resource specific
	if properties := resp.Properties; properties != nil {
		if sku := properties.Sku; sku != nil {
			d.Set("sku", sku.Name)
		}
		d.Set("state", string(properties.State))

		if err := d.Set("quota", flattenAzureArmSchedulerJobCollectionQuota(properties.Quota)); err != nil {
			return fmt.Errorf("Error flattening quota for Job Collection %q (Resource Group %q): %+v", resp.Name, resourceGroup, err)
		}
	}

	if err := flattenAndSetTags(d, &resp.Tags); err != nil {
		return fmt.Errorf("Error flattening `tags`: %+v", err)
	}

	return nil
}

func resourceArmSchedulerJobCollectionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).schedulerJobCollectionsClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	name := id.Path["jobCollections"]
	resourceGroup := id.ResourceGroup

	log.Printf("[DEBUG] Deleting Scheduler Job Collection %q (resource group %q)", name, resourceGroup)

	future, err := client.Delete(ctx, resourceGroup, name)
	if err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("Error issuing delete request for Scheduler Job Collection %q (Resource Group %q): %+v", name, resourceGroup, err)
		}
	}

	err = future.WaitForCompletion(ctx, client.Client)
	if err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("Error waiting for deletion of Scheduler Job Collection %q (Resource Group %q): %+v", name, resourceGroup, err)
		}
	}

	return nil
}

func expandAzureArmSchedulerJobCollectionQuota(d *schema.ResourceData) *scheduler.JobCollectionQuota {
	if qb, ok := d.Get("quota").([]interface{}); ok && len(qb) > 0 {
		quota := scheduler.JobCollectionQuota{
			MaxRecurrence: &scheduler.JobMaxRecurrence{},
		}

		quotaBlock := qb[0].(map[string]interface{})

		if v, ok := quotaBlock["max_job_count"].(int); ok {
			quota.MaxJobCount = utils.Int32(int32(v))
		}
		if v, ok := quotaBlock["max_recurrence_frequency"].(string); ok {
			quota.MaxRecurrence.Frequency = scheduler.RecurrenceFrequency(v)
		}
		if v, ok := quotaBlock["max_retry_interval"].(int); ok {
			quota.MaxRecurrence.Interval = utils.Int32(int32(v))
		}

		return &quota
	}

	return nil
}

func flattenAzureArmSchedulerJobCollectionQuota(quota *scheduler.JobCollectionQuota) []interface{} {
	if quota == nil {
		return nil
	}

	quotaBlock := make(map[string]interface{})

	if v := quota.MaxJobCount; v != nil {
		quotaBlock["max_job_count"] = *v
	}
	if recurrence := quota.MaxRecurrence; recurrence != nil {
		if v := recurrence.Interval; v != nil {
			quotaBlock["max_retry_interval"] = *v
		}

		quotaBlock["max_recurrence_frequency"] = string(recurrence.Frequency)
	}

	return []interface{}{quotaBlock}
}