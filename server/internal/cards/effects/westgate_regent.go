package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Westgate Regent — Creature — Vampire {3}{B}{B}, 4/4 (Adventures in the
// Forgotten Realms):
//
//	"Flying
//	 Ward—Discard a card.
//	 Whenever this creature deals combat damage to a player, put that many
//	 +1/+1 counters on it."
//
// Flying is a printed keyword, the ward is Ward(WardDiscard(1)), and the
// counters are as many as the combat damage dealt, read off the damage event;
// they land only while the Regent is still the permanent that triggered
// (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2df30a75-4957-4eab-8b0b-4c973f00642d",
		Name:            "Westgate Regent",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Ward(WardDiscard(1), "Westgate Regent — ward, discard a card"),
			WheneverThisDealsCombatDamageToAPlayer("Westgate Regent — put that many +1/+1 counters on it",
				Do(Instruction(func(ctx *Context) error {
					n := ctx.Trigger().Event.Amount
					if n <= 0 || !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
						return nil
					}
					return CounterOnThis{Kind: game.CounterPlusOne, N: n}.Apply(ctx)
				}))),
		},
	})
}
