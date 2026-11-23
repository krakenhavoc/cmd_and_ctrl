package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gorbag of Minas Morgul — Legendary Creature — Orc Soldier {1}{B}, 2/2 (The
// Lord of the Rings: Tales of Middle-earth):
//
//	"Whenever a Goblin or Orc you control deals combat damage to a player,
//	 you may sacrifice it. When you do, choose one —
//	 • Draw a card.
//	 • Create a Treasure token."
//
// One trigger for each Goblin or Orc that connects (CR 603.2c), the creature
// the event's source. The "you may sacrifice it" is a confirmation, the
// sacrifice is the creature itself, and "when you do" is a CR 603.12
// reflexive trigger created only if it really was sacrificed
// (SacrificePermanent's tail), whose mode is chosen as it goes on the stack
// (ReflexiveModalBody, CR 603.3c). The creature types are read through the
// source's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e77e97d7-58ff-4454-ba49-66182e550f12",
		Name:         "Gorbag of Minas Morgul",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.CombatDamageToPlayerBy(ev, source.Controller, g) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.Source)
				return ok && (c.HasSubtype(source.TextSubtype("Goblin")) || c.HasSubtype(source.TextSubtype("Orc")))
			}, "Gorbag of Minas Morgul — you may sacrifice it; when you do, draw a card or create a Treasure token",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Gorbag of Minas Morgul — sacrifice it?",
						Then:     gorbagSacrificeAnswer,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var gorbagSacrificeAnswer = effects.MayYes("confirm/gorbag-sacrifice", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	if ctx.Item == nil || ctx.Item.Trigger == nil {
		return nil
	}
	target := ctx.Item.Trigger.Event.Source
	if !effects.OnBattlefield(ctx.Game, target) {
		return nil
	}
	return effects.SacrificePermanent{Target: target, ThenTail: ctx.Tail(gorbagSacrificed)}.Apply(ctx)
})

// gorbagSacrificed is "when you do": only a creature that really was
// sacrificed. An on-disk key: never renamed, never reused.
var gorbagSacrificed = effects.CardTail("then/gorbag-sacrificed", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.ReflexiveTrigger{Label: "Gorbag of Minas Morgul — choose one", Body: gorbagModes}.Apply(ctx)
})

var gorbagModes = game.ReflexiveModalBody("gorbag-of-minas-morgul/modes", func(_ uuid.UUID, _ game.EffectParams) *game.ModeSpec {
	return effects.ChooseOne(
		effects.ModeDoing("Draw a card.", nil, func(_ *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.DrawCards{N: 1}.Apply(ctx)
		}),
		effects.ModeDoing("Create a Treasure token.", nil, func(_ *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(ctx)
		}),
	)
})
