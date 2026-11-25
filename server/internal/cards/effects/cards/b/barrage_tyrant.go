package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Barrage Tyrant — Creature — Eldrazi {4}{R}, 5/3 (Battle for Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 {2}{R}, Sacrifice another colorless creature: This creature deals damage
//	 equal to the sacrificed creature's power to any target."
//
// Bloodshot Cyclops' fling: the amount is the sacrificed creature's power as it
// last existed on the battlefield, floored at zero (CR 107.1b). Only another
// colorless creature may be sacrificed (judged through the layers).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f3b70d61-e33e-4243-aac9-e7bd9ba40acb",
		Name:            "Barrage Tyrant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, Sacrifice another colorless creature: This creature deals damage equal to the sacrificed creature's power to any target.",
			Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.SacrificeN(1, "another colorless creature", effects.Creature(), effects.Colorless())),
			Targets: effects.TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DamageEqualToSacrificedPower(item, effects.NewContext(g, item))
			},
		}},
	})
}
