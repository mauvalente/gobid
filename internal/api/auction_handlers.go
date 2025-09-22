package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mauvalente/go-bid/internal/jsonutils"
	"github.com/mauvalente/go-bid/internal/services"
)

func (api *Api) handleSubscribeUserToAuction(w http.ResponseWriter, r *http.Request) {
	rawProductId := chi.URLParam(r, "product_id")

	productId, err := uuid.Parse(rawProductId)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"message": "invalid product id, must be a valid id",
		})
		return
	}

	_, err = api.ProductService.GetProductById(r.Context(), productId)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			jsonutils.EncodeJson(w, r, http.StatusNotFound, map[string]any{
				"error": "no product with given id",
			})
			return
		}
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "unexpected error, try again later",
		})
		return
	}

	userId, ok := api.Sessions.Get(r.Context(), "AuthenticatedUserId").(uuid.UUID)
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "unexpected error, try again later",
		})
		return
	}

	api.AuctionLobby.Lock()
	room, ok := api.AuctionLobby.Rooms[productId]
	api.AuctionLobby.Unlock()
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"message": "the auction has ended",
		})
		return
	}

	conn, err := api.WsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "could not upgrade connection to a websocket protocol",
		})
		return
	}

	client := services.NewClient(room, conn, userId)

	room.Register <- client
	go client.ReadEventLoop()
	go client.WriteEventLoop()

	for {

	}

}
