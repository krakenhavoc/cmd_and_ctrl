package game

import "github.com/google/uuid"

// damage_each_keyed.go — #4253: "~ deals N damage to each of these, then
// …" and "~ deals N damage to target creature, then …" with the rest of
// the effect as data.
//
// DealDamageEachEachThenForEffect walks its recipients one damage event at
// a time, each from the previous event's continuation, so a CR 616 ordering
// prompt over one recipient's damage waits with the others behind it. Its
// step is a closure over the remaining recipients and the running total.
// DealDamageEachKeyedForEffect walks the same batch on a data tail
// (damageTail.each, a damageEachRun), the damage twin of
// LoseLifeEachKeyedForEffect's lifeEachRun, and finishes with a registered
// tail told the total dealt. The leg's damage event is then data
// (damageEventIsData), so the prompt is a restore point.
//
// DealDamageToCreatureKeyedForEffect is the permanent-target sibling of
// DealDamageToPlayerKeyedForEffect (damage_keyed.go).

// damageThen is the caller's half of one damage event's tail, in any of its
// forms: a closure (damageTail.then), a registered tail (damageTail.keyed)
// or the rest of a keyed Each walk (damageTail.each). At most one is set.
type damageThen struct {
	fn    func(g *Game, dealt int) error
	keyed KeyedTail
	each  *damageEachRun
}

// onto writes the continuation onto a damage event's tail.
func (t damageThen) onto(tail *damageTail) {
	tail.then = t.fn
	if t.fn == nil {
		tail.keyed = t.keyed
		tail.each = t.each
	}
}

// runDamageThenNowLocked runs a continuation whose damage never became an
// event (zero damage, CR 120.8; a player who has left, CR 800.4a), told
// `dealt` (zero) and the target.
//
// Caller must hold g.mu (write).
func (g *Game) runDamageThenNowLocked(t damageThen, target uuid.UUID, dealt int) error {
	switch {
	case t.fn != nil:
		return t.fn(g, dealt)
	case t.each != nil:
		return g.dealDamageEachKeyedStepLocked(t.each.after(dealt))
	}
	return g.runKeyedTailLocked(t.keyed, []uuid.UUID{target}, dealt)
}

// DealDamageToCreatureKeyedForEffect is damage to a battlefield permanent
// with a CONTINUATION as data — the permanent-target sibling of
// DealDamageToPlayerKeyedForEffect, with the identical contract: `then`
// is a registered tail (RegisterTail), told the permanent as
// TailOutcome.IDs and the post-replacement amount as its Amount, once, and
// told 0 when the damage was fully prevented, replaced away, or the
// permanent left between a CR 616 prompt and its answer. A zero tail is no
// continuation. A prompt the damage pauses on is a restore point.
//
// "~ deals damage equal to its power to target creature. You gain that
// much life" is the sentence this exists for; a card that only deals the
// damage keeps using DealDamageToCreatureForEffect.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) DealDamageToCreatureKeyedForEffect(source, cardID uuid.UUID, amount int, then KeyedTail) error {
	return g.dealDamageToPermanentTailLocked(source, nil, cardID, amount, DamageMarks{}, 0, damageThen{keyed: then})
}

// damageEachRun is a keyed Each walk, riding the damage tail of the leg now
// being dealt: the recipients still to be dealt damage, how much each is
// dealt, the running total, the walk's instance and whether the walk ended
// it, and the effect's continuation. Never written after it is built, so an
// undo copy of a tail may share it. Additive within snapshot v9.
type damageEachRun struct {
	Source   uuid.UUID          `json:"source"`
	Rest     []uuid.UUID        `json:"rest,omitempty"`
	Amount   int                `json:"amount"`
	Dealt    int                `json:"dealt,omitempty"`
	Instance DamageInstance     `json:"instance"`
	Owned    bool               `json:"owned,omitempty"`
	Then     *keyedTailSnapshot `json:"then,omitempty"`
}

// after is the walk once the leg in hand has settled with `dealt`.
func (r *damageEachRun) after(dealt int) *damageEachRun {
	out := cloneDamageEachRun(r)
	out.Dealt += dealt
	return out
}

// resumeKeyedDamageWalk is a divided keyed walk's start, once the division
// is made. The divide_shield prompt's own resume is a closure frame, so a
// table waiting on that prompt is not a restore point whatever the walk is.
func resumeKeyedDamageWalk(r *damageEachRun) func(*Game) error {
	run := cloneDamageEachRun(r)
	return func(g *Game) error { return g.dealDamageEachKeyedStepLocked(run) }
}

// dealDamageEachKeyedStepLocked deals the head of the walk's recipients its
// damage, with the rest of the walk as that event's tail, or, past the
// last, ends the walk's instance and runs its tail told the total.
// A seated player who has left the game, and anything that is neither a
// seated player nor a battlefield permanent any more, is skipped
// (CR 800.4a, #808), as the closure walk skips them.
//
// Caller must hold g.mu (write).
func (g *Game) dealDamageEachKeyedStepLocked(r *damageEachRun) error {
	rest := r.Rest
	for len(rest) > 0 {
		next := rest[0]
		rest = rest[1:]
		leg := cloneDamageEachRun(r)
		leg.Rest = copyUUIDs(rest)
		if p := g.playerByIDLocked(next); p != nil {
			if p.Eliminated {
				continue
			}
			return g.dealDamageToPlayerTailLocked(r.Source, nil, next, r.Amount, DamageMarks{}, r.Instance, damageThen{each: leg})
		}
		if findBattlefieldCard(g, next) != nil {
			return g.dealDamageToPermanentTailLocked(r.Source, nil, next, r.Amount, DamageMarks{}, r.Instance, damageThen{each: leg})
		}
	}
	paused := r.Owned && g.endDamageInstanceLocked(r.Instance)
	then := restoreKeyedTail(r.Then)
	if then.IsZero() {
		return nil
	}
	if paused {
		// #2040: a follow-up of the walk's instance is asking; the total
		// waits for the answer, as data (followUpPark.keyed, no target).
		g.parkFollowUpKeyedLocked(r.Instance, then, uuid.Nil, r.Dealt)
		return nil
	}
	return g.runKeyedTailLocked(then, nil, r.Dealt)
}

// cloneDamageEachRun copies a walk with its own slices.
func cloneDamageEachRun(r *damageEachRun) *damageEachRun {
	if r == nil {
		return nil
	}
	out := *r
	out.Rest = copyUUIDs(r.Rest)
	if r.Then != nil {
		out.Then = snapshotKeyedTail(restoreKeyedTail(r.Then))
	}
	return &out
}
