package model

import (
	"os"
	"strings"
)

type (
	UserRole   uint8 // 0: normal, 1: admin
	UserStatus uint8 // 0: active (joined), 1: inactive (left)
)

const (
	UserRoleNormal UserRole = iota
	UserRoleAdmin
	UserRoleBot = 100
)

const (
	UserStatusActive UserStatus = iota
	UserStatusInactive
)

type (
	Record struct {
		ID     uint64 `badgerhold:"key"`
		GameID string
		Data   any
	}

	Player struct {
		ID         string `badgerhold:"key"`
		TelegramID string `badgerhold:"index"`
		Name       string
		UserRole   UserRole
		UserStatus UserStatus
		Balance    int64
	}

	Following struct {
		ID         uint64 `badgerhold:"key"`
		FollowerID string `badgerhold:"index"`
		FolloweeID string `badgerhold:"index"`
	}
)

func (p Player) GetID() string {
	return p.ID
}

func (p Player) GetName() string {
	return p.Name
}

func (p Player) IsAdmin() bool {
	return p.UserRole == UserRoleAdmin || len(p.TelegramID) > 0 && p.TelegramID == os.Getenv("ADMIN_TELEGRAM_ID")
}

func (p Player) IsBot() bool {
	return p.UserRole == UserRoleBot || strings.HasPrefix(p.TelegramID, "BOT_")
}

func (p Player) IsActive() bool {
	return p.UserStatus == UserStatusActive
}

func (st UserStatus) String() string {
	switch st {
	case UserStatusActive:
		return "active"
	case UserStatusInactive:
		return "inactive"
	default:
		return "unknown"
	}
}
