package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Diversion Unit — Artifact Creature — Robot {1}{U}, 2/1 (Aetherdrift,
// #4464):
//
//	"Flying
//	 {U}, Sacrifice this creature: Counter target instant or sorcery spell
//	 unless its controller pays {3}."
//
// Malevolent Hermit's sacrifice-to-tax with the target an instant or
// sorcery spell. The spell is chosen before the Unit is sacrificed; as the
// ability resolves the spell's controller is asked to pay {3}, and it is
// countered if they don't (CR 118.12a). A spell that has left the stack
// by then makes the ability do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4d739edc-4e02-4b35-a2a0-e9685c6e2889",
		Name:            "Diversion Unit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, Sacrifice this creature: Counter target instant or sorcery spell unless its controller pays {3}.",
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.SacrificeThis()),
			Targets: effects.TargetSpell("target instant or sorcery spell", effects.Or(effects.Instant(), effects.Sorcery())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				targets := ctx.LegalTargets()
				if len(targets) == 0 {
					return nil
				}
				return effects.CounterUnlessPaid{
					StackID:  targets[0].ID,
					Cost:     "{3}",
					Question: "Diversion Unit — pay {3} or your spell is countered",
				}.Apply(ctx)
			},
		}},
	})
}
