package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mona Lisa, Ever Adaptable — {3}{G} Legendary Creature — Lizard Mutant
// 4/4 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"Whenever a player casts a creature spell, you create a Mutagen token."
//
// Any player's creature spell, Mona Lisa's controller's included, and the
// token is always hers. The trigger resolves before the spell, and even if
// the spell is countered (the rulings). The token is the catalog's
// CR 111.10v Mutagen (MutagenToken).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9fa38462-57b6-4f2f-a73c-f14a66f56946",
		Name:         "Mona Lisa, Ever Adaptable",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, tmcB8APlayerCast(Creature()),
				"Mona Lisa, Ever Adaptable — create a Mutagen token",
				Do(CreateToken{Template: MutagenToken(), N: 1})),
		},
	})
}
