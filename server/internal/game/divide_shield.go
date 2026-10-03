package game

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
)

// divide_shield.go — ADR 0108 §7 decision 6, owner decision 1 (#1904):
// a charged shield that meets several damage events of one instance is
// divided by the player it protects, before any of them is applied.
//
// THE RULE. CR 615.7: "If damage would be dealt to the shielded permanent
// or player by two or more applicable sources at the same time, the
// player or the controller of the permanent chooses which damage the
// shield prevents." The Harm's Way ruling says the same of a charged
// redirection: "You choose which 2 damage is redirected." "At the same
// time" is one damage instance (damage_instance.go, owner decision 4).
//
// WHAT IS CHARGED. Every shield whose Replace spends a CR 615.7 charge:
// ModPreventDamage (Mending Hands) and ModPreventFromSource with an
// Amount (Healing Grace). chargedShieldCharge is the one list; a later
// charged kind (§9's redirectDamage with an Amount) is one more case
// there, and the rest of this file serves it unchanged.
//
// WHEN IT IS ASKED. Only when the shield meets two or more events of the
// instance and their total is more than the charge left. When the charge
// covers them all there is nothing to choose, and nothing is asked. An
// event whose damage can't be prevented is not offered: the shield
// prevents none of it whatever the player says (CR 615.12), and is not
// reduced by it.
//
// THE EVENTS UP FRONT. The question comes before any event of the
// instance is applied, so the instance's events have to be known first.
// The engine deals an instance as a sequence of calls, so this file
// collects them at the three places an instance of several events is
// opened:
//
//   - an Each walk (DealDamageEachThenForEffect and its per-recipient
//     sibling) knows its recipients before its first leg: the walk is
//     asked about up front (walkDivisionEventsLocked) and starts once the
//     division is made;
//   - a DamageInstanceForEffect scope (a printed "each" written as a
//     loop) STAGES each event a live charged shield could meet, and
//     deals it when the scope ends, after the question. Only those: an
//     event no charged shield could meet is dealt as it always was, so a
//     card body that reads the board after a deal sees what it saw before
//     (a CR 616 pause already defers a deal the same way);
//   - a combat damage step stages every event of the step while a
//     charged shield is live, so the step's damage is still dealt all at
//     once (CR 510.2) after the question. A CR 510.1c assignment prompt
//     answered later in the step is its own group.
//
// A staged event is plain data (stagedDamage): the event's fields and its
// tail's snapshot of the source, taken as it was opened, exactly as a
// CR 616 pause keeps them. Only an event with no continuation is staged —
// a deal whose caller waits on it (a `...ThenForEffect` form) goes
// through the pipeline at once, because its continuation is the rest of
// a card. The Each walk is the one such caller that is asked about up
// front instead.
//
// THE ANSWER. A division names, for each event (by source and
// recipient), how much of the charge it gets; the shares add up to the
// charge, and none is more than its event. It is written to
// Game.shieldDivisions (plain data) for the instance, and the shield's
// AppliesTo and Replace read it: an event given nothing is not met by the
// shield at all, and an event given N has at most N prevented. The
// shares are spent as the events land, copy on write, so an undo rewinds
// them, and dropped as the instance ends.
//
// THE AMOUNTS. The events are offered at the amounts they are opened
// with, before any other replacement has applied to them: a doubler the
// affected player applies first may make an event larger than its share,
// and then the share is what the shield prevents (CR 615.7 counts only
// damage). The share is a cap, never a promise.

// PendingChoiceDivideShield is "divide this shield among the damage it
// meets" (CR 615.7). Answered with resolve_choice's distribution payload:
// {distribution: {entry id: share}}, the shares adding up to the charge.
const PendingChoiceDivideShield PendingChoiceKind = "divide_shield"

// chargedShieldCharge reports whether m is a CR 615.7 charged shield, and
// the charge it has left.
func chargedShieldCharge(m Mod) (int, bool) {
	switch m.Kind {
	case ModPreventDamage, ModPreventFromSource:
		return m.Amount, m.Amount > 0
	}
	return 0, false
}

// --- the division registry -----------------------------------------

