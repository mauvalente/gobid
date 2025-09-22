package services

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MessageKind int

const (
	//Request
	PlaceBid MessageKind = iota

	// Ok/Success
	SuccessfullyPlaceBid

	// Info
	NewBidPlaced
	AuctionFinished
	InvalidJSON

	//Errors
	FailedToPlaceBid
)

type Message struct {
	Message string      `json:"message,omitempty"`
	Kind    MessageKind `json:"kind"`
	UserId  uuid.UUID   `json:"user_id,omitempty"`
	Amount  float64     `json:"amount,omitempty"`
}

type AuctionLobby struct {
	sync.Mutex
	Rooms map[uuid.UUID]*AuctionRoom
}

type AuctionRoom struct {
	Id         uuid.UUID
	Context    context.Context
	Broadcast  chan Message
	Register   chan *Client
	Unregister chan *Client
	Clients    map[uuid.UUID]*Client

	BidService BidService
}

func NewAuctionRoom(ctx context.Context, id uuid.UUID, BidService BidService) *AuctionRoom {
	return &AuctionRoom{
		Id:         id,
		Context:    ctx,
		Broadcast:  make(chan Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[uuid.UUID]*Client),
		BidService: BidService,
	}
}

func (r *AuctionRoom) registerClient(c *Client) {
	slog.Info("New user Connected", "Client", c)
	r.Clients[c.UserId] = c
}

func (r *AuctionRoom) unregisterClient(c *Client) {
	slog.Info("New user Disconnected", "Client", c)
	delete(r.Clients, c.UserId)
}

func (r *AuctionRoom) broadcastMessage(m Message) {
	slog.Info("New message received", "RoomId", r.Id, "message", m.Message, "user_id", m.UserId)
	switch m.Kind {
	case PlaceBid:
		bid, err := r.BidService.PlaceBid(r.Context, r.Id, m.UserId, m.Amount)
		if err != nil {
			if errors.Is(err, ErrBidIsTooLow) {
				if client, ok := r.Clients[m.UserId]; ok {
					client.Send <- Message{Kind: FailedToPlaceBid, Message: ErrBidIsTooLow.Error(), UserId: m.UserId}
				}
				return
			}
		}

		if client, ok := r.Clients[m.UserId]; ok {
			client.Send <- Message{Kind: SuccessfullyPlaceBid, Message: "Your bid was Successfully placed.", UserId: m.UserId}
		}

		for id, client := range r.Clients {
			if id == m.UserId {
				continue
			}
			newBidMessage := Message{Kind: NewBidPlaced, Message: "A new bid was placed", Amount: bid.BidAmount, UserId: m.UserId}
			client.Send <- newBidMessage
		}
	case InvalidJSON:
		client, ok := r.Clients[m.UserId]
		if !ok {
			slog.Info("Client not found in hashmap", "UserId", m.UserId)
			return
		}
		client.Send <- m

	}
}

func (r *AuctionRoom) Run() {
	slog.Info("Auction has begun,", "auctionId", r.Id)

	defer func() {
		close(r.Broadcast)
		close(r.Register)
		close(r.Unregister)
	}()

	for {
		select {
		case client := <-r.Register:
			r.registerClient(client)
		case client := <-r.Unregister:
			r.unregisterClient(client)
		case message := <-r.Broadcast:
			r.broadcastMessage(message)
		case <-r.Context.Done():
			slog.Info("Auction has ended.", "auctionId", r.Id)
			for _, client := range r.Clients {
				client.Send <- Message{Kind: AuctionFinished, Message: "Auction has been finished"}
			}
			return
		}
	}
}

type Client struct {
	Room   *AuctionRoom
	Conn   *websocket.Conn
	Send   chan Message
	UserId uuid.UUID
}

func NewClient(room *AuctionRoom, conn *websocket.Conn, userId uuid.UUID) *Client {
	return &Client{
		Room:   room,
		Conn:   conn,
		Send:   make(chan Message, 512),
		UserId: userId,
	}
}

const (
	maxMessageSize = 512
	readDeadLine   = 60 * time.Second
	writeWait      = 10 * time.Second
	pingPeriod     = (readDeadLine * 9) / 10
)

func (c *Client) ReadEventLoop() {
	defer func() {
		c.Room.Unregister <- c
		c.Conn.Close()
	}()

	// Limitanto o maximo de bytes pra receber do usuário
	c.Conn.SetReadLimit(maxMessageSize)
	// Tempo que a connexão aceita mensagens
	c.Conn.SetReadDeadline(time.Now().Add(readDeadLine))

	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(readDeadLine))
		return nil
	})

	for {
		var m Message
		m.UserId = c.UserId
		err := c.Conn.ReadJSON(&m)
		if err != nil {
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("Unexpected Close error", "error", err)
				return
			}

			c.Room.Broadcast <- Message{
				Kind:    InvalidJSON,
				Message: "this message should be a valid json",
				UserId:  m.UserId,
			}
			continue
		}
		c.Room.Broadcast <- m
	}
}

func (c *Client) WriteEventLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteJSON(Message{
					Kind:    websocket.CloseMessage,
					Message: "Closing websocket message",
				})
				return
			}
			if message.Kind == AuctionFinished {
				close(c.Send)
				return
			}
			c.Conn.SetReadDeadline(time.Now().Add(writeWait))

			err := c.Conn.WriteJSON(message)
			if err != nil {
				c.Room.Unregister <- c
				return

			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Error("Unexpected write error", "error", err)
				return
			}
		}

	}
}
