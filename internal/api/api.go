package api

import (
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/mauvalente/go-bid/internal/services"
)

type Api struct {
	Router       *chi.Mux
	Sessions     *scs.SessionManager
	WsUpgrader   websocket.Upgrader
	AuctionLobby services.AuctionLobby

	UserService    services.UserService
	ProductService services.ProductService
	BidService     services.BidService
}
