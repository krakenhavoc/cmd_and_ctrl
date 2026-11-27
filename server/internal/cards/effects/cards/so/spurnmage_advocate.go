package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spurnmage Advocate — Creature — Human Nomad {W}, 1/1 (Commander 2011, S82
// older sets batch 453, #5574):
//
//	"{T}: Return two target cards from an opponent's graveyard to their
//	 hand. Destroy target attacking creature."
//
// Deadeye Tracker's "two target cards from an opponent's graveyard" (both from
// one graveyard) plus a second clause for the attacking creature; the cards go
// to their owner's hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0477b923-9f32-4e7f-ac08-c715cca819a4",
		Name:         "Spurnmage Advocate",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Return two target cards from an opponent's graveyard to their hand. Destroy target attacking creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.TapCost(),
			Targets: effects.Clauses(
				effects.TargetCardInGraveyard("two target cards from an opponent's graveyard", effects.Not(effects.YouOwn())).
					WithCount(2, 2).AllShare(effects.FromASingleGraveyard()),
				effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
			),
			Effect: effects.S82OldB453SpurnmageAdvocate,
		}},
	})
}
