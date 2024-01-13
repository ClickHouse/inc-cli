package client

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
)

func New(ctx context.Context, apiKey, apiEndpoint, version string, opts ...ClientOption) (*ClientWithResponses, error) {
	bearerTokenProvider, bearerTokenProviderErr := securityprovider.NewSecurityProviderBearerToken(apiKey)
	if bearerTokenProviderErr != nil {
		return nil, bearerTokenProviderErr
	}

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3

	base := retryClient.StandardClient()

	// The generated client won't turn validation errors into actual errors, so we do this
	// inside of a generic middleware.
	base.Transport = Wrap(cleanhttp.DefaultTransport(), func(req *http.Request, next http.RoundTripper) (*http.Response, error) {
		resp, err := next.RoundTrip(req)
		if err == nil && resp.StatusCode > 299 {
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("status %d: no response body", resp.StatusCode)
			}

			return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(data))
		}

		return resp, err
	})

	clientOpts := append([]ClientOption{
		WithHTTPClient(base),
		WithRequestEditorFn(bearerTokenProvider.Intercept),
		// Add a user-agent so we can tell which version these requests came from.
		WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Add("user-agent", fmt.Sprintf("tap-incident/%s", version))
			return nil
		}),
	}, opts...)

	client, err := NewClientWithResponses(apiEndpoint, clientOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "creating client")
	}

	return client, nil
}

// WithReadOnly restricts the client to GET requests only, useful when creating a client
// for the purpose of dry-running.
func WithReadOnly() ClientOption {
	return WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		if req.Method != http.MethodGet {
			return fmt.Errorf("read-only client tried to make mutating request: %s %s", req.Method, req.URL.String())
		}

		return nil
	})
}

// RoundTripperFunc wraps a function to implement the RoundTripper interface, allowing
// easy wrapping of existing round-trippers.
type RoundTripperFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// Wrap allows easy wrapping of an existing RoundTripper with a function that can
// optionally call the original, or do its own thing.
func Wrap(next http.RoundTripper, apply func(req *http.Request, next http.RoundTripper) (*http.Response, error)) http.RoundTripper {
	return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return apply(req, next)
	})
}
