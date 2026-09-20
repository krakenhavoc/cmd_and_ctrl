package game

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

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
// WHAT THE lifeTail STRUCT IS FOR, AND WHAT IT IS NOT. damageTail
// carries what the ENGINE still owes a settled damage event: which side
// of the CR 120.3 split the target is on, and the deathtouch, lifelink
// and CR 903.10a commander state snapshotted off a source that may be
// in a graveyard by the time a paused event resumes. A life change owes
// none of that — LifePlayer, LifeDelta and Source are the whole of the
// engine's half, and the ReplacementEvent already carries all three,
// which is what makes the resume faithful (Source is the difference
// between the log reading "Rhox Faithmender: +6" and a bare "+6").
//
// What lifeTail carries instead is what the event owes its CALLER: the
// rest of the effect that asked for the change, run with the amount
// that actually moved. #793, and the struct's own comment below has the
// argument. Until #482 there was nothing to carry, because a life
// change could not pause and a caller could just read the total back.
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
		//
		// The continuation still runs, with zero: "you gain life equal
		// to the life lost this way" gains nothing when the loss was
		// replaced away, and a drain adding up several players' losses
		// would otherwise wait forever for this one.
		return false, g.runLifeTailLocked(ev, 0)
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
// Returns ErrPlayerNotFound when the player is not seated and
// ErrPlayerEliminated when they have left the game (CR 800.4a) — both
// "the player is gone", and both still a TERMINAL outcome, so the
// caller's continuation runs with zero before the error comes back,
// exactly as applyResolvedDamageLocked does for a vanished target. The
// entry points surface the error; the resume treats it as "the life
// change simply does not happen" rather than failing an action whose
// prompt is already dequeued.
//
// #808: an eliminated seat is never removed from g.Seats, so the
// nil check alone let a player who conceded between a CR 616 prompt
// and its answer lose life — and a drain then gained its caster life
// from somebody no longer in the game.
//
// Caller must hold g.mu.
func (g *Game) applyResolvedLifeChangeLocked(ev *ReplacementEvent) error {
	if ev == nil {
		return nil
	}
	p := g.playerByIDLocked(ev.LifePlayer)
	if p == nil || p.Eliminated {
		gone := ErrPlayerNotFound
		if p != nil {
			gone = ErrPlayerEliminated
		}
		if tailErr := g.runLifeTailLocked(ev, 0); tailErr != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				ErrorMsg: "life continuation failed: " + tailErr.Error(),
			})
		}
		return gone
	}
	p.ChangeLife(ev.LifeDelta)
	// #1117: a static keyed on a life total (Serra Ascendant) is
	// stale from this instant until something drops the cache.
	g.invalidateLayersForLifeChangeLocked()
	g.EmitEvent(Event{
		Kind:   EventChangeLife,
		Source: ev.Source,
		Target: ev.LifePlayer,
		Amount: ev.LifeDelta,
	})
	// #793: the rest of the effect, with the amount that actually
	// moved. Runs after the event so a continuation that reads the log
	// — or queues its own life change — sees this one already recorded,
	// and before the caller's state-based sweep, because the loss and
	// the gain of a drain are both part of one resolution.
	return g.runLifeTailLocked(ev, ev.LifeDelta)
}

// lifeTail is the life half's answer to damageTail: what a life change
// still owes once the CR 614 window has settled the amount. Where the
// damage tail carries what the ENGINE owes the event (which side of the
// CR 120.3 split, deathtouch, lifelink, the CR 903.10a tally), the life
// tail carries what the event owes its CALLER — the rest of the effect
// that asked for the change.
//
// #793: every life change has run the CR 614 window since #482, which
// means every life change can PAUSE on a CR 616 ordering prompt. A
// caller that reads the life total back on the next line to find out
// how much actually moved — "each opponent loses X life, you gain life
// equal to the life lost this way" (Exsanguinate, Debt to the
// Deathless, Gray Merchant, Kokusho) — reads it before the prompt is
// answered and sees nothing. The continuation is the fix, and it is the
// shape the search, scry and confirm prompts already use: the frame
// carries the rest of the effect and the engine runs it at the one
// place the change lands.
//
// Unexported engine plumbing. The catalog never builds one; it hands a
// `then` to ChangePlayerLifeThenForEffect or LoseLifeEachThenForEffect
// and the entry point fills this in before the pipeline runs, for the
// same reason damageTail goes on first: a CR 616 prompt returns without
// landing anything, and the resume has nothing else to go on.
type lifeTail struct {
	// then runs once the change has settled, with the amount that
	// ACTUALLY moved: the post-replacement delta, signed the way the
	// event is (negative for a loss), and zero for a change that was
	// replaced away (CR 614.10) or whose player has left the game.
	//
	// It takes the live *Game rather than capturing one, on the same
	// undo-safety contract confirmFrame and StackItem.Effect follow —
	// an undo restores this game's fields in place, so a *Game argument
	// is always the right game and a captured *Player would not be. It
	// runs with g.mu held, so it may start the next life change or
	// queue the next prompt itself.
	then func(g *Game, applied int) error
}

