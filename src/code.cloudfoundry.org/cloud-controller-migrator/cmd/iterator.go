package cmd

import (
	"context"
	"io"
	"net/http"

	"fmt"
	"net/http/httputil"

	"code.cloudfoundry.org/cloud-controller-migrator/cloudcontroller"
	"code.cloudfoundry.org/cloud-controller-migrator/messages"
	"code.cloudfoundry.org/lager"
)

func IterateOverCloudControllerEntities(ctx context.Context, logger lager.Logger, w io.Writer, client *http.Client, url string) error {
	logger = logger.Session("iterate-over-cloud-controller-entities").WithData(lager.Data{
		"url": url,
	})

	rg := cloudcontroller.NewRequestGenerator(url)

	routerLogger := logger.WithData(lager.Data{
		"routes": rg.Routes,
	})
	route := cloudcontroller.Info
	req, err := rg.CreateRequest(route, nil, nil)
	if err != nil {
		routerLogger.Error(messages.FailedToCreateRequest, err, lager.Data{
			"route": route,
		})
		return err
	}

	res, err := client.Do(req)
	if err != nil {
		routerLogger.Error(messages.FailedToPerformRequest, err, lager.Data{
			"route": route,
		})

		return err
	}

	b, err := httputil.DumpResponse(res, true)
	if err != nil {
		return err
	}

	fmt.Fprintln(w, string(b))

	return nil
}
