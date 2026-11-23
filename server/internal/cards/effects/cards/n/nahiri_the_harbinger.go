package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nahiri, the Harbinger — Legendary Planeswalker — Nahiri {2}{R}{W},
// loyalty 4 (Commander Masters, #5948):
//
//	"+2: You may discard a card. If you do, draw a card.
//	 −2: Exile target enchantment, tapped artifact, or tapped creature.
//	 −8: Search your library for an artifact or creature card, put it onto
//	 the battlefield, then shuffle. It gains haste. Return it to your hand at
//	 the beginning of the next end step."
//
// The −2's tapped clauses are part of the target restriction (CR 115.1), so
// an untapped artifact or creature is not a legal target. The −8's haste has
// no duration (CR 611.2a lasts as long as it stays); the return is a delayed
// trigger tied to the object that entered (CR 603.7), so a card that has left
// by the end step is not returned from wherever it went (CR 400.7).
//
// No simplification.
func init() {
	tapped := func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Tapped }
	effects.Register(effects.Spec{
		OracleID:        "d492eb9a-1f3b-4bf2-9a5d-0f4a4433f33f",
		Name:            "Nahiri, the Harbinger",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+2: You may discard a card. If you do, draw a card.",
				Cost:  effects.LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.B39MayDiscardThenDraw(1, false, "Nahiri, the Harbinger — you may discard a card; if you do, draw a card",
						effects.DrawOneIfDiscarded)(effects.NewContext(g, item))
				},
			},
			{
				Label: "−2: Exile target enchantment, tapped artifact, or tapped creature.",
				Cost:  effects.LoyaltyCost(-2),
				Targets: effects.TargetPermanent("target enchantment, tapped artifact, or tapped creature",
					effects.Or(effects.Enchantment(), effects.And(effects.Artifact(), tapped), effects.And(effects.Creature(), tapped))),
				Effect: effects.ExileFirstLegalCardTarget,
			},
			{
				Label: "−8: Search your library for an artifact or creature card, put it onto the battlefield, then shuffle. It gains haste. Return it to your hand at the beginning of the next end step.",
				Cost:  effects.LoyaltyCost(-8),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.SearchLibrary{
						Player:    ctx.Controller(),
						Predicate: func(c game.Card) bool { return c.IsCreature() || c.HasCardType("artifact") },
						Dest:      game.ZoneBattlefield,
						Limit:     1,
						Shuffle:   true,
						Reason:    "Nahiri, the Harbinger — search for an artifact or creature card",
						Source:    ctx.Source(),
						ThenTail:  ctx.ItemTail(nahiriTheHarbingerFetched),
					}.Apply(ctx)
				})),
			},
		},
	})
}

// nahiriTheHarbingerFetched is the rest of the −8, told the card that entered:
// haste with no duration, and the delayed return to hand. An on-disk key:
// never renamed, never reused.
var nahiriTheHarbingerFetched = effects.CardTail("then/nahiri-the-harbinger-fetched", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := effects.BattlefieldCardForEffect(ctx.Game, id)
		if !ok {
			continue
		}
		if err := (effects.ScopedEffectFor{
			Target:   id,
			Mods:     []game.Mod{game.AddKeywordsMod("haste")},
			Duration: game.IndefiniteDuration(),
			Label:    "Nahiri, the Harbinger — haste",
		}).Apply(ctx); err != nil {
			return err
		}
		if err := (effects.ScheduleDelayedTrigger{
			At:     game.StepEnd,
			Label:  "Nahiri, the Harbinger — return " + c.Name + " to your hand",
			Cards:  []uuid.UUID{id},
			Body:   effects.ZaraReturnToHandBody,
			Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
