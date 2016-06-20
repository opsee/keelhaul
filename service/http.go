package service

import (
	"crypto/tls"
	"github.com/opsee/basic/grpcutil"
	"github.com/opsee/basic/schema"
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/basic/tp"
	"golang.org/x/net/context"
	"golang.org/x/net/http2"
	"net/http"
	"time"
)

const (
	serviceKey = iota
	userKey
	requestKey
	paramsKey
)

func (s *service) StartMux(addr, certfile, certkeyfile string) error {
	router := tp.NewHTTPRouter(context.Background())

	router.CORS(
		[]string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		[]string{`https?://localhost:8080`, `https?://localhost:8008`, `https://(\w+\.)?(opsy\.co|opsee\.co|opsee\.com)`},
	)

	// swagger
	router.Handle("GET", "/api/swagger.json", []tp.DecodeFunc{}, s.swagger())

	// json api
	router.Handle("GET", "/vpcs/bastions", decoders(schema.User{}, ListBastionsRequest{}), s.listBastions())
	router.Handle("POST", "/bastions/authenticate", []tp.DecodeFunc{tp.RequestDecodeFunc(requestKey, opsee.AuthenticateBastionRequest{})}, s.authenticateBastion())

	// websocket
	router.HandlerFunc("GET", "/stream/", s.websocketHandlerFunc)

	// set a big timeout bc aws be slow
	router.Timeout(5 * time.Minute)

	httpServer := &http.Server{
		Addr:      addr,
		Handler:   grpcutil.GRPCHandlerFunc(s.grpcServer, router),
		TLSConfig: &tls.Config{},
	}

	if err := http2.ConfigureServer(httpServer, nil); err != nil {
		return err
	}

	return httpServer.ListenAndServeTLS(certfile, certkeyfile)
}

func decoders(userType interface{}, requestType interface{}) []tp.DecodeFunc {
	return []tp.DecodeFunc{
		tp.AuthorizationDecodeFunc(userKey, userType),
		tp.RequestDecodeFunc(requestKey, requestType),
	}
}

func (s *service) swagger() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		return swaggerMap, http.StatusOK, nil
	}
}

func (s *service) listBastions() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		request, ok := ctx.Value(requestKey).(*ListBastionsRequest)
		if !ok {
			return nil, http.StatusBadRequest, errBadRequest
		}

		user, ok := ctx.Value(userKey).(*schema.User)
		if !ok {
			return nil, http.StatusUnauthorized, errUnauthorized
		}

		bastions, err := s.ListBastions(user, request)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		return bastions, http.StatusOK, nil
	}
}

func (s *service) authenticateBastion() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		request, ok := ctx.Value(requestKey).(*opsee.AuthenticateBastionRequest)
		if !ok {
			return nil, http.StatusBadRequest, errBadRequest
		}

		resp, err := s.AuthenticateBastion(ctx, request)
		if err != nil {
			return nil, http.StatusUnauthorized, err
		}

		return resp, http.StatusOK, nil
	}
}
