package cloudcontroller

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"net/http"

	"fmt"

	"time"

	"code.cloudfoundry.org/cc-to-perm-migrator/messages"
	"code.cloudfoundry.org/lager"
)

type APIClient struct {
	Host           string
	HTTPClient     *http.Client
	RequestTimeout time.Duration
}

func NewAPIClient(host string, client *http.Client, timeout time.Duration) *APIClient {
	return &APIClient{
		Host:           host,
		HTTPClient:     client,
		RequestTimeout: timeout,
	}
}

func (c *APIClient) MakePaginatedGetRequest(logger lager.Logger, route string, bodyCallback func(lager.Logger, io.Reader) error) error {
	rg := NewRequestGenerator(c.Host)

	var (
		paginatedResponse PaginatedResponse

		routeLogger lager.Logger
	)

	for {
		nextURL, err := func() (*string, error) {
			routeLogger = logger.WithData(lager.Data{
				"route": route,
			})

			ctx, cancelFunc := context.WithTimeout(context.Background(), c.RequestTimeout)
			defer cancelFunc()

			res, err := makeAPIRequest(ctx, routeLogger.Session("make-api-request"), c.HTTPClient, rg, route)
			if err != nil {
				return nil, err
			}
			defer res.Body.Close()

			var body []byte
			buf := bytes.NewBuffer(body)
			r := io.TeeReader(res.Body, buf)

			err = json.NewDecoder(r).Decode(&paginatedResponse)
			if err != nil {
				routeLogger.Error("failed-to-decode-response", err)
				return nil, err
			}

			err = bodyCallback(routeLogger, buf)
			if err != nil {
				return nil, err
			}

			if paginatedResponse.NextURL == nil {
				return nil, nil
			}

			return paginatedResponse.NextURL, nil
		}()

		if err != nil {
			return err
		}

		if nextURL == nil {
			break
		}

		route = *nextURL
	}

	return nil
}

func makeAPIRequest(ctx context.Context, logger lager.Logger, client *http.Client, rg *RequestGenerator, route string) (*http.Response, error) {
	req, err := rg.NewGetRequest(logger.Session("new-get-request"), route)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	logger.Debug("making-request")
	res, err := client.Do(req)
	if err != nil {
		logger.Error(messages.FailedToPerformRequest, err)
		return nil, err
	}

	if res.StatusCode >= 400 {
		err = fmt.Errorf("HTTP bad response: %d", res.StatusCode)
		logger.Error("failed-to-ping-cloudcontroller", err)
		return nil, err
	}

	return res, nil
}
