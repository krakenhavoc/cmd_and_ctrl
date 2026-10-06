package game

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
)

// shield_source_filter.go — #2026 (ADR 0108 §7, amendment of
// 2026-10-05): a ModPreventFromSource shield that names its sources by
// what they are NOT, by their power, by whether they are attacking or
// unblocked, by their counters, by being colourless, or by who controls
// them. "Prevent all combat damage that would be dealt this turn by
// non-Spider creatures" (Arachnogenesis), "by creatures with power 3 or
// less" (Fog of War), "by attacking creatures without flying"
// (Al-abara's Carpet), "by creatures your opponents control" (Thwart the
// Enemy).
//
// THE RULES.
//
//   - CR 609.7b / 615.9: a shield against sources with certain
//     properties rechecks the source's properties when the source would
//     deal damage. If they no longer match, the damage isn't prevented
//     and the shield isn't used up.
//   - CR 611.2c fixes the affected set of a resolved effect only when the
//     effect modifies characteristics or changes control. A prevention
//     effect does neither, so its set of sources is never locked in: a
//     creature that enters, turns nongreen, loses trample or starts
//     attacking after the shield resolved is caught, and one that stops
//     matching is not. Every card's ruling says the same (Hindervines,
//     Vine Snare, Tanglesap, Moonmist, Hunter's Ambush, Encircling
//     Fissure, Lithomancer's Focus, Haze Frog, Inspire Awe).
//   - CR 608.2h: a source that has left the battlefield is read as it
//     last existed there. Heavy Fog's ruling: noncombat damage from a
//     creature that has left is prevented "if it was an attacking
//     creature at the time it left".
//   - CR 509.1h: "unblocked" is decided as blockers are declared and
//     lasts until the creature is removed from combat or the combat
//     phase ends; UnblockedAttackerForEffect is its reader.
//   - CR 105.2c: a colorless object has no color.
//
// THE SHAPE. One plain-data DamageSourceFilter beside the shield's
// Queries, in Mod.SourceFilter (at most one entry, refused on every
// other kind). Every field is a closed vocabulary or plain data, so a
// restore point restores it exactly and an older binary refuses a file
// carrying it (an unknown mod key, ADR 0041 P4). Nothing in it is a
// func, so the closure ratchet has no new route.
//
// WHEN IT IS READ. As the damage would be dealt, every time, by
// fromSourceMeetsLocked — never as the shield is made. The one
// exception a card can print is a set it locks in itself (Terrifying
// Presence's "target creature", Haze Frog's "other creatures"), and
// those are objects pinned as the shield resolves (ExceptObjects).

// SourceCombatStatus is the filter's combat test: "attacking creatures",
// "unblocked creatures". A closed vocabulary; restore refuses a value
// this binary does not know.
type SourceCombatStatus string

const (
	// SourceCombatAny asks nothing about combat — the zero value.
	SourceCombatAny SourceCombatStatus = ""
	// SourceCombatAttacking is "attacking creatures" (Harmless Assault,
	// Heavy Fog): an attacking creature (CR 508.1k) as the damage would
	// be dealt, or, for a source that has left the battlefield, as it
	// last existed there (CR 608.2h, Heavy Fog's ruling).
	SourceCombatAttacking SourceCombatStatus = "attacking"
	// SourceCombatUnblocked is "unblocked creatures" (Snag): an
	// unblocked attacking creature (CR 509.1h), read as the damage would
	// be dealt. A creature is unblocked only on the battlefield, in
	// combat, once blockers are declared.
	SourceCombatUnblocked SourceCombatStatus = "unblocked"
)

// Known reports whether this binary can interpret s.
func (s SourceCombatStatus) Known() bool {
	switch s {
	case SourceCombatAny, SourceCombatAttacking, SourceCombatUnblocked:
		return true
	}
	return false
}

// SourceControllerRel is the filter's controller test, relative to the
// shield's controller ("you", the record's Controller). A closed
// vocabulary.
type SourceControllerRel string

const (
	// SourceControllerAny asks nothing about the controller — the zero
	// value.
	SourceControllerAny SourceControllerRel = ""
	// SourceControllerNotYou is "sources you don't control" (Channel
	// Harm, Comeuppance): a source with a controller other than you.
	SourceControllerNotYou SourceControllerRel = "notYou"
	// SourceControllerOpponents is "sources your opponents control",
	// "creatures your opponents control" (Thwart the Enemy, Judgment of
	// Alexander). Every other player is an opponent (CR 102.3; the
	// engine plays free-for-all), so it is a controller other than you.
	// Kept apart from SourceControllerNotYou because the two are
	// different printed words, and a team variant would split them.
	SourceControllerOpponents SourceControllerRel = "opponents"
	// SourceControllerPlayer is "creatures target opponent controls"
	// (Encircling Fissure): the source's controller is
	// DamageSourceFilter.ControllerPlayer, chosen as the shield was made.
	SourceControllerPlayer SourceControllerRel = "player"
)

