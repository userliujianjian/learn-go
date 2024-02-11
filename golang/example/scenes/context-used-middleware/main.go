package goexperments

import (
	"context"
	"net/http"
)

type HandlerMiddleware interface {
	HandleHTTPC(ctx context.Context, rw http.ResponseWriter, req *http.Request, next http.Handler)
}

var function1 HandlerMiddleware = nil
var function2 HandlerMiddleware = nil

func addUserID(rw http.ResponseWriter, req *http.Request, next http.Handler) {
	ctx := context.WithValue(req.Context(), "userid", req.Header.Get("userid"))
	req = req.WithContext(ctx)
	next.ServeHTTP(rw, req)
}

func userUserID(rw http.ResponseWriter, req *http.Request, next http.Handler) {
	uid := req.Context().Value("userid")
	rw.Write([]byte(uid))
}

func makeChain(chain ...HandlerMiddleware) http.Handler { return nil }

type Server struct{}

func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req = req.WithContext(context.Background())
	chain := makeChain(addUserID, function1, function2, useUserID)
	chain.ServeHTTP(rw, req)
}
