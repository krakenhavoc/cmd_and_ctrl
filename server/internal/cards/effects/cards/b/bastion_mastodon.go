package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bastion Mastodon — Artifact Creature — Elephant {5}, 4/5 (Kaladesh, #5539):
//
//	"{W}: This creature gains vigilance until end of turn."
//
// Femeref Knight's self grant: a layer-6 vigilance pinned to this object
// until the turn ends (CR 613.1f, CR 514.2) through the shared
// thisGainsKeywordUntilEndOfTurn body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "48a01671-da1f-4875-bcd5-ca99566ff399",
		Name:         "Bastion Mastodon",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}: This creature gains vigilance until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"vigilance"}}},
			Cost:    effects.ManaCost("{W}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("vigilance", "Bastion Mastodon — vigilance until end of turn"),
		}},
	})
}
