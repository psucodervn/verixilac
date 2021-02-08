package game

import (
	"testing"
)

func TestCard_String(t *testing.T) {
	tests := []struct {
		name string
		id   int
		want string
	}{
		{id: 0, want: "A♥"},
		{id: 9, want: "10♥"},
		{id: 38, want: "K♣"},
		{id: 51, want: "K♠"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Card{
				id: tt.id,
			}
			if got := c.String(); got != tt.want {
				t.Errorf("String() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestCard_Value(t *testing.T) {
	tests := []struct {
		name string
		id   int
		want int
	}{
		{id: 0, want: 1},
		{id: 1, want: 2},
		{id: 9, want: 10},
		{id: 10, want: 10},
		{id: 51, want: 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Card{
				id: tt.id,
			}
			if got := c.Value(); got != tt.want {
				t.Errorf("Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCards_Value(t *testing.T) {
	tests := []struct {
		name string
		cs   Cards
		want int
	}{
		{cs: NewCards(0, 0, 5), want: 18},
		{cs: NewCards(0, 0, 7), want: 20},
		{cs: NewCards(0, 0, 8), want: 21},
		{cs: NewCards(0, 0, 9), want: 21},
		{cs: NewCards(0, 0, 10), want: 21},
		{cs: NewCards(0, 0, 5, 6), want: 15},
		{cs: NewCards(5, 6), want: 13},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cs.Value(); got != tt.want {
				t.Errorf("Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCards_IsBlackJack(t *testing.T) {
	tests := []struct {
		name string
		cs   Cards
		want bool
	}{
		{cs: NewCards(0, 3), want: false},
		{cs: NewCards(0, 9), want: true},
		{cs: NewCards(9, 0), want: true},
		{cs: NewCards(0, 12), want: true},
		{cs: NewCards(0, 12, 2), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cs.IsBlackJack(); got != tt.want {
				t.Errorf("IsBlackJack() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCards_IsDoubleBlackJack(t *testing.T) {
	tests := []struct {
		name string
		cs   Cards
		want bool
	}{
		{cs: NewCards(0, 0), want: true},
		{cs: NewCards(0, 1), want: false},
		{cs: NewCards(1, 0), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cs.IsDoubleBlackJack(); got != tt.want {
				t.Errorf("IsDoubleBlackJack() = %v, want %v", got, tt.want)
			}
		})
	}
}
