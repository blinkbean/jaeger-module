package jaegerhttp

import (
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"io"
	"net/http"
)

func WrapClient(c *http.Client) *http.Client {
	if c == nil {
		c = http.DefaultClient
	}
	copied := *c
	copied.Transport = wrapRoundTripper(copied.Transport)
	return &copied
}

func wrapRoundTripper(r http.RoundTripper) http.RoundTripper {
	if r == nil {
		r = http.DefaultTransport
	}
	rt := &roundTripper{r: r}
	return rt
}

type roundTripper struct {
	r http.RoundTripper
}

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	sp, ctx := opentracing.StartSpanFromContext(ctx, req.Method)
	ext.HTTPMethod.Set(sp, req.Method)
	ext.HTTPUrl.Set(sp, req.URL.String())
	//ext.PeerAddress.Set(sp, req.URL.Host)

	_ = sp.Tracer().Inject(sp.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))

	resp, err := r.r.RoundTrip(req) // real request
	if err != nil {
		sp.Finish()
		return resp, err
	}
	ext.HTTPStatusCode.Set(sp, uint16(resp.StatusCode))
	if resp.StatusCode >= http.StatusInternalServerError {
		ext.Error.Set(sp, true)
	}
	if req.Method == "HEAD" {
		sp.Finish()
	} else {
		readWriteCloser, ok := resp.Body.(io.ReadWriteCloser)
		if ok {
			resp.Body = writeCloseTracker{readWriteCloser, sp}
		} else {
			resp.Body = closeTracker{resp.Body, sp}
		}
	}
	return resp, err
}

type closeTracker struct {
	io.ReadCloser
	sp opentracing.Span
}

func (c closeTracker) Close() error {
	err := c.ReadCloser.Close()
	//c.sp.LogFields(log.String("event","CloseBody"))
	c.sp.Finish()
	return err
}

type writeCloseTracker struct {
	io.ReadWriteCloser
	sp opentracing.Span
}

func (c writeCloseTracker) Close() error {
	err := c.ReadWriteCloser.Close()
	//c.sp.LogFields(log.String("event", "CloseBody"))
	c.sp.Finish()
	return err
}
