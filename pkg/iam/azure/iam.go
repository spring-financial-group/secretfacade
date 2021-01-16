package azure

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/pkg/errors"
)

type Credentials interface {
	ClientID() string
	ClientSecret() string
	TenantID() string
	SubscriptionID() string
	UseManagedIdentity() bool
}

func Environment() (*azure.Environment, error) {
	settings, err := auth.GetSettingsFromEnvironment()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to determine Azure environment from environment")
	}
	return &settings.Environment, nil
}

// OAuthGrantType specifies which grant type to use.
type OAuthGrantType int

const (
	// OAuthGrantTypeServicePrincipal for client credentials flow
	OAuthGrantTypeServicePrincipal OAuthGrantType = iota
	// OAuthGrantTypeManagedIdentity for aad-pod-identity
	OAuthGrantTypeManagedIdentity
)

// GrantType returns what grant type has been configured.
func grantType(creds Credentials) OAuthGrantType {
	if creds.UseManagedIdentity() {
		return OAuthGrantTypeManagedIdentity
	}
	return OAuthGrantTypeServicePrincipal
}

var keyvaultAuthorizer autorest.Authorizer

// GetKeyvaultAuthorizer gets an OAuthTokenAuthorizer for use with Key Vault
// keys and secrets. Note that Key Vault *Vaults* are managed by Azure Resource
// Manager.
func GetKeyvaultAuthorizer(creds Credentials) (autorest.Authorizer, error) {
	if keyvaultAuthorizer != nil {
		return keyvaultAuthorizer, nil
	}

	environment, err := Environment()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to determine environment")
	}
	// BUG: default value for KeyVaultEndpoint is wrong
	vaultEndpoint := strings.TrimSuffix(environment.KeyVaultEndpoint, "/")
	// BUG: alternateEndpoint replaces other endpoints in the configs below
	alternateEndpoint, _ := url.Parse(
		"https://login.windows.net/" + creds.TenantID() + "/oauth2/token")

	var a autorest.Authorizer

	switch grantType(creds) {
	case OAuthGrantTypeServicePrincipal:
		oauthconfig, err := adal.NewOAuthConfig(
			environment.ActiveDirectoryEndpoint, creds.TenantID())
		if err != nil {
			return a, err
		}
		oauthconfig.AuthorizeEndpoint = *alternateEndpoint

		token, err := adal.NewServicePrincipalToken(
			*oauthconfig, creds.ClientID(), creds.ClientSecret(), vaultEndpoint)
		if err != nil {
			return a, err
		}

		a = autorest.NewBearerAuthorizer(token)

	case OAuthGrantTypeManagedIdentity:
		MIEndpoint, err := adal.GetMSIVMEndpoint()
		if err != nil {
			return nil, err
		}

		token, err := adal.NewServicePrincipalTokenFromMSI(MIEndpoint, vaultEndpoint)
		if err != nil {
			return nil, err
		}

		a = autorest.NewBearerAuthorizer(token)

	default:
		return a, fmt.Errorf("invalid grant type specified")
	}

	keyvaultAuthorizer = a

	return keyvaultAuthorizer, err
}
