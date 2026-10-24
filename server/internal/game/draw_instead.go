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
// # "Draw, then …" (#2391, CR 608.2c)
//
// A clause printed after a draw ("draw two cards, then discard two")
// must not act before the draw has resolved, and a draw that paused on
// a prompt has not. DrawNThenForEffect carries the rest of the card as
// DATA — a registered body's key plus a player, a source card and a
// count (DrawThen) — through the same resume frame drawTail already
// rides, and runs it once the last draw of the instruction is
// finished: inline when nothing paused, from the answer's resume when
// something did. It runs even when a draw found the library empty, as
// the continuation of any other effect does: the loss is a state-based
// action, not a reason to skip the sentence after "then".
//
// # Declared simplification, weaker than printed
//
// Every draw introduced by an amount replacement gets the same
// treatment: one replacement window, one substitute, then the next
// descendant, followed by the original instruction's remaining draws.

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

// DrawThenFunc is the body of a "draw …, then X" continuation. It runs
// with g.mu held once every draw of the instruction has finished.
type DrawThenFunc func(g *Game, d DrawThen) error

// DrawThenRef names a registered DrawThenFunc; like DrawInsteadRef the
// key is unexported so a func never reaches an event.
type DrawThenRef struct{ key string }

// DrawThen is the "then" of a draw: which registered body, and the
// scalars it needs. Plain data, copied by value into a resume frame.
type DrawThen struct {
	Ref    DrawThenRef
	Player uuid.UUID
	Source uuid.UUID
	N      int
}

var (
	drawThenMu     sync.RWMutex
	drawThenBodies = map[string]DrawThenFunc{}
)

// RegisterDrawThen registers a continuation body under key. Idempotent
// for one key. The body must capture nothing that lives in the game.
func RegisterDrawThen(key string, body DrawThenFunc) DrawThenRef {
	if key == "" || body == nil {
		panic("game.RegisterDrawThen: a key and a body are required")
	}
	drawThenMu.Lock()
	drawThenBodies[key] = body
	drawThenMu.Unlock()
	return DrawThenRef{key: key}
}

// runDrawThenLocked runs a recorded continuation, if any.
//
// Caller must hold g.mu.
func (g *Game) runDrawThenLocked(d DrawThen) error {
	if d.Ref.key == "" {
		return nil
	}
	drawThenMu.RLock()
	body := drawThenBodies[d.Ref.key]
	drawThenMu.RUnlock()
	if body == nil {
		return nil
	}
	return body(g, d)
}

// DrawNThenForEffect draws n cards for playerID and then runs `then`:
// "draw two cards, then discard two cards". The then waits for any
// prompt a draw pauses on (a dredge offer, a draw-instead pick, a CR
// 616 ordering prompt) instead of running as the call returns.
func (g *Game) DrawNThenForEffect(playerID uuid.UUID, n int, then DrawThen) error {
	err := g.drawRunLocked(playerID, n, then)
	if errors.Is(err, ErrZoneEmpty) {
		return nil
	}
	return err
}

