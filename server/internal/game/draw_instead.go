package game

import (
	"errors"
	"sync"

	"github.com/google/uuid"
)

// draw_instead.go — a draw replaced by an EFFECT that may need to ask
// something (#2168, #2127): dredge (CR 702.52), Underrealm Lich's "look
// at the top three cards of your library, then put one into your hand
// and the rest into your graveyard", Forbidden Crypt's "return a card
// from your graveyard to your hand instead".
//
// # The shape
//
// The CR 614 window on a draw already settled a COUNT, a drawing player
// and a cancellation. What it could not do is run a body that stops to
// ask — a Replace runs inside the apply-loop and nothing can pause
// there. So a replacement that replaces the draw with something
// declares ReplacementEffect.DrawInstead instead of a Replace: applying
// it cancels the draw (nothing is drawn, no EventDrawCard — CR 614.6,
// the replaced event never happens) and records which body to run. Once
// the window has SETTLED, the body runs, and it may queue prompts of its
// own.
//
// # The rest of the instruction (CR 121.6, 614.11)
//
// "Draw three cards" is three individual draws (CR 121.2), and a
// replaced one has to be finished before the next begins. The body is
// handed `done`, and calls it from wherever it finishes — inline, or
// from the Then of a prompt it queued — and `done` is what performs
// the draws still owed. So a paused draw holds the rest of its
// instruction behind the prompt, and the next draw opens its own
// window, where dredge is offered again.
//
// A paused CR 616 ordering prompt or "may" prompt over a plain draw
// owes the same tail, and pays it the same way: the count of draws
// still owed rides ReplacementEvent.drawTail, a plain int the resume
// frame copies by value, so undo and a snapshot taken mid-pause replay
// the answer against the same remaining count.
//
// The closures a body queues must capture scalars only (the
// StackItem.Effect contract): `done` captures a player ID and a
// count, and nothing else. The body itself is not stored on the event:
// it records a registered key (drawInsteadRun) and the body is looked
// up when the window settles.
//
// # Declared simplification, weaker than printed
//
// Dredge is offered on each individual draw as it begins, not on the
// extra draws a doubler (Thought Reflection) adds to it: those are
// drawn ordinarily once the dredge has replaced the first.

// DrawInsteadFunc is the body of a draw replaced by an effect. It runs
// with g.mu held once the replacement window has settled, may queue
// prompts, and must call done exactly once when it has finished.
// `source` is the card the effect belongs to.
type DrawInsteadFunc func(g *Game, drawer, source uuid.UUID, done func(g *Game) error) error

// DrawInsteadRef names a registered DrawInsteadFunc. The key is
// unexported, so the only way to hold one is RegisterDrawInstead, which
// keeps a func out of ReplacementEffect (and so out of what the closure
// census has to account for): the body lives in a registry and the
// effect holds its name.
type DrawInsteadRef struct{ key string }

var (
	drawInsteadMu     sync.RWMutex
	drawInsteadBodies = map[string]DrawInsteadFunc{}
)

// RegisterDrawInstead registers a body under key and returns the ref a
// ReplacementEffect declares. Idempotent for one key, so a card
// constructor called by several cards (Dredge(3)) registers the same
// body each time. A body must capture scalars only, and the key must
// identify everything it captures ("dredge-3").
func RegisterDrawInstead(key string, body DrawInsteadFunc) DrawInsteadRef {
	if key == "" || body == nil {
		panic("game.RegisterDrawInstead: a key and a body are required")
	}
	drawInsteadMu.Lock()
	drawInsteadBodies[key] = body
	drawInsteadMu.Unlock()
	return DrawInsteadRef{key: key}
}

// drawInsteadRun is a recorded, not yet run, substituted draw: which
// body, and which card it belongs to. Plain data, so a resume frame
// holding it is copied by value.
type drawInsteadRun struct {
	ref    DrawInsteadRef
	source uuid.UUID
	// each: the substitute is mandatory, so it replaces every draw of a
	// doubled event rather than one of them.
	each bool
}

// body looks the registered body up. Nil when nothing was recorded.
func (r drawInsteadRun) body() DrawInsteadFunc {
	if r.ref.key == "" {
		return nil
	}
	drawInsteadMu.RLock()
	defer drawInsteadMu.RUnlock()
	return drawInsteadBodies[r.ref.key]
}

// recordDrawInsteadLocked is what applying a DrawInstead replacement
// does: cancel the draw and remember which body to run.
func (g *Game) recordDrawInsteadLocked(ev *ReplacementEvent, a activeReplacement) {
	ev.Canceled = true
	run := drawInsteadRun{ref: a.effect.DrawInstead, each: !a.effect.Optional}
	if a.source != nil {
		run.source = a.source.InstanceID
	}
	ev.drawInstead = run
}

