package game

import (
	"testing"
)

func Test_generateRoomID(t *testing.T) {
	for i := 0; i < 10; i++ {
		if got := generateRoomID(); len(got) != 2 {
			t.Errorf("generateRoomID() = %v, len = %d", got, len(got))
		}
	}
}
