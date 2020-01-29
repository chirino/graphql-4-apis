package gateway

import (
    "github.com/chirino/graphql/resolvers"
    "net/http"
    "time"
)

type root byte

func (root) Login(rctx resolvers.ExecutionContext, args struct{ Token string }) string {
    ctx := rctx.GetContext()
    if r := ctx.Value("net/http.ResponseWriter"); r != nil {
        if r, ok := r.(http.ResponseWriter); ok {
            http.SetCookie(r, &http.Cookie{
                Name:    "Authorization",
                Value:   "Bearer " + args.Token,
                Path:    "/",
                Expires: time.Now().Add(1 * time.Hour),
            })
        }
    }
    return "ok"
}

func (root) Logout(rctx resolvers.ExecutionContext) string {
    ctx := rctx.GetContext()
    if r := ctx.Value("net/http.ResponseWriter"); r != nil {
        if r, ok := r.(http.ResponseWriter); ok {
            http.SetCookie(r, &http.Cookie{
                Name:    "Authorization",
                Value:   "",
                Path:    "/",
                Expires: time.Now().Add(-10000 * time.Hour),
            })
        }
    }
    return "ok"
}

func (root) Api() string {
    return "ok"
}
