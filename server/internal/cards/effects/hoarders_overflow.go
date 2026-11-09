package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hoarder's Overflow — Enchantment {1}{R} (Bloomburrow, #4467):
//
//	"When this enchantment enters and whenever you expend 4, put a stash
//	 counter on it.
//	 {1}{R}, Sacrifice this enchantment: Discard your hand, then draw cards
//	 equal to the number of stash counters on this enchantment."
//
// The enters half and the expend half are two registered triggers over one
// body. The count is read as the ability resolves from the sacrificed
// enchantment's last-known counters (CR 608.2h), the sacrifice being a cost.
//
// No simplification.
func init() {
	const putStash = "Hoarder's Overflow — put a stash counter on it"
	put := Do(CounterOnThis{Kind: "stash", N: 1})
	Register(Spec{
		OracleID:       "75f7d9c7-5d52-4bfe-a730-62fef187b8b8",
		Name:           "Hoarder's Overflow",
		Completeness:   CompletenessFull,
		ReadsManaSpent: game.ReadsYourCasts,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(putStash, put),
			WheneverYouExpend(4, putStash, put),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}, Sacrifice this enchantment: Discard your hand, then draw cards equal to the number of stash counters on it.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}{R}"), SacrificeThis()),
			Effect: Do(
				Instruction(func(ctx *Context) error {
					n := CountersOnThisAsItResolves("stash")(ctx)
					if _, err := discardWholeHand(ctx.Game, ctx.Controller()); err != nil {
						return err
					}
					return DrawCards{N: n}.Apply(ctx)
				}),
			),
		}},
	})
}
