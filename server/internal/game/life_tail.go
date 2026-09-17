package game

import "errors"

// life_tail.go is damage_tail.go's sibling for the other half of
// CR 118: what still has to happen once the CR 614 replacement
// pipeline has settled a LIFE-CHANGE event, and the guarantee that
// every writer of a life total goes through it.
//
// Issue #482: the CR 614 window on a life change — the window the
// "if you would gain life, you gain twice that much instead" family
// lives in — ran on exactly one entry point, the public
// ChangePlayerLife sandbox verb a player reaches by dragging their own
// life counter. The two paths a GAME actually takes did not:
//
//   - ChangePlayerLifeForEffect — every catalog GainLife, every drain,
//     every "pay N life" cost. It called p.ChangeLife and emitted
//     EventChangeLife itself, three lines with no pipeline in them.
//   - creditLifelinkLocked (damage_tail.go) — the CR 702.15b lifelink
//     credit, which is life gain and therefore replaceable. The same
//     three lines again.
//
// So Rhox Faithmender ("If you would gain life, you gain twice that
// much life instead") doubled a life total typed in by hand and
// nothing else: not a Lightning Helix, not a Soul Warden, and not its
// own printed lifelink. It shipped CompletenessFull because the test
// that covered it only ever exercised the sandbox verb.
//
// The fix is the shape #694 used for damage: there is exactly ONE
// function that lands a settled life change, every entry point reaches
// it through one shared "run the window, then land it" helper, and the
// CR 616 resume in applyResolvedReplacementEventLocked calls that same
// function instead of carrying its own copy of the tail.
//
// WHY THERE IS NO lifeTail STRUCT. damageTail exists because a settled
// damage event still has to choose between a player and a permanent,
// and still owes deathtouch, lifelink and the CR 903.10a commander
// tally — all snapshotted off a source that may be in a graveyard by
// the time a paused event resumes. A life change owes none of that:
// LifePlayer, LifeDelta and Source are the whole of the tail, and the
// ReplacementEvent already carries all three. Carrying them there is
// what makes the resume faithful — Source is the difference between
// the log reading "Rhox Faithmender: +6" and reading a bare "+6".
//
// WHAT THIS TAIL DELIBERATELY DOES NOT DO.
//
//   - It never runs state-based actions. The effect paths get their
//     sweep from the resolution bookend around resolveTopOfStackLocked
//     and the combat paths from the end of the damage step (CR 510.2),
//     exactly as they did before. The CR 616 resume runs its own,
//     because answering a prompt is an action boundary like every
//     other Resolve* handler.
//   - DAMAGE does not fire this window. Damage dealt to a player
//     reduces life directly once the DAMAGE replacements have settled
//     (CR 120.3), inside applyResolvedDamageToPlayerLocked, and a
//     life-change replacement must not get a second bite at the same
//     event. Lifelink is the one part of a damage event that IS a life
//     change (CR 702.15b: "causes that source's controller to gain
//     that much life") and it is the one part routed here.

// changeLifeThroughReplacementsLocked runs the CR 614 window on a
// life-change event and lands it through the one tail. The shared body
// behind ChangePlayerLife, ChangePlayerLifeForEffect and the lifelink
// credit.
//
// Reports paused=true when a CR 616 ordering prompt was queued (two
// life replacements applied to one event and the affected player has
// to order them). The event is not lost: it lands from
// applyResolvedReplacementEventLocked when the prompt is answered. A
// caller that owes its own caller a life total has nothing to report
// until then.
//
// Caller must hold g.mu.
func (g *Game) changeLifeThroughReplacementsLocked(ev *ReplacementEvent) (paused bool, err error) {
	if ev == nil {
		return false, nil
	}
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		return true, nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return false, err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		// CR 614.10 with a null replacement — "your life total can't
		// change". No mutation, and no EventChangeLife for a "whenever
		// you gain life" trigger to see, because nothing happened.
		return false, nil
	}
	return false, g.applyResolvedLifeChangeLocked(out)
}

// applyResolvedLifeChangeLocked performs the underlying mutation for a
// life ReplacementEvent whose replacement pipeline has settled. This is
// the ONLY place a life total changes outside the CR 120.3 damage path:
// every entry point calls it on the unpaused path and the CR 616 resume
// calls it on the paused one, so the two cannot drift.
//
// The caller is responsible for state-based actions — see the file
// comment.
//
// Returns ErrPlayerNotFound when the player is no longer seated. The
// entry points surface that; the resume treats it as "the player left,
// the life change simply does not happen" rather than failing an
// action whose prompt is already dequeued.
//
// Caller must hold g.mu.
func (g *Game) applyResolvedLifeChangeLocked(ev *ReplacementEvent) error {
	if ev == nil {
		return nil
	}
	p := g.playerByIDLocked(ev.LifePlayer)
	if p == nil {
		return ErrPlayerNotFound
	}
	p.ChangeLife(ev.LifeDelta)
	g.EmitEvent(Event{
		Kind:   EventChangeLife,
		Source: ev.Source,
		Target: ev.LifePlayer,
		Amount: ev.LifeDelta,
	})
	return nil
}
