package game

import "github.com/google/uuid"

// counter_shield_grants.go — ADR 0106 §4 decisions 2 to 4 (#1806),
// delivery PR 3: the three "can't be countered" shapes that are not a
// printed static, and where each one is stored.
//
//	Shape                    Printed example                          Stored in
//	a grant for a turn       Veil of Summer, "Spells you control      Player.Statics, until cleanup
//	                         can't be countered this turn"
//	a one-use promise        Insist, "The next creature spell you     Player.Statics until it is
//	                         cast this turn can't be countered"       spent, then a mark
//	a mark on one spell      Vexing Shusher, "Target spell can't      StackItem.CantBeCountered
//	                         be countered"
//
// Every one is a rule-modifying effect, not a characteristic of the
// spell (CR 613.11), so none of it is touched by the layer pass. The
// gate (spellCantBeCounteredLocked, cant_be_countered.go) reads the
// turn grants (its source 5) and the marks (its source 3) when
// something tries to counter a spell; a resolved spell's rule
// modification covers spells cast after it resolved (CR 611.2c),
// which reading the grant at the gate gives for free.
//
// THE PROMISE IS NOT READ AT THE GATE. "The next creature spell you
// cast this turn" is decided once, at the moment a spell becomes cast
// (CR 601.2i): castSpellLocked calls spendCounterShieldPromisesLocked
// just before EventCast, and every live promise of the caster's whose
// filter the spell matches is taken off Player.Statics and written
// onto that spell's item as a mark. So:
//
//   - the first matching spell spends it, whether or not anything was
//     ever going to counter that spell;
//   - a spell that does not match (a sorcery under Insist) leaves it;
//   - two promises that match one spell are both spent on it, each its
//     own mark;
//   - a copy is not cast (CR 707.10) and a land is not cast (CR 305.1),
//     so neither ever reaches the spend;
//   - a promise nobody spends ends at cleanup (CR 514.2) like any
//     "this turn" grant, through sweepPlayerStaticsLocked.
//
// PLAIN DATA throughout, so a table holding any of the three is still
// a restore point: the grant rides seats[].statics[] and the mark rides
// the stack item. No closure is added anywhere.

// CounterShieldGrant is the CantBeCountered payload of a PlayerStatic:
// a "this turn" grant, or (NextOnly) a one-use promise. Every field is
// plain data, written once when the grant is made.
type CounterShieldGrant struct {
	// Active is the presence bit. The zero grant is NOT "nothing to
	// say": every spell, "you control", is Veil of Summer's real one,
	// so a payload needs a bit of its own (CastBanRule.Kind's
	// argument).
	Active bool `json:"active,omitempty"`

	// Whose is whose spells the grant covers, judged against the
	// player the grant is on: the spell's current controller
	// (CounterShieldYouControl, Veil of Summer) or its caster
	// (CounterShieldYouCast, Domri's "creature spells you cast this
	// turn", never a copy — CR 707.10).
	Whose CounterShieldWhose `json:"whose,omitempty"`

	// Filter narrows which spells: "creature spells", "instant or
	// sorcery spells". The zero filter is every spell.
	//
	// No `omitzero`, for PlayerStatic.Timing's reason (#1492).
	Filter PermissionFilter `json:"filter"`

	// NextOnly marks the one-use promise: it is spent by the first
	// matching spell its player casts and is never read at the gate.
	// A promise is always "you cast" (GrantCounterShieldForEffect sets
	// Whose to say so).
	NextOnly bool `json:"nextOnly,omitempty"`

	// Except names the one spell the grant does not cover — the spell
	// that made it, for "OTHER spells you control can't be countered
	// this turn" (Bound // Determined). Matched by object: a card
	// that left the stack and was cast again is a new object (CR 400.7)
	// and is covered. The zero ref excepts nothing.
	Except ObjectRef `json:"except"`

	// Text is the clause as printed ("The next creature spell you cast
	// this turn can't be countered."), for the player panel's line.
	// Not a rules input.
	Text string `json:"text,omitempty"`
}

