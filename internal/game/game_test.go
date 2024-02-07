package game

import (
	"fmt"
	"sync"
	"testing"

	"go.uber.org/atomic"

	"github.com/psucodervn/verixilac/internal/model"
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
		{args: args{aIds: []int{7, 7, 7}, bIds: []int{7, 7, 8}}, want: Draw},
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

func TestGetReward(t *testing.T) {
	type args struct {
		aIds   []int
		bIds   []int
		ruleID string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{args: args{aIds: []int{7, 7, 7}, bIds: []int{7, 7, 8}}, want: 0},
		{args: args{aIds: []int{10, 12}, bIds: []int{5, 0, 11}}, want: 1},
		{args: args{aIds: []int{1, 2}, bIds: []int{0, 5, 7}}, want: 0},
		{args: args{aIds: []int{1, 2}, bIds: []int{0, 5, 9}}, want: -1},
		{args: args{aIds: []int{9, 5}, bIds: []int{0, 5, 7}}, want: 1},
		{args: args{aIds: []int{7, 8}, bIds: []int{0, 5, 7}}, want: 1},
		{args: args{aIds: []int{7, 8}, bIds: []int{0, 13}}, want: -2},
		{args: args{aIds: []int{7, 8}, bIds: []int{0, 13}, ruleID: "2"}, want: -3},
		{args: args{aIds: []int{0, 13}, bIds: []int{26, 39}}, want: 0},
		{args: args{aIds: []int{0, 13}, bIds: []int{1, 2}}, want: 1},
		{args: args{aIds: []int{0, 13}, bIds: []int{1, 2}, ruleID: "2"}, want: 3},
		{args: args{aIds: []int{0, 10}, bIds: []int{1, 2}, ruleID: "2"}, want: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := DefaultRules[DefaultRuleID]
			if tt.args.ruleID != "" {
				rule = DefaultRules[tt.args.ruleID]
			}
			pa := &PlayerInGame{cards: NewCards(tt.args.aIds...), isDealer: *atomic.NewBool(true)}
			pb := &PlayerInGame{cards: NewCards(tt.args.bIds...), isDealer: *atomic.NewBool(false), betAmount: *atomic.NewUint64(1)}
			if got := GetReward(&rule, pa, pb); got != tt.want {
				t.Errorf("GetReward() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGame_Deal(t *testing.T) {
	g := &Game{
		dealer: NewPlayerInGame(&model.Player{}, 0, true),
		players: []*PlayerInGame{
			NewPlayerInGame(&model.Player{Name: "Bot #1"}, 10, false),
			NewPlayerInGame(&model.Player{Name: "Bot #2"}, 20, false),
		},
	}
	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			if err := g.Deal(); err != nil {
				t.Logf("Deal() error = %v", err)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	fmt.Println(g.CurrentBoard())
}
