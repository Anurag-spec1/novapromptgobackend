package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Anurag-spec1/novapromptgobackend/internal/models"
)

var ErrNotFound = errors.New("not found")

type ImageRepo struct {
	pool *pgxpool.Pool
}

func NewImageRepo(pool *pgxpool.Pool) *ImageRepo { return &ImageRepo{pool: pool} }

func (r *ImageRepo) Create(ctx context.Context, img *models.Image) error {
	const q = `
		INSERT INTO images (title, cloudinary_url, cloudinary_public_id, category_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, q, img.Title, img.CloudinaryURL, img.CloudinaryPublicID, img.CategoryID).
		Scan(&img.ID, &img.CreatedAt, &img.UpdatedAt)
}

func (r *ImageRepo) GetByID(ctx context.Context, id string) (*models.Image, error) {
	const q = `
		SELECT id, title, cloudinary_url, cloudinary_public_id, category_id, created_at, updated_at
		FROM images WHERE id = $1`
	var img models.Image
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&img.ID, &img.Title, &img.CloudinaryURL, &img.CloudinaryPublicID,
		&img.CategoryID, &img.CreatedAt, &img.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get image: %w", err)
	}
	return &img, nil
}

func (r *ImageRepo) List(ctx context.Context, limit, offset int) ([]models.Image, error) {
	const q = `
		SELECT id, title, cloudinary_url, cloudinary_public_id, category_id, created_at, updated_at
		FROM images
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}
	defer rows.Close()

	out := []models.Image{}
	for rows.Next() {
		var img models.Image
		if err := rows.Scan(
			&img.ID, &img.Title, &img.CloudinaryURL, &img.CloudinaryPublicID,
			&img.CategoryID, &img.CreatedAt, &img.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, img)
	}
	return out, rows.Err()
}

func (r *ImageRepo) Update(ctx context.Context, img *models.Image) error {
	const q = `
		UPDATE images
		SET title = $1, cloudinary_url = $2, cloudinary_public_id = $3, category_id = $4
		WHERE id = $5
		RETURNING updated_at`
	err := r.pool.QueryRow(ctx, q,
		img.Title, img.CloudinaryURL, img.CloudinaryPublicID, img.CategoryID, img.ID).
		Scan(&img.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (r *ImageRepo) Delete(ctx context.Context, id string) (string, error) {
	const q = `DELETE FROM images WHERE id = $1 RETURNING COALESCE(cloudinary_public_id, '')`
	var publicID string
	err := r.pool.QueryRow(ctx, q, id).Scan(&publicID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("delete image: %w", err)
	}
	return publicID, nil
}

func (r *ImageRepo) Search(ctx context.Context, tags []string, q string, limit, offset int) ([]models.Image, int, error) {
	args := []any{}
	var where []string
	
	if len(tags) > 0 {
		args = append(args, tags)
		where = append(where, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM image_tags it
			JOIN tags t ON t.id = it.tag_id
			WHERE it.image_id = images.id AND t.name = ANY($%d)
		)`, len(args)))
	}
	
	if q != "" {
		args = append(args, q)
		where = append(where, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM descriptions d
			WHERE d.image_id = images.id AND d.body %% $%d
		)`, len(args)))
	}
	
	var whereSQL string
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}
	
	countSQL := fmt.Sprintf(`SELECT count(*) FROM images %s`, whereSQL)
	var total int
	if err := r.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count search: %w", err)
	}
	
	if total == 0 {
		return []models.Image{}, 0, nil
	}
	
	args = append(args, limit, offset)
	itemsSQL := fmt.Sprintf(`
		SELECT id, title, cloudinary_url, cloudinary_public_id, category_id, created_at, updated_at
		FROM images
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereSQL, len(args)-1, len(args),
	)
	rows, err := r.pool.Query(ctx, itemsSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("search items: %w", err)
	}
	defer rows.Close()
	
	out := []models.Image{}
	for rows.Next() {
		var img models.Image
		if err := rows.Scan(
			&img.ID, &img.Title, &img.CloudinaryURL, &img.CloudinaryPublicID,
			&img.CategoryID, &img.CreatedAt, &img.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, img)
	}
	return out, total, rows.Err()
}
