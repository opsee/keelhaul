package service

import (
	"github.com/opsee/basic/schema"
	"github.com/opsee/basic/tp"
	"golang.org/x/net/context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	serviceKey = iota
	userKey
	requestKey
	paramsKey
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
	router.Handle("POST", "/vpcs/scan", decoders(schema.User{}, ScanVPCsRequest{}), s.scanVPCs())
	router.Handle("POST", "/vpcs/launch", decoders(schema.User{}, LaunchBastionsRequest{}), s.launchBastions())
	router.Handle("GET", "/vpcs/bastions", decoders(schema.User{}, ListBastionsRequest{}), s.listBastions())
	router.Handle("POST", "/bastions/authenticate", []tp.DecodeFunc{tp.RequestDecodeFunc(requestKey, AuthenticateBastionRequest{})}, s.authenticateBastion())

	router.Handle("GET", "/admin/bastions", []tp.DecodeFunc{tp.QueryDecoder(paramsKey)}, s.listTrackerStates())

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

		user, ok := ctx.Value(userKey).(*schema.User)
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

		user, ok := ctx.Value(userKey).(*schema.User)
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

		user, ok := ctx.Value(userKey).(*schema.User)
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

func (s *service) listTrackerStates() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		var (
			err      error
			limit    = 0
			offset   = 0
			bastions []string
			resp     interface{}
		)

		// TODO: make this use the general purporse parameter decoding
		params, _ := ctx.Value(paramsKey).(url.Values)

		bastions = params["bastionID"]
		if len(bastions) > 0 {
			resp, err = s.ListBastionStates(bastions)
		} else {
			l := params.Get("limit")
			o := params.Get("offset")
			if l != "" {
				limit, err = strconv.Atoi(l)
				if err != nil {
					return nil, http.StatusInternalServerError, err
				}
			}
			if o != "" {
				offset, err = strconv.Atoi(l)
				if err != nil {
					return nil, http.StatusInternalServerError, err
				}
			}
			resp, err = s.ListTrackerStates(offset, limit)
		}

		if err == nil {
			return resp, http.StatusOK, nil
		}
		return nil, http.StatusInternalServerError, err
	}
}
