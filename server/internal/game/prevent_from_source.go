package game

import (
	"fmt"

	"github.com/google/uuid"
)

// prevent_from_source.go — ADR 0108 §7 (#1904): a prevention shield
// against a source that is NOT spent by the first instance of damage.
//
// THE RULES.
//
//   - CR 615.1a: an effect that says "prevent" is a prevention effect.
//   - CR 615.7: a shield that refers to a specific amount of damage ("the
//     next 3 damage") prevents each 1 damage until it is reduced to 0. If
//     two or more applicable sources would deal damage to the shielded
//     permanent or player at the same time, the player or the controller
//     of the permanent chooses which damage the shield prevents
//     (divide_shield.go).
//   - CR 615.9 / 609.7b: a shield against "a source of your choice with
//     certain properties" rechecks the properties as the source would deal
//     damage; a shield that prevents no damage is not used up.
//   - CR 609.7a: the source is chosen as the effect is created, and a
//     permanent spell chosen as a source covers the permanent it becomes.
//   - CR 615.5: the additional effect happens immediately after the
//     prevention, and may count what was prevented.
//   - CR 615.13: "whenever damage is prevented" triggers once each time a
//     prevention effect is applied to one or more simultaneous damage
//     events and prevents some of it — the follow-up queue's one entry per
//     record and damage instance.
//
// THE SHAPE (ADR 0108 §7 decision 1). One scoped replacement kind,
// ModPreventFromSource, plain data in a ScopedEffect record. It reads the
// same source and protection fields as ModPreventNextFromSource
// (prevent_next_from_source.go), so a source is pinned and rechecked the
// same way on both:
//
//   - the source, Objects[0] with SourceZone (a chosen or a targeted
//     source, CR 400.7 / 609.7a), or Queries alone ("damage black sources
//     and red sources would deal", Prismatic Strands), or neither ("all
//     damage that would be dealt to X this turn");
//   - the protected player (Player, with Types for "you and/or permanents
//     you control"), a pinned permanent (the record's Affected set), or
//     nothing ("would deal this turn");
//   - CombatOnly, for "all combat damage" (Maze of Ith);
//   - Amount: 0 is "all damage this turn" and the shield is never spent;
//     N is CR 615.7's charge, replaced copy-on-write as it is spent, as
//     ModPreventDamage's is (setShieldChargeLocked), and divided by the
//     protected player when it meets several events of one damage
//     instance it cannot cover (divide_shield.go);
//   - Then, the CR 615.5 follow-up body, owed once per damage instance
//     with what was prevented (PreventionFollowUp).
//
// A charged shield with no source at all stays ModPreventDamage: Register
// refuses a preventFromSource with an Amount and neither a source nor a
// property, so the two kinds never describe the same shield (§7 decision
// 3).
//
// "Prevent all combat damage that would be dealt to and dealt by that
// creature" (Maze of Ith) is ONE prevention effect, so it is one record
// with one mod (ADR 0108 Delivery PR 7): the creature is pinned as the
// record's affected object and Mod.AndDealtBy makes the pin name both
// what the shield protects and the source it stops. Two records would be
// two effects: an event from one pinned creature to another would meet
// both, against CR 614.5's one opportunity per event, and a "whenever
// damage is prevented" trigger would see two applications to one
// instance where CR 615.13 sees one. The affected set may hold several
// objects ("those creatures", Energy Arc; "up to two target creatures",
// Redeem), still one effect.

// ModPreventFromSource is "prevent all [combat] damage [<source>] would
// deal [to <protected>] this turn" (Amount 0), or "prevent the next N
// damage <source> would deal [to <protected>] this turn" (Amount N,
// CR 615.7). Reads Objects, SourceZone, Queries, Player, Types,
// CombatOnly, Amount and Then. Scope ScopeGame for a protected player or
// none; pinned (ScopeNone) to a protected permanent.
const ModPreventFromSource ModKind = "preventFromSource"

