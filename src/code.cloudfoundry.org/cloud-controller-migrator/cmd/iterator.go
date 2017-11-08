package cmd

import (
	"context"
	"io"
	"net/http"

	"fmt"
	"net/http/httputil"

	"code.cloudfoundry.org/cloud-controller-migrator/cloudcontroller"
	"code.cloudfoundry.org/lager"
)

func IterateOverCloudControllerEntities(ctx context.Context, logger lager.Logger, w io.Writer, client *http.Client, url string) error {
	rg := cloudcontroller.NewRequestGenerator(url)

	req, err := rg.CreateRequest(cloudcontroller.Info, nil, nil)
	if err != nil {
		return err
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	b, err := httputil.DumpResponse(res, true)
	if err != nil {
		return err
	}

	fmt.Fprintln(w, string(b))

	return nil
}
