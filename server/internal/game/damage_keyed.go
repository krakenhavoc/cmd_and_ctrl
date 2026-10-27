package game

import "github.com/google/uuid"

// damage_keyed.go — #4253: damage to a player with the rest of the effect
// as data, and the damage event's frame written to a restore point.
//
// The closure form (DealDamageToPlayerThenForEffect, deleted in #4253's
// cleanup once only tests called it) handed its caller's continuation to the
// damage event as a closure (damageTail.then), and before this no damage
// event's frame was ever written: damageTail held the closure and the
// source's last-known traits, and every damage field of ReplacementEvent
// was a gate (replacementEventFields). So a table paused on a CR 616
// ordering prompt (or a "may" replacement's question) over a damage event
// was not a restore point whatever its caller passed.
//
// Now a non-combat damage event to a player or a permanent is written with
// its frame (replacementEventSnapshot's damage fields and DamageTail, under
// ReplacementResumeDamage) when its tail is data (since #4361 S8 every
// tail's continuation is) and it has no redirected parts. A division or a
// redirection order held for its instance, and its release from a
// division, are written beside it since #4361 S16 (divide_shield_keyed.go),
// so none of them keeps the frame off the disk. The caller's continuation
// is damageTail.keyed, a registered tail told the target and the damage
// dealt, or damageTail.each, the rest of a keyed Each walk
// (damage_each_keyed.go). Combat damage stays a closure frame: the step it
// belongs to is not data.
//
// When the event ends its instance and a prevention follow-up of that
// instance asks a question (#2040), the keyed continuation is parked as
// data too (followUpPark.keyed, GameSnapshot.FollowUpParks), so that
// question is a restore point when its own prompt is keyed. The closure
// park (followUpPark.run) is gone with damageTail.then (#4361 S8).

// ReplacementResumeDamage is the frame key of a damage event (#4253): its
// own key, as no binary before it keyed a damage event at all.
const ReplacementResumeDamage = "replacement-resume/damage"

// DealDamageToPlayerKeyedForEffect is damage to a player with a
// CONTINUATION as data: `then` is a registered tail (RegisterTail), told
// the player as TailOutcome.IDs and the damage that actually landed as
// TailOutcome.Amount, once the CR 614 window has settled it. A zero tail
// is no continuation. A prompt the damage pauses on is a restore point.
//
// It is how a card says "~ deals N damage to target opponent. You gain
// life equal to the damage dealt this way"; DealDamageEachKeyedForEffect
// is the each-opponent version of the same sentence.
//
// WHAT `then` IS TOLD. The post-replacement amount, non-negative: 6 for a
// Lightning Bolt under Angrath's Marauders, 2 for one under a "prevent 1"
// shield. Zero means nothing landed — fully prevented, replaced away
// (CR 614.10, a Fog), or the player has left the game. `then` is told
// either way, because a caller adding up what several opponents took
// would otherwise wait forever on the one that took nothing.
//
// WHEN IT RUNS. Immediately, on the ordinary path, before this function
// returns. After the affected player answers the CR 616 ordering prompt
// when two DIFFERENT damage replacements apply to one event — the case
// #807 exists for, and the reason `then` is a continuation rather than a
// return value. Either way it runs with g.mu held and with the event's
// own EventDealDamage already emitted, so it may start the next damage
// event or queue the next prompt itself.
//
// Returns ErrPlayerNotFound without running `then` when the target is not
// seated at all, exactly as ChangePlayerLifeThenForEffect does. A seated
// player who has been eliminated is not an error: nothing is dealt and
// `then` runs with zero (CR 800.4a, #808). The batch form skips both
// rather than failing on them.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) DealDamageToPlayerKeyedForEffect(source, playerID uuid.UUID, amount int, then KeyedTail) error {
	return g.dealDamageToPlayerTailLocked(source, nil, playerID, amount, DamageMarks{}, 0, damageThen{keyed: then})
}