// Known reports whether this binary can interpret r.
func (r SourceControllerRel) Known() bool {
	switch r {
	case SourceControllerAny, SourceControllerNotYou, SourceControllerOpponents, SourceControllerPlayer:
		return true
	}
	return false
}

// DamageSourceFilter is the part of a preventFromSource shield's source
// description that a PermanentQuery cannot say. Every set field must
// hold, beside the record's Queries (any of them). The zero value asks
// nothing.
type DamageSourceFilter struct {
	// Except are the properties a source must NOT have, any of them:
	// "non-Spider creatures" is Queries {creature} with Except
	// {Subtypes: Spider}; "nongreen", "non-Human sources", "creatures
	// without trample", "creatures other than Werewolves and Wolves",
	// and Inspire Awe's "except … enchanted creatures and enchantment
	// creatures" (Enchanted is read off the battlefield, as the source
	// is or last was there).
	Except []PermanentQuery `json:"except,omitempty"`

	// ExceptObjects are objects whose damage the shield does not
	// prevent, pinned as the shield is made (CR 400.7): Haze Frog's
	// "other creatures" (the Frog itself) and Terrifying Presence's
	// "creatures other than target creature". A pinned card that leaves
	// and comes back is a new object, and its damage is prevented.
	ExceptObjects []ObjectRef `json:"exceptObjects,omitempty"`

	// PowerBounded and PowerAtMost are "with power N or less" (Fog of
	// War, Vine Snare): the source's power, counters included (CR
	// 613.4c), is at most PowerAtMost. PowerAtMost is read only when
	// PowerBounded is set, so "power 0 or less" is sayable.
	PowerBounded bool `json:"powerBounded,omitempty"`
	PowerAtMost  int  `json:"powerAtMost,omitempty"`

	// Colorless is "colorless sources" (Lithomancer's Focus): a source
	// with no colour (CR 105.2c).
	Colorless bool `json:"colorless,omitempty"`

	// NoCounters are counter kinds the source must have none of: "with
	// no +1/+1 counters on them" (Hindervines) is {CounterPlusOne}.
	NoCounters []string `json:"noCounters,omitempty"`

	// Combat is "attacking" or "unblocked".
	Combat SourceCombatStatus `json:"combat,omitempty"`

	// Controller is "you don't control", "your opponents control" or
	// "that player controls"; ControllerPlayer is that player, read only
	// with SourceControllerPlayer.
	Controller       SourceControllerRel `json:"controller,omitempty"`
	ControllerPlayer uuid.UUID           `json:"controllerPlayer,omitempty"`
}

// IsZero reports whether f asks nothing.
func (f DamageSourceFilter) IsZero() bool {
	return len(f.Except) == 0 && len(f.ExceptObjects) == 0 && !f.PowerBounded && f.PowerAtMost == 0 &&
		!f.Colorless && len(f.NoCounters) == 0 && f.Combat == SourceCombatAny &&
		f.Controller == SourceControllerAny && f.ControllerPlayer == uuid.Nil
}

// clone deep-copies f.
func (f DamageSourceFilter) clone() DamageSourceFilter {
	f.Except = clonePermanentQueries(f.Except)
	f.ExceptObjects = append([]ObjectRef(nil), f.ExceptObjects...)
	if len(f.ExceptObjects) == 0 {
		f.ExceptObjects = nil
	}
	f.NoCounters = copyStrings(f.NoCounters)
	return f
}

// cloneSourceFilters deep-copies a mod's SourceFilter slice, preserving
// nil.
func cloneSourceFilters(in []DamageSourceFilter) []DamageSourceFilter {
	if len(in) == 0 {
		return nil
	}
	out := make([]DamageSourceFilter, len(in))
	for i := range in {
		out[i] = in[i].clone()
	}
	return out
}

// sourceFilterProblem is registration's and restore's check on one
// filter. "" when it is sound.
func sourceFilterProblem(f DamageSourceFilter) string {
	switch {
	case f.IsZero():
		return "a preventFromSource source filter asks nothing"
	case !f.Combat.Known():
		return fmt.Sprintf("a preventFromSource source filter names combat status %q", f.Combat)
	case !f.Controller.Known():
		return fmt.Sprintf("a preventFromSource source filter names controller %q", f.Controller)
	case (f.Controller == SourceControllerPlayer) != (f.ControllerPlayer != uuid.Nil):
		return "a preventFromSource source filter names a player exactly when its controller test is that player"
	case !f.PowerBounded && f.PowerAtMost != 0:
		return "a preventFromSource source filter carries a power bound it does not read"
	}
	for _, r := range f.ExceptObjects {
		if r.ID == uuid.Nil {
			return "a preventFromSource source filter excepts an empty object"
		}
	}
	for _, k := range f.NoCounters {
		if k == "" {
			return "a preventFromSource source filter names an empty counter kind"
		}
	}
	return ""
}

