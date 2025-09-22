-- name: CreateProduct :one
INSERT INTO products (
    seller_id, product_name, description,
    baseprice, auction_end, is_sold
) VALUES (
    $1,$2,$3,$4,$5,$6
)
RETURNING id;


-- name: GetProductById :one
SELECT * FROM products
WHERE id = $1;


-- name: GetAllAvailableProducts :many
SELECT * FROM products
WHERE auction_end > now();
