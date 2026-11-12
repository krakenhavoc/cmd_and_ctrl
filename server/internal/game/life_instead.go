package game

import "github.com/google/uuid"

// life_instead.go: "If your life total would be reduced to 0 or less,
// instead ... your life total becomes N. Then ... you lose the game"
// (Enduring Angel, #5678).
//
// It needs no window of its own. Every way a life total falls already
// runs the CR 614 life window as a RepEventLife event with a negative
// LifeDelta: a loss and a set-life effect (CR 119.5,
// ChangePlayerLifeForEffect), a payment (CR 119.4, payLifeAsCostLocked)
// and the life that damage costs (CR 120.3a, ADR 0169). So the clause is
// an ordinary ReplacementEffect on that window, and this file is the
// three things its body asks of the engine:
//
//   - the test, read off the event as it stands now (a doubler applied
//     first has already made the loss larger);
//   - the rewrite, "your life total becomes N", which is a gain or a loss
//     of the difference (CR 119.5) and stays in the window as one, so a
//     later replacement sees it and the caller's "life lost this way" is
//     told what really moved;
//   - "Then ... you lose the game", which the event carries as
//     LifeThenLoses and the place it lands carries out.
//
// The body's other half (transform this creature) is the card's: it asks
// CanTransformForEffect and calls TransformPermanentForEffect.

// LifeWouldBeReducedToZeroOrLessForEffect reports whether the life event
// `ev` would reduce its player's life total to 0 or less: it is a loss
// (a negative delta) and the total after it is 0 or less. A player
// already at 0 or less who loses more is having their total reduced to 0
// or less; one who gains is not.
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) LifeWouldBeReducedToZeroOrLessForEffect(ev *ReplacementEvent) bool {
	if ev == nil || ev.Kind != RepEventLife || ev.LifeDelta >= 0 {
		return false
	}
	p := g.playerByIDLocked(ev.LifePlayer)
	return p != nil && !p.Eliminated && p.Life+ev.LifeDelta <= 0
}

// LifeTotalBecomesInsteadForEffect rewrites the life event `ev` into
// "that player's life total becomes `total`": the event now gains or
// loses the difference (CR 119.5). A player who can't gain life does not
// gain the difference, so a total below `total` stays where it is. A
// difference of zero moves no life and lands as no event.
//
// Caller must hold g.mu (a read lock is enough: only `ev` is written).
func (g *Game) LifeTotalBecomesInsteadForEffect(ev *ReplacementEvent, total int) {
	if ev == nil || ev.Kind != RepEventLife {
		return
	}
	p := g.playerByIDLocked(ev.LifePlayer)
	if p == nil {
		return
	}
	delta := total - p.Life
	if delta > 0 && g.playerCantGainLifeLocked(p) {
		delta = 0
	}
	ev.LifeDelta = delta
}

// ReplacementIsPreviewForEffect reports whether `ev` is a read-only
// preview of a window (lifeChangePreviewLocked, the auto-tapper's
// question) and not an event that is happening: a preview never enters
// applyReplacementsLocked, so its ID was never minted. A Replace that
// does more than rewrite the event (transforms a permanent) asks this
// first and only rewrites.
func (g *Game) ReplacementIsPreviewForEffect(ev *ReplacementEvent) bool {
	return ev != nil && ev.ID == 0
}

// loseGameAfterLifeEventLocked carries out a settled life event's "Then
// ... you lose the game" (ReplacementEvent.LifeThenLoses), once, and
// reports whether the player left. Every terminal outcome of the event
// calls it after the life has landed, or has turned out not to move.
//
// The loss is CR 104.3e's: the "can't lose" gate is read now and the
// player leaves now (loseGameLocked). The game-over check and the
// rotation are deferred to the next state-based action pass through
// ActiveSeatLeftPending whoever the player is, because this runs inside
// a replacement window, which may be inside a resolution or a batch of
// combat damage (ADR 0057 Decision 3, amendment 2026-11-12).
//
// Caller must hold g.mu.
func (g *Game) loseGameAfterLifeEventLocked(ev *ReplacementEvent) bool {
	if ev == nil || ev.LifeThenLoses == uuid.Nil {
		return false
	}
	source := ev.LifeThenLoses
	ev.LifeThenLoses = uuid.Nil
	if g.State != StateActive {
		return false
	}
	p := g.playerByIDLocked(ev.LifePlayer)
	if p == nil || !g.loseGameLocked(p, LossEffect, source) {
		return false
	}
	g.ActiveSeatLeftPending = true
	return true
}

// resolvingControllerIsLocked reports whether `player` controls the
// spell or ability resolving right now.
func (g *Game) resolvingControllerIsLocked(player uuid.UUID) bool {
	r := g.resolving
	return r != nil && r.item != nil && r.item.Controller == player
}

// finishDamageLifeInsteadLocked lands what the life window made of the
// life a damage event cost, beyond a plain loss (ADR 0169, amendment
// 2026-11-12): a loss a replacement rewrote into a GAIN ("your life
// total becomes 3" at 2 life) is gained, and a "Then ... you lose the
// game" the event carries is carried out. `settled` is nil when the
// window left neither.
//
// Caller must hold g.mu.
func (g *Game) finishDamageLifeInsteadLocked(settled *ReplacementEvent) {
	if settled == nil {
		return
	}
	if !settled.Canceled && settled.LifeDelta > 0 {
		// The one tail lands the gain, and the loss after it. Its error
		// is "the player is gone" or a clean stop: nothing to report.
		_ = g.applyResolvedLifeChangeLocked(settled)
		return
	}
	g.loseGameAfterLifeEventLocked(settled)
}
