package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Primo, the Unbounded — Legendary Creature — Fractal Wolf {X}{G}{G}{U},
// 0/0 (Secrets of Strixhaven Commander, #4834):
//
//	"Trample
//	 Primo enters with twice X +1/+1 counters on it.
//	 Whenever one or more creatures you control with base power 0 deal combat
//	 damage to a player, create a 0/0 green and blue Fractal creature token.
//	 Put a number of +1/+1 counters on it equal to the damage dealt."
//
// The counters are a CR 614.1c entry clause over the announced X (Benevolent
// Hydra's XCounters, doubled), so a cast for X=0 is a 0/0 that dies as a
// state-based action. "Base power 0" is the power after the layer 7b step,
// without pumps or counters (CR 208.4b, Card.BasePower): a Fractal with
// counters or Primo himself with his counters still has base power 0 and
// counts. The trigger is the one-per-player batch of CR 603.2c, so two such
// creatures hitting one player make one Fractal, and the Fractal's counters
// are the damage those creatures dealt that player in that combat damage step
// (socB10BasePowerZeroDamage). The Fractal enters with its counters (CR 122.6),
// so it never exists as a 0/0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "65fb0e20-61f8-4c61-bc08-7f6f718ee9ee",
		Name:            "Primo, the Unbounded",
		XMatters:        true,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{{
			Kind:  game.CounterPlusOne,
			Count: func(cast game.CastCounts) int { return 2 * cast.X },
		}},
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(
				socB10BasePowerZero,
				"Primo, the Unbounded — create a Fractal with that many +1/+1 counters",
				Do(Instruction(func(ctx *Context) error {
					n := socB10BasePowerZeroDamage(ctx)
					_, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(TokenCard("0/0 green and blue Fractal")), 1,
						game.TokenEntryOptions{Counters: map[string]int{game.CounterPlusOne: n}})
					return err
				}))),
		},
	})
}
