package opencensus

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"bytes"
	"strings"
	"net/http/httputil"

	"github.com/luraproject/lura/v2/config"
	transport "github.com/luraproject/lura/v2/transport/http/client"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
)

var defaultClient = &http.Client{Transport: &ochttp.Transport{}}

func NewHTTPClient(ctx context.Context) *http.Client {
	if !IsBackendEnabled() {
		return transport.NewHTTPClient(ctx)
	}
	return defaultClient
}

func HTTPRequestExecutor(clientFactory transport.HTTPClientFactory) transport.HTTPRequestExecutor {
	return HTTPRequestExecutorFromConfig(clientFactory, nil)
}

type ComputationRequest struct {
	ProductId string
}

func HTTPRequestExecutorFromConfig(clientFactory transport.HTTPClientFactory, cfg *config.Backend) transport.HTTPRequestExecutor {
	if !IsBackendEnabled() {
		return transport.DefaultHTTPRequestExecutor(clientFactory)
	}

	pathExtractor := GetAggregatedPathForBackendMetrics(cfg)

	return func(ctx context.Context, req *http.Request) (*http.Response, error) {
		httpClient := clientFactory(ctx)

		if _, ok := httpClient.Transport.(*Transport); ok {
			return httpClient.Do(req.WithContext(trace.NewContext(ctx, fromContext(ctx))))
		}

		c := &http.Client{
			Transport: &Transport{
				Base: httpClient.Transport,
				tags: []tagGenerator{
					func(r *http.Request) tag.Mutator { return tag.Upsert(ochttp.KeyClientHost, req.Host) },
					func(r *http.Request) tag.Mutator {
						return tag.Upsert(ochttp.KeyClientPath, pathExtractor(r))
					},
					func(r *http.Request) tag.Mutator {
						requestDump, err := httputil.DumpRequest(r, true)
						if err != nil {
						fmt.Println(err)
						}
						fmt.Println(string(requestDump))
						tenant := strings.Trim(strings.Split(strings.SplitAfter(r.URL.Path, "tenants/")[1], "/")[0], "")
						fmt.Println(tenant)
						return tag.Upsert(tag.MustNewKey("http_client_tenant"), tenant)
					},
					func(r *http.Request) tag.Mutator {
						b, err := ioutil.ReadAll(r.Body)
						r.Body = ioutil.NopCloser(bytes.NewBuffer(b))

						// defer r.Body.Close()
						if err != nil {
							fmt.Println(err.Error())
						}

						var cr ComputationRequest
						err = json.Unmarshal(b, &cr)
						if err != nil {
							fmt.Println(err.Error())
						}
						fmt.Println(cr.ProductId)
						
						return tag.Upsert(tag.MustNewKey("http_client_product"), string(cr.ProductId))
					},
					func(r *http.Request) tag.Mutator { return tag.Upsert(ochttp.KeyClientMethod, req.Method) },
				},
			},
			CheckRedirect: httpClient.CheckRedirect,
			Jar:           httpClient.Jar,
			Timeout:       httpClient.Timeout,
		}
		return c.Do(req.WithContext(trace.NewContext(ctx, fromContext(ctx))))
	}
}