// ShieldDivision is one share of a divided shield: the most the shield
// (Seq, Mod) may prevent of the damage Source deals Target in Instance.
// Plain data, never captured: the events it divides land in the action
// that answered the prompt, or wait on a CR 616 prompt, which is not a
// restore point either.
type ShieldDivision struct {
	Seq      int64
	Mod      int
	Instance DamageInstance
	Source   uuid.UUID
	Target   uuid.UUID
	Amount   int
}

// divisionDecidedLocked reports whether shield (seq, mod) has been divided
// for instance `inst`.
//
// Caller must hold g.mu.
func (g *Game) divisionDecidedLocked(seq int64, mod int, inst DamageInstance) bool {
	if inst == 0 {
		return false
	}
	return slices.ContainsFunc(g.shieldDivisions, func(d ShieldDivision) bool {
		return d.Seq == seq && d.Mod == mod && d.Instance == inst
	})
}

// divisionShareLocked is the share of a divided shield an event may still
// use, and whether the shield was divided for its instance at all.
//
// Caller must hold g.mu.
func (g *Game) divisionShareLocked(seq int64, mod int, ev *ReplacementEvent) (share int, divided bool) {
	if !g.divisionDecidedLocked(seq, mod, ev.DamageInstance) {
		return 0, false
	}
	for _, d := range g.shieldDivisions {
		if d.Seq == seq && d.Mod == mod && d.Instance == ev.DamageInstance &&
			d.Source == ev.DamageSource && d.Target == ev.DamageTarget {
			share += d.Amount
		}
	}
	return share, true
}

// chargedShieldOpenToLocked is the division's half of a charged shield's
// AppliesTo: an undivided shield meets every event its kind does; a
// divided one only the events it was given a share of (CR 615.7: the
// player chose which damage the shield prevents).
//
// Caller must hold g.mu.
func (g *Game) chargedShieldOpenToLocked(seq int64, mod int, ev *ReplacementEvent) bool {
	share, divided := g.divisionShareLocked(seq, mod, ev)
	return !divided || share > 0
}

// applyChargedShieldLocked is the CR 615.7 arithmetic every charged
// shield shares: prevent as much of the event as the charge — and the
// share, when the shield was divided — allows, reduce the charge by it,
// and report what was prevented. A 4-point shield facing 6 damage
// prevents 4 and lets 2 through; facing 3 it prevents all 3 and keeps 1.
//
// Caller must hold g.mu (write).
func (g *Game) applyChargedShieldLocked(e ScopedEffect, mod int, m Mod, ev *ReplacementEvent) int {
	limit := m.Amount
	if share, divided := g.divisionShareLocked(e.Seq, mod, ev); divided {
		limit = min(limit, share)
	}
	prevented := min(limit, ev.DamageAmount)
	if prevented <= 0 {
		return 0
	}
	if prevented == ev.DamageAmount {
		ev.Cancel()
	} else {
		ev.DamageAmount -= prevented
	}
	g.setShieldChargeLocked(e.Seq, mod, m.Amount-prevented)
	g.spendDivisionLocked(e.Seq, mod, ev, prevented)
	return prevented
}

// spendDivisionLocked takes `n` off the event's share. COPY ON WRITE, the
// registry's rule: a clone taken before the damage shares the slice.
//
// Caller must hold g.mu (write).
func (g *Game) spendDivisionLocked(seq int64, mod int, ev *ReplacementEvent, n int) {
	if n <= 0 || !g.divisionDecidedLocked(seq, mod, ev.DamageInstance) {
		return
	}
	next := slices.Clone(g.shieldDivisions)
	for i := range next {
		d := &next[i]
		if n == 0 {
			break
		}
		if d.Seq != seq || d.Mod != mod || d.Instance != ev.DamageInstance ||
			d.Source != ev.DamageSource || d.Target != ev.DamageTarget {
			continue
		}
		take := min(n, d.Amount)
		d.Amount -= take
		n -= take
	}
	g.shieldDivisions = next
}