// CounterShieldMark is one "this spell can't be countered" mark on a
// stack item (StackItem.CantBeCountered).
type CounterShieldMark struct {
	// Source is the card whose effect made the mark (Vexing Shusher,
	// Insist), for the log. Never read by a rule.
	Source uuid.UUID `json:"source,omitempty"`
	// SourceName is that card's name, for the stack chip's tooltip.
	SourceName string `json:"sourceName,omitempty"`
	// Label is the clause that made it, for the log.
	Label string `json:"label,omitempty"`
}

// copyCounterShieldMarks gives a mark slice its own backing array. Nil
// for none, so a snapshot omits the key and an unmarked item stays the
// zero value it always was.
func copyCounterShieldMarks(in []CounterShieldMark) []CounterShieldMark {
	if len(in) == 0 {
		return nil
	}
	return append([]CounterShieldMark(nil), in...)
}

// GrantCounterShieldForEffect gives `player` a "this turn" grant or a
// one-use promise — Veil of Summer's "Spells you control can't be
// countered this turn", Insist's "The next creature spell you cast
// this turn can't be countered". It lasts until end of turn
// (CR 514.2) in both cases, because every printed card of the shape
// says "this turn"; a promise that is spent before then is gone
// sooner.
//
// `label` is the attribution (the source's name) and `source` the card
// that granted it, exactly as on every other PlayerStatic. A grant
// without Active set is made Active: a card file that built one meant
// it. A promise is made "you cast", which is what every printed
// promise says.
//
// The *ForEffect surface: caller must hold g.mu (write), which a
// resolution frame already does.
func (g *Game) GrantCounterShieldForEffect(player uuid.UUID, grant CounterShieldGrant, label string, source uuid.UUID) {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return
	}
	grant.Active = true
	if grant.NextOnly {
		grant.Whose = CounterShieldYouCast
	}
	p.Statics = append(p.Statics, PlayerStatic{
		CantBeCountered: grant,
		Source:          source,
		Label:           label,
		Duration:        g.UntilEndOfTurnDuration(),
	})
}

// counterShieldGrantedLocked is the gate's source 5: does a live "this
// turn" grant of the spell's controller, or of its caster, cover this
// spell? `item` is the spell's stack item; nil (a hand-built test
// spell) covers nothing, since there is no player to ask about.
//
// A NextOnly entry is skipped: a promise is spent at the cast and is
// read at the gate only as the mark it became. The duration is tested
// here as well as in the sweep, for playerAbilityTokensLocked's
// reason: the sweep is hygiene, the reader is the truth.
//
// Caller must hold g.mu.
func (g *Game) counterShieldGrantedLocked(item *StackItem, spell Card) bool {
	if item == nil {
		return false
	}
	controller, caster := item.Controller, item.BaseController
	if caster == uuid.Nil {
		caster = controller
	}
	for i, id := range [2]uuid.UUID{controller, caster} {
		if id == uuid.Nil || (i == 1 && id == controller) {
			continue
		}
		p := g.playerByIDLocked(id)
		if p == nil {
			continue
		}
		for _, s := range p.Statics {
			gr := s.CantBeCountered
			if !gr.Active || gr.NextOnly {
				continue
			}
			if g.durationExpiredLocked(s.Duration, false) {
				continue
			}
			if !gr.covers(p.ID, item, spell) {
				continue
			}
			return true
		}
	}
	return false
}

// covers reports whether a grant on player `you` covers this spell:
// whose it is, which spell it is, and the one spell it excepts.
func (gr CounterShieldGrant) covers(you uuid.UUID, item *StackItem, spell Card) bool {
	switch gr.Whose {
	case CounterShieldYouControl:
		if item.Controller != you {
			return false
		}
	case CounterShieldYouCast:
		caster := item.BaseController
		if caster == uuid.Nil {
			caster = item.Controller
		}
		if item.IsCopy || caster != you {
			return false
		}
	case CounterShieldAnyPlayer:
	default:
		return false
	}
	if gr.Except.ID != uuid.Nil && gr.Except.ID == spell.InstanceID && gr.Except.Epoch == spell.ObjectEpoch {
		return false
	}
	return gr.Filter.Matches(spell)
}

