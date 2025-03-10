package transfer

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceAccess() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccessCreate,
		ReadWithoutTimeout:   resourceAccessRead,
		UpdateWithoutTimeout: resourceAccessUpdate,
		DeleteWithoutTimeout: resourceAccessDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"external_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},

			"home_directory": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},

			"home_directory_mappings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 50,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"entry": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
						"target": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
					},
				},
			},

			"home_directory_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      transfer.HomeDirectoryTypePath,
				ValidateFunc: validation.StringInSlice(transfer.HomeDirectoryType_Values(), false),
			},

			"policy": {
				Type:                  schema.TypeString,
				Optional:              true,
				ValidateFunc:          verify.ValidIAMPolicyJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},

			"posix_profile": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gid": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"uid": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"secondary_gids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeInt},
							Optional: true,
						},
					},
				},
			},

			"role": {
				Type: schema.TypeString,
				// Although Role is required in the API it is not currently returned on Read.
				// Required:     true,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},

			"server_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validServerID,
			},
		},
	}
}

func resourceAccessCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn()

	externalID := d.Get("external_id").(string)
	serverID := d.Get("server_id").(string)
	id := AccessCreateResourceID(serverID, externalID)
	input := &transfer.CreateAccessInput{
		ExternalId: aws.String(externalID),
		ServerId:   aws.String(serverID),
	}

	if v, ok := d.GetOk("home_directory"); ok {
		input.HomeDirectory = aws.String(v.(string))
	}

	if v, ok := d.GetOk("home_directory_mappings"); ok {
		input.HomeDirectoryMappings = expandHomeDirectoryMappings(v.([]interface{}))
	}

	if v, ok := d.GetOk("home_directory_type"); ok {
		input.HomeDirectoryType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("policy"); ok {
		policy, err := structure.NormalizeJsonString(v.(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", v.(string), err)
		}

		input.Policy = aws.String(policy)
	}

	if v, ok := d.GetOk("posix_profile"); ok {
		input.PosixProfile = expandUserPOSIXUser(v.([]interface{}))
	}

	if v, ok := d.GetOk("role"); ok {
		input.Role = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Transfer Access: %s", input)
	_, err := conn.CreateAccessWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Transfer Access (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceAccessRead(ctx, d, meta)...)
}

func resourceAccessRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn()

	serverID, externalID, err := AccessParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing Transfer Access ID: %s", err)
	}

	access, err := FindAccessByServerIDAndExternalID(ctx, conn, serverID, externalID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transfer Access (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Transfer Access (%s): %s", d.Id(), err)
	}

	d.Set("external_id", access.ExternalId)
	d.Set("home_directory", access.HomeDirectory)
	if err := d.Set("home_directory_mappings", flattenHomeDirectoryMappings(access.HomeDirectoryMappings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting home_directory_mappings: %s", err)
	}
	d.Set("home_directory_type", access.HomeDirectoryType)

	if err := d.Set("posix_profile", flattenUserPOSIXUser(access.PosixProfile)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting posix_profile: %s", err)
	}
	// Role is currently not returned via the API.
	// d.Set("role", access.Role)

	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), aws.StringValue(access.Policy))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Transfer Access (%s): %s", d.Id(), err)
	}

	d.Set("policy", policyToSet)

	d.Set("server_id", serverID)

	return diags
}

func resourceAccessUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn()

	serverID, externalID, err := AccessParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing Transfer Access ID: %s", err)
	}

	input := &transfer.UpdateAccessInput{
		ExternalId: aws.String(externalID),
		ServerId:   aws.String(serverID),
	}

	if d.HasChange("home_directory") {
		input.HomeDirectory = aws.String(d.Get("home_directory").(string))
	}

	if d.HasChange("home_directory_mappings") {
		input.HomeDirectoryMappings = expandHomeDirectoryMappings(d.Get("home_directory_mappings").([]interface{}))
	}

	if d.HasChange("home_directory_type") {
		input.HomeDirectoryType = aws.String(d.Get("home_directory_type").(string))
	}

	if d.HasChange("policy") {
		policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", d.Get("policy").(string), err)
		}

		input.Policy = aws.String(policy)
	}

	if d.HasChange("posix_profile") {
		input.PosixProfile = expandUserPOSIXUser(d.Get("posix_profile").([]interface{}))
	}

	if d.HasChange("role") {
		input.Role = aws.String(d.Get("role").(string))
	}

	log.Printf("[DEBUG] Updating Transfer Access: %s", input)
	_, err = conn.UpdateAccessWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Transfer Access (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAccessRead(ctx, d, meta)...)
}

func resourceAccessDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn()

	serverID, externalID, err := AccessParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing Transfer Access ID: %s", err)
	}

	log.Printf("[DEBUG] Deleting Transfer Access: %s", d.Id())
	_, err = conn.DeleteAccessWithContext(ctx, &transfer.DeleteAccessInput{
		ExternalId: aws.String(externalID),
		ServerId:   aws.String(serverID),
	})

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Transfer Access (%s): %s", d.Id(), err)
	}

	return diags
}
