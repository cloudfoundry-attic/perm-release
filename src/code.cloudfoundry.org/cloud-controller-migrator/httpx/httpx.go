package httpx

import (
	"net/url"

	"code.cloudfoundry.org/lager"
)

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
