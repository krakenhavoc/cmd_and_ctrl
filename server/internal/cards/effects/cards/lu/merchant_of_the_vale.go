package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Merchant of the Vale // Haggle — Creature — Human Peasant {2}{R}, 2/3 //
// Instant — Adventure {R} (Throne of Eldraine, #5516):
//
//	Merchant of the Vale — "{2}{R}, Discard a card: Draw a card."
//	Haggle — "You may discard a card. If you do, draw a card."
//
// Both faces register: the creature under the oracle id and the Adventure
// under "<oracle_id>#1". The discard of the creature's ability is a cost
// (CR 118.8, Rummaging Goblin's); Haggle is Keldon Raider's optional discard
// whose draw follows only from it. The Adventure lifecycle is the engine's.
//
// No simplification.
const merchantOfTheValeOracleID = "f88a096e-7497-4297-ab65-f18769b1cc47"

func init() {
	effects.Register(effects.Spec{
		OracleID:     merchantOfTheValeOracleID,
		Name:         "Merchant of the Vale",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, Discard a card: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.DiscardACard()),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
	const label = "Haggle — you may discard a card; if you do, draw a card"
	effects.Register(effects.Spec{
		OracleID:     merchantOfTheValeOracleID + "#1",
		Name:         "Haggle",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B39MayDiscardThenDraw(1, false, label, effects.DrawAsManyAsDiscarded)(ctx)
		},
	})
}
