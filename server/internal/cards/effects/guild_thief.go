package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "bf061d7e-46eb-4037-968c-3cfc5ca94552",
		Name:         "Guild Thief",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Guild Thief — put a +1/+1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					return CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Cunning Action — {3}{U}: This creature can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{3}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
					Label: "Guild Thief — can't be blocked"}.Apply(NewContext(g, item))
			},
		}},
	})
}
