package game

import (
	"errors"
)

var (
	ErrPlayerNotFound          = errors.New("không có thông tin người chơi")
	ErrYouAlreadyInAnotherRoom = errors.New("bạn đã ở trong phòng khác")
	ErrPlayerAlreadyInRoom     = errors.New("người chơi đã ở trong phòng")
	ErrYouAlreadyInRoom        = errors.New("bạn đã ở trong phòng")
	ErrRoomNotFound            = errors.New("không có thông tin phòng")
	ErrNotInRoom               = errors.New("bạn chưa vào phòng")
	ErrGameIsExisted           = errors.New("đang có ván diễn ra trong phòng")
	ErrYouAlreadyInGame        = errors.New("bạn đã ở trong ván")
	ErrGameAlreadyStarted      = errors.New("ván đang diễn ra")
	ErrYouNotPlaying           = errors.New("bạn chưa tới lượt")
	ErrYouArePlayed            = errors.New("bạn đã qua lượt")
	ErrTooLow                  = errors.New("chưa đủ tẩy")
	ErrEmptyGame               = errors.New("chưa có người tham gia")
	ErrPlayerNotStandYet       = errors.New("người chơi chưa rút xong")
	ErrPlayerIsDone            = errors.New("đã tính rồi")
	ErrYouCannotHit            = errors.New("bạn không thể rút thêm")
	ErrYouCannotStand          = errors.New("bạn chưa thể úp bài")
)
