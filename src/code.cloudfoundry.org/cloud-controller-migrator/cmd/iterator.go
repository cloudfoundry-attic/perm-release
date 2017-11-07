package cmd

import (
	"context"
	"io"
	"net/http"

	"fmt"
	"net/http/httputil"

	"code.cloudfoundry.org/cloud-controller-migrator/httpx"
	"code.cloudfoundry.org/lager"
)

func IterateOverCloudControllerEntities(ctx context.Context, logger lager.Logger, w io.Writer, client *http.Client, url string) error {
	infoURL, err := httpx.JoinURL(logger, url, "/v2/info")
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", infoURL.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")

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
