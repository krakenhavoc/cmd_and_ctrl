package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desperate Castaways — Creature — Human Pirate {1}{B}, 2/3 (Ixalan, #5532):
//
//	"This creature can't attack unless you control an artifact."
//
// An attack restriction read live as attackers are declared (CR 508.1c)
// over the controller's permanents: any artifact counts, an artifact
// creature included, and losing the last one before the declaration
// refuses the attack. Once declared, the attack stands.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0432e07d-baf6-4c20-8646-08338c7dc950",
		Name:         "Desperate Castaways",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.Labelled(effects.CantAttackUnlessYouControl(game.PermanentQuery{Types: []string{"artifact"}}),
				"This creature can't attack unless you control an artifact."),
		},
	})
}
