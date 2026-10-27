package game

import (
	"slices"

	"github.com/google/uuid"
)

// divide_shield_keyed.go — #4361 S16 (owner decision D3): a divide_shield
// prompt, and the CR 616 prompts the divided damage meets, written to a
// restore point.
//
// Before this, the prompt's answer was a closure pair (confirmResume)
// around the damage pipeline's landing step, and the divisions it wrote
// (Game.shieldDivisions) and the redirection orders a CR 616 answer
// remembers for an instance (Game.redirectOrders) restored empty. So a
// table waiting on a divide_shield prompt was not a restore point, and a
// damage event's frame was kept off the disk whenever a division or an
// order was held for its instance.
//
// Now the landing step is data (divideShieldLand): the staged events of a
// DamageInstanceForEffect scope or a combat step (stagedDamage), with the
// scope's instance to end once they have landed, or a keyed Each walk to
// start (damageEachRun). The prompt carries it, with the shields still to
// divide after it, and both are written with the prompt
// (pendingChoiceSnapshot.DivideShield). The divisions and the orders are
// written too (GameSnapshot.ShieldDivisions, GameSnapshot.RedirectOrders),
// and a damage event released from a division carries the release
// (damageTailSnapshot.Released). All additive within v9.
//
// A written divide_shield prompt carries DivideShieldLandKey in PickThen,
// so a binary from before it refuses the file rather than restoring a
// prompt it can only leave open.

// DivideShieldLandKey marks a divide_shield prompt written with its
// landing step as data. An on-disk identity: never renamed, never reused.
const DivideShieldLandKey = "divide-shield/land"

// divideShieldLand is what a division leads to once every shield of the
// instance is divided: the staged events dealt in the order they were
// opened, then `end` (a scope's instance) ended; or a keyed Each walk
// started. Plain data, never mutated once built, so a Clone shares it.
type divideShieldLand struct {
	staged []stagedDamage
	end    DamageInstance
	walk   *damageEachRun
}

// landDivisionLocked lands `l`. The staged events are rebuilt on every
// run, so an undo across the prompt and a second answer deal them again,
// once. A target that has gone deals nothing (the resume's rule); the
// instruction that opened it has returned already. `end`, when set, is a
// scope's instance, ended once its events have landed (CR 615.5).
//
// Caller must hold g.mu (write).
func (g *Game) landDivisionLocked(l divideShieldLand) error {
	if l.walk != nil {
		return g.dealDamageEachKeyedStepLocked(cloneDamageEachRun(l.walk))
	}
	for _, r := range l.staged {
		_, _ = g.damageThroughReplacementsLocked(r.event())
	}
	if l.end != 0 {
		g.endDamageInstanceLocked(l.end)
	}
	return nil
}

// --- on disk ------------------------------------------------------------

// divideShieldSnapshot is a DivideShieldPrompt on disk: the prompt, the
// shield it divides, the shields after it and the landing step.
type divideShieldSnapshot struct {
	Label    string               `json:"label"`
	Charge   int                  `json:"charge"`
	Entries  []DivideShieldEntry  `json:"entries"`
	Seq      int64                `json:"seq"`
	Mod      int                  `json:"mod"`
	Instance DamageInstance       `json:"instance"`
	Rest     []divideNeedSnapshot `json:"rest,omitempty"`
	// Staged and End, or Walk: the landing step (divideShieldLand).
	Staged []stagedDamageSnapshot `json:"staged,omitempty"`
	End    DamageInstance         `json:"end,omitempty"`
	Walk   *damageEachRun         `json:"walk,omitempty"`
}

// divideNeedSnapshot is a divideNeed on disk.
type divideNeedSnapshot struct {
	Seq      int64               `json:"seq"`
	Mod      int                 `json:"mod"`
	Instance DamageInstance      `json:"instance"`
	Chooser  uuid.UUID           `json:"chooser"`
	Source   uuid.UUID           `json:"source"`
	Label    string              `json:"label"`
	Verb     string              `json:"verb"`
	Charge   int                 `json:"charge"`
	Entries  []DivideShieldEntry `json:"entries"`
}