// drawRunLocked performs n individual draws for playerID, each through
// its own CR 614 window (CR 121.2), stopping at the first that pauses
// for a prompt: that draw's resume owns the rest (drawTail). Returns
// ErrZoneEmpty on the first draw from an empty library, with the
// loss flag already set, exactly as one draw does.
//
// Caller must hold g.mu.
func (g *Game) drawRunLocked(playerID uuid.UUID, n int) error {
	for i := 0; i < n; i++ {
		paused, err := g.drawOneLocked(playerID, n-i-1)
		if err != nil || paused {
			return err
		}
	}
	return nil
}

// drawOneLocked opens the window on one draw. `tail` is how many draws
// of the same instruction follow it. paused is true when something
// else now owns those: a queued prompt, or a body that will call done.
//
// Caller must hold g.mu.
func (g *Game) drawOneLocked(playerID uuid.UUID, tail int) (paused bool, err error) {
	ev := &ReplacementEvent{
		Kind:       RepEventDraw,
		Actor:      playerID,
		DrawPlayer: playerID,
		// #1222: the amount. Always ONE here — CR 121.2 makes "draw
		// three cards" three individual card draws — so a draw-amount
		// replacement (Thought Reflection, Alhammarret's Archive)
		// doubles EACH of them rather than the instruction.
		DrawCount: 1,
		drawTail:  tail,
	}
	out, aerr := g.applyReplacementsLocked(ev)
	if errors.Is(aerr, errReplacementPending) {
		// CR 616 / "may" prompt queued; its resume re-enters the
		// pipeline, runs the draw, and pays the tail.
		return true, nil
	}
	if aerr != nil && !errors.Is(aerr, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return false, aerr
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		if ev.drawInstead.body() != nil {
			return true, g.runDrawInsteadLocked(ev)
		}
		// Cancelled by a replacement that substitutes nothing.
		return false, nil
	}
	return false, g.actuallyDrawCardsLocked(out.DrawPlayer, out.DrawCount)
}

// runDrawInsteadLocked runs a cancelled draw's body, with a `done` that
// pays the tail of the instruction.
//
// A draw a doubler made into N is N draws of one event. A MANDATORY
// substitute (Underrealm Lich, Forbidden Crypt) replaces every one of
// them, so its body runs N times, one after the other. An optional one
// (dredge) replaces a single draw, and the other N-1 are ordinary
// draws that follow it.
//
// Caller must hold g.mu.
func (g *Game) runDrawInsteadLocked(ev *ReplacementEvent) error {
	drawer, player, tail := ev.Actor, ev.DrawPlayer, ev.drawTail
	total := max(ev.DrawCount, 1)
	run := ev.drawInstead
	ev.drawInstead = drawInsteadRun{}
	body := run.body()

	finish := func(g *Game) error {
		err := g.drawRunLocked(drawer, tail)
		if errors.Is(err, ErrZoneEmpty) {
			// Flag set; the SBA pass handles the loss, as it does for
			// DrawNForEffect.
			return nil
		}
		return err
	}
	if body == nil {
		return finish(g)
	}
	rounds := 1
	if run.each {
		rounds = total
	}
	// step captures scalars and a registered, stateless body — nothing
	// that lives in the game — so it may sit in a prompt's continuation.
	var step func(g *Game, left int) error
	step = func(g *Game, left int) error {
		if left == 0 {
			if extra := total - rounds; extra > 0 {
				if err := g.actuallyDrawCardsLocked(player, extra); err != nil && !errors.Is(err, ErrZoneEmpty) {
					return err
				}
			}
			return finish(g)
		}
		return body(g, player, run.source, func(g *Game) error { return step(g, left-1) })
	}
	return step(g, rounds)
}

// finishDrawTailLocked is what a RESUMED draw owes once its own card
// has been drawn (or its replacement cancelled with nothing instead):
// the rest of its instruction.
//
// Caller must hold g.mu.
func (g *Game) finishDrawTailLocked(ev *ReplacementEvent) error {
	err := g.drawRunLocked(ev.Actor, ev.drawTail)
	if errors.Is(err, ErrZoneEmpty) {
		return nil
	}
	return err
}

// settleResumedDrawLocked finishes a paused draw whose window has now
// settled — the RepEventDraw arm of both resume paths. A settled event
// that was not cancelled draws and then pays the tail; a cancelled one
// runs its body, which pays the tail itself, or pays it directly.
//
// Caller must hold g.mu.
func (g *Game) settleResumedDrawLocked(ev *ReplacementEvent, cancelled bool) error {
	if cancelled {
		if ev.drawInstead.body() != nil {
			return g.runDrawInsteadLocked(ev)
		}
		return g.finishDrawTailLocked(ev)
	}
	if err := g.actuallyDrawCardsLocked(ev.DrawPlayer, ev.DrawCount); err != nil {
		if errors.Is(err, ErrZoneEmpty) {
			return nil
		}
		return err
	}
	return g.finishDrawTailLocked(ev)
}

