package game

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// multiply_damage.go — ADR 0108 §3 (#1890): damage doubled or tripled
// by a resolved spell or ability, for the rest of the turn or until
// someone's next turn.
//
// THE RULES.
//
//   - CR 614.1a: "it deals double that damage instead" is a replacement
//     effect. It is not a prevention effect (CR 615), so "damage can't be
//     prevented" leaves it alone.
//   - CR 616.1, 616.1f: several multipliers on one event are ordered by
//     the affected player (the one being dealt the damage, or the
//     controller of the permanent), one at a time; each applies once. Two
//     Insults are ×4 (the Insult // Injury ruling), and either order gives
//     the same number, so the order is never asked between two of these.
//   - CR 120.8: a source that would deal 0 damage deals no damage at all,
//     so there is nothing to multiply and a "next time" record is not
//     spent by it.
//   - CR 611.2c: "a source you control", "an opponent", "a permanent an
//     opponent controls" are read as the damage would be dealt, not when
//     the effect began — the effect changes no characteristic — so a
//     creature you gain control of after Insult resolved is doubled too.
//     The source is read through its last-known information (CR 608.2h),
//     as Angrath's Marauders reads it.
//   - CR 615.8's "next time", which Desperate Gambit and Impulsive
//     Maneuvers print for a multiplier: the next INSTANCE of damage from
//     that source, however many events it is (a trampler's damage to its
//     blocker and to the player is one instance).
//
// THE SHAPE (ADR 0108 §3 decision 1). One ScopedEffect replacement kind,
// ModMultiplyDamage, plain data, so a table holding one is a restore
// point. Its mod reads:
//
//   - Amount, the multiplier (2 or 3);
//   - which damage: Objects[0] + SourceZone for one source, read by
//     damageFromChosenSourceLocked exactly as preventNextFromSource reads
//     its chosen source (ADR 0108 "one source vocabulary"); or Sources,
//     DamageSourcesYours / DamageSourcesCreatures; or neither, which is
//     any source (Lightning, Army of One);
//   - to what: Recipients, the DamageRecipients vocabulary below, with
//     Player for DamageRecipientsPlayerAndTheirPermanents; empty is
//     anything;
//   - CombatOnly, and Next — "the next time", spent per instance through
//     SpentInstance / SpentBatch, the next-damage shield's own fields and
//     rule (nextShieldOpenToLocked, markNextShieldSpentLocked).
//
// The record's Controller is the "you" of every clause.

// ModMultiplyDamage is "if <source> would deal damage [to <recipient>]
// this turn, it deals double (triple) that damage instead" (ADR 0108 §3,
// CR 614.1a). Reads Amount, Objects, SourceZone, Sources, Recipients,
// Player, CombatOnly, Next, SpentBatch and SpentInstance. Scope
// ScopeGame.
const ModMultiplyDamage ModKind = "multiplyDamage"

// DamageSources is a multiplier's "which sources" when it names no one
// object. A closed vocabulary: restore refuses a value this binary does
// not know (ErrUnknownEffectKey).
type DamageSources string

const (
	// DamageSourcesAny is every source — the zero value.
	DamageSourcesAny DamageSources = ""
	// DamageSourcesYours is "a source you control": the source's
	// controller, as it last existed, is the record's controller
	// (Insult, Isengard Unleashed, Goblin Goliath, Quest for Pure Flame).
	DamageSourcesYours DamageSources = "yours"
	// DamageSourcesCreatures is "a creature" (Blind Fury).
	DamageSourcesCreatures DamageSources = "creatures"
)

// Known reports whether this binary can interpret s.
func (s DamageSources) Known() bool {
	switch s {
	case DamageSourcesAny, DamageSourcesYours, DamageSourcesCreatures:
		return true
	}
	return false
}

// DamageRecipients is a multiplier's "to what". A closed vocabulary, as
// DamageSources is.
type DamageRecipients string

