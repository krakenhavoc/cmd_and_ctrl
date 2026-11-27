package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rayne, Academy Chancellor — Legendary Creature — Human Wizard {2}{U}, 1/1
// (Urza's Destiny, S82 #5621):
//
//	"Whenever you or a permanent you control becomes the target of a spell or
//	 ability an opponent controls, you may draw a card. You may draw an
//	 additional card if Rayne is enchanted."
//
// Altanak's watch on EventBecomesTarget, widened to the controller (a player
// target) and to every permanent they control. The engine emits one event per
// target slot, so a spell that targets the same object twice would emit two;
// the check lets only the first of them trigger (CR 115.3: it becomes the
// target once). Two different objects of yours targeted by one spell trigger
// twice, as the ruling says.
//
// "You may draw a card" is asked as the trigger resolves; Rayne being enchanted
// is read then (CR 303.4), and only then is the second, additional, draw
// offered.
//
// No simplification.
func init() {
	const label = "Rayne, Academy Chancellor — you may draw a card; you may draw an additional card if Rayne is enchanted"
	effects.Register(effects.Spec{
		OracleID:     "6933a7bf-10ed-46c8-a49c-ce200bce3eea",
		Name:         "Rayne, Academy Chancellor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, rayneWatches, label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				word := ""
				if effects.IsEnchantedByAura(g, item.SourceCardID) {
					word = "enchanted"
				}
				return effects.MayChoice{
					Question: "Rayne, Academy Chancellor — draw a card?",
					Draws:    1,
					YesLabel: "Draw",
					NoLabel:  "Decline",
					Word:     word,
					Then:     rayneFirstDraw,
				}.Apply(ctx)
			}),
		},
	})
}

var rayneFirstDraw = effects.MayAnswer("s82-old-b705/rayne-first-draw", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if r.Accepted {
		if err := (effects.DrawCards{Player: ctx.Controller(), N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	if r.Word != "enchanted" {
		return nil
	}
	return effects.MayChoice{
		Question: "Rayne, Academy Chancellor — draw an additional card?",
		Draws:    1,
		YesLabel: "Draw",
		NoLabel:  "Decline",
		Then:     rayneAdditionalDraw,
	}.Apply(ctx)
})

var rayneAdditionalDraw = effects.MayYes("s82-old-b705/rayne-additional-draw", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
})

// rayneWatches is "you or a permanent you control becomes the target of a
// spell or ability an opponent controls", once per object per spell or ability.
func rayneWatches(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || source == nil {
		return false
	}
	if ev.Actor == uuid.Nil || ev.Actor == source.Controller || g.PlayerByIDForEffect(ev.Actor) == nil {
		return false
	}
	if ev.CardID == uuid.Nil {
		if ev.Target != source.Controller {
			return false
		}
	} else if !effects.B40ControlsA(g, source.Controller, func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID == ev.CardID }) {
		return false
	}
	for i := len(g.Events) - 1; i >= 0; i-- {
		earlier := g.Events[i]
		if earlier.Seq >= ev.Seq {
			continue
		}
		switch earlier.Kind {
		case game.EventBecomesTarget:
			if earlier.StackItemID == ev.StackItemID && earlier.Target == ev.Target {
				return false
			}
		case game.EventTrigger, game.EventTriggerPlaced:
			continue
		default:
			return true
		}
	}
	return true
}
