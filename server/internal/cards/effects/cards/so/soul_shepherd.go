package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soul Shepherd — Creature — Human Cleric {1}{W}, 2/1 (Weatherlight):
//
//	"{W}, Exile a creature card from your graveyard: You gain 1 life."
//
// Any creature card of the activator's graveyard pays the cost; the life is
// gained once the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78dbcc18-291c-4b2c-9576-f8dc93914936",
		Name:         "Soul Shepherd",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, Exile a creature card from your graveyard: You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.ExileFromGraveyard(1, "a creature card", effects.MatchCreature)),
			Effect:  effects.GainOneLifeEffect,
		}},
	})
}
