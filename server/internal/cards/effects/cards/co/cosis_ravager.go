package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cosi's Ravager — Creature — Elemental {3}{R}, 2/2 (Worldwake, #5582):
//
//	"Landfall — Whenever a land you control enters, you may have this
//	 creature deal 1 damage to target player or planeswalker."
//
// Lava Axe's "target player or planeswalker", chosen as the trigger
// goes on the stack (CR 603.3d); the "you may" is asked as it
// resolves (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1253de62-6218-401c-9d98-a5ec8617e0d8",
		Name:         "Cosi's Ravager",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.Landfall("Cosi's Ravager — 1 damage to target player or planeswalker (landfall)", effects.SourceDealsOneToFirstTarget),
				effects.TargetPlayerOrPlaneswalker()),
				"Cosi's Ravager — deal 1 damage to target player or planeswalker?"),
		},
	})
}
