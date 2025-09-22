package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/mauvalente/go-bid/internal/jsonutils"
)

func (api *Api) HandlerGetCSRFToken(w http.ResponseWriter, r *http.Request) {
	fmt.Println("entrei")
	token := csrf.Token(r)
	jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{
		"csrf_token": token,
	})
}

func (api *Api) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !api.Sessions.Exists(r.Context(), "AuthenticatedUserId") {
			jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]string{
				"message": "must be loggerd in",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
