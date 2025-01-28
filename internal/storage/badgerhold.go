package storage

import (
	"context"

	"github.com/timshannon/badgerhold/v4"

	"github.com/psucodervn/verixilac/internal/model"
)

type BadgerHoldStorage struct {
	store *badgerhold.Store
}

func (b *BadgerHoldStorage) ResetBalance(ctx context.Context, newBalance int64) error {
	err := b.store.UpdateMatching(&model.Player{}, nil, func(record interface{}) error {
		p := record.(*model.Player)
		p.Balance = newBalance
		return nil
	})
	return err
}

func (b *BadgerHoldStorage) UpdatePlayerStatus(ctx context.Context, id string, status model.UserStatus) (*model.Player, error) {
	p := (*model.Player)(nil)
	err := b.store.UpdateMatching(&model.Player{}, badgerhold.Where("ID").Eq(id), func(record interface{}) error {
		p = record.(*model.Player)
		p.UserStatus = status
		return nil
	})
	if p == nil {
		return nil, badgerhold.ErrNotFound
	}
	return p, err
}

func (b *BadgerHoldStorage) AddPlayerBalance(ctx context.Context, id string, amount int64) (*model.Player, error) {
	p := (*model.Player)(nil)
	err := b.store.UpdateMatching(&model.Player{}, badgerhold.Where("ID").Eq(id), func(record interface{}) error {
		p = record.(*model.Player)
		p.Balance += amount
		return nil
	})
	if p == nil {
		return nil, badgerhold.ErrNotFound
	}
	return p, err
}

func (b *BadgerHoldStorage) ListPlayers(ctx context.Context) ([]model.Player, error) {
	var players []model.Player
	q := &badgerhold.Query{}
	err := b.store.Find(&players, q.SortBy("UserStatus", "Balance", "Name"))
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
	return b.store.Upsert(p.ID, p)
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

func (b *BadgerHoldStorage) ListRecords(ctx context.Context, playerID string, limit int) ([]model.Record, error) {
	var records []model.Record
	err := b.store.Find(&records, badgerhold.Where("PlayerID").Eq(playerID).SortBy("GameID").Limit(limit).Reverse())
	return records, err
}

func (b *BadgerHoldStorage) GetPlayerByID(ctx context.Context, id string) (*model.Player, error) {
	var p model.Player
	err := b.store.Get(id, &p)
	return &p, err
}
