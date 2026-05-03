package usecase

import "context"

type Publisher interface {
	Publish(ctx context.Context, subject string, payload any) error
}
