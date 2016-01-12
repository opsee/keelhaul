package service

import (
	"github.com/opsee/basic/com"
	"github.com/opsee/basic/tp"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

const (
	serviceKey = iota
	userKey
	requestKey
)

func (s *service) StartHTTP(addr string) {
	router := tp.NewHTTPRouter(context.Background())

	router.CORS(
		[]string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		[]string{`https?://localhost:8080`, `https?://localhost:8008`, `https://(\w+\.)?(opsy\.co|opsee\.co|opsee\.com)`},
	)

	// swagger
	router.Handle("GET", "/api/swagger.json", []tp.DecodeFunc{}, s.swagger())

	// json api
	router.Handle("POST", "/vpcs/scan", decoders(com.User{}, ScanVPCsRequest{}), s.scanVPCs())
	router.Handle("POST", "/vpcs/launch", decoders(com.User{}, LaunchBastionsRequest{}), s.launchBastions())
	router.Handle("GET", "/vpcs/bastions", decoders(com.User{}, ListBastionsRequest{}), s.listBastions())
	router.Handle("POST", "/bastions/authenticate", []tp.DecodeFunc{tp.RequestDecodeFunc(requestKey, AuthenticateBastionRequest{})}, s.authenticateBastion())

	// websocket
	router.HandlerFunc("GET", "/stream/", s.websocketHandlerFunc)

	// set a big timeout bc aws be slow
	router.Timeout(5 * time.Minute)

	http.ListenAndServe(addr, router)
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

func (s *service) scanVPCs() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		request, ok := ctx.Value(requestKey).(*ScanVPCsRequest)
		if !ok {
			return ctx, http.StatusBadRequest, errUnknown
		}

		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errUnknown
		}

		vpcs, err := s.ScanVPCs(user, request)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		return vpcs, http.StatusOK, nil
	}
}

func (s *service) launchBastions() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		request, ok := ctx.Value(requestKey).(*LaunchBastionsRequest)
		if !ok {
			return ctx, http.StatusBadRequest, errUnknown
		}

		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errUnknown
		}

		bastions, err := s.LaunchBastions(user, request)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		return bastions, http.StatusOK, nil
	}
}

func (s *service) listBastions() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		request, ok := ctx.Value(requestKey).(*ListBastionsRequest)
		if !ok {
			return ctx, http.StatusBadRequest, errUnknown
		}

		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errUnknown
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
		request, ok := ctx.Value(requestKey).(*AuthenticateBastionRequest)
		if !ok {
			return ctx, http.StatusBadRequest, errUnknown
		}

		resp, err := s.AuthenticateBastion(request)
		if err != nil {
			return nil, http.StatusUnauthorized, err
		}

		return resp, http.StatusOK, nil
	}
}
