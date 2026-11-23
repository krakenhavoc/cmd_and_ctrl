package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Theorist's Proxy — Creature — Illusion {1}{U}, 0/3 (Reality Fracture,
// tracker #2795):
//
//	"Flash
//	 When this creature enters, empower Jace 3.
//	 {U}, Sacrifice this creature: The next spell you cast this turn can't
//	 be countered."
//
// Empower Jace is the keyword action (ADR 0139). The sacrifice ability is
// the one-use promise of ADR 0106 §4 (Insist's shape) with no spell
// filter: spent by the next spell its player casts, or gone at cleanup.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0089acfe-da66-4dd7-b1e5-4d7407f58257",
		Name:            "Theorist's Proxy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Theorist's Proxy — empower Jace 3", effects.Do(effects.EmpowerJace{N: 3})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, Sacrifice this creature: The next spell you cast this turn can't be countered.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantCounterShield{
					From: "Theorist's Proxy",
					Grant: effects.NextSpellYouCastCantBeCountered(
						"The next spell you cast this turn can't be countered.", game.PermissionFilter{}),
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
