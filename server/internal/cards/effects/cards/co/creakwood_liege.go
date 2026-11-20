package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Creakwood Liege — Creature — Horror {1}{B/G}{B/G}{B/G}, 2/2:
//
//	"Other black creatures you control get +1/+1.
//	 Other green creatures you control get +1/+1.
//	 At the beginning of your upkeep, you may create a 1/1 black and
//	 green Worm creature token."
//
// Glen Elendra Liege's two statics in black and green: each ability
// applies on its own, so a creature that is both gets +2/+2 (the Worm
// tokens among them), colours are read through the layered view, and
// "other" is by instance. The upkeep trigger is optional, asked as it
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1e67b4eb-e5d2-4368-8284-4ba5c4c42455",
		Name:         "Creakwood Liege",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(effects.G2OtherCreaturesYouControlOfColor("B"), 1, 1),
			effects.B16Anthem(effects.G2OtherCreaturesYouControlOfColor("G"), 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Creakwood Liege — create a 1/1 black and green Worm creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 black and green Worm"), N: 1})),
				"Creakwood Liege — create a 1/1 black and green Worm creature token?"),
		},
	})
}
