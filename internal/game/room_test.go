package game

import (
	"testing"
)

func TestRoom_GetID(t *testing.T) {
	r := NewRoom("123")
	if got, want := r.ID(), r.id; got != want {
		t.Errorf("ID() = %v, want %v", got, want)
	}
}
