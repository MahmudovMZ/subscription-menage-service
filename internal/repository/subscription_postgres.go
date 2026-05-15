package repository

import (
	"context"
	"database/sql"
	"log"
	"subscriptions/internal/models"
	"time"
)

type Repository interface {
	Create(ctx context.Context, sub *models.Subscription) error
	GetTotalStats(ctx context.Context, userId, serviceName string, from, to time.Time) (int, error)
	GetSubByID(ctx context.Context, userID, subID string) (models.Subscription, error)
	GetSubList(ctx context.Context, userID string) ([]models.Subscription, error)
	DeleteSubByID(ctx context.Context, userID, subId string) error
	UpdateSubByID(ctx context.Context, userID, subId string, sub *models.Subscription) error
}

type SubscriptionRepo struct {
	DB *sql.DB
}

func NewSubscriptionRepo(db *sql.DB) *SubscriptionRepo {
	return &SubscriptionRepo{DB: db}
}

func (s *SubscriptionRepo) Create(ctx context.Context, sub *models.Subscription) error {
	log.Printf("Repository: Creating subscription for UserID: %s", sub.UserID)

	query := `INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date) 
              VALUES ($1, $2, $3, $4, $5) RETURNING id`

	err := s.DB.QueryRowContext(
		ctx,
		query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
	).Scan(&sub.ID)

	if err != nil {
		log.Printf("Repository Error (Create): %v", err)
	}
	return err
}

func (s *SubscriptionRepo) GetTotalStats(ctx context.Context, userId, serviceName string, from, to time.Time) (int, error) {
	log.Printf("Repository: Calculating stats for UserID: %s (%s to %s)", userId, from.Format("2006-01-02"), to.Format("2006-01-02"))

	var total sql.NullInt64
	query := `SELECT SUM(price) FROM subscriptions 
              WHERE service_name = $1 AND user_id = $2 AND start_date >= $3 AND start_date < $4`

	err := s.DB.QueryRowContext(ctx, query, serviceName, userId, from, to).Scan(&total)
	if err != nil {
		log.Printf("Repository Error (GetTotalStats): %v", err)
		return 0, err
	}

	if !total.Valid {
		return 0, nil
	}
	return int(total.Int64), nil
}

func (s *SubscriptionRepo) GetSubByID(ctx context.Context, userID, subID string) (models.Subscription, error) {
	log.Printf("Repository: Fetching subscription %s for user %s", subID, userID)

	var sub models.Subscription
	query := `SELECT id, service_name, price, user_id, start_date, end_date 
              FROM subscriptions WHERE id = $1 AND user_id = $2`

	err := s.DB.QueryRowContext(ctx, query, subID, userID).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.Subscription{}, sql.ErrNoRows
		}
		log.Printf("Repository Error (GetSubByID): %v", err)
		return models.Subscription{}, err
	}
	return sub, nil
}

func (s *SubscriptionRepo) GetSubList(ctx context.Context, userID string) ([]models.Subscription, error) {
	log.Printf("Repository: Fetching all subscriptions for user %s", userID)

	query := `SELECT id, service_name, price, user_id, start_date, end_date 
              FROM subscriptions WHERE user_id = $1`

	rows, err := s.DB.QueryContext(ctx, query, userID)
	if err != nil {
		log.Printf("Repository Error (GetSubList): %v", err)
		return nil, err
	}
	defer rows.Close()

	var subs []models.Subscription
	for rows.Next() {
		var sub models.Subscription
		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&sub.EndDate,
		)
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

func (s *SubscriptionRepo) DeleteSubByID(ctx context.Context, userID, subId string) error {
	log.Printf("Repository: Deleting subscription %s for user %s", subId, userID)

	query := `DELETE FROM subscriptions WHERE user_id = $1 AND id = $2`
	_, err := s.DB.ExecContext(ctx, query, userID, subId)
	if err != nil {
		log.Printf("Repository Error (Delete): %v", err)
	}
	return err
}

func (s *SubscriptionRepo) UpdateSubByID(ctx context.Context, userID, subId string, sub *models.Subscription) error {
	log.Printf("Repository: Updating subscription %s for user %s", subId, userID)
	query := `UPDATE subscriptions SET service_name = $1, price = $2, start_date = $3, end_date = $4 
              WHERE user_id = $5 AND id = $6`

	_, err := s.DB.ExecContext(ctx, query, sub.ServiceName, sub.Price, sub.StartDate, sub.EndDate, userID, subId)
	if err != nil {
		log.Printf("Repository Error (Update): %v", err)
	}
	return err
}
