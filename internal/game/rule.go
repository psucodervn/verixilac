package game

import (
	"fmt"
	"sort"
	"strings"

	"github.com/psucodervn/verixilac/internal/model"
)

type Rule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Multipliers map[PlayerType]map[model.ResultType]int64
}

var (
	DefaultRuleID = "1"
	DefaultRules  = map[string]Rule{
		"1": {
			ID:          "1",
			Name:        "Hai Dinh",
			Description: `Xì lác, ngũ linh: x2. Xì bàn: x3. Con cái như nhau.`,
			Multipliers: map[PlayerType]map[model.ResultType]int64{
				Dealer: {
					model.TypeDoubleBlackJack: 3,
					model.TypeHighFive:        2,
					model.TypeBlackJack:       2,
				},
				Participant: {
					model.TypeDoubleBlackJack: 3,
					model.TypeHighFive:        2,
					model.TypeBlackJack:       2,
				},
			},
		},
		"2": {
			ID:          "2",
			Name:        "Normal",
			Description: `Xì bàn: con x2, cái x1`,
			Multipliers: map[PlayerType]map[model.ResultType]int64{
				Participant: {
					model.TypeDoubleBlackJack: 2,
				},
			},
		},
	}
	DefaultRule   = DefaultRules[DefaultRuleID]
	SortedRuleIDs []string
	RuleListText  string
)

func init() {
	for id := range DefaultRules {
		SortedRuleIDs = append(SortedRuleIDs, id)
	}
	sort.Strings(SortedRuleIDs)

	var bf strings.Builder
	bf.WriteString(`Danh sách rules:`)
	for _, id := range SortedRuleIDs {
		bf.WriteString(fmt.Sprintf("\n\n - Rule: %s, ID: %s", DefaultRules[id].Name, id))
		bf.WriteString(fmt.Sprintf("\n%s", DefaultRules[id].Description))
	}
	RuleListText = bf.String()
}
