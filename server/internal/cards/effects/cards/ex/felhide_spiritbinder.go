package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Felhide Spiritbinder — {3}{R} Creature — Minotaur Shaman 3/4 (Born of the Gods):
//
//	"Inspired — Whenever this creature becomes untapped, you may pay {1}{R}. If you do,
//	 create a token that's a copy of another target creature, except it's an enchantment
//	 in addition to its other types. It gains haste. Exile it at the beginning of the next
//	 end step."
//
// The target is chosen as the trigger goes on the stack; the payment is asked as it
// resolves. The enchantment type is part of the copy (CR 707.9b); the haste is a separate
// gain with no duration, and the exile is one delayed trigger for the next end step, as
// Molten Duplication's sacrifice is.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91e3f057-c24f-42b5-8525-74b44bdefc83",
		Name:         "Felhide Spiritbinder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventUntapCard, effects.Self, "Felhide Spiritbinder — you may pay {1}{R}; if you do, create a token copy of another target creature that is also an enchantment, with haste, and exile it at the next end step (inspired)",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}{R}",
						Question: "Felhide Spiritbinder — pay {1}{R} to create a token copy of the target creature?",
						Then:     felhideSpiritbinderPaid,
					}.Apply(effects.NewContext(g, item))
				}),
				effects.Another(effects.TargetCreature("another target creature"))),
		},
	})
}

// felhideSpiritbinderPaid is the paid branch, keyed: an on-disk identity, never renamed,
// never reused.
var felhideSpiritbinderPaid = effects.PayUnlessBranches("pay-unless/felhide-spiritbinder-copy",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		id, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return effects.CreateTokenCopiesTogether{
			Controller: ctx.Controller(),
			Copies:     []uuid.UUID{id},
			Except:     effects.TokenCopyIsAlsoAnEnchantment,
			Then:       ctx.ItemTail(felhideSpiritbinderRest),
		}.Apply(ctx)
	}, nil)

// felhideSpiritbinderRest is "It gains haste. Exile it at the beginning of the next end
// step." as a keyed tail, told the tokens the creation made. An on-disk key.
var felhideSpiritbinderRest = effects.CardTail("felhide-spiritbinder/haste-and-exile-copy", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	for _, id := range o.IDs {
		if err := (effects.ScopedEffectFor{Target: id, Mods: []game.Mod{game.AddKeywordsMod("haste")},
			Duration: game.IndefiniteDuration(), Label: "Felhide Spiritbinder — the token gains haste"}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: ctx.Item.Controller,
		Label:      "Felhide Spiritbinder — exile the token",
		Cards:      o.IDs,
		Body:       effects.ExileListedCardsBody,
	}.Apply(ctx)
})