// --- dredge's home: the graveyard ------------------------------------

// graveyardReplacementIDBase is where graveyard replacements' IDs
// start. Inside the self-replacement range's headroom (its users sit in
// the first few strides above selfReplacementIDBase) and far below the
// scoped range. An ID packs (seat, graveyard index, slot), so two
// dredge cards in one graveyard are two effects and one card with two
// of them is two.
const graveyardReplacementIDBase = selfReplacementIDBase + 1<<40

const graveyardIndexStride = 1024

func encodeGraveyardReplacementID(seat, cardIdx, slot int) (ReplacementEffectID, bool) {
	if seat < 0 || cardIdx < 0 || cardIdx >= graveyardIndexStride || slot < 0 || slot >= MaxCatalogReplacementSlots {
		return 0, false
	}
	return graveyardReplacementIDBase +
		ReplacementEffectID((seat*graveyardIndexStride+cardIdx)*MaxCatalogReplacementSlots+slot), true
}

func decodeGraveyardReplacementID(id ReplacementEffectID) (seat, cardIdx, slot int, ok bool) {
	if id < graveyardReplacementIDBase || id >= graveyardReplacementIDBase+1<<39 {
		return 0, 0, 0, false
	}
	raw := int(id - graveyardReplacementIDBase)
	slot = raw % MaxCatalogReplacementSlots
	raw /= MaxCatalogReplacementSlots
	return raw / graveyardIndexStride, raw % graveyardIndexStride, slot, true
}

// gatherGraveyardReplacementsLocked appends the FromGraveyard
// replacements of the drawing player's graveyard that apply to ev.
// Only a draw consults a graveyard, and only the DRAWER's: dredge says
// "your graveyard".
//
// The source is a copy of the card, never a pointer into the zone's
// slice, because a prompt can outlive the slice it was gathered from.
//
// Caller must hold g.mu.
func (g *Game) gatherGraveyardReplacementsLocked(ev *ReplacementEvent, applied map[ReplacementEffectID]bool, out []activeReplacement) []activeReplacement {
	if CatalogReplacements == nil || ev == nil || ev.Kind != RepEventDraw || ev.DrawCount <= 0 {
		return out
	}
	seat := -1
	for i, p := range g.Seats {
		if p != nil && p.ID == ev.DrawPlayer {
			seat = i
			break
		}
	}
	if seat < 0 || g.Seats[seat].Graveyard == nil {
		return out
	}
	gy := g.Seats[seat].Graveyard
	for cardIdx := range gy.Cards {
		card := &gy.Cards[cardIdx]
		key := catalogAbilityKeyOf(card)
		reps := CatalogReplacements(key)
		for repIdx := range reps {
			eff := reps[repIdx]
			if !eff.FromGraveyard {
				continue
			}
			id, ok := encodeGraveyardReplacementID(seat, cardIdx, repIdx)
			if !ok || applied[id] {
				continue
			}
			if !eventKindMatches(eff.Watches, ev.Kind) {
				continue
			}
			src := *card
			if eff.AppliesTo != nil && !eff.AppliesTo(ev, g, &src) {
				continue
			}
			out = append(out, activeReplacement{
				effect: eff,
				source: &src,
				id:     id,
				// A card in a graveyard is its own declaration: two
				// Life from the Loams are two dredges, and which is
				// returned is observable, so they are never
				// interchangeable.
			})
		}
	}
	return out
}

// graveyardReplacementMetaLocked names a graveyard replacement for the
// prompt view: its label and its source card.
func (g *Game) graveyardReplacementMetaLocked(id ReplacementEffectID) (label string, card uuid.UUID, ok bool) {
	seat, cardIdx, slot, ok := decodeGraveyardReplacementID(id)
	if !ok {
		return "", uuid.Nil, false
	}
	if seat >= len(g.Seats) || g.Seats[seat] == nil || g.Seats[seat].Graveyard == nil ||
		cardIdx >= len(g.Seats[seat].Graveyard.Cards) {
		return "", uuid.Nil, true
	}
	c := &g.Seats[seat].Graveyard.Cards[cardIdx]
	reps := CatalogReplacements(catalogAbilityKeyOf(c))
	if slot >= len(reps) {
		return "", c.InstanceID, true
	}
	return reps[slot].Label, c.InstanceID, true
}

// allGraveyardOptions reports whether every gathered replacement is a
// graveyard "may" — dredge cards. Ordering them is not a question.
func allGraveyardOptions(applicable []activeReplacement) bool {
	for _, a := range applicable {
		if !a.effect.FromGraveyard || !a.effect.Optional {
			return false
		}
	}
	return len(applicable) > 0
}
