package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reflexive.go — the card-side vocabulary for CR 603.12 reflexive
// triggered abilities (#636). The engine half is
// server/internal/game/reflexive.go, which explains the rule and the
// design; this file is how a card says it.
//
// A reflexive trigger is the second sentence of "<do something>. When
// you do, <do something else>": a triggered ability created by the
// first sentence while it resolves. It is applied from INSIDE the
// parent's Effect, once the condition has actually been met, and it
// composes with Do(...) like any other primitive:
//
//	func riveteersOverlookSacrifice(g *game.Game, item *game.StackItem) error {
//	    ctx := NewContext(g, item)
//	    if err := (SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
//	        return err
//	    }
//	    // "When you do" — it was sacrificed, so the trigger happens.
//	    return WhenYouDo("Riveteers Overlook — fetch a basic tapped, gain 1 life",
//	        riveteersOverlookFetch).Apply(ctx)
//	}
//
// The three rules for using it:
//
//   - **Apply it only once the condition held.** "When you do" is
//     conditional on the doing: a land that was bounced in response
//     and could not be sacrificed never creates the trigger. Nothing
//     in the engine re-checks that — the `if` is the card's.
//   - **The Effect reads everything off the item**, exactly as a
//     harvested trigger's does. The trigger's own target is
//     item.Targets (chosen when it goes on the stack, CR 603.3d, not
//     when the parent resolved); what the parent needs to tell it
//     rides Cards / Players and comes back as ctx.Payload().
//     Capturing a *Card or a *Game in that closure is the same
//     mistake it always was.
//   - **The "you may" usually belongs to the PARENT.** In "you may
//     sacrifice another creature. When you do, …" the choice is the
//     parent ability's optional prompt; the reflexive half is
//     mandatory once you did. Set Optional only when the printed
//     reflexive sentence itself says "you may".

// ReflexiveTrigger creates a CR 603.12 reflexive triggered ability
// from a resolving effect. It goes on the stack at the next priority
// boundary — above the parent, which is the printed order — so every
// player gets a response window, and its target is chosen as it goes
// there rather than when the parent was announced.
//
// Applying it is a no-op outside a resolving stack item (an AsEnters
// context has none) and for a declaration with no Label or no Effect.
// It never returns an error: a trigger with no legal target is
// removed per CR 603.3d, which is the printed outcome, not a failure.
type ReflexiveTrigger struct {
	// Label is the stack-overlay copy, phrased like every other
	// trigger's: "<card> — <what happens>".
	Label string

	// Targets is the reflexive trigger's own target clause. The
	// legal set is computed when the trigger is created and the
	// controller picks from it then (CR 603.3d); an empty set drops
	// the trigger with no prompt. Nil for an untargeted follow-up.
	Targets *game.TargetSpec

	// Optional is the question for a reflexive trigger whose own
	// printed text says "you may". Empty means mandatory, which is
	// what "when you do" almost always is — see the file comment.
	Optional string

	// Cards is the payload: what the parent resolution has to tell
	// the trigger. The creatures that were tapped, the card that was
	// sacrificed, the cards that were revealed. Read back with
	// ctx.PayloadCards().
	Cards []uuid.UUID

	// Players is the same for player payloads — the player the
	// parent chose. Read back off ctx.Payload() by Kind.
	Players []uuid.UUID

	// Effect is what the trigger does on resolution. Same contract
	// as every other trigger's: read the controller, source, targets
	// and payload off the item, capture nothing.
	Effect Effect
}

func (r ReflexiveTrigger) Apply(ctx *Context) error {
	if ctx == nil || ctx.Item == nil {
		return nil
	}
	var optional *game.TriggerOptionalPrompt
	if r.Optional != "" {
		optional = &game.TriggerOptionalPrompt{Question: r.Optional}
	}
	payload := make([]game.TargetRef, 0, len(r.Cards)+len(r.Players))
	for _, id := range r.Cards {
		if id != uuid.Nil {
			payload = append(payload, game.TargetRef{Kind: game.TargetCard, ID: id})
		}
	}
	for _, id := range r.Players {
		if id != uuid.Nil {
			payload = append(payload, game.TargetRef{Kind: game.TargetPlayer, ID: id})
		}
	}
	ctx.Game.QueueReflexiveTriggerForEffect(ctx.Item, game.ReflexiveTrigger{
		Label:    r.Label,
		Targets:  r.Targets,
		Optional: optional,
		Payload:  payload,
		Effect:   r.Effect,
	})
	return nil
}

// WhenYouDo is the common shape: a mandatory, untargeted "when you
// do, <do X>" with nothing to carry over. Sugar for a
// ReflexiveTrigger with only Label and Effect set, so the card file
// reads like the card.
//
// Add a clause by taking the value and setting the field —
// `t := WhenYouDo(label, effect); t.Targets = TargetAny()` — or write
// the struct literal, which is what a card with a target clause or a
// payload should do anyway.
func WhenYouDo(label string, effect Effect) ReflexiveTrigger {
	return ReflexiveTrigger{Label: label, Effect: effect}
}
