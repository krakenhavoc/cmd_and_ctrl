package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scepter of Empires — Artifact {3} (Magic 2012, S82 OLD batch 452, #5573):
//
//	"{T}: This artifact deals 1 damage to target player or planeswalker.
//	 It deals 3 damage instead if you control artifacts named Crown of
//	 Empires and Throne of Empires."
//
// The amount is read as the ability resolves: 3 when the controller
// controls an artifact named Crown of Empires and one named Throne of
// Empires (two different permanents), else 1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "377d9245-9a70-4e35-9ed1-308565fc1acc",
		Name:         "Scepter of Empires",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This artifact deals 1 damage to target player or planeswalker. It deals 3 damage instead if you control artifacts named Crown of Empires and Throne of Empires.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayerOrPlaneswalker(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				amount := 1
				if scepterNamed(g, item.Controller, "Crown of Empires") && scepterNamed(g, item.Controller, "Throne of Empires") {
					amount = 3
				}
				return effects.DealDamageToTheTarget(amount)(g, item)
			},
		}},
	})
}

func scepterNamed(g *game.Game, you uuid.UUID, name string) bool {
	return effects.CountControlled(g, you, func(c game.Card) bool { return c.IsArtifact() && c.Name == name }) >= 1
}