const (
	// DamageRecipientsAny is anything — the zero value ("to a permanent
	// or player").
	DamageRecipientsAny DamageRecipients = ""
	// DamageRecipientsOpponents is "to an opponent": a player other
	// than the record's controller (Goblin Goliath).
	DamageRecipientsOpponents DamageRecipients = "opponents"
	// DamageRecipientsOpponentsAndTheirPermanents is "to an opponent or
	// a permanent an opponent controls" (Isengard Unleashed).
	DamageRecipientsOpponentsAndTheirPermanents DamageRecipients = "opponentsAndTheirPermanents"
	// DamageRecipientsPlayerAndTheirPermanents is "to that player or a
	// permanent that player controls", the player being Mod.Player
	// (Lightning, Army of One).
	DamageRecipientsPlayerAndTheirPermanents DamageRecipients = "playerAndTheirPermanents"
	// DamageRecipientsCreatures is "to a creature" (Blind Fury).
	DamageRecipientsCreatures DamageRecipients = "creatures"
)

// Known reports whether this binary can interpret r.
func (r DamageRecipients) Known() bool {
	switch r {
	case DamageRecipientsAny, DamageRecipientsOpponents, DamageRecipientsOpponentsAndTheirPermanents,
		DamageRecipientsPlayerAndTheirPermanents, DamageRecipientsCreatures:
		return true
	}
	return false
}

// DamageMultiplier is the queue-side description of a ModMultiplyDamage
// record.
type DamageMultiplier struct {
	// EffectSource is the card whose spell or ability made the effect;
	// Controller is its "you".
	EffectSource uuid.UUID
	Controller   uuid.UUID

	// Factor is the multiplier: 2 is "double", 3 "triple".
	Factor int

	// Source is one source (a target, a chosen source, an attacking
	// creature), built with DamageSourceRefLocked; the zero ref is "the
	// sources Sources names".
	Source     ObjectRef
	SourceZone ZoneKind
	Sources    DamageSources

	// Recipients is "to what"; Player is the player
	// DamageRecipientsPlayerAndTheirPermanents names.
	Recipients DamageRecipients
	Player     uuid.UUID

	// CombatOnly is "would deal combat damage"; Next is "the next time"
	// (CR 615.8's instance).
	CombatOnly bool
	Next       bool

	// UntilNextTurnOf, when set, makes the effect last "until <that
	// player>'s next turn" (Lightning, Army of One) instead of "this
	// turn" (CR 514.2).
	UntilNextTurnOf uuid.UUID

	// Label is the record's label.
	Label string
}

// MultiplyDamageForEffect registers the multiplier. Reports whether a
// record was written: nothing is registered for a factor below 2, a
// vocabulary value this binary does not know, a "next time" with no one
// source to name, or a player-scoped recipient with no player.
//
// Caller must hold g.mu (write) — every caller is a resolving effect.
func (g *Game) MultiplyDamageForEffect(d DamageMultiplier) bool {
	m := Mod{
		Kind:       ModMultiplyDamage,
		Amount:     d.Factor,
		Sources:    d.Sources,
		Recipients: d.Recipients,
		Player:     d.Player,
		CombatOnly: d.CombatOnly,
		Next:       d.Next,
	}
	if d.Source.ID != uuid.Nil {
		m.Objects = []ObjectRef{d.Source}
		m.SourceZone = d.SourceZone
	}
	if multiplyDamageModProblem(m) != "" {
		return false
	}
	label := d.Label
	if label == "" {
		label = "damage multiplied"
	}
	dur := g.UntilEndOfTurnDuration()
	if d.UntilNextTurnOf != uuid.Nil {
		dur = g.UntilYourNextTurnDuration(d.UntilNextTurnOf)
	}
	return g.RegisterScopedRuleEffectForEffect(d.EffectSource, ScopeGame, d.Controller, []Mod{m}, dur, label)
}

