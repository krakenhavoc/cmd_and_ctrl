package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rage Weaver (Invasion, S82 OLD batch 646, #5616):
//
//	Creature — Human Wizard {1}{R}, 2/1:
//	"{2}: Target black or green creature gains haste until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9777a54b-1c60-4d6b-bb83-c57cb3a1d7e6",
		Name:         "Rage Weaver",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.S82OldB646TargetGainsRow(effects.ManaCost("{2}"), "{2}", "black or green creature", "haste", "haste",
				effects.Or(effects.OfColor("B"), effects.OfColor("G"))),
		},
	})
}
