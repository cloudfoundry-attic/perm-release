package main

import (
	"crypto/tls"
	"os"

	"context"

	"net/http"

	"net/url"

	"fmt"
	"net/http/httputil"

	"code.cloudfoundry.org/cloud-controller-migrator/cmd"
	flags "github.com/jessevdk/go-flags"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type options struct {
	ConfigFilePath string `long:"config-file-path" description:"Path to the config file for the CloudController migrator" required:"true"`
}

func main() {
	parserOpts := &options{}
	parser := flags.NewParser(parserOpts, flags.Default)
	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}

	f, err := os.Open(parserOpts.ConfigFilePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	config, err := cmd.NewConfig(f)
	if err != nil {
		os.Exit(1)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	sslcli := &http.Client{Transport: tr}

	tokenURL, err := JoinURL(config.UAA.URL, "/oauth/token")
	if err != nil {
		panic(err)
	}

	uaaConfig := &clientcredentials.Config{
		ClientID:     config.CloudController.ClientID,
		ClientSecret: config.CloudController.ClientSecret,
		TokenURL:     tokenURL.String(),
		Scopes:       []string{"cloud_controller.admin_read_only"},
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, sslcli)

	client := uaaConfig.Client(ctx)

	infoURL, err := JoinURL(config.CloudController.URL, "/v2/info")
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

func JoinURL(base string, path string) (*url.URL, error) {
	b, err := url.Parse(base)
	if err != nil {
		return nil, err
	}

	p, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	return b.ResolveReference(p), nil
}
