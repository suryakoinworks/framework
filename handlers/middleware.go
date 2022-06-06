package handlers

import (
	"net/http"
	"sort"

	configs "github.com/KejawenLab/bima/v2/configs"
	events "github.com/KejawenLab/bima/v2/events"

	"github.com/CAFxX/httpcompression"
	"github.com/CAFxX/httpcompression/contrib/andybalholm/brotli"
	"github.com/CAFxX/httpcompression/contrib/compress/zlib"
	"github.com/CAFxX/httpcompression/contrib/klauspost/pgzip"
)

type Middleware struct {
	Dispatcher     *events.Dispatcher
	Middlewares    []configs.Middleware
	MuxMiddlewares []configs.MuxMiddlewares
	Logger         *Logger
}

func (m *Middleware) Register(middlewares []configs.Middleware) {
	for _, v := range middlewares {
		m.Add(v)
	}
}

func (m *Middleware) Add(middlware configs.Middleware) {
	m.Middlewares = append(m.Middlewares, middlware)
}

func (m *Middleware) Sort() {
	sort.Slice(m.Middlewares, func(i, j int) bool {
		return m.Middlewares[i].Priority() > m.Middlewares[j].Priority()
	})
}

func (m *Middleware) Attach(handler http.Handler) http.Handler {
	internal := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		for _, middleware := range m.Middlewares {
			stop := middleware.Attach(request, response)
			if stop {
				return
			}
		}

		m.Dispatcher.Dispatch(events.REQUEST_EVENT, &events.Request{
			HttpRequest: request,
		})

		m.Dispatcher.Dispatch(events.RESPONSE_EVENT, &events.Response{
			ResponseWriter: response,
		})

		handler.ServeHTTP(response, request)
	})

	deflateEncoder, err := zlib.New(zlib.Options{})
	if err != nil {
		m.Logger.Fatal(err.Error())
	}

	brotliEncoder, err := brotli.New(brotli.Options{})
	if err != nil {
		m.Logger.Fatal(err.Error())
	}

	gzipEncoder, err := pgzip.New(pgzip.Options{
		Level:     pgzip.DefaultCompression,
		BlockSize: 1 << 20,
		Blocks:    4,
	})
	if err != nil {
		m.Logger.Fatal(err.Error())
	}

	compress, err := httpcompression.Adapter(
		httpcompression.Compressor(brotli.Encoding, 2, brotliEncoder),
		httpcompression.Compressor(pgzip.Encoding, 1, gzipEncoder),
		httpcompression.Compressor(zlib.Encoding, 0, deflateEncoder),
		httpcompression.Prefer(httpcompression.PreferServer),
		httpcompression.MinSize(165),
	)

	if err != nil {
		m.Logger.Fatal(err.Error())
	}

	last := compress(internal)
	for _, middleware := range m.MuxMiddlewares {
		last = middleware(last)
	}

	return last
}