// damageSourceView is a damage source as a source filter reads it: its
// characteristics and controller, and the facts only a permanent has —
// power with its counters, its counters, its combat status, whether it
// is enchanted, and which object it is.
type damageSourceView struct {
	ch *Characteristic
	// controller is the source's controller (Characteristic.Controller,
	// or the permanent's as it last existed).
	controller uuid.UUID
	// power is the source's power, counters included for a permanent.
	power int
	// permanent reports that the facts below are a permanent's, live
	// or last-known; false for a spell or a card in another zone.
	permanent bool
	ref       ObjectRef
	counters  map[string]int
	attacking bool
	unblocked bool
	enchanted bool
}

// damageSourceViewLocked reads the event's source as the damage would be
// dealt (CR 609.7b):
//
//   - on the battlefield: its characteristics as the event was opened
//     (ev.SourceLKI, the reading every property shield shares) and the
//     permanent's own facts read now;
//   - departed from the battlefield this turn and not since cast: the
//     permanent as it last existed there (CR 608.2h), characteristics
//     and controller included — Frontline Strategist's ruling: "If the
//     creature that is dealing the damage is not on the battlefield, use
//     its creature type right before it left the battlefield";
//   - otherwise (a spell, an emblem, a card in another zone): its
//     characteristics only. It is not a permanent, so it has no
//     counters, is not attacking and is not enchanted.
//
// False when the engine cannot see the source at all: the filter then
// matches nothing, the weaker direction.
//
// Caller must hold g.mu with fresh layers, as every replacement
// applicability check does.
func (g *Game) damageSourceViewLocked(ev *ReplacementEvent) (damageSourceView, bool) {
	id := ev.DamageSource
	if id == uuid.Nil {
		return damageSourceView{}, false
	}
	if c := findBattlefieldCard(g, id); c != nil {
		ch := ev.SourceLKI
		if ch == nil {
			ch = SourceCharacteristics(c)
		}
		return damageSourceView{
			ch:         ch,
			controller: c.Controller,
			power:      c.PowerForComparison(),
			permanent:  true,
			ref:        ObjectRef{ID: id, Epoch: c.ObjectEpoch},
			counters:   c.Counters,
			attacking:  c.AttackingTarget != uuid.Nil,
			unblocked:  g.UnblockedAttackerForEffect(id),
			enchanted:  g.isEnchantedLocked(id),
		}, true
	}
	if z, _ := g.locateCardLocked(id); z == nil || z.Kind != ZoneStack {
		if rec, ok := g.LastKnownPermanentForEffect(id); ok {
			ch := rec.Characteristic
			ch.Controller = rec.Controller
			return damageSourceView{
				ch:         &ch,
				controller: rec.Controller,
				power:      rec.Power,
				permanent:  true,
				ref:        ObjectRef{ID: id, Epoch: rec.Epoch},
				counters:   rec.Counters,
				attacking:  rec.Attacking,
				enchanted:  rec.Enchanted,
			}, true
		}
	}
	ch := ev.SourceLKI
	if ch == nil {
		ch = g.damageSourceLKILocked(id)
	}
	if ch == nil {
		return damageSourceView{}, false
	}
	return damageSourceView{ch: ch, controller: ch.Controller, power: ch.Power}, true
}

// sourceFilterMatches reports whether the source, viewed as the damage
// would be dealt, passes the record's filter. "You" is the record's
// controller.
func sourceFilterMatches(e ScopedEffect, f DamageSourceFilter, v damageSourceView) bool {
	for _, q := range f.Except {
		if v.matches(q) {
			return false
		}
	}
	for _, r := range f.ExceptObjects {
		if v.permanent && v.ref == r {
			return false
		}
	}
	if f.PowerBounded && v.power > f.PowerAtMost {
		return false
	}
	if f.Colorless && len(v.ch.Colors) != 0 {
		return false
	}
	for _, k := range f.NoCounters {
		if v.counters[k] > 0 {
			return false
		}
	}
	switch f.Combat {
	case SourceCombatAttacking:
		if !v.attacking {
			return false
		}
	case SourceCombatUnblocked:
		if !v.unblocked {
			return false
		}
	}
	switch f.Controller {
	case SourceControllerNotYou, SourceControllerOpponents:
		if v.controller == uuid.Nil || e.Controller == uuid.Nil || v.controller == e.Controller {
			return false
		}
	case SourceControllerPlayer:
		if v.controller == uuid.Nil || v.controller != f.ControllerPlayer {
			return false
		}
	}
	return true
}

// matches is PermanentQuery.matchesCharacteristic over the view, with
// Enchanted read off the battlefield: a permanent with an Aura attached
// to it now, or as it last existed there.
func (v damageSourceView) matches(q PermanentQuery) bool {
	if q.Enchanted && !(v.permanent && v.enchanted) {
		return false
	}
	q.Enchanted = false
	return q.matchesCharacteristic(v.ch)
}

// queriesMatchView is queriesMatchCharacteristic over the view, for a
// filtered shield's positive Queries, so they and the filter read one
// source.
func queriesMatchView(qs []PermanentQuery, v damageSourceView) bool {
	return slices.ContainsFunc(qs, v.matches)
}
