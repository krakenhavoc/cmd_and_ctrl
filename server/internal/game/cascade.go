package game

import (
	"github.com/google/uuid"
)

// cascade.go — S28: CR 702.85.
//
//	"Cascade (When you cast this spell, exile cards from the top of
//	 your library until you exile a nonland card that costs less. You
//	 may cast it without paying its mana cost. Put the exiled cards on
//	 the bottom of your library in a random order.)"
//
// Three engine facts the keyword needs and the engine did not have:
//
//  1. **A triggered ability of a SPELL.** Cascade triggers when the
//     spell is cast, which is to say while it is on the stack and
//     nowhere near the battlefield the trigger harvester scans. That
//     is why TriggeredAbility grew a FromStack flag and the harvester
//     grew a narrow cast-only stack scan (triggers.go) rather than a
//     blanket one: a permanent's ETB-watching trigger must not start
//     firing while its own spell is still on the stack.
//
//  2. **Reproducible randomness.** "In a random order" has to come
//     out the same after an undo or a restore, so the order draws from
//     the game's keyed RNG (rng.go, ADR 0054) on the owner's
//     "random_order" stream. Reaching for math/rand directly would
//     make it neither rewindable nor persistable.
//
//  3. **A yes/no during a resolution.** "You MAY cast it" is a
//     decision taken while the trigger is resolving, which is the
//     PendingChoice machinery's job (PendingChoiceMayCast).
//
// SANDBOX SIMPLIFICATION — the free cast is a GRANT, not an inline
// cast. Printed cascade casts the card then and there, as part of the
// trigger's resolution, ignoring timing. Here, answering "yes" stamps
// a free-cast permission on the exiled card and the player casts it
// with an ordinary cast_spell action, which means it goes on the
// stack after the trigger has finished resolving rather than on top
// of it, and it obeys ordinary timing.
//
// That is weaker than printed in the case that matters (a cascaded
// sorcery during an opponent's turn simply cannot be cast) and the
// alternative is worse: an inline cast would have to collect targets,
// modes and X from inside a resolution, and the announce path has no
// frame for a half-validated cast — see CastSpellParams.Face on why
// PendingChoice is deliberately not that frame.
//
// The one place the grant could have been STRONGER than printed is
// closed: a card left uncast goes to the bottom of its owner's
// library at the beginning of the next end step, via a delayed
// trigger. Without that, declining to cast would leave the card
// parked in exile — permanently available, which cascade never is.

// CascadeForEffect performs one cascade for `controller` off a spell
// with mana value `lessThan`, per CR 702.85a. `source` is the
// cascading spell, for attribution on the prompt and the event log.
//
// Returns nil in every "nothing happened" case — an empty library, a
// library with no qualifying card — because cascade does as much as
// it can (CR 608.2c) and a deck with nothing cheap in it is a legal
// board state, not an error.
//
// Caller must hold g.mu (this runs from a resolving trigger's
// Effect).
func (g *Game) CascadeForEffect(controller, source uuid.UUID, lessThan int) error {
	p := g.playerByIDLocked(controller)
	if p == nil || p.Library == nil || g.Exile == nil {
		return nil
	}
	var pile []uuid.UUID
	hit := uuid.Nil
	for p.Library.Size() > 0 {
		// The library's top is the LAST element — PopTop, every draw
		// and every mill take it from there.
		top := p.Library.Cards[len(p.Library.Cards)-1]
		id := top.InstanceID
		if _, err := MoveCard(p.Library, g.Exile, id); err != nil {
			return err
		}
		// Exiled face up: cascade reveals what it passes over, and a
		// card nobody can see is a card nobody can choose to cast.
		g.markCardKnownInZoneLocked(g.Exile, id)
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			Actor:   controller,
			CardID:  id,
			OldZone: ZoneLibrary,
			NewZone: ZoneExile,
		})
		if cascadeHit(top, lessThan) {
			hit = id
			break
		}
		pile = append(pile, id)
	}
	if hit == uuid.Nil {
		// Ran the library out without finding anything. Everything
		// exiled goes back to the bottom; the deck is reordered but
		// not lost.
		g.bottomInRandomOrderLocked(p, pile)
		return nil
	}
	hitName := "the exiled card"
	if c, ok := g.cardInZoneLocked(g.Exile, hit); ok {
		hitName = c.Name
	}
	// "You MAY cast it." Declining is a real choice — a cascade into
	// a card you do not want cast (an opponent's Bojuka Bog trigger
	// waiting, a creature that would die to a board wipe already on
	// the stack) is a decision, not a formality.
	return g.QueueMayCastForEffect(controller, source, hit,
		"Cascade — cast "+hitName+" without paying its mana cost?",
		func(g *Game) error {
			g.grantFreeCastLocked(controller, hit)
			g.scheduleCascadeBottomLocked(controller, source, hit, hitName)
			g.bottomInRandomOrderLocked(p, pile)
			return nil
		},
		func(g *Game) error {
			// Declined: the hit joins the rest of the pile and the
			// whole lot goes to the bottom in a random order.
			g.bottomInRandomOrderLocked(p, append(pile, hit))
			return nil
		})
}