// drawInsteadRun is a recorded, not yet run, substituted draw: which
// body, and which card it belongs to. Plain data, so a resume frame
// holding it is copied by value.
type drawInsteadRun struct {
	ref    DrawInsteadRef
	source uuid.UUID
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
	run := drawInsteadRun{ref: a.effect.DrawInstead}
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
func (g *Game) drawRunLocked(playerID uuid.UUID, n int, then DrawThen) error {
	if n <= 0 {
		return g.runDrawThenLocked(then)
	}
	_, err := g.drawOneLocked(playerID, n-1, then)
	return err
}

// pendingDraw is a descendant of an amount replacement. Applied is the
// inherited once-per-event history, copied to its new event ID when it
// begins. Remaining siblings wait until this draw and its substitute finish.
type pendingDraw struct {
	player  uuid.UUID
	count   int
	applied []ReplacementEffectID
}

func (g *Game) splitExpandedDrawLocked(ev *ReplacementEvent) {
	if ev.Kind != RepEventDraw || ev.Canceled || ev.DrawCount <= 1 {
		return
	}
	p := pendingDraw{player: ev.DrawPlayer, count: ev.DrawCount - 1}
	for id, applied := range g.replacementsAppliedThisEvent[ev.ID] {
		if applied {
			p.applied = append(p.applied, id)
		}
	}
	ev.DrawCount = 1
	ev.drawPending = append([]pendingDraw{p}, ev.drawPending...)
}

// drawOneLocked opens the window on one draw. Once that draw settles it
// performs its amount-replacement descendants, then the original tail.
func (g *Game) drawOneLocked(playerID uuid.UUID, tail int, then DrawThen) (paused bool, err error) {
	ev := &ReplacementEvent{
		Kind: RepEventDraw, Actor: playerID, DrawPlayer: playerID,
		DrawCount: 1, drawTail: tail, drawThen: then,
	}
	return g.drawEventLocked(ev)
}

func (g *Game) drawEventLocked(ev *ReplacementEvent) (paused bool, err error) {
	// #3927, CR 121.2b / 614.17: a draw its player can't make (Narset's
	// "can't draw more than one card each turn") is not performed and
	// opens no CR 614 window, so no replacement sees it. The rest of the
	// instruction goes on, each later draw asking again (draw_limit.go).
	if g.drawForbiddenLocked(ev) {
		if ev.ID != 0 {
			defer g.clearReplacementEventLocked(ev.ID)
		}
		return false, g.finishDrawTailLocked(ev)
	}
	out, aerr := g.applyReplacementsLocked(ev)
	if errors.Is(aerr, errReplacementPending) {
		return true, nil
	}
	if aerr != nil && !errors.Is(aerr, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return false, aerr
	}
	defer g.clearReplacementEventLocked(ev.ID)
	return false, g.settleResumedDrawLocked(ev, out == nil || out.Canceled)
}

// runDrawInsteadLocked runs a cancelled draw's body, with a `done` that
// pays the tail of the instruction.
//
// Each expanded draw has already been split into an individual event.
// Its substitute runs once, and its continuation carries plain event
// data so a prompt can finish the siblings and original tail later.
//
// Caller must hold g.mu.
func (g *Game) runDrawInsteadLocked(ev *ReplacementEvent) error {
	run := ev.drawInstead
	ev.drawInstead = drawInsteadRun{}
	// Capture the frame by value: a substitute may pause, and its
	// continuation must not mutate a frame shared with an undo clone.
	tail := *ev
	finish := func(g *Game) error { return g.finishDrawTailLocked(&tail) }
	if body := run.body(); body != nil {
		return body(g, ev.DrawPlayer, run.source, finish)
	}
	return finish(g)
}

// finishDrawTailLocked is what a RESUMED draw owes once its own card
// has been drawn (or its replacement cancelled with nothing instead):
// the rest of its instruction.
//
// Caller must hold g.mu.
func (g *Game) finishDrawTailLocked(ev *ReplacementEvent) error {
	if len(ev.drawPending) > 0 {
		p := ev.drawPending[0]
		next := &ReplacementEvent{
			ID:   ReplacementEventID(g.nextReplacementEventID.Add(1)),
			Kind: RepEventDraw, Actor: ev.Actor, DrawPlayer: p.player, DrawCount: p.count,
			drawPending: ev.drawPending[1:], drawTail: ev.drawTail, drawThen: ev.drawThen,
		}
		if g.replacementsAppliedThisEvent == nil {
			g.replacementsAppliedThisEvent = make(map[ReplacementEventID]map[ReplacementEffectID]bool)
		}
		inherited := make(map[ReplacementEffectID]bool, len(p.applied))
		for _, id := range p.applied {
			inherited[id] = true
		}
		g.replacementsAppliedThisEvent[next.ID] = inherited
		_, err := g.drawEventLocked(next)
		return err
	}
	err := g.drawRunLocked(ev.Actor, ev.drawTail, ev.drawThen)
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
			// A failed draw still finishes its instruction. A later
			// descendant may be replaced by another player drawing or
			// by returning a graveyard card instead.
			if tailErr := g.finishDrawTailLocked(ev); tailErr != nil {
				return tailErr
			}
		}
		return err
	}
	return g.finishDrawTailLocked(ev)
}

