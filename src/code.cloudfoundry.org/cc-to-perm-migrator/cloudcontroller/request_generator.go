package cloudcontroller

import (
	"net/http"

	"code.cloudfoundry.org/cc-to-perm-migrator/httpx"
	"code.cloudfoundry.org/cc-to-perm-migrator/messages"
	"code.cloudfoundry.org/lager"
)

type RequestGenerator struct {
	Host string
}

func NewRequestGenerator(host string) *RequestGenerator {
	return &RequestGenerator{
		Host: host,
	}
}

func (rg *RequestGenerator) NewGetRequest(logger lager.Logger, route string) (*http.Request, error) {
	u, err := httpx.JoinURL(logger.Session("join-url"), rg.Host, route)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		logger.Error(messages.FailedToCreateRequest, err)
		return nil, err

	}

	req.Header.Add("Accept", "application/json")

	return req, nil
}
