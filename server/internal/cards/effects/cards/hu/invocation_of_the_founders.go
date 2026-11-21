package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Invocation of the Founders — Enchantment, the back face of Invasion of
// Arcavios (March of the Machine, S82 batch 23):
//
//	"Whenever you cast an instant or sorcery spell from your hand, you may
//	 copy that spell. You may choose new targets for the copy."
//
// The trigger watches casts from the hand only (CR 601.2a records the zone
// the spell was cast from), and the "may" is asked as the trigger resolves.
// The copy is made from last-known information if the spell has already
// left the stack (CR 608.2h) and its controller is asked for new targets
// (CR 707.10c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     effects.InvasionOfArcaviosOracleID + "#1",
		Name:         "Invocation of the Founders",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && (c.IsInstant() || c.IsSorcery()) && ev.OldZone == game.ZoneHand
			}, "Invocation of the Founders — you may copy that spell",
				func(g *game.Game, item *game.StackItem) error {
					spellID := effects.NewContext(g, item).Trigger().Event.CardID
					g.HoldSpellForEffect(spellID)
					return effects.MayChoice{
						Question: "Invocation of the Founders — copy that spell?",
						Then:     invocationOfTheFoundersCopy,
						Carry:    []uuid.UUID{spellID},
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// invocationOfTheFoundersCopy is the "yes": copy the spell, choosing new
// targets. Keyed; the spell rides on the prompt as its Carry.
var invocationOfTheFoundersCopy = effects.MayAnswer("confirm/invocation-of-the-founders-copy", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if !r.Accepted || len(r.Carry) == 0 {
		return nil
	}
	return effects.CopySpell{
		StackID:          r.Carry[0],
		Controller:       ctx.Controller(),
		FromLastKnown:    true,
		ChooseNewTargets: true,
	}.Apply(ctx)
})
