package provider

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"

	"github.com/camptocamp/go-freeipa/freeipa"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	version = "dev"
)

type Provider struct {
	dataSources []func() datasource.DataSource
	resources   []func() resource.Resource

	client *freeipa.Client
}

type Model struct {
	Host               types.String `tfsdk:"host"`
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure"`
	KerberosEnabled    types.Bool   `tfsdk:"kerberos_enabled"`
	KerberosPrincipal  types.String `tfsdk:"kerberos_principal"`
	KerberosRealm      types.String `tfsdk:"kerberos_realm"`
	Krb5ConfPath       types.String `tfsdk:"krb5_conf_path"`
	KeytabPath         types.String `tfsdk:"keytab_path"`
}

func (p *Provider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "freeipa"
	resp.Version = version
}

func (p *Provider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:    true,
				Description: "FreeIPA host to connect to",
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "Username to use for connection",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Password to use for connection",
			},
			"insecure": schema.BoolAttribute{
				Optional:    true,
				Description: "Set to true to disable FreeIPA host TLS certificate verification",
			},
			"kerberos_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Use Kerberos/keytab authentication instead of username/password",
			},
			"kerberos_principal": schema.StringAttribute{
				Optional:    true,
				Description: "Kerberos principal to use when kerberos_enabled is true",
			},
			"kerberos_realm": schema.StringAttribute{
				Optional:    true,
				Description: "Kerberos realm to use when kerberos_enabled is true",
			},
			"krb5_conf_path": schema.StringAttribute{
				Optional:    true,
				Description: "Path to krb5.conf to use for Kerberos authentication",
			},
			"keytab_path": schema.StringAttribute{
				Optional:    true,
				Description: "Path to keytab file to use for Kerberos authentication",
			},
		},
	}
}

func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config Model

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("FREEIPA_HOST")
	username := os.Getenv("FREEIPA_USERNAME")
	password := os.Getenv("FREEIPA_PASSWORD")
	insecureSkipVerify := false

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	if !config.InsecureSkipVerify.IsNull() {
		insecureSkipVerify = config.InsecureSkipVerify.ValueBool()
	}

	kerberosEnabled := false
	if !config.KerberosEnabled.IsNull() {
		kerberosEnabled = config.KerberosEnabled.ValueBool()
	} else if os.Getenv("FREEIPA_KERBEROS_ENABLED") == "true" {
		kerberosEnabled = true
	}

	kerberosPrincipal := os.Getenv("FREEIPA_KERBEROS_PRINCIPAL")
	if !config.KerberosPrincipal.IsNull() {
		kerberosPrincipal = config.KerberosPrincipal.ValueString()
	}

	kerberosRealm := os.Getenv("FREEIPA_KERBEROS_REALM")
	if !config.KerberosRealm.IsNull() {
		kerberosRealm = config.KerberosRealm.ValueString()
	}

	krb5ConfPath := os.Getenv("FREEIPA_KRB5_CONF")
	if krb5ConfPath == "" {
		krb5ConfPath = "/etc/krb5.conf"
	}
	if !config.Krb5ConfPath.IsNull() {
		krb5ConfPath = config.Krb5ConfPath.ValueString()
	}

	keytabPath := os.Getenv("FREEIPA_KEYTAB")
	if keytabPath == "" {
		keytabPath = "/etc/krb5.keytab"
	}
	if !config.KeytabPath.IsNull() {
		keytabPath = config.KeytabPath.ValueString()
	}

	if host == "" {
		resp.Diagnostics.AddAttributeError(path.Root("host"), "Missing FreeIPA host",
			`Host is required to establish a connection to FreeIPA.`,
		)
	}

	if kerberosEnabled {
		if kerberosPrincipal == "" {
			resp.Diagnostics.AddAttributeError(path.Root("kerberos_principal"), "Missing Kerberos principal",
				`Kerberos principal is required when kerberos_enabled is true.`,
			)
		}
		if kerberosRealm == "" {
			resp.Diagnostics.AddAttributeError(path.Root("kerberos_realm"), "Missing Kerberos realm",
				`Kerberos realm is required when kerberos_enabled is true.`,
			)
		}
		if keytabPath == "" {
			resp.Diagnostics.AddAttributeError(path.Root("keytab_path"), "Missing keytab path",
				`Path to keytab file is required when kerberos_enabled is true.`,
			)
		}
	} else {
		if username == "" {
			resp.Diagnostics.AddAttributeError(path.Root("username"), "Missing FreeIPA username",
				`Username is required to establish a connection to FreeIPA.`,
			)
		}

		if password == "" {
			resp.Diagnostics.AddAttributeError(path.Root("password"), "Missing FreeIPA password",
				`Password is required to establish a connection to FreeIPA.`,
			)
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tspt := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecureSkipVerify,
		},
	}

	var err error

	if kerberosEnabled {
		krb5ConfFile, err := os.Open(krb5ConfPath)
		if err != nil {
			resp.Diagnostics.AddError("Failed to open krb5.conf", "Reason: "+err.Error())
			return
		}
		defer krb5ConfFile.Close()

		keytabFile, err := os.Open(keytabPath)
		if err != nil {
			resp.Diagnostics.AddError("Failed to open keytab file", "Reason: "+err.Error())
			return
		}
		defer keytabFile.Close()

		kerberosOpts := &freeipa.KerberosConnectOptions{
			Krb5ConfigReader: krb5ConfFile,
			KeytabReader:     keytabFile,
			Username:         kerberosPrincipal,
			Realm:            kerberosRealm,
		}

		p.client, err = freeipa.ConnectWithKerberos(host, tspt, kerberosOpts)
		if err != nil {
			resp.Diagnostics.AddError("Failed to connect to FreeIPA", "Reason: "+err.Error())
			return
		}
	} else {
		p.client, err = freeipa.Connect(host, tspt, username, password)
		if err != nil {
			resp.Diagnostics.AddError("Failed to connect to FreeIPA", "Reason: "+err.Error())
			return
		}
	}

	tflog.Info(ctx, "Successfully connected to FreeIPA", map[string]any{
		"host":             host,
		"username":         username,
		"kerberos_enabled": kerberosEnabled,
	})
}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return p.dataSources
}

func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	return p.resources
}

func (p *Provider) Client() *freeipa.Client {
	return p.client
}

func NewFactory(ds []func(p *Provider) datasource.DataSource, rs []func(p *Provider) resource.Resource) func() provider.Provider {
	return func() provider.Provider {
		p := &Provider{}

		p.dataSources = make([]func() datasource.DataSource, len(ds))

		for i, d := range ds {
			d := d

			p.dataSources[i] = func() datasource.DataSource {
				return d(p)
			}
		}

		p.resources = make([]func() resource.Resource, len(rs))

		for i, r := range rs {
			r := r

			p.resources[i] = func() resource.Resource {
				return r(p)
			}
		}

		var _ provider.Provider = p

		return p
	}
}