// damageEventIsData reports whether a damage event can be written with its
// frame: see the file comment.
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) damageEventIsData(ev *ReplacementEvent) bool {
	t := ev.damageTail
	if t == nil || t.combat || t.combatStep != "" || ev.IsCombatDamage || len(t.redirected) > 0 {
		return false
	}
	return t.kind == damageTailPlayer || t.kind == damageTailPermanent
}

// damageTailSnapshot is a damageTail's data half: present exactly when the
// event had a tail. A tail of combat damage or with redirected parts is
// never written.
type damageTailSnapshot struct {
	Kind                damageTailKind     `json:"kind"`
	Actor               *uuid.UUID         `json:"actor,omitempty"`
	SourceLKI           *Characteristic    `json:"sourceLki,omitempty"`
	Deathtouch          bool               `json:"deathtouch,omitempty"`
	LifelinkTo          *uuid.UUID         `json:"lifelinkTo,omitempty"`
	CommanderSource     *uuid.UUID         `json:"commanderSource,omitempty"`
	Result              DamageResultSource `json:"result,omitempty"`
	Controller          *uuid.UUID         `json:"controller,omitempty"`
	Marks               DamageMarks        `json:"marks,omitempty"`
	SourceUnpreventable bool               `json:"sourceUnpreventable,omitempty"`
	SourceChecked       bool               `json:"sourceChecked,omitempty"`
	EndsInstance        bool               `json:"endsInstance,omitempty"`
	// Released marks an event a shield division released (its staged
	// events, landing after the answer): it is never staged again.
	// Additive within v9 (#4361 S16).
	Released bool               `json:"released,omitempty"`
	Then     *keyedTailSnapshot `json:"then,omitempty"`
	// Each is the rest of a keyed Each walk the event is a leg of
	// (damage_each_keyed.go). Additive within v9.
	Each *damageEachRun `json:"each,omitempty"`
}

// snapshotDamageEvent writes a damage event's fields onto its mirror.
func snapshotDamageEvent(ev *ReplacementEvent, out *replacementEventSnapshot) {
	out.DamageSource = uuidPtrOrNil(ev.DamageSource)
	out.DamageTarget = uuidPtrOrNil(ev.DamageTarget)
	out.DamageAmount = ev.DamageAmount
	out.DamageInstance = ev.DamageInstance
	out.SourceLKI = copyCharacteristic(ev.SourceLKI)
	t := ev.damageTail
	if t == nil {
		return
	}
	out.DamageTail = &damageTailSnapshot{
		Kind: t.kind, Actor: uuidPtrOrNil(t.actor), SourceLKI: copyCharacteristic(t.sourceLKI),
		Deathtouch: t.deathtouch, LifelinkTo: uuidPtrOrNil(t.lifelinkTo), CommanderSource: uuidPtrOrNil(t.commanderSource),
		Result: t.result, Controller: uuidPtrOrNil(t.controller), Marks: t.marks,
		SourceUnpreventable: t.sourceUnpreventable, SourceChecked: t.sourceChecked, EndsInstance: t.endsInstance,
		Released: t.released, Then: snapshotKeyedTail(t.keyed), Each: cloneDamageEachRun(t.each),
	}
}

// restoreDamageEvent is snapshotDamageEvent's inverse.
func restoreDamageEvent(s *replacementEventSnapshot, ev *ReplacementEvent) {
	ev.DamageSource = uuidOrNil(s.DamageSource)
	ev.DamageTarget = uuidOrNil(s.DamageTarget)
	ev.DamageAmount = s.DamageAmount
	ev.DamageInstance = s.DamageInstance
	ev.SourceLKI = copyCharacteristic(s.SourceLKI)
	t := s.DamageTail
	if t == nil {
		return
	}
	ev.damageTail = &damageTail{
		kind: t.Kind, actor: uuidOrNil(t.Actor), sourceLKI: copyCharacteristic(t.SourceLKI),
		deathtouch: t.Deathtouch, lifelinkTo: uuidOrNil(t.LifelinkTo), commanderSource: uuidOrNil(t.CommanderSource),
		result: t.Result, controller: uuidOrNil(t.Controller), marks: t.Marks,
		sourceUnpreventable: t.SourceUnpreventable, sourceChecked: t.SourceChecked, endsInstance: t.EndsInstance,
		released: t.Released, keyed: restoreKeyedTail(t.Then), each: cloneDamageEachRun(t.Each),
	}
}