// multiplyDamageModProblem is registration's (and restore's) check on the
// kind's parameters, and the refusal of its own fields on every other
// kind (a newer binary's shape). "" when the mod is sound.
func multiplyDamageModProblem(m Mod) string {
	if m.Kind != ModMultiplyDamage {
		if m.Sources != "" || m.Recipients != "" || m.Next {
			return fmt.Sprintf("mod %q carries a field only multiplyDamage reads", m.Kind)
		}
		return ""
	}
	switch {
	case m.Amount < 2:
		return fmt.Sprintf("a multiplyDamage effect needs a multiplier of at least 2, got %d", m.Amount)
	case !m.Sources.Known():
		return fmt.Sprintf("damage sources %q", m.Sources)
	case !m.Recipients.Known():
		return fmt.Sprintf("damage recipients %q", m.Recipients)
	case len(m.Objects) > 1:
		return "a multiplyDamage effect names more than one source"
	case len(m.Objects) == 1 && m.Objects[0].ID == uuid.Nil:
		return "a multiplyDamage effect names an empty source"
	case len(m.Objects) == 1 && m.Sources != DamageSourcesAny:
		return "a multiplyDamage effect names both one source and a set of sources"
	case m.Next && len(m.Objects) == 0:
		return "a \"next time\" multiplyDamage effect names no source"
	case (m.Recipients == DamageRecipientsPlayerAndTheirPermanents) != (m.Player != uuid.Nil):
		return "a multiplyDamage effect names a player exactly when its recipients are that player and their permanents"
	case len(m.Queries) != 0:
		return "a multiplyDamage effect carries a property recheck it does not read"
	case m.Then != "":
		return "a multiplyDamage effect carries a follow-up it does not read"
	}
	return ""
}

// multiplyDamageAppliesLocked is the kind's AppliesTo.
//
// Caller must hold g.mu.
func (g *Game) multiplyDamageAppliesLocked(e ScopedEffect, m Mod, ev *ReplacementEvent) bool {
	// CR 120.8: a source that would deal 0 damage deals none.
	if ev.Kind != RepEventDamage || ev.DamageAmount <= 0 || m.Amount < 2 {
		return false
	}
	if m.CombatOnly && !ev.IsCombatDamage {
		return false
	}
	if m.Next && !nextShieldOpenToLocked(g, m, ev) {
		return false
	}
	if len(m.Objects) == 1 && !g.damageFromChosenSourceLocked(m, ev.DamageSource) {
		return false
	}
	if !g.damageSourcesMatchLocked(e, m.Sources, ev) {
		return false
	}
	return g.damageRecipientsMatchLocked(e, m, ev.DamageTarget)
}

// damageSourcesMatchLocked reads "a source you control" / "a creature"
// off the source as it was when the damage event was opened (CR 608.2h).
// A source the engine cannot see matches neither, which is the weaker
// direction.
//
// Caller must hold g.mu.
func (g *Game) damageSourcesMatchLocked(e ScopedEffect, s DamageSources, ev *ReplacementEvent) bool {
	if s == DamageSourcesAny {
		return true
	}
	ch := ev.SourceLKI
	if ch == nil {
		ch = g.damageSourceLKILocked(ev.DamageSource)
	}
	if ch == nil {
		return false
	}
	switch s {
	case DamageSourcesYours:
		return e.Controller != uuid.Nil && ch.Controller == e.Controller
	case DamageSourcesCreatures:
		return characteristicHasType(ch, "Creature")
	}
	return false
}

// damageRecipientsMatchLocked reads "to an opponent", "to a permanent an
// opponent controls", "to a creature" as the damage would be dealt
// (CR 611.2c): a player by identity, a permanent as it is on the
// battlefield now.
//
// Caller must hold g.mu.
func (g *Game) damageRecipientsMatchLocked(e ScopedEffect, m Mod, target uuid.UUID) bool {
	if m.Recipients == DamageRecipientsAny {
		return true
	}
	if target == uuid.Nil {
		return false
	}
	if p := g.playerByIDLocked(target); p != nil {
		switch m.Recipients {
		case DamageRecipientsOpponents, DamageRecipientsOpponentsAndTheirPermanents:
			return e.Controller != uuid.Nil && p.ID != e.Controller
		case DamageRecipientsPlayerAndTheirPermanents:
			return p.ID == m.Player
		}
		return false
	}
	c := findBattlefieldCard(g, target)
	if c == nil {
		return false
	}
	switch m.Recipients {
	case DamageRecipientsOpponentsAndTheirPermanents:
		return e.Controller != uuid.Nil && c.Controller != uuid.Nil && c.Controller != e.Controller
	case DamageRecipientsPlayerAndTheirPermanents:
		return c.Controller == m.Player
	case DamageRecipientsCreatures:
		return c.IsCreature()
	}
	return false
}

