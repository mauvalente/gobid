package api

import (
	"net/http"

	"github.com/mauvalente/go-bid/internal/jsonutils"
)

func (api *Api) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{
		"message": "OK",
	})
}
