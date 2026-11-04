package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "1e67b4eb-e5d2-4368-8284-4ba5c4c42455",
		Name:         "Creakwood Liege",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16Anthem(g2OtherCreaturesYouControlOfColor("B"), 1, 1),
			b16Anthem(g2OtherCreaturesYouControlOfColor("G"), 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			Optional(AtYourUpkeep("Creakwood Liege — create a 1/1 black and green Worm creature token",
				Do(CreateToken{Template: TokenCard("1/1 black and green Worm"), N: 1})),
				"Creakwood Liege — create a 1/1 black and green Worm creature token?"),
		},
	})
}
