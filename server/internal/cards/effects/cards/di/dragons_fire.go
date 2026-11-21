package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragon's Fire — Instant {1}{R} (Adventures in the Forgotten Realms):
//
//	"As an additional cost to cast this spell, you may reveal a Dragon card
//	 from your hand or choose a Dragon you control.
//	 Dragon's Fire deals 3 damage to target creature or planeswalker. If you
//	 revealed a Dragon card or chose a Dragon as you cast this spell, Dragon's
//	 Fire deals damage equal to the power of that card or creature instead."
//
// The optional cost is the optional behold (a Dragon you control or a Dragon
// card in hand, CR 701.4) with the pick recorded on the payment. The
// damage is the power of the object it named, read as the spell resolves
// (CR 608.2h): a Dragon still on the battlefield counts with its current
// power, one that has left counts as it last existed, and a revealed card with
// its printed power. Declined, it deals 3. The spell is the damage source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "7237d607-2613-4786-96d6-55792c2e02b9",
		Name:          "Dragon's Fire",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove}, // damages a creature or planeswalker
		OptionalCosts: []game.AdditionalCost{effects.AfrB18OptionalBeholdChosen("a", "Dragon")},
		Targets:       effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			n := 3
			if ctx.WasBeheld() {
				if p, ok := effects.B25ChosenPower(ctx); ok {
					n = p
				}
			}
			return effects.DealDamage{Source: ctx.Source(), Target: target, Amount: n}.Apply(ctx)
		},
	})
}