// dropShieldDivisionsLocked forgets the divisions of instance `inst` once
// it has ended — or, with zero, of every instance that is not still
// waiting on a prompt (as play moves on, beginEventBatchLocked).
//
// Caller must hold g.mu (write).
func (g *Game) dropShieldDivisionsLocked(inst DamageInstance) {
	if len(g.shieldDivisions) == 0 {
		return
	}
	var kept []ShieldDivision
	for _, d := range g.shieldDivisions {
		ended := d.Instance == inst
		if inst == 0 {
			ended = !g.damageInstancePausedLocked(d.Instance)
		}
		if !ended {
			kept = append(kept, d)
		}
	}
	g.shieldDivisions = kept
}

// --- staging an instance's events -----------------------------------

// damageStage is an open group of one instance's damage events being
// collected so a charged shield can be divided among them before any is
// dealt. Plain data; set and cleared inside one mutation.
type damageStage struct {
	inst DamageInstance
	// all stages every event of the group (a combat damage step), not
	// only those a charged shield could meet (a scope).
	all bool
	// scope marks a DamageInstanceForEffect group, whose instance ends
	// once its staged events have landed.
	scope  bool
	events []stagedDamage
}

// stagedDamage is one staged damage event: the event's fields and its
// tail's snapshot of the source (damageTail), minus the continuation —
// an event with one is never staged. TestStagedDamageCarriesTheWholeTail
// holds the copy to damageTail's fields.
type stagedDamage struct {
	source       uuid.UUID
	damageSource uuid.UUID
	target       uuid.UUID
	amount       int
	combat       bool
	inst         DamageInstance
	lki          *Characteristic

	kind                damageTailKind
	tailCombat          bool
	combatStep          string
	actor               uuid.UUID
	sourceLKI           *Characteristic
	deathtouch          bool
	lifelinkTo          uuid.UUID
	commanderSource     uuid.UUID
	result              DamageResultSource
	controller          uuid.UUID
	marks               DamageMarks
	sourceUnpreventable bool
	sourceChecked       bool
}

func stagedDamageOf(ev *ReplacementEvent) stagedDamage {
	t := ev.damageTail
	return stagedDamage{
		source:              ev.Source,
		damageSource:        ev.DamageSource,
		target:              ev.DamageTarget,
		amount:              ev.DamageAmount,
		combat:              ev.IsCombatDamage,
		inst:                ev.DamageInstance,
		lki:                 ev.SourceLKI,
		kind:                t.kind,
		tailCombat:          t.combat,
		combatStep:          t.combatStep,
		actor:               t.actor,
		sourceLKI:           t.sourceLKI,
		deathtouch:          t.deathtouch,
		lifelinkTo:          t.lifelinkTo,
		commanderSource:     t.commanderSource,
		result:              t.result,
		controller:          t.controller,
		marks:               t.marks,
		sourceUnpreventable: t.sourceUnpreventable,
		sourceChecked:       t.sourceChecked,
	}
}

// event rebuilds the staged event, fresh each time, marked released so it
// is never staged again.
func (s stagedDamage) event() *ReplacementEvent {
	return &ReplacementEvent{
		Kind:           RepEventDamage,
		Source:         s.source,
		DamageSource:   s.damageSource,
		DamageTarget:   s.target,
		DamageAmount:   s.amount,
		IsCombatDamage: s.combat,
		DamageInstance: s.inst,
		SourceLKI:      s.lki,
		damageTail: &damageTail{
			kind:                s.kind,
			combat:              s.tailCombat,
			combatStep:          s.combatStep,
			actor:               s.actor,
			sourceLKI:           s.sourceLKI,
			deathtouch:          s.deathtouch,
			lifelinkTo:          s.lifelinkTo,
			commanderSource:     s.commanderSource,
			result:              s.result,
			controller:          s.controller,
			marks:               s.marks,
			sourceUnpreventable: s.sourceUnpreventable,
			sourceChecked:       s.sourceChecked,
			released:            true,
		},
	}
}

