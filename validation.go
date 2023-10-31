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

type responseValidatorLastModified struct {
	rt   http.RoundTripper
	next ResponseValidator
}

func newIfLastModifiedResponseValidator(rt http.RoundTripper, next ResponseValidator) ResponseValidator {
	return &responseValidatorLastModified{rt: rt, next: next}
}

func (v *responseValidatorLastModified) Validate(cachedResponse *http.Response, r *http.Request) (*http.Response, error) {
	// if there is no last modified header, then we can't validate
	lastModified, err := lastModifiedFromHeader(cachedResponse.Header)
	if err != nil {
		return v.next.Validate(cachedResponse, r)
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

type responseValidatorEtag struct {
	rt   http.RoundTripper
	next ResponseValidator
}

func newEtagResponseValidator(rt http.RoundTripper, next ResponseValidator) ResponseValidator {
	return &responseValidatorEtag{rt: rt, next: next}
}

func (v *responseValidatorEtag) Validate(cachedResponse *http.Response, r *http.Request) (*http.Response, error) {
	etag, err := etagFromHeader(cachedResponse.Header)
	if err != nil {
		return v.next.Validate(cachedResponse, r)
	}

	r.Header = withIfNoneMatchHeader(r.Header, etag)
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

func newResponseValidator(transport http.RoundTripper) ResponseValidator {
	return newEtagResponseValidator(
		transport,
		newIfLastModifiedResponseValidator(
			transport,
			newRevalidator(transport),
		),
	)
}

type revalidator struct {
	transport http.RoundTripper
}

func newRevalidator(transport http.RoundTripper) ResponseValidator {
	return &revalidator{transport: transport}
}

func (v *revalidator) Validate(cachedResponse *http.Response, r *http.Request) (*http.Response, error) {
	return v.transport.RoundTrip(r)
}
