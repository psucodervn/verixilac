package game

import (
	"context"

	"github.com/psucodervn/verixilac/internal/model"
)

type Storage interface {
	SaveRecord(ctx context.Context, r *model.Record) error
	ListRecords(ctx context.Context, playerID string, limit int) ([]model.Record, error)
	GetPlayerByID(ctx context.Context, id string) (*model.Player, error)
	SavePlayer(ctx context.Context, p *model.Player) error
	ListPlayers(ctx context.Context) ([]model.Player, error)
	ListActivePlayers(ctx context.Context) ([]model.Player, error)
	AddPlayerBalance(ctx context.Context, id string, amount int64) (*model.Player, error)
	UpdatePlayerStatus(ctx context.Context, id string, status model.UserStatus) (*model.Player, error)
	ResetBalance(ctx context.Context, newBalance int64) error
}
