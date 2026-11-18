package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Myojin of Blooming Dawn — Legendary Creature — Spirit {5}{W}{W}{W},
// 4/6 (Kamigawa: Neon Dynasty Commander, #5493):
//
//	"Myojin of Blooming Dawn enters with an indestructible counter on it
//	 if you cast it from your hand.
//	 Remove an indestructible counter from Myojin of Blooming Dawn: Create
//	 a 1/1 colorless Spirit creature token for each permanent you
//	 control."
//
// The counter is Myojin of Night's Reach's entry clause. The permanents
// are counted as the ability resolves (CR 608.2h), the Myojin among them
// while it is still on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "bfb75b28-0e51-4829-b104-f1679e63c126",
		Name:                       "Myojin of Blooming Dawn",
		Completeness:               CompletenessFull,
		EntersWithCountersFromCast: neoMyojinEntry(),
		Activated: []ActivatedAbility{{
			Label:   "Remove an indestructible counter from Myojin of Blooming Dawn: Create a 1/1 colorless Spirit creature token for each permanent you control.",
			Cost:    neoRemoveIndestructible(),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Effect: func(g *game.Game, item *game.StackItem) error {
				n := countControlled(g, item.Controller, func(game.Card) bool { return true })
				return CreateToken{Template: TokenCard("1/1 colorless Spirit"), N: n}.Apply(NewContext(g, item))
			},
		}},
	})
}
