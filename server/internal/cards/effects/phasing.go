package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// phasing.go — the card-facing half of CR 702.26 (#1199, ADR 0084).
//
// Three primitives, one per printed shape:
//
//	"Target creature phases out."                      PhaseOut
//	"… phases out until this enchantment leaves
//	 the battlefield."                                 PhaseOutUntilLeaves
//	"It can't phase in for as long as you control
//	 this Saga."                                       CantPhaseInWhileYouControl
//
// Everything a card author might expect to have to say is the
// engine's job and is said nowhere here: the Auras and Equipment
// attached to the target come with it (CR 702.26g), the permanent
// keeps its counters, damage and tapped state (CR 702.26d), it is
// removed from combat (CR 506.4), nothing enters or leaves the
// battlefield so no trigger fires (CR 702.26d), and it comes back
// during its controller's untap step (CR 502.1) with no delayed
// trigger to schedule. See game/phasing.go.
//
// The KEYWORD needs nothing here at all. "Phasing" is a canonical
// keyword the deck importer stamps (game/keywords.go), so a printed
// phasing permanent phases in and out on its own with no Spec, and a
// card that GRANTS it — Shimmer's "each land of the chosen type has
// phasing", an Aura's "enchanted permanent has phasing" — is an
// ordinary layer-6 keyword grant.

// PhaseOut is "[those permanents] phase out" (CR 702.26) — Teferi's
// Protection's "all permanents you control phase out", Vodalian
// Illusionist's "target creature phases out", Clever Concealment's
// "any number of target nonland permanents you control phase out".
//
// Every named permanent phases out SIMULTANEOUSLY, together with
// everything attached to any of them; a name that is not on the
// battlefield right now is skipped, which is CR 608.2b's per-slot
// existence re-check.
type PhaseOut struct {
	// Targets are the permanents that phase out directly. Pass the
	// whole list in one call rather than calling once per card: the
	// rule is simultaneous, and a card attached to another card in the
	// same list must not be dragged out twice.
	Targets []uuid.UUID
}

func (p PhaseOut) Apply(ctx *Context) error {
	targets := ctx.withoutNewSourceObject(p.Targets) // #1432
	if len(targets) == 0 {
		return nil
	}
	return ctx.Game.PhaseOutForEffect(ctx.Source(), targets...)
}

// PhaseOutUntilLeaves is "[that permanent] phases out until ~ leaves
// the battlefield" — Oubliette, Out of Time.
//
// The permanent does NOT come back at its controller's untap step. It
// comes back the moment `Until` stops being on the battlefield, which
// is what the cards do: Out of Time's whole point is that every
// creature returns at once when the last time counter comes off,
// mid-turn.
//
// TapOnPhaseIn is Oubliette's "Tap that creature as it phases in this
// way", which is what stops the creature blocking the turn it comes
// back.
type PhaseOutUntilLeaves struct {
	Targets      []uuid.UUID
	Until        uuid.UUID
	TapOnPhaseIn bool
}

func (p PhaseOutUntilLeaves) Apply(ctx *Context) error {
	// #1432: "until THIS leaves the battlefield" with a source that
	// already left and came back is over before it starts (CR 611.2b)
	// — the new object is not the one the duration names.
	targets := ctx.withoutNewSourceObject(p.Targets)
	if len(targets) == 0 || p.Until == uuid.Nil || ctx.isNewSourceObjectAsThis(p.Until) {
		return nil
	}
	return ctx.Game.PhaseOutUntilLeavesForEffect(
		ctx.Source(), p.Until, p.TapOnPhaseIn, targets...)
}

// CantPhaseInWhileYouControl is "It can't phase in for as long as you
// control this Saga" (The Phasing of Zhalfir, #5664), said after the
// PhaseOut that put Targets out.
//
// A prohibition, not a clock: while the resolving ability's controller
// controls Source the permanents sit out every untap step, and once
// that stops (Source leaves, phases out or changes controller) they
// phase in during the next untap step of the player they phased out
// under (CR 502.1), not at once. If that player does not control Source
// as this applies, the duration never starts (CR 611.2b) and the
// permanents are on the ordinary clock.
type CantPhaseInWhileYouControl struct {
	Targets []uuid.UUID
	Source  uuid.UUID
}

func (p CantPhaseInWhileYouControl) Apply(ctx *Context) error {
	// A Source that left and came back is a new object, which is not
	// the one the sentence names (CR 400.7).
	if len(p.Targets) == 0 || p.Source == uuid.Nil || ctx.isNewSourceObjectAsThis(p.Source) {
		return nil
	}
	ctx.Game.CantPhaseInWhileControlledForEffect(p.Source, ctx.Controller(), p.Targets...)
	return nil
}

// legalTargetCards is the list of this item's card targets that are
// still legal at resolution (CR 608.2b, checked per slot), in
// announce order.
//
// A list rather than b16FirstLegalTargetCard's single answer, because
// every phasing card but one names more than one permanent and the
// rule is simultaneous: collecting first and phasing once is what
// keeps an Equipment named beside the creature it is attached to from
// being pulled out twice.
func legalTargetCards(item *game.StackItem, g *game.Game) []uuid.UUID {
	var out []uuid.UUID
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || !g.TargetStillLegalForEffect(item, t) {
			continue
		}
		out = append(out, t.ID)
	}
	return out
}
