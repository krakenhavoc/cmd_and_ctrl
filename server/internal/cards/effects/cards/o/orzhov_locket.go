package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orzhov Locket — Artifact {3}:
//
//	"{T}: Add {W} or {B}.
//	 {W/B}{W/B}{W/B}{W/B}, {T}, Sacrifice this artifact: Draw two cards."
//
// Haunted Screen's {W|B} mana ability and Blitzball's tap-and-sacrifice
// draw; each hybrid symbol is paid with {W} or {B}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8aaf248e-66cf-447b-b062-9aa2ab1ec864",
		Name:         "Orzhov Locket",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|B}",
			Label:    "{T}: Add {W} or {B}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W/B}{W/B}{W/B}{W/B}, {T}, Sacrifice this artifact: Draw two cards.",
			Cost:    effects.Plus(effects.ManaCost("{W/B}{W/B}{W/B}{W/B}"), effects.TapCost(), effects.SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 2},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 2}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
