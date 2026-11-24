package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Marath, Will of the Wild — Legendary Creature — Elemental Beast {R}{G}{W},
// 0/0 (Commander Legends):
//
//	"Marath enters with a number of +1/+1 counters on it equal to the amount of
//	 mana spent to cast it.
//	 {X}, Remove X +1/+1 counters from Marath: Choose one —
//	 • Put X +1/+1 counters on target creature. X can't be 0.
//	 • Marath deals X damage to any target. X can't be 0.
//	 • Create an X/X green Elemental creature token. X can't be 0."
//
// The entry clause reads the amount of mana spent (CR 601.2h) off the
// payment, as Dyadrine's does. The ability is Talon of Pain's {X} and
// "Remove X counters" with one announced X; "X can't be 0" bars an
// announcement of 0 (MinX), since each mode needs a nonzero X. The token's
// printed card is */*, so its template is a 0/0 and the mode sets X/X as it
// creates it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "fae87115-8749-4d25-a594-7139dd01a034",
		Name:           "Marath, Will of the Wild",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		XMatters:       true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{{
			Kind:  game.CounterPlusOne,
			Count: func(cast game.CastCounts) int { return cast.ManaSpent },
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}, Remove X +1/+1 counters from Marath: Choose one — put X +1/+1 counters on target creature; or Marath deals X damage to any target; or create an X/X green Elemental creature token. X can't be 0.",
			Cost:    effects.Plus(effects.ManaCost("{X}"), effects.RemoveCountersXFromThisSameX(game.CounterPlusOne), effects.MinX(1)),
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Modes: effects.ChooseOne(
				effects.ModeDoing("Put X +1/+1 counters on target creature.", effects.TargetCreature("target creature"),
					func(_ *game.StackItem, ctx *effects.Context, occ int) error {
						for _, t := range ctx.ModeTargets(occ) {
							return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: ctx.X()}.Apply(ctx)
						}
						return nil
					}),
				effects.ModeDoing("Marath deals X damage to any target.", effects.TargetAny(),
					func(_ *game.StackItem, ctx *effects.Context, occ int) error {
						for _, t := range ctx.ModeTargets(occ) {
							return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: ctx.X()}.Apply(ctx)
						}
						return nil
					}),
				effects.ModeDoing("Create an X/X green Elemental creature token.", nil,
					func(_ *game.StackItem, ctx *effects.Context, _ int) error {
						x := ctx.X()
						tok := ctx.TextToken(effects.TokenCard("0/0 green Elemental"))
						tok.Power, tok.Toughness = x, x
						return ctx.Game.CreateTokenForEffect(ctx.Controller(), tok, 1)
					}),
			),
		}},
	})
}
