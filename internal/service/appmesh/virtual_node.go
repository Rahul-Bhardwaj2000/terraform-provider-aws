package appmesh

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVirtualNode() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceVirtualNodeCreate,
		ReadWithoutTimeout:   resourceVirtualNodeRead,
		UpdateWithoutTimeout: resourceVirtualNodeUpdate,
		DeleteWithoutTimeout: resourceVirtualNodeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVirtualNodeImport,
		},

		SchemaVersion: 1,
		MigrateState:  resourceVirtualNodeMigrateState,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"mesh_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"mesh_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},

			"spec": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"backend": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 0,
							MaxItems: 50,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"virtual_service": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"virtual_service_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 255),
												},

												"client_policy": VirtualNodeClientPolicySchema(),
											},
										},
									},
								},
							},
						},

						"backend_defaults": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"client_policy": VirtualNodeClientPolicySchema(),
								},
							},
						},

						"listener": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"connection_pool": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"grpc": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max_requests": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
												},

												"http": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max_connections": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},

															"max_pending_requests": {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
												},

												"http2": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max_requests": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
												},

												"tcp": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max_connections": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
												},
											},
										},
									},

									"health_check": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"healthy_threshold": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(2, 10),
												},

												"interval_millis": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(5000, 300000),
												},

												"path": {
													Type:     schema.TypeString,
													Optional: true,
												},

												"port": {
													Type:         schema.TypeInt,
													Optional:     true,
													Computed:     true,
													ValidateFunc: validation.IsPortNumber,
												},

												"protocol": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(appmesh.PortProtocol_Values(), false),
												},

												"timeout_millis": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(2000, 60000),
												},

												"unhealthy_threshold": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(2, 10),
												},
											},
										},
									},

									"outlier_detection": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"base_ejection_duration": {
													Type:     schema.TypeList,
													Required: true,
													MinItems: 1,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"unit": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
															},

															"value": {
																Type:     schema.TypeInt,
																Required: true,
															},
														},
													},
												},

												"interval": {
													Type:     schema.TypeList,
													Required: true,
													MinItems: 1,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"unit": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
															},

															"value": {
																Type:     schema.TypeInt,
																Required: true,
															},
														},
													},
												},

												"max_ejection_percent": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(0, 100),
												},

												"max_server_errors": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
											},
										},
									},

									"port_mapping": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"port": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IsPortNumber,
												},

												"protocol": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(appmesh.PortProtocol_Values(), false),
												},
											},
										},
									},

									"timeout": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"grpc": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"idle": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"unit": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
																		},

																		"value": {
																			Type:     schema.TypeInt,
																			Required: true,
																		},
																	},
																},
															},

															"per_request": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"unit": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
																		},

																		"value": {
																			Type:     schema.TypeInt,
																			Required: true,
																		},
																	},
																},
															},
														},
													},
												},

												"http": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"idle": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"unit": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
																		},

																		"value": {
																			Type:     schema.TypeInt,
																			Required: true,
																		},
																	},
																},
															},

															"per_request": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"unit": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
																		},

																		"value": {
																			Type:     schema.TypeInt,
																			Required: true,
																		},
																	},
																},
															},
														},
													},
												},

												"http2": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"idle": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"unit": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
																		},

																		"value": {
																			Type:     schema.TypeInt,
																			Required: true,
																		},
																	},
																},
															},

															"per_request": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"unit": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
																		},

																		"value": {
																			Type:     schema.TypeInt,
																			Required: true,
																		},
																	},
																},
															},
														},
													},
												},

												"tcp": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"idle": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"unit": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
																		},

																		"value": {
																			Type:     schema.TypeInt,
																			Required: true,
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},

									"tls": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"certificate": {
													Type:     schema.TypeList,
													Required: true,
													MinItems: 1,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"acm": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"certificate_arn": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: verify.ValidARN,
																		},
																	},
																},
															},

															"file": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"certificate_chain": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringLenBetween(1, 255),
																		},

																		"private_key": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringLenBetween(1, 255),
																		},
																	},
																},
															},

															"sds": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"secret_name": {
																			Type:     schema.TypeString,
																			Required: true,
																		},
																	},
																},
															},
														},
													},
												},

												"mode": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(appmesh.ListenerTlsMode_Values(), false),
												},

												"validation": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"subject_alternative_names": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"match": {
																			Type:     schema.TypeList,
																			Required: true,
																			MinItems: 1,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"exact": {
																						Type:     schema.TypeSet,
																						Required: true,
																						Elem:     &schema.Schema{Type: schema.TypeString},
																						Set:      schema.HashString,
																					},
																				},
																			},
																		},
																	},
																},
															},

															"trust": {
																Type:     schema.TypeList,
																Required: true,
																MinItems: 1,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"file": {
																			Type:     schema.TypeList,
																			Optional: true,
																			MinItems: 0,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"certificate_chain": {
																						Type:         schema.TypeString,
																						Required:     true,
																						ValidateFunc: validation.StringLenBetween(1, 255),
																					},
																				},
																			},
																		},

																		"sds": {
																			Type:     schema.TypeList,
																			Optional: true,
																			MinItems: 0,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"secret_name": {
																						Type:         schema.TypeString,
																						Required:     true,
																						ValidateFunc: validation.StringLenBetween(1, 255),
																					},
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},

						"logging": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"access_log": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"file": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"path": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringLenBetween(1, 255),
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},

						"service_discovery": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"aws_cloud_map": {
										Type:          schema.TypeList,
										Optional:      true,
										MinItems:      0,
										MaxItems:      1,
										ConflictsWith: []string{"spec.0.service_discovery.0.dns"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"attributes": {
													Type:     schema.TypeMap,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},

												"namespace_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 1024),
												},

												"service_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 1024),
												},
											},
										},
									},

									"dns": {
										Type:          schema.TypeList,
										Optional:      true,
										MinItems:      0,
										MaxItems:      1,
										ConflictsWith: []string{"spec.0.service_discovery.0.aws_cloud_map"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"hostname": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.NoZeroValues,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"resource_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tftags.TagsSchema(),

			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

// VirtualNodeClientPolicySchema returns the schema for `client_policy` attributes.
func VirtualNodeClientPolicySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 0,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"tls": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 0,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"certificate": {
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 0,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"file": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 0,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"certificate_chain": {
														Type:         schema.TypeString,
														Required:     true,
														ValidateFunc: validation.StringLenBetween(1, 255),
													},

													"private_key": {
														Type:         schema.TypeString,
														Required:     true,
														ValidateFunc: validation.StringLenBetween(1, 255),
													},
												},
											},
										},

										"sds": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 0,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"secret_name": {
														Type:     schema.TypeString,
														Required: true,
													},
												},
											},
										},
									},
								},
							},

							"enforce": {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  true,
							},

							"ports": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeInt},
								Set:      schema.HashInt,
							},

							"validation": {
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"subject_alternative_names": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 0,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"match": {
														Type:     schema.TypeList,
														Required: true,
														MinItems: 1,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"exact": {
																	Type:     schema.TypeSet,
																	Required: true,
																	Elem:     &schema.Schema{Type: schema.TypeString},
																	Set:      schema.HashString,
																},
															},
														},
													},
												},
											},
										},

										"trust": {
											Type:     schema.TypeList,
											Required: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"acm": {
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 0,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"certificate_authority_arns": {
																	Type:     schema.TypeSet,
																	Required: true,
																	Elem:     &schema.Schema{Type: schema.TypeString},
																	Set:      schema.HashString,
																},
															},
														},
													},

													"file": {
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 0,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"certificate_chain": {
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: validation.StringLenBetween(1, 255),
																},
															},
														},
													},

													"sds": {
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 0,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"secret_name": {
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: validation.StringLenBetween(1, 255),
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceVirtualNodeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	req := &appmesh.CreateVirtualNodeInput{
		MeshName:        aws.String(d.Get("mesh_name").(string)),
		VirtualNodeName: aws.String(d.Get("name").(string)),
		Spec:            expandVirtualNodeSpec(d.Get("spec").([]interface{})),
		Tags:            Tags(tags.IgnoreAWS()),
	}
	if v, ok := d.GetOk("mesh_owner"); ok {
		req.MeshOwner = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating App Mesh virtual node: %s", req)
	resp, err := conn.CreateVirtualNodeWithContext(ctx, req)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating App Mesh virtual node: %s", err)
	}

	d.SetId(aws.StringValue(resp.VirtualNode.Metadata.Uid))

	return append(diags, resourceVirtualNodeRead(ctx, d, meta)...)
}

func resourceVirtualNodeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	req := &appmesh.DescribeVirtualNodeInput{
		MeshName:        aws.String(d.Get("mesh_name").(string)),
		VirtualNodeName: aws.String(d.Get("name").(string)),
	}
	if v, ok := d.GetOk("mesh_owner"); ok {
		req.MeshOwner = aws.String(v.(string))
	}

	var resp *appmesh.DescribeVirtualNodeOutput

	err := resource.RetryContext(ctx, propagationTimeout, func() *resource.RetryError {
		var err error

		resp, err = conn.DescribeVirtualNodeWithContext(ctx, req)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		resp, err = conn.DescribeVirtualNodeWithContext(ctx, req)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
		log.Printf("[WARN] App Mesh Virtual Node (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Virtual Node: %s", err)
	}

	if resp == nil || resp.VirtualNode == nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Virtual Node: empty response")
	}

	if aws.StringValue(resp.VirtualNode.Status.Status) == appmesh.VirtualNodeStatusCodeDeleted {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading App Mesh Virtual Node: %s after creation", aws.StringValue(resp.VirtualNode.Status.Status))
		}

		log.Printf("[WARN] App Mesh Virtual Node (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	arn := aws.StringValue(resp.VirtualNode.Metadata.Arn)
	d.Set("name", resp.VirtualNode.VirtualNodeName)
	d.Set("mesh_name", resp.VirtualNode.MeshName)
	d.Set("mesh_owner", resp.VirtualNode.Metadata.MeshOwner)
	d.Set("arn", arn)
	d.Set("created_date", resp.VirtualNode.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", resp.VirtualNode.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("resource_owner", resp.VirtualNode.Metadata.ResourceOwner)
	err = d.Set("spec", flattenVirtualNodeSpec(resp.VirtualNode.Spec))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting spec: %s", err)
	}

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for App Mesh virtual node (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceVirtualNodeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()

	if d.HasChange("spec") {
		_, v := d.GetChange("spec")
		req := &appmesh.UpdateVirtualNodeInput{
			MeshName:        aws.String(d.Get("mesh_name").(string)),
			VirtualNodeName: aws.String(d.Get("name").(string)),
			Spec:            expandVirtualNodeSpec(v.([]interface{})),
		}
		if v, ok := d.GetOk("mesh_owner"); ok {
			req.MeshOwner = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Updating App Mesh virtual node: %s", req)
		_, err := conn.UpdateVirtualNodeWithContext(ctx, req)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating App Mesh virtual node (%s): %s", d.Id(), err)
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, arn, o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating App Mesh virtual node (%s) tags: %s", arn, err)
		}
	}

	return append(diags, resourceVirtualNodeRead(ctx, d, meta)...)
}

func resourceVirtualNodeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()

	log.Printf("[DEBUG] Deleting App Mesh Virtual Node: %s", d.Id())
	_, err := conn.DeleteVirtualNodeWithContext(ctx, &appmesh.DeleteVirtualNodeInput{
		MeshName:        aws.String(d.Get("mesh_name").(string)),
		VirtualNodeName: aws.String(d.Get("name").(string)),
	})

	if tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting App Mesh virtual node (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceVirtualNodeImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'mesh-name/virtual-node-name'", d.Id())
	}

	mesh := parts[0]
	name := parts[1]
	log.Printf("[DEBUG] Importing App Mesh virtual node %s from mesh %s", name, mesh)

	conn := meta.(*conns.AWSClient).AppMeshConn()

	resp, err := conn.DescribeVirtualNodeWithContext(ctx, &appmesh.DescribeVirtualNodeInput{
		MeshName:        aws.String(mesh),
		VirtualNodeName: aws.String(name),
	})
	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(resp.VirtualNode.Metadata.Uid))
	d.Set("name", resp.VirtualNode.VirtualNodeName)
	d.Set("mesh_name", resp.VirtualNode.MeshName)

	return []*schema.ResourceData{d}, nil
}