// anyChargedShieldLocked reports whether a live record holds a charged
// shield — the cheap gate in front of every stage, so a game with none
// deals its damage exactly as it always has.
//
// Caller must hold g.mu.
func (g *Game) anyChargedShieldLocked() bool {
	for i := range g.ScopedEffects {
		if g.ScopedEffects[i].Seq == 0 {
			continue
		}
		for _, m := range g.ScopedEffects[i].Mods {
			if _, ok := chargedShieldCharge(m); ok {
				return true
			}
		}
	}
	return false
}

// openDamageStageLocked opens a group for instance `inst` and returns the
// one it replaces, for closeDamageStageLocked to restore. Nothing is
// opened while no charged shield is live.
//
// Caller must hold g.mu (write).
func (g *Game) openDamageStageLocked(inst DamageInstance, all, scope bool) (prev *damageStage, opened bool) {
	prev = g.damageStage
	if inst == 0 || !g.anyChargedShieldLocked() {
		return prev, false
	}
	g.damageStage = &damageStage{inst: inst, all: all, scope: scope}
	return prev, true
}

// stageDamageEventLocked stages `ev` when an open group collects it: an
// event of the group's instance, with no continuation, that a charged
// shield could meet — or any such event, for a combat group. Reports
// whether it was staged; the caller then treats it as paused.
//
// Caller must hold g.mu (write).
func (g *Game) stageDamageEventLocked(ev *ReplacementEvent) bool {
	st := g.damageStage
	t := ev.damageTail
	if st == nil || t == nil || t.released || t.then != nil || t.endsInstance {
		return false
	}
	if ev.DamageInstance == 0 || ev.DamageInstance != st.inst || ev.DamageAmount <= 0 {
		return false
	}
	if !st.all && !g.chargedShieldCouldMeetLocked(ev) {
		return false
	}
	st.events = append(st.events, stagedDamageOf(ev))
	return true
}

// chargedShieldCouldMeetLocked reports whether a live, undivided charged
// shield meets `ev`.
//
// Caller must hold g.mu.
func (g *Game) chargedShieldCouldMeetLocked(ev *ReplacementEvent) bool {
	for i := range g.ScopedEffects {
		e := g.ScopedEffects[i]
		if e.Seq == 0 {
			continue
		}
		for j, m := range e.Mods {
			if _, ok := chargedShieldCharge(m); !ok {
				continue
			}
			if !g.divisionDecidedLocked(e.Seq, j, ev.DamageInstance) && scopedReplacementMeetsLocked(g, e, m, ev) {
				return true
			}
		}
	}
	return false
}

// closeDamageStageLocked closes the open group and deals what it staged:
// at once when no shield needs dividing, otherwise once the protected
// players have answered. Restores `prev`.
//
// Caller must hold g.mu (write).
func (g *Game) closeDamageStageLocked(prev *damageStage) {
	st := g.damageStage
	g.damageStage = prev
	if st == nil || len(st.events) == 0 {
		return
	}
	records := slices.Clone(st.events)
	var end DamageInstance
	if st.scope {
		end = st.inst
	}
	evs := make([]*ReplacementEvent, len(records))
	for i, r := range records {
		evs[i] = r.event()
	}
	land := landStagedDamage(records, end)
	if err := g.askShieldDivisionsLocked(g.shieldDivisionNeedsLocked(evs), land); err != nil {
		g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: "staged damage: " + err.Error()})
	}
}

// landStagedDamage deals staged events in the order they were opened,
// each through the CR 614 pipeline and its own tail. A package-level
// constructor capturing plain data: each run rebuilds the events, so an
// undo across the prompt and a second answer deal them again, once.
// `end`, when set, is a scope's instance, ended once its events have
// landed (CR 615.5).
func landStagedDamage(records []stagedDamage, end DamageInstance) func(*Game) error {
	return func(g *Game) error {
		for _, r := range records {
			// A target that has gone deals nothing (the resume's rule);
			// the instruction that opened it has returned already.
			_, _ = g.damageThroughReplacementsLocked(r.event())
		}
		if end != 0 {
			g.endDamageInstanceLocked(end)
		}
		return nil
	}
}

// --- what needs dividing --------------------------------------------

