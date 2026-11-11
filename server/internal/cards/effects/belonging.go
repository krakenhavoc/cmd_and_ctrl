package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0ec88e35-191b-411b-9eb0-2dfef7ae4174",
		Name:         "Belonging",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Belonging — create three 1/1 colorless Shapeshifter creature tokens with changeling",
				Do(CreateToken{Template: TokenCard("1/1 colorless Shapeshifter with changeling"), N: 3})),
		},
		Activated: []ActivatedAbility{Encore("{6}{W}{W}")},
	})
}