// spendCounterShieldPromisesLocked is CR 601.2i for the one-use
// promise: the spell `spellID` has just become cast by `caster`, so
// every live promise of the caster's whose filter it matches is taken
// off Player.Statics and written onto the spell's item as a mark. A
// spell that matches none changes nothing.
//
// Called from castSpellLocked just before EventCast, and from nowhere
// else: that is the one place a spell becomes cast, so a copy (CR
// 707.10) and a land (CR 305.1) never get here.
//
// Replaces the statics slice rather than compacting it, and the mark
// slice rather than appending into it, for the reason
// sweepPlayerStaticsLocked gives: an undo snapshot shares the backing
// arrays.
//
// Caller must hold g.mu (write).
func (g *Game) spendCounterShieldPromisesLocked(caster, spellID uuid.UUID) {
	p := g.playerByIDLocked(caster)
	item := g.StackMeta[spellID]
	spell := g.spellCardOnStackLocked(spellID)
	if p == nil || item == nil || spell == nil || item.IsCopy {
		return
	}
	var (
		kept  []PlayerStatic
		marks []CounterShieldMark
	)
	for _, s := range p.Statics {
		gr := s.CantBeCountered
		if gr.Active && gr.NextOnly && !g.durationExpiredLocked(s.Duration, false) && gr.Filter.Matches(*spell) {
			marks = append(marks, CounterShieldMark{Source: s.Source, SourceName: s.Label, Label: gr.Text})
			continue
		}
		kept = append(kept, s)
	}
	if len(marks) == 0 {
		return
	}
	p.Statics = kept
	item.CantBeCountered = append(copyCounterShieldMarks(item.CantBeCountered), marks...)
}

// MarkSpellCantBeCounteredForEffect is "Target spell can't be
// countered" (Vexing Shusher): it marks the spell `spellID` on the
// stack, for as long as that object stays there (CR 400.7). False when
// `spellID` is not a spell on the stack — it left in response, or it
// is an ability, which this statement does not reach.
//
// The *ForEffect surface: caller must hold g.mu (write).
func (g *Game) MarkSpellCantBeCounteredForEffect(spellID uuid.UUID, mark CounterShieldMark) bool {
	item := g.StackMeta[spellID]
	if item == nil || item.Kind != StackItemSpell || g.spellCardOnStackLocked(spellID) == nil {
		return false
	}
	item.CantBeCountered = append(copyCounterShieldMarks(item.CantBeCountered), mark)
	return true
}

// CounterShieldGrantSource is one live "this turn" grant or unspent
// promise on a player, for the player panel's `counter_shields` line
// (ADR 0106 §4 decision 6).
type CounterShieldGrantSource struct {
	// Source is the card that made it.
	Source uuid.UUID
	// SourceName is that card's name (the grant's attribution label).
	SourceName string
	// Text is the clause as printed.
	Text string
	// NextOnly marks an unspent one-use promise.
	NextOnly bool
}

// CounterShieldGrantsForEffect lists the live "can't be countered"
// grants and unspent promises on a player, in the order they were
// made. Nil for none, which is nearly every seat.
//
// Caller must hold g.mu (read or write).
func (g *Game) CounterShieldGrantsForEffect(p *Player) []CounterShieldGrantSource {
	if p == nil {
		return nil
	}
	var out []CounterShieldGrantSource
	for _, s := range p.Statics {
		gr := s.CantBeCountered
		if !gr.Active || g.durationExpiredLocked(s.Duration, false) {
			continue
		}
		out = append(out, CounterShieldGrantSource{
			Source:     s.Source,
			SourceName: s.Label,
			Text:       gr.Text,
			NextOnly:   gr.NextOnly,
		})
	}
	return out
}
