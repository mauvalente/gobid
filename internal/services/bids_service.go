package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mauvalente/go-bid/internal/store/pgstore"
)

type BidService struct {
	pool    *pgxpool.Pool
	queries *pgstore.Queries
}

func NewBidService(pool *pgxpool.Pool) BidService {
	return BidService{
		pool:    pool,
		queries: pgstore.New(pool),
	}
}

var ErrBidIsTooLow = errors.New("the bid value is too low")

func (bs *BidService) PlaceBid(ctx context.Context, product_id, bidder_id uuid.UUID, amount float64) (pgstore.Bid, error) {
	product, err := bs.queries.GetProductById(ctx, product_id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Bid{}, err
		}
		return pgstore.Bid{}, err
	}

	highestBid, err := bs.queries.GetHighestBidByProductId(ctx, product_id)
	if err != nil {
		// se nao encontrou linha é a primeira a ser inserida
		if !errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Bid{}, err
		}
	}

	if product.Baseprice >= amount || highestBid.BidAmount >= amount {
		return pgstore.Bid{}, ErrBidIsTooLow
	}

	highestBid, err = bs.queries.CreateBid(ctx, pgstore.CreateBidParams{
		ProductID: product_id,
		BidderID:  bidder_id,
		BidAmount: amount,
	})
	if err != nil {
		return pgstore.Bid{}, err
	}

	return highestBid, nil
}