// runLifeTailLocked runs a settled life event's continuation exactly
// once, with the amount that actually moved. The tail is cleared before
// it runs, so a continuation that re-enters the pipeline on the same
// event value cannot run itself twice.
//
// Every TERMINAL outcome of a life event goes through here — landed,
// replaced away, player gone — because a caller adding up "the life
// lost this way" has to be told even when the answer is zero, or it
// waits forever. A pause is not a terminal outcome: the resume reaches
// this function later, through applyResolvedLifeChangeLocked.
//
// Caller must hold g.mu.
func (g *Game) runLifeTailLocked(ev *ReplacementEvent, applied int) error {
	if ev == nil || ev.lifeTail == nil || ev.lifeTail.then == nil {
		return nil
	}
	then := ev.lifeTail.then
	ev.lifeTail = nil
	return then(g, applied)
}

// payLifeAsCostLocked pays `amount` life as a COST (CR 118.3) rather
// than as the effect of a spell or ability.
//
// WHAT THE RULES SAY, precisely, because #793 guessed the other way.
// CR 119.4: "If a player pays life, the payment is subtracted from
// their life total; in other words, the player loses that much life."
// Paying life IS losing life, so the CR 614 window runs on it exactly as it runs on a drain:
// a life-loss replacement (Bloodletter of Aclazotz) sees a Thoughtseize
// being paid for, and "your life total can't change" stops it. What a
// payment is NOT is life GAIN, so Rhox Faithmender and Alhammarret's
// Archive never touch it — their AppliesTo tests LifeDelta > 0 — which
// is the half #793 was actually worried about. Routing costs around the
// window would have been a rules change, not a fix.
//
// WHAT A COST MAY NOT DO IS PAUSE. CR 601.2h pays a spell's costs as
// one indivisible step of casting it, and CR 601.2 rewinds the whole
// announcement if they cannot all be paid; CR 602.2b says the same for
// an activated ability. A CR 616 ordering prompt in the middle of that
// leaves a spell on the stack with its cost half paid, a table waiting
// on a prompt, and nothing left that a rewind could take back. So the
// payment sets mustSettleNow and the pipeline settles it without
// asking: an ordering window applies in the order it was gathered (the
// same escape an eliminated chooser has taken since the S31 fuzzer
// finding), and an effect that would ask its own question is skipped
// un-applied, which is the weaker-never-stronger posture
// optionalReplacementResumableLocked already takes for an entry that
// cannot resume.
//
// A PAYMENT THE WINDOW CANCELS IS NOT PAID. CR 119.8: "if an effect
// says that a player can't lose life, ... a cost that involves having
// that player pay life can't be paid", and CR 614.17b: "If an event
// can't happen, a player can't choose to pay a cost that includes that
// event." A null replacement on the payment ("your life total can't
// change") therefore refuses the cost with ErrInvalidParam rather than
// handing the player whatever the cost bought for nothing (#808). No
// life moved, so there is nothing to put back. A replacement that
// merely CHANGES the amount (a doubler, a "lose 1 less") is still a
// payment: the replaced event is what was paid.
//
// Returns ErrInvalidParam when the payment cannot be made: CR 119.4
// allows it only from a life total at least as large as the payment,
// and CR 119.8 not at all while the player's life can't be lost. Every
// caller validates the first before it starts paying anything, so this
// is the backstop rather than the check. Nothing in the catalog makes
// a player unable to lose life today, so the second has no up-front
// check to back up yet; a card that adds one should add that check to
// the cost validators it reaches.
//
// Caller must hold g.mu.
func (g *Game) payLifeAsCostLocked(source, playerID uuid.UUID, amount int) error {
	if amount <= 0 {
		return nil
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if p.Life < amount {
		return ErrInvalidParam
	}
	ev := &ReplacementEvent{
		Kind:          RepEventLife,
		Source:        source,
		LifePlayer:    playerID,
		LifeDelta:     -amount,
		mustSettleNow: true,
	}
	if _, err := g.changeLifeThroughReplacementsLocked(ev); err != nil {
		return err
	}
	if ev.Canceled {
		return fmt.Errorf("%w: a life payment that cannot be made cannot be paid (CR 119.8)", ErrInvalidParam)
	}
	return nil
}