// DamageShield is the queue-side description of a ModPreventFromSource
// record. Its fields mean what NextDamageShield's do; the two below are
// the kind's own.
type DamageShield struct {
	// EffectSource is the card whose spell or ability made the shield.
	EffectSource uuid.UUID
	// Controller is "you": the shield's controller and its follow-up's.
	Controller uuid.UUID

	// Source is the chosen or targeted source, built with
	// DamageSourceRefLocked; the zero ref is "any source" (with Queries:
	// any source with those properties).
	Source     ObjectRef
	SourceZone ZoneKind

	// Queries is the CR 615.9 recheck. Empty is no recheck.
	Queries []PermanentQuery

	// ProtectPlayer, ProtectTypes and ProtectPermanent are what the
	// shield protects; all zero protects everything.
	ProtectPlayer    uuid.UUID
	ProtectTypes     []string
	ProtectPermanent uuid.UUID

	// ProtectPermanents are more protected permanents beside
	// ProtectPermanent, all of them in the one record: "up to two target
	// creatures" (Redeem), "those permanents" (Mutational Advantage). The
	// set is fixed as the shield is made (CR 611.2c), and each member is
	// pinned as the object it is now (CR 400.7).
	ProtectPermanents []uuid.UUID

	// AndDealtBy is "dealt to and dealt by": the protected permanents are
	// also the sources whose damage the shield prevents (Maze of Ith). It
	// needs a protected permanent and refuses a source, a property, a
	// player and a charge.
	AndDealtBy bool

	// CombatOnly narrows the shield to combat damage.
	CombatOnly bool

	// Amount is CR 615.7's charge: "the next N damage". Zero is all
	// damage this turn.
	Amount int

	// Then is the CR 615.5 follow-up, run with the amount prevented.
	Then BodyRef
	// To is the player or permanent the follow-up deals its damage to,
	// chosen as the shield was made (Refraction Trap's "any target", ADR
	// 0108 §9); uuid.Nil is none. Pinned to the object it is now.
	To uuid.UUID

	// UntilYourNextTurn is Gideon of the Trials' "until your next turn"
	// (CR 611.2b), "you" being Controller. False is "this turn" (CR
	// 514.2), which every other printed member says.
	UntilYourNextTurn bool

	// Label is the record's label.
	Label string
}

// PreventDamageFromSourceThisTurnForEffect registers the shield until
// cleanup (CR 514.2: "this turn"), or until its controller's next turn
// begins with UntilYourNextTurn (CR 611.2b, Gideon of the Trials). Reports
// whether a record was written: nothing is registered for a charged
// shield with neither a source nor a property, a negative charge, or a
// protected permanent that is not on the battlefield.
//
// Caller must hold g.mu (write) — every caller is a resolving effect.
func (g *Game) PreventDamageFromSourceThisTurnForEffect(s DamageShield) bool {
	if s.Amount < 0 || (s.Amount > 0 && s.Source.ID == uuid.Nil && len(s.Queries) == 0) {
		return false
	}
	protected := s.ProtectPermanents
	if s.ProtectPermanent != uuid.Nil {
		protected = append([]uuid.UUID{s.ProtectPermanent}, s.ProtectPermanents...)
	}
	if s.AndDealtBy && (len(protected) == 0 || s.Source.ID != uuid.Nil || len(s.Queries) > 0 ||
		s.ProtectPlayer != uuid.Nil || s.Amount != 0) {
		return false
	}
	m := Mod{
		Kind:       ModPreventFromSource,
		SourceZone: s.SourceZone,
		Queries:    clonePermanentQueries(s.Queries),
		Player:     s.ProtectPlayer,
		Types:      copyStrings(s.ProtectTypes),
		CombatOnly: s.CombatOnly,
		Amount:     s.Amount,
		Then:       s.Then.key,
		AndDealtBy: s.AndDealtBy,
	}
	if s.Source.ID != uuid.Nil {
		m.Objects = []ObjectRef{s.Source}
	} else {
		m.SourceZone = ""
	}
	ShieldFollowUp{Body: s.Then, To: s.To}.mod(g, &m)
	label := s.Label
	if label == "" {
		label = "prevent damage from a source this turn"
	}
	d := g.UntilEndOfTurnDuration()
	if s.UntilYourNextTurn {
		d = g.UntilYourNextTurnDuration(s.Controller)
	}
	if len(protected) > 0 {
		affected := g.PinnedObjectsLocked(protected...)
		if len(affected) == 0 {
			return false
		}
		// One protected permanent: the record ends as that object does.
		// Several: each member stops matching as it stops being the
		// object it was (CR 400.7), and the record lasts its duration.
		if len(affected) == 1 {
			d = g.PinnedTo(d, affected[0].ID)
		}
		return g.appendScopedEffectLocked(s.EffectSource, affected, ScopeNone, s.Controller, []Mod{m},
			d, label, timeNowUnixNano())
	}
	return g.RegisterScopedRuleEffectForEffect(s.EffectSource, ScopeGame, s.Controller, []Mod{m}, d, label)
}