// --- the parked continuation ---------------------------------------------

// parkFollowUpKeyedLocked parks the rest of a damage instruction behind its
// instance's follow-up prompt (prevention_then_pause.go, decision 3), as
// data: a damage event's keyed tail, told the target and the damage dealt
// when the park resumes.
//
// Caller must hold g.mu (write).
func (g *Game) parkFollowUpKeyedLocked(inst DamageInstance, then KeyedTail, target uuid.UUID, dealt int) {
	if then.IsZero() {
		return
	}
	k := copyKeyedTail(then)
	g.followUpParks = append(cloneFollowUpParks(g.followUpParks), followUpPark{inst: inst, keyed: &k, target: target, dealt: dealt})
}

// cloneFollowUpParks copies the park list for a copy-on-write append.
func cloneFollowUpParks(in []followUpPark) []followUpPark {
	if len(in) == 0 {
		return nil
	}
	return append([]followUpPark(nil), in...)
}

// runFollowUpParkLocked runs one parked continuation.
//
// Caller must hold g.mu (write).
func (g *Game) runFollowUpParkLocked(p followUpPark) error {
	if p.keyed == nil {
		return nil
	}
	// A keyed Each walk's total names no one recipient.
	var ids []uuid.UUID
	if p.target != uuid.Nil {
		ids = []uuid.UUID{p.target}
	}
	return g.runKeyedTailLocked(*p.keyed, ids, p.dealt)
}

// followUpParkSnapshot is a followUpPark on disk.
type followUpParkSnapshot struct {
	Instance DamageInstance     `json:"instance"`
	Then     *keyedTailSnapshot `json:"then"`
	Target   uuid.UUID          `json:"target"`
	Dealt    int                `json:"dealt,omitempty"`
}

// snapshotFollowUpParks writes the parks. Every park is data (#4361 S8),
// so a table holding one is a restore point when the prompt it waits on is.
func snapshotFollowUpParks(parks []followUpPark) []followUpParkSnapshot {
	var out []followUpParkSnapshot
	for _, p := range parks {
		if p.keyed == nil {
			continue
		}
		out = append(out, followUpParkSnapshot{Instance: p.inst, Then: snapshotKeyedTail(*p.keyed), Target: p.target, Dealt: p.dealt})
	}
	return out
}

// restoreFollowUpParks rebuilds the keyed parks. A file written before them
// restores none.
func restoreFollowUpParks(in []followUpParkSnapshot) []followUpPark {
	if len(in) == 0 {
		return nil
	}
	out := make([]followUpPark, 0, len(in))
	for _, s := range in {
		k := restoreKeyedTail(s.Then)
		out = append(out, followUpPark{inst: s.Instance, keyed: &k, target: s.Target, dealt: s.Dealt})
	}
	return out
}

// maxParkedDamageInstance is the largest instance a restored park names.
func maxParkedDamageInstance(parks []followUpPark) uint64 {
	var n uint64
	for _, p := range parks {
		if uint64(p.inst) > n {
			n = uint64(p.inst)
		}
	}
	return n
}

// unknownFollowUpParkKeys is checkEffectKeys' line for a park whose tail
// this binary does not register.
func unknownFollowUpParkKeys(in []followUpParkSnapshot) []string {
	var out []string
	for _, s := range in {
		out = append(out, unknownKeyedTail(s.Then)...)
	}
	return out
}
