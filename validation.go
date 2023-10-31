package webcache

import (
	"io"
	"net/http"
)

// https://developer.mozilla.org/en-US/docs/Web/HTTP/Caching#validation
type ResponseValidator interface {
	// Validate
	// validate that the cached response is still fresh against the incoming request
	Validate(cachedResponse *http.Response, r *http.Request) (*http.Response, error)
}

type responseValidator struct {
	rt http.RoundTripper
}

func newIfLastModifiedResponseValidator(rt http.RoundTripper) ResponseValidator {
	return &responseValidator{rt: rt}
}

func (v *responseValidator) Validate(cachedResponse *http.Response, r *http.Request) (*http.Response, error) {
	// if there is no last modified header, then we can't validate
	lastModified, err := lastModifiedFromHeader(cachedResponse.Header)
	if err != nil {
		resp, err := v.rt.RoundTrip(r)
		return resp, err
	}

	r.Header = withIFModifiedSinceHeader(r.Header, lastModified)

	response, err := v.rt.RoundTrip(r)
	if err != nil {
		return response, err
	}

	if response.StatusCode == http.StatusNotModified {
		_, _ = io.Copy(io.Discard, response.Body)
		response.Body.Close()
		cachedResponse.Header = withCacheHitHeader(cachedResponse.Header)
		return cachedResponse, nil
	}

	return response, nil
}
