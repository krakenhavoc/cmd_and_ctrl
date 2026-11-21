package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Belonging — Creature — Elemental Incarnation {5}{W}, 6/6 (Lorwyn
// Eclipsed Commander):
//
//	"When this creature enters, create three 1/1 colorless Shapeshifter
//	 creature tokens with changeling.
//	 Encore {6}{W}{W} (...)"
//
// Three of the shared changeling Shapeshifter token. Encore is the shared
// keyword (encore.go); each token copy makes three more.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ec88e35-191b-411b-9eb0-2dfef7ae4174",
		Name:         "Belonging",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Belonging — create three 1/1 colorless Shapeshifter creature tokens with changeling",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Shapeshifter with changeling"), N: 3})),
		},
		Activated: []effects.ActivatedAbility{effects.Encore("{6}{W}{W}")},
	})
}
