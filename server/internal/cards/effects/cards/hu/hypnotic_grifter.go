package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hypnotic Grifter — Creature — Human Rogue {U}, 1/2 (Streets of New
// Capenna, #5491):
//
//	"{3}: This creature connives. (Draw a card, then discard a card. If you
//	 discarded a nonland card, put a +1/+1 counter on this creature.)"
//
// An activated ability with no timing limit, declared as a pump that also
// draws (its counter comes only for a nonland discard); the connive is connive.go's
// ConniveEach on the Grifter (sourceConnives).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5df80355-b12a-4989-94e5-1e0fa27350e8",
		Name:         "Hypnotic Grifter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{3}: This creature connives.",
			Cost:  game.AbilityCost{Mana: "{3}"},
			// Instant speed: the +1/+1 counter can save the Grifter from
			// damage or a -1/-1 effect, so it answers as a pump (ADR 0142).
			Purpose: game.Purpose{Answers: game.AnswerPump, Draws: 1},
			Effect:  effects.ConniveGrantedCreature,
		}},
	})
}
