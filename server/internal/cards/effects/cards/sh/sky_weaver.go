package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sky Weaver (Invasion, S82 OLD batch 645, #5616):
//
//	Creature — Metathran Wizard {1}{U}, 2/1:
//	"{2}: Target white or black creature gains flying until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25933e8b-fa48-4152-9eae-388aef10ad37",
		Name:         "Sky Weaver",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.S82OldB645TargetGainsRow(effects.ManaCost("{2}"), "{2}", "white or black creature", "flying",
				effects.Or(effects.OfColor("W"), effects.OfColor("B"))),
		},
	})
}
