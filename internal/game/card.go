package game

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

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

var (
	CardValueNames = []rune("A23456789_JQK")
	CardKindNames  = []rune("♥♦♣♠")
)

type Card struct {
	id int
}

func (c Card) Value() int {
	v := c.id % 13
	if v > 9 {
		v = 9
	}
	return v + 1
}

func (c Card) String() string {
	v := c.id % 13
	bf := bytes.NewBuffer(nil)
	if v == 9 {
		bf.WriteString("10")
	} else {
		bf.WriteRune(CardValueNames[v])
	}
	bf.WriteRune(CardKindNames[c.id/13])
	return bf.String()
}

type Cards []Card

func NewCards(ids ...int) Cards {
	var cs Cards
	for _, id := range ids {
		cs = append(cs, Card{id: id})
	}
	return cs
}

func (cs Cards) IsBlackJack() bool {
	if len(cs) != 2 {
		return false
	}
	return cs[0].Value() == 1 && cs[1].Value() == 10 || cs[0].Value() == 10 && cs[1].Value() == 1
}

func (cs Cards) IsDoubleBlackJack() bool {
	return len(cs) == 2 && cs[0].Value() == 1 && cs[1].Value() == 1
}

func (cs Cards) IsHighFive() bool {
	return len(cs) == 5 && cs.Value() <= 21
}

func (cs Cards) Value() int {
	aCnt := 0
	sum := 0
	for _, c := range cs {
		if c.Value() == 1 {
			aCnt++
		} else {
			sum += c.Value()
		}
	}
	if aCnt == 0 {
		return sum
	}
	if sum >= 12 || len(cs) >= 4 {
		return sum + aCnt
	}
	if sum+11+(aCnt-1) <= 21 {
		return sum + 11 + (aCnt - 1)
	}
	if sum+10+(aCnt-1) <= 21 {
		return sum + 10 + (aCnt - 1)
	}
	return sum + aCnt
}

func (cs Cards) String(censor bool, isDealer ...bool) string {
	if censor {
		return strings.Repeat("**, ", len(cs)-1) + " ** (" + strconv.Itoa(len(cs)) + " lá)"
	}
	s := make([]string, len(cs))
	for i := range cs {
		s[i] = cs[i].String()
	}
	return strings.Join(s, ", ") + " (" + cs.TypeString(isDealer...) + ")"
}

func (cs Cards) Type(isDealer ...bool) ResultType {
	if cs.IsDoubleBlackJack() {
		return TypeDoubleBlackJack
	} else if cs.IsBlackJack() {
		return TypeBlackJack
	} else if cs.IsHighFive() {
		return TypeHighFive
	}
	val := cs.Value()
	min := 16
	if len(isDealer) > 0 && isDealer[0] {
		min = 15
	}
	if val < min {
		return TypeTooLow
	} else if val >= 28 {
		return TypeTooHigh
	} else if val > 21 {
		return TypeBusted
	}
	return TypeNormal
}

func (cs Cards) TypeString(isDealer ...bool) string {
	switch cs.Type(isDealer...) {
	case TypeHighFive:
		return fmt.Sprintf("ngũ linh: %d điểm ⚡️", cs.Value())
	case TypeBusted:
		return fmt.Sprintf("toang: %d điểm 💥", cs.Value())
	case TypeBlackJack:
		return "xì lác ⚡️"
	case TypeDoubleBlackJack:
		return "xì bàn ⚡️"
	case TypeTooLow:
		return fmt.Sprintf("chưa đủ tẩy: %d điểm", cs.Value())
	case TypeTooHigh:
		return fmt.Sprintf("đền: %d điểm", cs.Value())
	default:
		return fmt.Sprintf("%d điểm", cs.Value())
	}
}
