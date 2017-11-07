package main

import (
	"crypto/tls"
	"os"

	"context"

	"net/http"

	"net/url"

	"fmt"
	"net/http/httputil"

	"crypto/x509"

	"errors"

	"bytes"

	"code.cloudfoundry.org/cloud-controller-migrator/cmd"
	"code.cloudfoundry.org/lager"
	flags "github.com/jessevdk/go-flags"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type options struct {
	ConfigFilePath cmd.FileOrStringFlag `long:"config-file-path" description:"Path to the config file for the CloudController migrator" required:"true"`
}

func main() {
	parserOpts := &options{}
	parser := flags.NewParser(parserOpts, flags.Default)
	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		// Necessary to not panic because this is how the parser prints Help messages
		os.Exit(1)
	}

	configFileContents, err := parserOpts.ConfigFilePath.Bytes(cmd.OS, cmd.IOReader)
	if err != nil {
		panic(err)
	}

	config, err := cmd.NewConfig(bytes.NewReader(configFileContents))
	if err != nil {
		panic(err)
	}

	logger, _ := config.Logger.Logger("cloud-controller-migrator")

	uaaCACert, err := config.UAA.CACertPath.Bytes(cmd.OS, cmd.IOReader)
	if err != nil {
		logger.Error("failed-to-read-uaa-ca-cert", err)
		panic(err)
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(uaaCACert)
	if !ok {
		logger.Error("failed-to-append-certs-from-pem", errors.New("could not append certs"), lager.Data{
			"path": config.UAA.CACertPath,
		})
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}
	sslcli := &http.Client{Transport: tr}

	tokenURL, err := JoinURL(logger, config.UAA.URL, "/oauth/token")
	if err != nil {
		panic(err)
	}

	uaaConfig := &clientcredentials.Config{
		ClientID:     config.CloudController.ClientID,
		ClientSecret: config.CloudController.ClientSecret,
		TokenURL:     tokenURL.String(),
		Scopes:       config.CloudController.ClientScopes,
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, sslcli)

	client := uaaConfig.Client(ctx)

	infoURL, err := JoinURL(logger, config.CloudController.URL, "/v2/info")
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("GET", infoURL.String(), nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	b, err := httputil.DumpResponse(res, true)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}

func JoinURL(logger lager.Logger, base string, path string) (*url.URL, error) {
	logger = logger.Session("join-url").WithData(lager.Data{
		"base": base,
		"path": path,
	})

	b, err := url.Parse(base)
	if err != nil {
		logger.Error("failed-to-parse-base", err)
		return nil, err
	}

	p, err := url.Parse(path)
	if err != nil {
		logger.Error("failed-to-parse-path", err)
		return nil, err
	}

	return b.ResolveReference(p), nil
}
