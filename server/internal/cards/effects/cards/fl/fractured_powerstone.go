package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fractured Powerstone — Artifact {2}:
//
//	"{T}: Add {C}.
//	 {T}: Roll the planar die. Activate only as a sorcery."
//
// The mana ability is the ordinary one. The roll is the planar die as a
// card says it (RollPlanarDieForEffect, CR 901.9): the same roll as the
// special action, with no {N} cost and not counted toward it. At a table
// not playing Planechase there is no planar die and the roll does
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af4f192a-9565-4377-b359-835fe4396e85",
		Name:         "Fractured Powerstone",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}: Roll the planar die. Activate only as a sorcery.",
			Cost:         effects.TapCost(),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				if g.Planechase == nil {
					return nil
				}
				return g.RollPlanarDieForEffect(item.Controller, item.SourceCardID)
			},
		}},
	})
}
