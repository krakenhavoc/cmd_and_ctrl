package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s47_helpers.go — batch 47 (token copies, token replacement and
// gifts). Card-side helpers only; every name starts cmdr2S47.

// cmdr2S47ExileThenCopyTrigger is "Whenever another nontoken creature you
// control dies, you may exile it. If you do, create a token that's a copy
// of that creature, except …" — Nightmare Shepherd and Brenard, Ginger
// Sculptor, Hofri Ghostforge's shape with the "may" of its exile. `tailKey`
// names the "if you do" tail (an on-disk identity, namespaced by card) and
// `except` is the card's "except" clause on the copiable values.
func cmdr2S47ExileThenCopyTrigger(label, question, tailKey string, except func(*game.Card)) game.TriggeredAbility {
	tail := CardTail(tailKey, func(ctx *Context, o game.TailOutcome) error {
		if !oneWentThisWay(o) {
			return nil
		}
		obj := ctx.Trigger().Object
		if obj == nil {
			return nil
		}
		tmpl, ok := ctx.Game.TokenCopyTemplateForObjectForEffect(obj.Ref())
		if !ok {
			return nil
		}
		except(&tmpl)
		return createTokenCopies(ctx, ctx.Controller(), tmpl, 1)
	})
	return Optional(On(game.EventLTB, hofriGhostforgeAnotherNontokenCreatureDied, label,
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			dead := item.Trigger.Event.CardID
			if !diedCardStillInGraveyard(ctx, dead) {
				return nil
			}
			return ExileTarget{Target: dead, ThenTail: ctx.ItemTail(tail)}.Apply(ctx)
		}), question)
}

var (
	// "Whenever a nontoken creature an opponent controls enters this turn"
	// (Theoretical Duplication): the creature an EventETB names, while it
	// is a nontoken creature an opponent of the delayed trigger's
	// controller controls.
	cmdr2S47OpponentNontokenCreatureEntersCondition = game.DelayedEachCondition(
		"etb/each-nontoken-creature-an-opponent-controls", cmdr2S47OpponentNontokenCreatureEntered)

	// "Create a token that's a copy of that creature" (Theoretical
	// Duplication): the payload's creature, copied while it is still that
	// object on the battlefield.
	cmdr2S47CopyThePayloadBody = game.SimpleDelayedBody("copy/token-copy-of-the-payload", cmdr2S47CopyThePayload)
)

func cmdr2S47OpponentNontokenCreatureEntered(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) []uuid.UUID {
	if ev.Kind != game.EventETB || ev.CardID == uuid.Nil {
		return nil
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !c.IsCreature() || c.IsToken() || c.Controller == dt.Controller {
		return nil
	}
	if !eachOpponent(g, dt.Controller, func(opp uuid.UUID) bool { return opp == c.Controller }) {
		return nil
	}
	return []uuid.UUID{ev.CardID}
}

func cmdr2S47CopyThePayload(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	obj, ok := payloadObject(ctx)
	if !ok {
		return nil
	}
	c, ok := battlefieldCardForEffect(g, obj.ID)
	if !ok || c.ObjectEpoch != obj.Epoch {
		return nil
	}
	return CreateTokenCopy{Controller: item.Controller, Copy: obj.ID, N: 1}.Apply(ctx)
}