// divideNeed is one charged shield that has to be divided among the
// events of one instance.
type divideNeed struct {
	seq     int64
	mod     int
	inst    DamageInstance
	chooser uuid.UUID
	source  uuid.UUID
	label   string
	charge  int
	entries []DivideShieldEntry
}

// shieldDivisionNeedsLocked lists the live charged shields that meet two
// or more of `evs` (all of one instance), with more damage between them
// than the charge left, in record order. Events of one source and
// recipient are one entry.
//
// Caller must hold g.mu (write: entry names are looked up).
func (g *Game) shieldDivisionNeedsLocked(evs []*ReplacementEvent) []divideNeed {
	if len(evs) < 2 {
		return nil
	}
	inst := evs[0].DamageInstance
	if inst == 0 {
		return nil
	}
	var needs []divideNeed
	for i := range g.ScopedEffects {
		e := g.ScopedEffects[i]
		if e.Seq == 0 {
			continue
		}
		for j, m := range e.Mods {
			charge, ok := chargedShieldCharge(m)
			if !ok || g.divisionDecidedLocked(e.Seq, j, inst) {
				continue
			}
			var entries []DivideShieldEntry
			total := 0
			for _, ev := range evs {
				if ev.DamageAmount <= 0 || ev.DamageInstance != inst || !scopedReplacementMeetsLocked(g, e, m, ev) {
					continue
				}
				if g.damageUnpreventableLocked(ev) {
					continue
				}
				total += ev.DamageAmount
				if k := slices.IndexFunc(entries, func(x DivideShieldEntry) bool {
					return x.Source == ev.DamageSource && x.Target == ev.DamageTarget
				}); k >= 0 {
					entries[k].Amount += ev.DamageAmount
					continue
				}
				entries = append(entries, g.divideShieldEntryLocked(ev))
			}
			if len(entries) < 2 || total <= charge {
				continue
			}
			needs = append(needs, divideNeed{
				seq:     e.Seq,
				mod:     j,
				inst:    inst,
				chooser: g.divideShieldChooserLocked(e, entries),
				source:  e.Source.ID,
				label:   scopedEffectDisplayName(&e),
				charge:  charge,
				entries: entries,
			})
		}
	}
	return needs
}

// divideShieldEntryLocked describes one event for the prompt.
//
// Caller must hold g.mu.
func (g *Game) divideShieldEntryLocked(ev *ReplacementEvent) DivideShieldEntry {
	en := DivideShieldEntry{
		ID:     uuid.New(),
		Source: ev.DamageSource,
		Target: ev.DamageTarget,
		Amount: ev.DamageAmount,
		Combat: ev.IsCombatDamage,
	}
	// The names are public, so a face-down permanent is not named (the
	// prompt reaches every seat).
	if c, ok := g.LookupCardForEffect(ev.DamageSource); ok {
		if c.FaceDown {
			en.SourceName = "A face-down permanent"
		} else {
			en.SourceName = c.Name
		}
	} else if ev.SourceLKI != nil {
		en.SourceName = ev.SourceLKI.Name
	}
	if p := g.playerByIDLocked(ev.DamageTarget); p != nil {
		en.TargetName = p.Name
		en.TargetIsPlayer = true
	} else if c := findBattlefieldCard(g, ev.DamageTarget); c != nil {
		en.TargetName = c.Name
		if c.FaceDown {
			en.TargetName = "a face-down permanent"
		}
	}
	return en
}

// divideShieldChooserLocked is CR 615.7's "the player or the controller of
// the permanent": the one player every entry's recipient is or is
// controlled by. A shield whose recipients belong to different players
// (one that protects anything) is divided by its controller.
//
// Caller must hold g.mu.
func (g *Game) divideShieldChooserLocked(e ScopedEffect, entries []DivideShieldEntry) uuid.UUID {
	var who uuid.UUID
	for _, en := range entries {
		var owner uuid.UUID
		if p := g.playerByIDLocked(en.Target); p != nil {
			owner = p.ID
		} else if c := findBattlefieldCard(g, en.Target); c != nil {
			owner = c.Controller
		}
		switch {
		case who == uuid.Nil:
			who = owner
		case owner != who:
			if e.Controller != uuid.Nil {
				return e.Controller
			}
			return who
		}
	}
	return who
}

