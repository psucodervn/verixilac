package game

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func generateRoomID() string {
	bi, _ := rand.Int(rand.Reader, big.NewInt(100))
	return fmt.Sprintf("%02d", bi.Int64())
}
