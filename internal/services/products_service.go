package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mauvalente/go-bid/internal/store/pgstore"
)

type ProductService struct {
	pool    *pgxpool.Pool
	queries *pgstore.Queries
}

func NewProductService(pool *pgxpool.Pool) ProductService {
	return ProductService{
		pool:    pool,
		queries: pgstore.New(pool),
	}
}

func (ps *ProductService) CreateProduct(
	ctx context.Context,
	sellerId uuid.UUID,
	productName, description string,
	baseprice float64,
	auctionEnd time.Time,
) (uuid.UUID, error) {
	id, err := ps.queries.CreateProduct(ctx, pgstore.CreateProductParams{
		SellerID:    sellerId,
		ProductName: productName,
		Description: description,
		Baseprice:   baseprice,
		AuctionEnd:  auctionEnd,
		IsSold:      false,
	})
	if err != nil {
		return uuid.UUID{}, err
	}
	return id, nil
}

var ErrProductNotFound = errors.New("product not found")

func (ps *ProductService) GetProductById(ctx context.Context, product_id uuid.UUID) (pgstore.Product, error) {
	product, err := ps.queries.GetProductById(ctx, product_id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Product{}, ErrProductNotFound
		}
		return pgstore.Product{}, err
	}
	return product, nil
}

func (ps *ProductService) GetAllAvailableProducts(ctx context.Context) ([]pgstore.Product, error) {
	products, err := ps.queries.GetAllAvailableProducts(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []pgstore.Product{}, ErrProductNotFound
		}
		return []pgstore.Product{}, err
	}
	return products, nil
}
