package game

import (
	"testing"

	"go.uber.org/atomic"
)

func TestCompare(t *testing.T) {
	type args struct {
		aIds []int
		bIds []int
	}
	tests := []struct {
		name string
		args args
		want Result
	}{
		{args: args{aIds: []int{10, 12}, bIds: []int{5, 0, 11}}, want: Win},
		{args: args{aIds: []int{1, 2}, bIds: []int{0, 5, 7}}, want: Draw},
		{args: args{aIds: []int{1, 2}, bIds: []int{0, 5, 9}}, want: Lose},
		{args: args{aIds: []int{9, 5}, bIds: []int{0, 5, 7}}, want: Win},
		{args: args{aIds: []int{7, 8}, bIds: []int{0, 5, 7}}, want: Win},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pa := &PlayerInGame{cards: NewCards(tt.args.aIds...), isDealer: *atomic.NewBool(true)}
			pb := &PlayerInGame{cards: NewCards(tt.args.bIds...), isDealer: *atomic.NewBool(false)}
			if got := Compare(pa, pb); got != tt.want {
				t.Errorf("Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_reverseResult(t *testing.T) {
	tests := []struct {
		name string
		res  Result
		want Result
	}{
		{res: Win, want: Lose},
		{res: Lose, want: Win},
		{res: Draw, want: Draw},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reverseResult(tt.res); got != tt.want {
				t.Errorf("reverseResult() = %v, want %v", got, tt.want)
			}
		})
	}
}
