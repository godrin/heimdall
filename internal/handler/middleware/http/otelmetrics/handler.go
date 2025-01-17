// Copyright 2023 Dimitrij Drus <dadrus@gmx.de>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package otelmetrics

import (
	"net/http"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"

	"github.com/dadrus/heimdall/internal/x"
	"github.com/dadrus/heimdall/internal/x/httpx"
)

const (
	instrumentationName = "github.com/dadrus/heimdall/internal/handler/middleware/http/otelmetrics"

	requestsActive = "http.server.active_requests"
)

func New(opts ...Option) func(http.Handler) http.Handler {
	conf := newConfig(opts...)

	meter := conf.provider.Meter(instrumentationName)

	activeRequests, err := meter.Float64UpDownCounter(
		requestsActive,
		metric.WithDescription("Measures the number of concurrent HTTP requests that are currently in-flight."),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		panic(err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if !conf.shouldProcess(req) {
				next.ServeHTTP(rw, req)

				return
			}

			labeler, _ := otelhttp.LabelerFromContext(req.Context())
			if conf.subsystem.Valid() {
				labeler.Add(conf.subsystem)
			}

			attributes := serverRequestMetrics(conf.server, req)
			attributes = append(labeler.Get(), attributes...)
			attributes = append(attributes, conf.attributes...)

			opt := metric.WithAttributes(attributes...)

			activeRequests.Add(req.Context(), 1, opt)
			defer func() {
				activeRequests.Add(req.Context(), -1, opt)
			}()

			next.ServeHTTP(rw, req)
		})
	}
}

func serverRequestMetrics(server string, req *http.Request) []attribute.KeyValue {
	attrsCount := 4 // Method, scheme, proto, and host name.

	var (
		host string
		port int
	)

	if server == "" {
		host, port = httpx.HostPort(req.Host)
	} else {
		// Prioritize the primary server name.
		host, port = httpx.HostPort(server)
		if port < 0 {
			_, port = httpx.HostPort(req.Host)
		}
	}

	hostPort := requiredHTTPPort(req.TLS != nil, port)
	if hostPort > 0 {
		attrsCount++
	}

	attrs := make([]attribute.KeyValue, 0, attrsCount)
	attrs = append(attrs, methodMetric(req.Method))
	attrs = append(attrs, x.IfThenElse(req.TLS != nil, semconv.HTTPSchemeHTTPS, semconv.HTTPSchemeHTTP))
	attrs = append(attrs, flavor(req.Proto))
	attrs = append(attrs, semconv.NetHostNameKey.String(host))

	if hostPort > 0 {
		attrs = append(attrs, semconv.NetHostPortKey.Int(hostPort))
	}

	return attrs
}

func methodMetric(method string) attribute.KeyValue {
	method = strings.ToUpper(method)
	switch method {
	case http.MethodConnect,
		http.MethodDelete,
		http.MethodGet,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodPost,
		http.MethodPut,
		http.MethodTrace:
	default:
		method = "_OTHER"
	}

	return semconv.HTTPMethodKey.String(method)
}

func flavor(proto string) attribute.KeyValue {
	switch proto {
	case "HTTP/1.0":
		return semconv.HTTPFlavorKey.String("1.0")
	case "HTTP/1.1":
		return semconv.HTTPFlavorKey.String("1.1")
	case "HTTP/2":
		return semconv.HTTPFlavorKey.String("2.0")
	case "HTTP/3":
		return semconv.HTTPFlavorKey.String("3.0")
	default:
		return semconv.HTTPFlavorKey.String(proto)
	}
}

func requiredHTTPPort(https bool, port int) int { // nolint:revive
	if https {
		if port > 0 && port != 443 {
			return port
		}
	} else {
		if port > 0 && port != 80 {
			return port
		}
	}

	return -1
}
