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
// trigger's resolution, ignoring timing (CR 608.2g). Here, answering
// "yes" stamps a free-cast permission on the exiled card and the player
// casts it with an ordinary cast_spell action, so it goes on the stack
// after the trigger has finished resolving rather than on top of it.
// An inline cast would have to collect targets, modes and X from inside
// a resolution, and the announce path has no frame for a half-validated
// cast — see CastSpellParams.Face on why PendingChoice is deliberately
// not that frame.
//
// ADR 0099 moved cascade onto the pieces discover shares (discover.go),
// so the grant now keeps the printed card's shape in three ways it used
// not to:
//
//   - TimingFlash (CR 608.2g). A cascaded sorcery is castable in combat
//     or on another player's turn. Before, it obeyed its own timing and
//     was simply lost there — weaker than printed.
//   - MaxSpellManaValue, one less than the cascading spell's. "If the
//     resulting spell's mana value is less than this spell's mana
//     value" (CR 702.85a) refuses a modal DFC's expensive back face or
//     an adventure half that costs more, which the grant used to allow.
//   - LapseOnPass. The window closes on the caster's NEXT priority pass
//     and the uncast hit goes to the bottom of the library then. It used
//     to stay castable until the end step and be bottomed by a delayed
//     trigger, which let a player hold a free spell for the rest of the
//     turn — stronger than printed.

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
	pile, hit, err := g.exileUntilLocked(controller, func(c Card) bool { return cascadeHit(c, lessThan) })
	if err != nil {
		return err
	}
	if hit == uuid.Nil {
		// Ran the library out without finding anything. Everything
		// exiled goes back to the bottom; the deck is reordered but
		// not lost.
		return g.PutOnBottomInRandomOrderForEffect(controller, ZoneExile, pile)
	}
	hitName := "the exiled card"
	if c, ok := g.cardInZoneLocked(g.Exile, hit); ok {
		hitName = c.Name
	}
	// The pile and the hit are IDs saved before the prompt. Every
	// bottom below names ZoneExile, so a card that left exile while
	// the question was open stays where it went rather than being
	// pulled back.
	//
	// "You MAY cast it." Declining is a real choice — a cascade into
	// a card you do not want cast is a decision, not a formality.
	return g.QueueMayCastPromptForEffect(MayCastPrompt{
		Chooser:      controller,
		Source:       source,
		Card:         hit,
		Question:     "Cascade — cast " + hitName + " without paying its mana cost?",
		Keyword:      MayCastKeywordCascade,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Put it on the bottom",
		OnAccept: func(g *Game) error {
			g.grantFreeCastLocked(controller, hit, lessThan)
			return g.PutOnBottomInRandomOrderForEffect(controller, ZoneExile, pile)
		},
		OnDecline: func(g *Game) error {
			// Declined: the hit joins the rest of the pile and the
			// whole lot goes to the bottom in a random order.
			return g.PutOnBottomInRandomOrderForEffect(controller, ZoneExile, append(pile, hit))
		},
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

// grantFreeCastLocked stamps the cascade hit's free cast.
//
// The price is spelled "{0}" rather than left empty because
// CastPermission.Cost treats the empty string as "no override, pay
// the printed cost" — which for a cascade hit would be the exact
// opposite of what the keyword grants. "{0}" parses to the zero cost,
// so the cast is free, and it locks X at 0 (CR 107.3b).
//
// CastOnly, because cascade says "you may CAST it". TimingFlash,
// because the cast happens during the trigger's resolution
// (CR 608.2g). MaxSpellManaValue is lessThan-1: the resulting spell's
// mana value must be LESS than the cascading spell's (CR 702.85a).
// LapseOnPass closes the window on the caster's next pass and bottoms
// the card then (ADR 0099 §4).
//
// Caller must hold g.mu.
func (g *Game) grantFreeCastLocked(controller, cardID uuid.UUID, lessThan int) {
	c, ok := g.cardInZoneLocked(g.Exile, cardID)
	if !ok {
		return
	}
	limit := lessThan - 1
	g.GrantCastPermissionToCardsForEffect(CastPermission{
		Player:            controller,
		Zone:              ZoneExile,
		Duration:          g.UntilEndOfTurnDuration(),
		Cost:              "{0}",
		Timing:            TimingFlash,
		CastOnly:          true,
		MaxSpellManaValue: &limit,
		LapseOnPass:       LapseToLibraryBottom,
		Label:             "Cascade — cast it without paying its mana cost",
	}, []Card{c})
}

// cascadeBottomBody is the end-step delayed trigger cascade scheduled
// before ADR 0099: at the beginning of the next end step, a hit still
// in exile under its free-cast grant went to the bottom of its owner's
// library. Nothing schedules it any more — the grant's window closes
// on the caster's next pass instead (LapseOnPass) — but the key stays
// registered, because a restore point written by an older binary may
// hold one, and an effect key is never deleted (ADR 0041 phase 3).
//
// The body is registered in init rather than by a var initialiser: it
// reaches the exit primitives, which reach the scheduler, and an
// initialiser would be an initialisation cycle. The card rides on the
// item's Targets and the controller is the item's, so it reads no
// params.
func init() { SimpleDelayedBody("cascade/bottom-if-not-cast", cascadeBottom) }

func cascadeBottom(g *Game, item *StackItem) error {
	controller := item.Controller
	for _, t := range item.Targets {
		if t.Kind != TargetCard || t.ID == uuid.Nil {
			continue
		}
		cardID := t.ID
		c, ok := g.cardInZoneLocked(g.Exile, cardID)
		if !ok {
			continue
		}
		perm := g.CastPermissionForLocked(controller, c, ZoneExile)
		if perm == nil || perm.Cost != "{0}" {
			continue
		}
		// The shared random bottom (random_bottom.go) with a pile
		// of one: it routes through the exit path, so the card
		// goes to its OWNER's library and a commander's owner is
		// asked about the command zone.
		if err := g.PutOnBottomInRandomOrderForEffect(controller, ZoneExile, []uuid.UUID{cardID}); err != nil {
			return err
		}
	}
	return nil
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
