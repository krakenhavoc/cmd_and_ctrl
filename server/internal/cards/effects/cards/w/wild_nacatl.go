package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wild Nacatl — Creature — Cat Warrior {G}, 1/1 (Shards of Alara, S82 OLD
// batch 511, #5588):
//
//	"This creature gets +1/+1 as long as you control a Mountain.
//	 This creature gets +1/+1 as long as you control a Plains."
//
// Two printed statics, so two rows: with both lands it is a 3/3.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "35d3d96a-d303-4bf5-b972-f311331ec07a",
		Name:         "Wild Nacatl",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "This creature gets +1/+1 as long as you control a Mountain.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: effects.SelfWhileYouControlA(effects.Subtype("Mountain")),
				Apply:     effects.PumpBy(1, 1),
			},
			{
				Label:     "This creature gets +1/+1 as long as you control a Plains.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: effects.SelfWhileYouControlA(effects.Subtype("Plains")),
				Apply:     effects.PumpBy(1, 1),
			},
		},
	})
}
