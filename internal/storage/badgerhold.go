package storage

import (
	"context"

	"github.com/timshannon/badgerhold/v4"

	"github.com/psucodervn/verixilac/internal/model"
)

type BadgerHoldStorage struct {
	store *badgerhold.Store
}

func (b *BadgerHoldStorage) UpdatePlayerStatus(ctx context.Context, id string, status model.UserStatus) (*model.Player, error) {
	p := (*model.Player)(nil)
	err := b.store.UpdateMatching(&model.Player{}, badgerhold.Where("ID").Eq(id), func(record interface{}) error {
		p = record.(*model.Player)
		p.UserStatus = status
		return nil
	})
	return p, err
}

func (b *BadgerHoldStorage) AddPlayerBalance(ctx context.Context, id string, amount int64) (*model.Player, error) {
	p := (*model.Player)(nil)
	err := b.store.UpdateMatching(&model.Player{}, badgerhold.Where("ID").Eq(id), func(record interface{}) error {
		p = record.(*model.Player)
		p.Balance += amount
		return nil
	})
	return p, err
}

func (b *BadgerHoldStorage) ListPlayers(ctx context.Context) ([]model.Player, error) {
	var players []model.Player
	q := &badgerhold.Query{}
	err := b.store.Find(&players, q)
	return players, err
}

func (b *BadgerHoldStorage) ListActivePlayers(ctx context.Context) ([]model.Player, error) {
	var players []model.Player
	err := b.store.Find(&players, badgerhold.Where("UserStatus").Eq(model.UserStatusActive))
	return players, err
}

func (b *BadgerHoldStorage) SavePlayer(ctx context.Context, p *model.Player) error {
	if len(p.ID) == 0 {
		p.ID = p.TelegramID
	}
	return b.store.Insert(p.ID, p)
}

func NewBadgerHoldStorage(dir string) *BadgerHoldStorage {
	opts := badgerhold.DefaultOptions
	opts.Dir = dir
	opts.ValueDir = dir
	opts.NumVersionsToKeep = 1
	store, err := badgerhold.Open(opts)
	if err != nil {
		panic(err)
	}

	return &BadgerHoldStorage{
		store: store,
	}
}

func (b *BadgerHoldStorage) Close() error {
	return b.store.Close()
}

func (b *BadgerHoldStorage) SaveRecord(ctx context.Context, r *model.Record) error {
	return b.store.Insert(badgerhold.NextSequence(), r)
}

func (b *BadgerHoldStorage) ListRecords(ctx context.Context, playerID string) ([]model.Record, error) {
	var records []model.Record
	q := &badgerhold.Query{}
	err := b.store.Find(&records, q.SortBy("GameID").Limit(10).Reverse())
	return records, err
}

func (b *BadgerHoldStorage) GetPlayerByID(ctx context.Context, id string) (*model.Player, error) {
	var p model.Player
	err := b.store.Get(id, &p)
	return &p, err
}
