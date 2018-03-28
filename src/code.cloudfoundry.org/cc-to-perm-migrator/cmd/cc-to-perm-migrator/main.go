package main

import (
	"crypto/tls"
	"log"
	"net"
	"os"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"context"

	"net/http"

	"crypto/x509"

	"errors"

	"bytes"

	"code.cloudfoundry.org/cc-to-perm-migrator/capi"
	"code.cloudfoundry.org/cc-to-perm-migrator/cmd"
	"code.cloudfoundry.org/cc-to-perm-migrator/httpx"
	"code.cloudfoundry.org/cc-to-perm-migrator/migrator"
	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/populator"
	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/reporter"
	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/retriever"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm-go"
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

	logger, _ := config.Logger.Logger("cc-to-perm-migrator")

	progressLogger := log.New(os.Stderr, "", log.LstdFlags|log.LUTC)

	uaaCACert, err := config.UAA.CACertPath.Bytes(cmd.OS, cmd.IOReader)
	if err != nil {
		logger.Error("failed-to-read-uaa-ca-cert", err)
		os.Exit(1)
	}

	uaaCACertPool := x509.NewCertPool()
	ok := uaaCACertPool.AppendCertsFromPEM(uaaCACert)
	if !ok {
		logger.Error("failed-to-append-certs-from-pem", errors.New("could not append certs"), lager.Data{
			"path": config.UAA.CACertPath,
		})
		os.Exit(1)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: uaaCACertPool,
		},
	}
	sslcli := &http.Client{Transport: tr}

	tokenURL, err := httpx.JoinURL(logger, config.UAA.URL, "/oauth/token")
	if err != nil {
		os.Exit(1)
	}

	uaaConfig := &clientcredentials.Config{
		ClientID:     config.CloudController.ClientID,
		ClientSecret: config.CloudController.ClientSecret,
		TokenURL:     tokenURL.String(),
		Scopes:       config.CloudController.ClientScopes,
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, sslcli)

	oauth2.RegisterBrokenAuthHeaderProvider(tokenURL.String())
	client := uaaConfig.Client(ctx)

	ccClient := capi.NewClient(config.CloudController.URL, client)

	permCACert, err := config.Perm.CACert.Bytes(cmd.OS, cmd.IOReader)
	if err != nil {
		logger.Error("failed-to-read-perm-ca-cert", err)
		os.Exit(1)
	}

	var dialOptions []grpc.DialOption

	if len(permCACert) != 0 {
		permCACertPool := x509.NewCertPool()
		if ok := permCACertPool.AppendCertsFromPEM(permCACert); !ok {
			logger.Error("failed-to-append-certs-from-pem", errors.New("could not append certs"))
			os.Exit(1)
		}

		creds := credentials.NewClientTLSFromCert(permCACertPool, config.Perm.Hostname)
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(creds))
	} else {
		dialOptions = append(dialOptions, grpc.WithInsecure())
	}

	permAddr := net.JoinHostPort(config.Perm.Hostname, strconv.Itoa(config.Perm.Port))
	permConn, err := grpc.Dial(permAddr, dialOptions...)
	if err != nil {
		logger.Error("failed-to-connect-to-perm", err)
		os.Exit(1)
	}
	defer permConn.Close()

	roleServiceClient := protos.NewRoleServiceClient(permConn)

	pop := populator.NewPopulator(roleServiceClient)

	migrator.NewMigrator(retriever.NewRetriever(ccClient), pop, &reporter.Reporter{}, uaaConfig.TokenURL).
		Migrate(logger, progressLogger, os.Stderr, config.DryRun)
}
