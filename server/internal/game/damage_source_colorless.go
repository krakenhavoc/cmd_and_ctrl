package game

import (
	"strings"

	"github.com/google/uuid"
)

// damage_source_colorless.go — #6230: "Black and/or red permanents and
// spells are colorless sources of damage." (Ghostly Flame). CR 120.2:
// any object can deal damage, and the colour of that SOURCE is what
// protection (CR 702.16e), prevention ("a red source of your choice",
// "damage black sources would deal") and source filters read. Under
// the rule a black or red permanent or spell, as a source of damage,
// has no colour (CR 105.2c). It is otherwise unchanged: it is still
// black or red for targeting, for "black creatures you control" and
// for everything that is not about the damage it deals.
//
// The rule is Characteristic.ColorlessDamageSourceColors, which the
// carrying permanent writes on itself through an ordinary static.
// It is read where a damage source's characteristics are taken:
//
//   - prepareDamageEventLocked, as a damage event is opened, rewrites
//     the event's SourceLKI (protection and every prevention shield
//     read that);
//   - damageSourceViewLocked's departed-permanent branch, which reads
//     the last-known record rather than SourceLKI;
//   - damageSourceCandidatesLocked, so "a red source of your choice"
//     (Circle of Protection: Red) can't choose one.
//
// Only permanents and spells: a card in a graveyard or an emblem that
// deals damage keeps its colour.

// colorlessDamageSourceColorsLocked is the union of every
// ColorlessDamageSourceColors rule on the battlefield, or "" for none.
// Caller must hold g.mu with fresh layers.
func (g *Game) colorlessDamageSourceColorsLocked() string {
	if g.Battlefield == nil {
		return ""
	}
	var b strings.Builder
	for i := range g.Battlefield.Cards {
		if cols := g.Battlefield.Cards[i].Effective().ColorlessDamageSourceColors; cols != "" {
			b.WriteString(cols)
		}
	}
	return b.String()
}

// damageSourceIsColorlessLocked reports whether a permanent or spell
// with colours `colors` is a colourless source of damage right now.
// Caller must hold g.mu with fresh layers.
func (g *Game) damageSourceIsColorlessLocked(colors []string) bool {
	if len(colors) == 0 {
		return false
	}
	rule := g.colorlessDamageSourceColorsLocked()
	if rule == "" {
		return false
	}
	for _, c := range colors {
		if c != "" && strings.Contains(rule, c) {
			return true
		}
	}
	return false
}

// asDamageSourceLocked is `ch` as a source of damage: a copy with no
// colour when the rule applies to it, `ch` itself otherwise.
// `permanentOrSpell` says whether the source is a permanent or a spell
// (on the battlefield, last known there, or on the stack as a spell).
// Caller must hold g.mu with fresh layers.
func (g *Game) asDamageSourceLocked(ch *Characteristic, permanentOrSpell bool) *Characteristic {
	if ch == nil || !permanentOrSpell || !g.damageSourceIsColorlessLocked(ch.Colors) {
		return ch
	}
	out := copyCharacteristic(ch)
	out.Colors = nil
	return out
}

// damageSourceIsPermanentOrSpellLocked reports whether the damage
// source `id`, described by `ch`, is a permanent or a spell: on the
// battlefield, a spell on the stack, or recorded as one in its
// last-known information.
// Caller must hold g.mu.
func (g *Game) damageSourceIsPermanentOrSpellLocked(id uuid.UUID, ch *Characteristic) bool {
	if ch != nil && (ch.SourceIsPermanent || ch.SourceIsSpell) {
		return true
	}
	if id == uuid.Nil {
		return false
	}
	if findCardOnBattlefield(g, id) >= 0 {
		return true
	}
	if item := g.StackMeta[id]; item != nil && item.Kind == StackItemSpell {
		return true
	}
	return false
}
