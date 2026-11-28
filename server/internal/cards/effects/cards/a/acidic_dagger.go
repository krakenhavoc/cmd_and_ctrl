package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Acidic Dagger — Artifact {4} (Mirage, #5633):
//
//	"{4}, {T}: Whenever target creature deals combat damage to a non-Wall
//	 creature this turn, destroy that non-Wall creature. When the targeted
//	 creature leaves the battlefield this turn, sacrifice this artifact. Activate
//	 only before blockers are declared."
//
// Two delayed triggers (CR 603.7), both about the targeted creature as the object
// it was (CR 400.7): a repeating one on its combat damage to a non-Wall creature
// and one on its leaving the battlefield that sacrifices the Dagger.
//
// No simplification.
func init() {
	const label = "{4}, {T}: Whenever target creature deals combat damage to a non-Wall creature this turn, destroy that non-Wall creature. When the targeted creature leaves the battlefield this turn, sacrifice this artifact. Activate only before blockers are declared."
	effects.Register(effects.Spec{
		OracleID:     "bf034a20-e362-4f38-9bf9-857777c62598",
		Name:         "Acidic Dagger",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: label,
			Cost:  effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Condition: func(g *game.Game, _, _ uuid.UUID) bool {
				return effects.OldB795BeforeBlockersAreDeclared(g)
			},
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.OldB795AcidicDaggerResolve,
		}},
	})
}