// characteristicHasType reports whether a characteristics snapshot has
// the card type t (case-insensitive).
func characteristicHasType(ch *Characteristic, t string) bool {
	for _, have := range ch.Types {
		if strings.EqualFold(have, t) {
			return true
		}
	}
	return false
}

// applyMultiplyDamageLocked is the kind's Replace: the amount is
// multiplied, and a "next time" record is marked spent with this
// instance, so it keeps applying to the instance's other events and to
// nothing after it (CR 615.8's reading, ADR 0108's one spend rule).
//
// Caller must hold g.mu (write).
func (g *Game) applyMultiplyDamageLocked(e ScopedEffect, mod int, m Mod, ev *ReplacementEvent) {
	ev.DamageAmount *= m.Amount
	if m.Next && m.SpentBatch == 0 && m.SpentInstance == 0 {
		g.markNextShieldSpentLocked(e.Seq, mod, g.currentEventBatchLocked(), ev.DamageInstance)
	}
}

// spentNextTimeMod reports whether m is a "next time" record already
// spent in an event batch other than `batch` — the instance it applied
// to is over, so the record is done (dropSpentNextShieldsLocked).
func spentNextTimeMod(m Mod, batch uint64) bool {
	if m.SpentBatch == 0 || m.SpentBatch == batch {
		return false
	}
	return isNextFromSourceKind(m.Kind) || (m.Kind == ModMultiplyDamage && m.Next)
}

// DamageMultiplierLines is the game banner's line for each live
// multiplier, oldest first (ADR 0108 §3 decision 4): "Alice's sources
// deal double damage this turn — Insult". Empty on nearly every turn.
//
// Caller must hold g.mu (read or write).
func (g *Game) DamageMultiplierLines() []string {
	var out []string
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		for _, m := range e.Mods {
			if m.Kind != ModMultiplyDamage {
				continue
			}
			out = append(out, g.damageMultiplierLineLocked(e, m))
		}
	}
	return out
}

// damageMultiplierLineLocked writes one record's banner line.
func (g *Game) damageMultiplierLineLocked(e *ScopedEffect, m Mod) string {
	factor := "double"
	switch m.Amount {
	case 2:
	case 3:
		factor = "triple"
	default:
		factor = fmt.Sprintf("%d times the", m.Amount)
	}
	kind := "damage"
	if m.CombatOnly {
		kind = "combat damage"
	}
	you := g.playerNameLocked(e.Controller)
	var b strings.Builder
	switch {
	case len(m.Objects) == 1:
		name := "That source"
		if c, ok := g.LookupCardForEffect(m.Objects[0].ID); ok && c.Name != "" {
			name = c.Name
		}
		b.WriteString(name)
		b.WriteString(" deals ")
	case m.Sources == DamageSourcesYours:
		b.WriteString(you)
		b.WriteString("'s sources deal ")
	case m.Sources == DamageSourcesCreatures:
		b.WriteString("Creatures deal ")
	default:
		b.WriteString("Sources deal ")
	}
	b.WriteString(factor)
	b.WriteString(" ")
	b.WriteString(kind)
	switch m.Recipients {
	case DamageRecipientsOpponents:
		b.WriteString(" to " + you + "'s opponents")
	case DamageRecipientsOpponentsAndTheirPermanents:
		b.WriteString(" to " + you + "'s opponents and their permanents")
	case DamageRecipientsPlayerAndTheirPermanents:
		b.WriteString(" to " + g.playerNameLocked(m.Player) + " and their permanents")
	case DamageRecipientsCreatures:
		b.WriteString(" to creatures")
	}
	switch {
	case m.Next:
		b.WriteString(" the next time it deals " + kind + " this turn")
	case e.Duration.Kind == UntilYourNextTurn:
		b.WriteString(" until " + g.playerNameLocked(e.Duration.Player) + "'s next turn")
	default:
		b.WriteString(" this turn")
	}
	if name := scopedEffectDisplayName(e); name != "" {
		b.WriteString(" — " + name)
	}
	line := b.String()
	return strings.ToUpper(line[:1]) + line[1:]
}