// --- dredge's home: the graveyard ------------------------------------

// graveyardReplacementIDBase is where graveyard replacements' IDs
// start. Inside the self-replacement range's headroom (its users sit in
// the first few strides above selfReplacementIDBase) and far below the
// scoped range. An ID packs (object serial, slot), so two dredge cards
// in one graveyard are two effects and one card with two of them is
// two. The serial, not the card's graveyard index: an ID outlives the
// dredge offer's pause, and an index shifts when an earlier card leaves
// the graveyard (#3482, replacement_serials.go).
const graveyardReplacementIDBase = selfReplacementIDBase + 1<<40

// graveyardReplacementIDSpan is the size of the graveyard range.
const graveyardReplacementIDSpan = 1 << 39

func encodeGraveyardReplacementID(serial, slot int) (ReplacementEffectID, bool) {
	if serial < 0 || slot < 0 || slot >= MaxCatalogReplacementSlots {
		return 0, false
	}
	raw := ReplacementEffectID(serial)*MaxCatalogReplacementSlots + ReplacementEffectID(slot)
	if raw >= graveyardReplacementIDSpan {
		return 0, false
	}
	return graveyardReplacementIDBase + raw, true
}

func decodeGraveyardReplacementID(id ReplacementEffectID) (serial, slot int, ok bool) {
	if id < graveyardReplacementIDBase || id >= graveyardReplacementIDBase+graveyardReplacementIDSpan {
		return 0, 0, false
	}
	raw := int(id - graveyardReplacementIDBase)
	return raw / MaxCatalogReplacementSlots, raw % MaxCatalogReplacementSlots, true
}

// gatherGraveyardReplacementsLocked appends the FromGraveyard
// replacements of the affected player's graveyard that apply to ev.
// Only a draw and a learn consult a graveyard, and only the affected
// player's: dredge says "your graveyard", and so does Retriever
// Phoenix.
//
// The source is a copy of the card, never a pointer into the zone's
// slice, because a prompt can outlive the slice it was gathered from.
//
// Caller must hold g.mu.
func (g *Game) gatherGraveyardReplacementsLocked(ev *ReplacementEvent, applied map[ReplacementEffectID]bool, serials *replacementSerialMinter, out []activeReplacement) []activeReplacement {
	if CatalogReplacements == nil || ev == nil {
		return out
	}
	var player uuid.UUID
	switch {
	case ev.Kind == RepEventDraw && ev.DrawCount > 0:
		player = ev.DrawPlayer
	case ev.Kind == RepEventKeywordAction && ev.KeywordAction == KeywordActionLearn:
		// #3525: the learning player's graveyard.
		player = ev.Actor
	default:
		return out
	}
	seat := -1
	for i, p := range g.Seats {
		if p != nil && p.ID == player {
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
		serial := -1
		for repIdx := range reps {
			eff := reps[repIdx]
			if !eff.FromGraveyard {
				continue
			}
			if serial < 0 {
				serial = serials.serial(card)
			}
			id, ok := encodeGraveyardReplacementID(serial, repIdx)
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
// prompt view: its label and its source card, found by the serial its
// ID packs among the graveyards' cards. A card that has left its
// graveyard names nothing.
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) graveyardReplacementMetaLocked(id ReplacementEffectID) (label string, card uuid.UUID, ok bool) {
	serial, slot, ok := decodeGraveyardReplacementID(id)
	if !ok {
		return "", uuid.Nil, false
	}
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for i := range p.Graveyard.Cards {
			c := &p.Graveyard.Cards[i]
			if s, known := g.replacementObjectSerialOfLocked(c); !known || s != serial {
				continue
			}
			reps := CatalogReplacements(catalogAbilityKeyOf(c))
			if slot >= len(reps) {
				return "", c.InstanceID, true
			}
			return reps[slot].Label, c.InstanceID, true
		}
	}
	return "", uuid.Nil, true
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