// cascadeHit reports whether an exiled card stops the cascade: a
// nonland card whose mana value is strictly less than the cascading
// spell's (CR 702.85a).
//
// A card whose printed cost the engine cannot READ is deliberately
// not a hit. Split and adventure cards import a joined cost
// ("{1}{R} // {1}{U}") that ParseCost rejects and Card.ManaValue
// reports as zero — which would make every one of them a hit for any
// cascade, and hand the player a free cast of a card whose real mana
// value might be higher than the cascading spell's. Skipping them
// keeps cascade weaker than printed (it may pass over a card it
// should have stopped on) rather than stronger, which is the side to
// err on. The loop still terminates: worst case it exiles the
// library, which is a legal cascade outcome.
func cascadeHit(c Card, lessThan int) bool {
	if c.IsLand() {
		return false
	}
	if _, err := ParseCost(c.ManaCost); err != nil {
		return false
	}
	return c.ManaValue() < lessThan
}

// grantFreeCastLocked stamps a "cast this, this turn, for nothing"
// permission on an exiled card.
//
// The price is spelled "{0}" rather than left empty because
// ExilePlayPermission.CostOverride treats the empty string as "no
// override, pay the printed cost" — which for a cascade hit would be
// the exact opposite of what the keyword grants. "{0}" parses to the
// zero cost, so the cast is free and the parser never sees the
// printed cost at all (which is what lets a card the engine can't
// price still be cast, should one ever get here).
//
// CastOnly, because cascade says "you may CAST it": a land is never a
// legal cascade hit, so the flag is belt-and-braces, but the grant
// should not silently become a land drop if the hit rule ever
// loosens.
//
// Caller must hold g.mu.
func (g *Game) grantFreeCastLocked(controller, cardID uuid.UUID) {
	if g.Exile == nil {
		return
	}
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID != cardID {
			continue
		}
		g.Exile.Cards[i].ExilePlay = ExilePlayPermission{
			Player:       controller,
			UntilTurn:    g.Turn.Number,
			CostOverride: "{0}",
			CastOnly:     true,
		}
		return
	}
}

// scheduleCascadeBottomLocked queues the "if you didn't cast it, it
// goes to the bottom anyway" cleanup: at the beginning of the next
// end step, a cascade hit still sitting in exile under its free-cast
// grant is put on the bottom of its owner's library.
//
// This is what keeps the grant-instead-of-inline-cast simplification
// from being STRONGER than printed. Without it, a player who answered
// "yes" and then never cast the card would have parked it in exile
// where nothing reclaims it, which is a permanent improvement on a
// keyword that gives you one window.
//
// Deliberately narrow: it fires only if the card is still in exile
// AND still carries the grant this cascade stamped. A card that was
// cast, or exiled again by something else, or re-granted by a
// different effect, is left alone.
//
// Caller must hold g.mu.
func (g *Game) scheduleCascadeBottomLocked(controller, source, cardID uuid.UUID, name string) {
	g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
		Controller:   controller,
		SourceCardID: source,
		Label:        "Cascade — put " + name + " on the bottom of its owner's library",
		At:           StepEnd,
		Cards:        []uuid.UUID{cardID},
		Effect: func(g *Game, _ *StackItem) error {
			c, ok := g.cardInZoneLocked(g.Exile, cardID)
			if !ok || !c.ExilePlay.Active(controller, g.Turn.Number) || c.ExilePlay.CostOverride != "{0}" {
				return nil
			}
			owner := g.playerByIDLocked(c.Owner)
			if owner == nil {
				return nil
			}
			// Not MoveCard: that pushes to the TOP, and cascade puts
			// what it did not cast on the BOTTOM. Same hand-rolled
			// remove-then-PushBottom bottomInRandomOrderLocked does,
			// and for the same reason.
			g.bottomInRandomOrderLocked(owner, []uuid.UUID{cardID})
			return nil
		},
	})
}

// bottomInRandomOrderLocked moves every card in `ids` from exile to
// the bottom of `p`'s library in a random order (CR 702.85a).
//
// The order is one draw on p's "random_order" stream (rng.go, ADR
// 0054 Decision 8), so an undone cascade bottoms the cards in the same
// order when it is redone, and a restored game continues the stream.
//
// Cards are stripped of their knower set on the way in: they were
// face up in exile, and a library is a hidden zone. Leaving KnownBy
// populated would hand every seat permanent knowledge of a handful of
// library cards and their positions.
//
// Caller must hold g.mu.
func (g *Game) bottomInRandomOrderLocked(p *Player, ids []uuid.UUID) {
	if p == nil || p.Library == nil || g.Exile == nil || len(ids) == 0 {
		return
	}
	order := append([]uuid.UUID(nil), ids...)
	swap := func(i, j int) { order[i], order[j] = order[j], order[i] }
	g.randForLocked(rngStream{kind: rngStreamRandomOrder, player: p.ID}).Shuffle(len(order), swap)
	for _, id := range order {
		c, err := g.Exile.Remove(id)
		if err != nil {
			continue
		}
		c.ExilePlay = ExilePlayPermission{}
		c.KnownBy = nil
		p.Library.PushBottom(c)
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			Actor:   p.ID,
			CardID:  id,
			OldZone: ZoneExile,
			NewZone: ZoneLibrary,
		})
	}
}

// cardInZoneLocked returns a copy of a card in the given zone.
// Caller must hold g.mu.
func (g *Game) cardInZoneLocked(z *Zone, id uuid.UUID) (Card, bool) {
	if z == nil {
		return Card{}, false
	}
	for i := range z.Cards {
		if z.Cards[i].InstanceID == id {
			return z.Cards[i], true
		}
	}
	return Card{}, false
}
