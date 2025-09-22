package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/mauvalente/go-bid/internal/jsonutils"
	"github.com/mauvalente/go-bid/internal/services"
	"github.com/mauvalente/go-bid/internal/usecase/product"
)

func (api *Api) HandleListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := api.ProductService.GetAllAvailableProducts(r.Context())
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			jsonutils.EncodeJson(w, r, http.StatusNotFound, map[string]any{
				"message": "no product was found",
			})
			return
		}
		jsonutils.EncodeJson(w, r, http.StatusNotFound, map[string]any{
			"err": "an unexpected error has occured, please come back later",
		})
		return
	}

	jsonutils.EncodeJson(w, r, http.StatusOK, products)
}

func (api *Api) HandleCreateProduct(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[product.CreateProductReq](r)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, problems)
		return
	}

	SellerID, ok := api.Sessions.Get(r.Context(), "AuthenticatedUserId").(uuid.UUID)
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "unexpected error, try again later",
		})
		return
	}

	productId, err := api.ProductService.CreateProduct(
		r.Context(),
		SellerID,
		data.ProductName,
		data.Description,
		data.Baseprice,
		data.AuctionEnd,
	)
	if err != nil {
		fmt.Println(err)
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error": "fail to create product auction, try again later",
		})
		return
	}

	ctx, _ := context.WithDeadline(context.Background(), data.AuctionEnd)

	auctionRoom := services.NewAuctionRoom(ctx, productId, api.BidService)

	go auctionRoom.Run()

	api.AuctionLobby.Lock()
	api.AuctionLobby.Rooms[productId] = auctionRoom
	api.AuctionLobby.Unlock()

	jsonutils.EncodeJson(w, r, http.StatusCreated, map[string]any{
		"message":    "Auction has started with success",
		"product_id": productId,
	})

}
