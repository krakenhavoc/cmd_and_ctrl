package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Minotaur Tactician — Creature — Minotaur {3}{R}, 1/1 (Apocalypse, S82 OLD,
// #5614):
//
//	"Haste
//	 This creature gets +1/+1 as long as you control a white creature.
//	 This creature gets +1/+1 as long as you control a blue creature."
//
// Wild Nacatl's two printed statics, so two rows: with both it is a 3/3.
// Haste is a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "197fd850-0a08-4e82-9249-a8d719436f77",
		Name:            "Minotaur Tactician",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static: []game.StaticAbility{
			{
				Label:     "This creature gets +1/+1 as long as you control a white creature.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: effects.SelfWhileYouControlA(effects.And(effects.Creature(), effects.OfColor("W"))),
				Apply:     effects.PumpBy(1, 1),
			},
			{
				Label:     "This creature gets +1/+1 as long as you control a blue creature.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: effects.SelfWhileYouControlA(effects.And(effects.Creature(), effects.OfColor("U"))),
				Apply:     effects.PumpBy(1, 1),
			},
		},
	})
}
