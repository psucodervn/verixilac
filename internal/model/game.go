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
		ID         uint64 `badgerhold:"key"`
		GameID     string
		PlayerID   string `badgerhold:"index"`
		Reward     int64
		ResultType ResultType
		Value      int
		IsDealer   bool
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

var (
	isAdmins = map[string]bool{}
)

func init() {
	for _, id := range strings.Split(os.Getenv("ADMIN_TELEGRAM_IDS"), ",") {
		if len(id) > 0 {
			isAdmins[id] = true
		}
	}
}

func (p Player) GetID() string {
	return p.ID
}

func (p Player) GetName() string {
	return p.Name
}

func (p Player) IsAdmin() bool {
	return p.UserRole == UserRoleAdmin || (len(p.TelegramID) > 0 && isAdmins[p.TelegramID])
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

type ResultType uint8

const (
	TypeDoubleBlackJack ResultType = iota
	TypeBlackJack
	TypeHighFive
	TypeNormal
	TypeBusted
	TypeTooHigh
	TypeTooLow
)

func (r ResultType) String() string {
	switch r {
	case TypeDoubleBlackJack:
		return "Xì bàn"
	case TypeBlackJack:
		return "Xì lác"
	case TypeHighFive:
		return "Ngũ linh"
	case TypeBusted:
		return "Toang"
	case TypeTooHigh:
		return "Đền"
	case TypeTooLow:
		return "Chưa đủ tẩy"
	default:
		return "normal"
	}
}
