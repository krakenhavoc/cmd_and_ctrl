package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guild Thief — Creature — Orc Rogue {1}{U}, 1/1 (Adventures in the
// Forgotten Realms):
//
//	"Whenever this creature deals combat damage to a player, put a +1/+1
//	 counter on it.
//	 Cunning Action — {3}{U}: This creature can't be blocked this turn."
//
// The counter goes on through the counter pipeline (CounterOnThis). The
// evasion lasts only while the Thief is still the permanent that was
// activated: one that left and came back is a new object (CR 400.7).
// "Cunning Action" has no rules meaning (CR 207.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bf061d7e-46eb-4037-968c-3cfc5ca94552",
		Name:         "Guild Thief",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Guild Thief — put a +1/+1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Cunning Action — {3}{U}: This creature can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{3}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
					Label: "Guild Thief — can't be blocked"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
