package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	put := effects.Do(effects.CounterOnThis{Kind: "stash", N: 1})
	effects.Register(effects.Spec{
		OracleID:       "75f7d9c7-5d52-4bfe-a730-62fef187b8b8",
		Name:           "Hoarder's Overflow",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsYourCasts,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(putStash, put),
			effects.WheneverYouExpend(4, putStash, put),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, Sacrifice this enchantment: Discard your hand, then draw cards equal to the number of stash counters on it.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.SacrificeThis()),
			Effect: effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					n := effects.CountersOnThisAsItResolves("stash")(ctx)
					if _, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller()); err != nil {
						return err
					}
					return effects.DrawCards{N: n}.Apply(ctx)
				}),
			),
		}},
	})
}
