package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodthorn Taunter — Creature — Human Scout {1}{R}, 1/1 (Shards of Alara,
// S82 OLD batch 510, #5588):
//
//	"Haste
//	 {T}: Target creature with power 5 or greater gains haste until end of
//	 turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b59da333-e658-447e-a8bd-8c0124d8b905",
		Name:            "Bloodthorn Taunter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{
			effects.S82OldB510TargetGainsRow(effects.TapCost(), "{T}", "with power 5 or greater", "haste", effects.PowerGE(5)),
		},
	})
}
