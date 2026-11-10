package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Infestation Sage — Creature — Elf Warlock {B}, 1/1 (Foundations, #4465):
//
//	"When this creature dies, create a 1/1 black and green Insect
//	 creature token with flying."
//
// Common Crook's dies trigger with The Swarmweaver's Insect from the
// token table. The token is created under the Sage's controller as it
// last existed (CR 603.10a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d398dca3-f3a5-44c9-998f-b4fce9e5c2bf",
		Name:         "Infestation Sage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Infestation Sage — create a 1/1 black and green Insect creature token with flying",
				Do(CreateToken{Template: TokenCard("1/1 black and green Insect with flying"), N: 1})),
		},
	})
}