// walkDivisionEventsLocked builds, without dealing them, the events an
// Each walk would open — one per recipient that is still a player in the
// game or a permanent — so the walk's shields can be divided before its
// first leg.
//
// Caller must hold g.mu (write).
func (g *Game) walkDivisionEventsLocked(source uuid.UUID, targets []uuid.UUID, amount int, inst DamageInstance) []*ReplacementEvent {
	if amount <= 0 || inst == 0 || len(targets) < 2 || !g.anyChargedShieldLocked() {
		return nil
	}
	var out []*ReplacementEvent
	for _, id := range targets {
		var kind damageTailKind
		if p := g.playerByIDLocked(id); p != nil {
			if p.Eliminated {
				continue
			}
			kind = damageTailPlayer
		} else if findBattlefieldCard(g, id) != nil {
			kind = damageTailPermanent
		} else {
			continue
		}
		ev := &ReplacementEvent{
			Kind:           RepEventDamage,
			Source:         source,
			DamageSource:   source,
			DamageTarget:   id,
			DamageAmount:   amount,
			DamageInstance: inst,
			damageTail:     g.effectDamageTailLocked(kind, source, nil),
		}
		g.prepareDamageEventLocked(ev)
		out = append(out, ev)
	}
	return out
}

// --- the prompt -----------------------------------------------------

// DivideShieldPrompt is a PendingChoiceDivideShield's payload.
type DivideShieldPrompt struct {
	// Label names the shield (its card); Charge is what it has left to
	// divide, the total the answer's shares must reach.
	Label  string
	Charge int
	// Entries are the damage events the shield meets, in the order they
	// were opened: one per source and recipient.
	Entries []DivideShieldEntry

	seq  int64
	mod  int
	inst DamageInstance
}

// DivideShieldEntry is one event a divided shield meets.
type DivideShieldEntry struct {
	// ID keys the answer's distribution.
	ID uuid.UUID
	// Source deals Amount to Target (a player when TargetIsPlayer);
	// SourceName and TargetName are what the table calls them.
	Source         uuid.UUID
	SourceName     string
	Target         uuid.UUID
	TargetName     string
	TargetIsPlayer bool
	Amount         int
	Combat         bool
}

// askShieldDivisionsLocked asks the first need's protected player to
// divide it, and so on down the list, then runs `land`. With no need
// left, `land` runs at once. A chooser who cannot be asked (gone from the
// game) gets the engine's order (DefaultShieldDivision), so the damage
// still lands.
//
// Caller must hold g.mu (write).
func (g *Game) askShieldDivisionsLocked(needs []divideNeed, land func(*Game) error) error {
	for len(needs) > 0 {
		n, rest := needs[0], needs[1:]
		if _, ok := g.scopedEffectIndexBySeqLocked(n.seq); !ok {
			needs = rest
			continue
		}
		prompt := &DivideShieldPrompt{Label: n.label, Charge: n.charge, Entries: n.entries, seq: n.seq, mod: n.mod, inst: n.inst}
		if who := g.playerByIDLocked(n.chooser); who == nil || who.Eliminated {
			g.writeShieldDivisionLocked(prompt, DefaultShieldDivision(prompt))
			needs = rest
			continue
		}
		g.QueueChoiceForEffect(PendingChoice{
			Kind:         PendingChoiceDivideShield,
			Chooser:      n.chooser,
			FromPlayer:   n.chooser,
			Source:       n.source,
			Reason:       fmt.Sprintf("%s — divide %d prevention among the damage", n.label, n.charge),
			DivideShield: prompt,
			confirmResume: &confirmFrame{
				onAccept:  continueShieldDivisions(rest, land),
				onDecline: defaultShieldDivisionThen(prompt, rest, land),
			},
		})
		return nil
	}
	if land == nil {
		return nil
	}
	return land(g)
}

// continueShieldDivisions is the answered prompt's continuation: the next
// need, or the damage.
func continueShieldDivisions(rest []divideNeed, land func(*Game) error) func(*Game) error {
	return func(g *Game) error { return g.askShieldDivisionsLocked(rest, land) }
}

