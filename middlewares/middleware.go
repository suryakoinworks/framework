package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"time"

	"github.com/KejawenLab/bima/v2/loggers"
	"github.com/fatih/color"

	"github.com/CAFxX/httpcompression"
	"github.com/CAFxX/httpcompression/contrib/andybalholm/brotli"
	"github.com/CAFxX/httpcompression/contrib/compress/zlib"
	"github.com/CAFxX/httpcompression/contrib/klauspost/pgzip"
)

type (
	Middleware interface {
		Attach(request *http.Request, response http.ResponseWriter) bool
		Priority() int
	}

	Factory struct {
		Debug       bool
		Middlewares []Middleware
		Logger      *loggers.Logger
	}

	responseWrapper struct {
		http.ResponseWriter
		statusCode int
	}
)

func (w *responseWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWrapper) StatusCode() int {
	if w.statusCode == 0 {
		return http.StatusOK
	}

	return w.statusCode
}

func (m *Factory) Register(middlewares []Middleware) {
	for _, v := range middlewares {
		m.Add(v)
	}
}

func (m *Factory) Add(middlware Middleware) {
	m.Middlewares = append(m.Middlewares, middlware)
}

func (m *Factory) Sort() {
	sort.Slice(m.Middlewares, func(i, j int) bool {
		return m.Middlewares[i].Priority() > m.Middlewares[j].Priority()
	})
}

func (m *Factory) Attach(handler http.Handler) http.Handler {
	ctx := context.WithValue(context.Background(), "scope", "middleware")
	start := time.Now()
	internal := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if !m.Debug {
			for _, middleware := range m.Middlewares {
				if stop := middleware.Attach(request, response); stop {
					return
				}
			}

			handler.ServeHTTP(response, request)

			elapsed := time.Since(start)

			m.Logger.Info(ctx, fmt.Sprintf("Execution time: %s", elapsed))

			return
		}

		wrapper := responseWrapper{ResponseWriter: response}
		for _, middleware := range m.Middlewares {
			if stop := middleware.Attach(request, response); stop {
				m.Logger.Debug(ctx, fmt.Sprintf("Middleware stopped by: %s", reflect.TypeOf(middleware).Name()))

				return
			}
		}

		handler.ServeHTTP(&wrapper, request)

		elapsed := time.Since(start)

		var statusCode string
		uri, _ := url.QueryUnescape(request.RequestURI)
		mGet := color.New(color.BgHiGreen, color.FgBlack)
		mPost := color.New(color.BgYellow, color.FgBlack)
		mPut := color.New(color.BgCyan, color.FgBlack)
		mDelete := color.New(color.BgRed, color.FgBlack)

		switch request.Method {
		case http.MethodPost:
			mPost.Print("[POST]")
		case http.MethodPatch:
			mPost.Print("[PATCH]")
		case http.MethodPut:
			mPut.Print("[PUT]")
		case http.MethodDelete:
			mDelete.Print("[DELETE]")
		default:
			mGet.Print("[GET]")
		}

		switch {
		case wrapper.StatusCode() < 300:
			statusCode = color.New(color.FgGreen, color.Bold).Sprintf("%d", wrapper.StatusCode())
		case wrapper.StatusCode() < 400:
			statusCode = color.New(color.FgYellow, color.Bold).Sprintf("%d", wrapper.StatusCode())
		default:
			statusCode = color.New(color.FgRed, color.Bold).Sprintf("%d", wrapper.StatusCode())
		}

		fmt.Printf("\t%s\t%s\t%s\n", statusCode, elapsed, uri)
	})

	deflateEncoder, _ := zlib.New(zlib.Options{})
	brotliEncoder, _ := brotli.New(brotli.Options{})
	gzipEncoder, _ := pgzip.New(pgzip.Options{
		Level:     pgzip.DefaultCompression,
		BlockSize: 1 << 20,
		Blocks:    4,
	})

	compress, _ := httpcompression.Adapter(
		httpcompression.Compressor(brotli.Encoding, 2, brotliEncoder),
		httpcompression.Compressor(pgzip.Encoding, 1, gzipEncoder),
		httpcompression.Compressor(zlib.Encoding, 0, deflateEncoder),
		httpcompression.Prefer(httpcompression.PreferServer),
		httpcompression.MinSize(165),
	)

	return compress(internal)
}