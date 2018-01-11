package cmd

import (
	"context"
	"net/http"

	"net/url"

	"fmt"

	"encoding/json"

	"code.cloudfoundry.org/lager"
	"golang.org/x/net/context/ctxhttp"
)

var OidcConfigurationRoute = &url.URL{
	Path:    ".well-known/openid-configuration",
	RawPath: ".well-known/openid-configuration",
}

func RetrieveIssuer(ctx context.Context, logger lager.Logger, client *http.Client, oidcProviderURL *url.URL) (string, error) {
	fullURL := oidcProviderURL.ResolveReference(OidcConfigurationRoute)

	res, err := ctxhttp.Get(ctx, client, fullURL.String())
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("HTTP bad response: %d", res.StatusCode)
		logger.Error("failed-to-retrieve-issuer", err)
		return "", err
	}

	var configuration OIDCProviderConfiguration
	err = json.NewDecoder(res.Body).Decode(&configuration)
	if err != nil {
		logger.Error("failed-to-parse-oidc-configuration", err)
		return "", err
	}

	issuer := configuration.Issuer
	logger.Info("succeeded", lager.Data{
		"issuer": issuer,
	})
	return issuer, nil
}

type OIDCProviderConfiguration struct {
	Issuer string `json:"issuer"`
}