// fromSourceModProblem is registration's and restore's check on a
// ModPreventFromSource mod. "" when it is sound.
func fromSourceModProblem(m Mod) string {
	if len(m.Objects) > 1 {
		return "a preventFromSource shield names more than one source"
	}
	if len(m.Objects) == 1 && m.Objects[0].ID == uuid.Nil {
		return "a preventFromSource shield names an empty source"
	}
	if len(m.Objects) == 0 && m.SourceZone != "" {
		return "a preventFromSource shield names a source zone and no source"
	}
	if m.Amount < 0 {
		return fmt.Sprintf("a preventFromSource shield has a negative charge, %d", m.Amount)
	}
	if m.Amount > 0 && len(m.Objects) == 0 && len(m.Queries) == 0 {
		return "a charged preventFromSource shield names neither a source nor a property (that shield is preventDamage)"
	}
	for _, q := range m.Queries {
		if q.Enchanted {
			return "a preventFromSource recheck cannot ask whether a source is enchanted"
		}
	}
	if m.SpentBatch != 0 || m.SpentInstance != 0 || m.Half {
		return "a preventFromSource shield carries a next-time field (spentBatch, spentInstance or half)"
	}
	if m.AndDealtBy && (len(m.Objects) != 0 || len(m.Queries) != 0 || m.Player != uuid.Nil || len(m.Types) != 0 || m.Amount != 0) {
		return "a to-and-by preventFromSource shield names a source, a property, a player or a charge; its pinned objects are both"
	}
	if m.Then != "" && !KnownEffectBody(m.Then) {
		return fmt.Sprintf("preventFromSource names follow-up body %q, which is not registered", m.Then)
	}
	return ""
}

// fromSourceMeetsLocked is the kind's applicability, before a CR 615.7
// division is consulted (scopedReplacementAppliesLocked adds that for a
// charged shield): damage this shield's source would deal to what it
// protects, rechecked as the damage would be dealt (CR 615.9).
//
// Caller must hold g.mu.
func (g *Game) fromSourceMeetsLocked(e ScopedEffect, m Mod, ev *ReplacementEvent) bool {
	if ev.Kind != RepEventDamage || ev.DamageAmount <= 0 {
		return false
	}
	if m.CombatOnly && !ev.IsCombatDamage {
		return false
	}
	if m.AndDealtBy {
		// "Dealt to and dealt by": a pinned object on either end of the
		// event. A record that is not pinned has no objects, and meets
		// nothing.
		return scopedAffectsLiveObjectLocked(g, e, ev.DamageTarget) ||
			scopedAffectsLiveObjectLocked(g, e, ev.DamageSource)
	}
	if len(m.Objects) == 1 && !g.damageFromChosenSourceLocked(m, ev.DamageSource) {
		return false
	}
	if len(m.Queries) > 0 && !queriesMatchCharacteristic(m.Queries, ev.SourceLKI) {
		return false
	}
	return g.nextFromSourceProtectsLocked(e, m, ev.DamageTarget)
}

// applyFromSourceLocked is the kind's Replace. All damage this turn
// cancels the event; a charge prevents what it can (CR 615.7, through the
// division when one was made). Either way the amount is added to the
// follow-up owed for the instance (CR 615.5).
//
// Caller must hold g.mu (write).
func (g *Game) applyFromSourceLocked(e ScopedEffect, mod int, m Mod, ev *ReplacementEvent) {
	damage := ev.DamageAmount
	var prevented int
	if m.Amount > 0 {
		prevented = g.applyChargedShieldLocked(e, mod, m, ev)
	} else {
		prevented = ev.DamageAmount
		ev.Cancel()
	}
	g.queuePreventionFollowUpLocked(e, m, ev, prevented, damage)
}

// DamageShieldLabels lists the live preventFromSource shields for the
// game banner (ADR 0108 "Shared machinery", the client), oldest first:
// the card that made each one and, when it names one, its source —
// "Pay No Heed (Goblin Guide)". Empty on nearly every turn.
//
// Caller must hold g.mu (read or write).
func (g *Game) DamageShieldLabels() []string {
	var out []string
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		for _, m := range e.Mods {
			if m.Kind != ModPreventFromSource {
				continue
			}
			line := scopedEffectDisplayName(e)
			if len(m.Objects) == 1 {
				if c, ok := g.LookupCardForEffect(m.Objects[0].ID); ok && c.Name != "" && !c.FaceDown {
					line += " (" + c.Name + ")"
				}
			}
			if m.Amount > 0 {
				line += fmt.Sprintf(" — %d left", m.Amount)
			}
			out = append(out, line)
		}
	}
	return out
}