// defaultShieldDivisionThen is the dropped prompt's continuation: the
// engine's order for this shield, then on as an answer would go.
func defaultShieldDivisionThen(p *DivideShieldPrompt, rest []divideNeed, land func(*Game) error) func(*Game) error {
	return func(g *Game) error {
		g.writeShieldDivisionLocked(p, DefaultShieldDivision(p))
		return g.askShieldDivisionsLocked(rest, land)
	}
}

// DefaultShieldDivision is the engine's order: each event in the order it
// was opened takes all it can until the charge is gone. Always a legal
// answer.
func DefaultShieldDivision(p *DivideShieldPrompt) map[uuid.UUID]int {
	if p == nil {
		return nil
	}
	out := make(map[uuid.UUID]int, len(p.Entries))
	left := p.Charge
	for _, en := range p.Entries {
		give := min(left, en.Amount)
		out[en.ID] = give
		left -= give
	}
	return out
}

// CheckShieldDivision reports whether `dist` divides the prompt's shield:
// entries of this prompt only, no share below zero or above its event,
// and shares adding up to the charge (CR 615.7: every 1 damage the shield
// can prevent, it does). Missing entries get nothing.
func CheckShieldDivision(p *DivideShieldPrompt, dist map[uuid.UUID]int) error {
	if p == nil {
		return ErrInvalidParam
	}
	total := 0
	for id, n := range dist {
		k := slices.IndexFunc(p.Entries, func(en DivideShieldEntry) bool { return en.ID == id })
		if k < 0 || n < 0 || n > p.Entries[k].Amount {
			return ErrInvalidParam
		}
		total += n
	}
	if total != p.Charge {
		return ErrInvalidParam
	}
	return nil
}

// writeShieldDivisionLocked records a division: one share per entry, the
// entries given nothing included, so the shield meets none of them.
//
// Caller must hold g.mu (write).
func (g *Game) writeShieldDivisionLocked(p *DivideShieldPrompt, dist map[uuid.UUID]int) {
	next := slices.Clone(g.shieldDivisions)
	for _, en := range p.Entries {
		next = append(next, ShieldDivision{
			Seq: p.seq, Mod: p.mod, Instance: p.inst,
			Source: en.Source, Target: en.Target, Amount: dist[en.ID],
		})
	}
	g.shieldDivisions = next
}

// ResolveDivideShield answers a PendingChoiceDivideShield with the shares
// the protected player gives each event. A division that does not add up
// is refused with the prompt still open.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveDivideShield(choiceID, chooserID uuid.UUID, dist map[uuid.UUID]int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx, choice := g.findChoiceLocked(choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	if choice.Kind != PendingChoiceDivideShield {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	if err := CheckShieldDivision(choice.DivideShield, dist); err != nil {
		return err
	}
	frame := choice.confirmResume
	source := choice.Source
	g.dequeueChoiceLocked(idx)
	g.writeShieldDivisionLocked(choice.DivideShield, dist)
	if frame != nil && frame.onAccept != nil {
		if err := frame.onAccept(g); err != nil {
			g.emitChoiceEffectErrorLocked(chooserID, source, err)
		}
	}
	g.runStateChecksLocked()
	return nil
}

// settleDroppedDivideShieldLocked is a dropped divide_shield's drop
// action: the engine's order, and the damage lands. The damage is owed to
// the table, not to the player who left, so it runs whoever left.
//
// Caller must hold g.mu (write).
func (g *Game) settleDroppedDivideShieldLocked(c *PendingChoice) {
	if c == nil || c.confirmResume == nil || c.confirmResume.onDecline == nil {
		return
	}
	if err := c.confirmResume.onDecline(g); err != nil {
		g.emitChoiceEffectErrorLocked(c.Chooser, c.Source, err)
	}
}

// cloneDamageStage deep-copies an open group for Clone.
func cloneDamageStage(st *damageStage) *damageStage {
	if st == nil {
		return nil
	}
	out := *st
	out.events = slices.Clone(st.events)
	return &out
}
