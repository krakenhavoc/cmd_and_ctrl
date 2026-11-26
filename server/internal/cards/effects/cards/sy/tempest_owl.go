package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tempest Owl — Creature — Bird {1}{U} 1/2 (Zendikar, S82 OLD batch 491,
// #5583):
//
//	"Kicker {4}{U}
//	 Flying
//	 When this creature enters, if it was kicked, tap up to three target
//	 permanents."
//
// An intervening "if" on the Owl's own kicked record, and an "up to three"
// target clause; each target still legal as it resolves is tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dfd2bc6e-32c8-4027-9c0d-8a3050b8eeda",
		Name:            "Tempest Owl",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{effects.Kicker("{4}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Tempest Owl — kicked, tap up to three target permanents",
					effects.Msc13TapTheLegalTargets),
				effects.TargetPermanent("up to three target permanents").WithCount(0, 3)),
		},
	})
}