// stagedDamageSnapshot is a stagedDamage on disk: the event's fields and
// its tail's data half, field for field.
type stagedDamageSnapshot struct {
	Source       uuid.UUID       `json:"source"`
	DamageSource uuid.UUID       `json:"damageSource"`
	Target       uuid.UUID       `json:"target"`
	Amount       int             `json:"amount"`
	Combat       bool            `json:"combat,omitempty"`
	Instance     DamageInstance  `json:"instance"`
	LKI          *Characteristic `json:"lki,omitempty"`

	Kind                damageTailKind     `json:"kind"`
	TailCombat          bool               `json:"tailCombat,omitempty"`
	CombatStep          string             `json:"combatStep,omitempty"`
	Actor               uuid.UUID          `json:"actor"`
	SourceLKI           *Characteristic    `json:"sourceLki,omitempty"`
	Deathtouch          bool               `json:"deathtouch,omitempty"`
	LifelinkTo          uuid.UUID          `json:"lifelinkTo"`
	CommanderSource     uuid.UUID          `json:"commanderSource"`
	Result              DamageResultSource `json:"result,omitempty"`
	Controller          uuid.UUID          `json:"controller"`
	Marks               DamageMarks        `json:"marks,omitempty"`
	SourceUnpreventable bool               `json:"sourceUnpreventable,omitempty"`
	SourceChecked       bool               `json:"sourceChecked,omitempty"`
}

// redirectOrderSnapshot is a redirectOrder on disk.
type redirectOrderSnapshot struct {
	Instance DamageInstance        `json:"instance"`
	Order    []ReplacementEffectID `json:"order"`
}

func snapshotStagedDamage(r stagedDamage) stagedDamageSnapshot {
	return stagedDamageSnapshot{
		Source: r.source, DamageSource: r.damageSource, Target: r.target, Amount: r.amount,
		Combat: r.combat, Instance: r.inst, LKI: copyCharacteristic(r.lki),
		Kind: r.kind, TailCombat: r.tailCombat, CombatStep: r.combatStep, Actor: r.actor,
		SourceLKI: copyCharacteristic(r.sourceLKI), Deathtouch: r.deathtouch, LifelinkTo: r.lifelinkTo,
		CommanderSource: r.commanderSource, Result: r.result, Controller: r.controller, Marks: r.marks,
		SourceUnpreventable: r.sourceUnpreventable, SourceChecked: r.sourceChecked,
	}
}

func restoreStagedDamage(s stagedDamageSnapshot) stagedDamage {
	return stagedDamage{
		source: s.Source, damageSource: s.DamageSource, target: s.Target, amount: s.Amount,
		combat: s.Combat, inst: s.Instance, lki: copyCharacteristic(s.LKI),
		kind: s.Kind, tailCombat: s.TailCombat, combatStep: s.CombatStep, actor: s.Actor,
		sourceLKI: copyCharacteristic(s.SourceLKI), deathtouch: s.Deathtouch, lifelinkTo: s.LifelinkTo,
		commanderSource: s.CommanderSource, result: s.Result, controller: s.Controller, marks: s.Marks,
		sourceUnpreventable: s.SourceUnpreventable, sourceChecked: s.SourceChecked,
	}
}

func snapshotDivideNeed(n divideNeed) divideNeedSnapshot {
	return divideNeedSnapshot{
		Seq: n.seq, Mod: n.mod, Instance: n.inst, Chooser: n.chooser, Source: n.source,
		Label: n.label, Verb: n.verb, Charge: n.charge, Entries: slices.Clone(n.entries),
	}
}

func restoreDivideNeed(s divideNeedSnapshot) divideNeed {
	return divideNeed{
		seq: s.Seq, mod: s.Mod, inst: s.Instance, chooser: s.Chooser, source: s.Source,
		label: s.Label, verb: s.Verb, charge: s.Charge, entries: slices.Clone(s.Entries),
	}
}

// snapshotDivideShield writes a divide_shield prompt's payload, nil for
// none.
func snapshotDivideShield(p *DivideShieldPrompt) *divideShieldSnapshot {
	if p == nil {
		return nil
	}
	out := &divideShieldSnapshot{
		Label: p.Label, Charge: p.Charge, Entries: slices.Clone(p.Entries),
		Seq: p.seq, Mod: p.mod, Instance: p.inst,
		End: p.land.end, Walk: cloneDamageEachRun(p.land.walk),
	}
	for _, n := range p.rest {
		out.Rest = append(out.Rest, snapshotDivideNeed(n))
	}
	for _, r := range p.land.staged {
		out.Staged = append(out.Staged, snapshotStagedDamage(r))
	}
	return out
}

// restoreDivideShield is snapshotDivideShield's inverse.
func restoreDivideShield(s *divideShieldSnapshot) *DivideShieldPrompt {
	if s == nil {
		return nil
	}
	out := &DivideShieldPrompt{
		Label: s.Label, Charge: s.Charge, Entries: slices.Clone(s.Entries),
		seq: s.Seq, mod: s.Mod, inst: s.Instance,
		land: divideShieldLand{end: s.End, walk: cloneDamageEachRun(s.Walk)},
	}
	for _, n := range s.Rest {
		out.rest = append(out.rest, restoreDivideNeed(n))
	}
	for _, r := range s.Staged {
		out.land.staged = append(out.land.staged, restoreStagedDamage(r))
	}
	return out
}

// maxDivideShieldInstance is the largest damage instance a written prompt
// names, for the restored instance counter.
func maxDivideShieldInstance(s *divideShieldSnapshot) uint64 {
	if s == nil {
		return 0
	}
	n := max(uint64(s.Instance), uint64(s.End))
	for _, r := range s.Rest {
		n = max(n, uint64(r.Instance))
	}
	for _, r := range s.Staged {
		n = max(n, uint64(r.Instance))
	}
	if s.Walk != nil {
		n = max(n, uint64(s.Walk.Instance))
	}
	return n
}

// unknownDivideShieldKeys is checkEffectKeys' line for a written
// divide_shield prompt: the marker, and a walk's tails.
func unknownDivideShieldKeys(c *pendingChoiceSnapshot) []string {
	if c.PickThen != "" && c.PickThen != DivideShieldLandKey {
		return []string{"divide-shield landing " + c.PickThen}
	}
	if c.DivideShield == nil {
		return nil
	}
	return unknownDamageEachKeys(c.DivideShield.Walk)
}

// snapshotShieldDivisions writes Game.shieldDivisions.
func snapshotShieldDivisions(in []ShieldDivision) []ShieldDivision {
	return slices.Clone(in)
}

// snapshotRedirectOrders writes Game.redirectOrders.
func snapshotRedirectOrders(in []redirectOrder) []redirectOrderSnapshot {
	var out []redirectOrderSnapshot
	for _, r := range in {
		out = append(out, redirectOrderSnapshot{Instance: r.instance, Order: slices.Clone(r.order)})
	}
	return out
}

// restoreRedirectOrders is snapshotRedirectOrders' inverse.
func restoreRedirectOrders(in []redirectOrderSnapshot) []redirectOrder {
	var out []redirectOrder
	for _, r := range in {
		out = append(out, redirectOrder{instance: r.Instance, order: slices.Clone(r.Order)})
	}
	return out
}

// maxHeldDivisionInstance is the largest instance a restored division or
// redirection order names.
func maxHeldDivisionInstance(divs []ShieldDivision, orders []redirectOrder) uint64 {
	var n uint64
	for _, d := range divs {
		n = max(n, uint64(d.Instance))
	}
	for _, r := range orders {
		n = max(n, uint64(r.instance))
	}
	return n
}
