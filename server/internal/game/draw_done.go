package game

import (
	"sort"

	"github.com/google/uuid"
)

// draw_done.go — #4253 slice 10c: the `done` a draw-instead body continues
// with is data, so a body that stops to ask (Underrealm Lich's pick, a
// graveyard return, a dredge) can hand it to a prompt whose table is a
// restore point.
//
// A draw-instead body (RegisterDrawInstead) runs once its replacement window
// has settled, may queue prompts, and must call `done` exactly once when it
// has finished: `done` pays what the replaced draw still owed, the rest of
// its instruction (draw_instead.go). That used to be a Go closure over a copy
// of the draw's frame, which a prompt's own continuation captured, so a table
// waiting on the prompt could be undone but not written to a restore point.
//
// DrawDone is a value with two forms. Inline, it holds the draw's frame by
// value and Run pays it. Parked, it is a handle: Park stores the frame on
// the game as plain data (Game.parkedDraws, written with the restore point
// as GameSnapshot.ParkedDraws) and returns an ID the body puts on its
// prompt's Carry; DrawDoneByID gives the handle back from the ID, in the
// restored binary, and Run pays the frame and forgets it. A parked frame no
// prompt ever answers (its chooser left) stays, as plain data, until the game
// ends: the closure it replaces was dropped with the prompt, and paid
// nothing either.

// DrawDone is what a draw-instead body calls when it has finished.
type DrawDone struct {
	frame *parkedDraw
	id    uuid.UUID
}

// parkedDraw is what a draw still owes once its substitute has finished: the
// drawing player, the expanded draws still to perform, the count of draws
// after this one and the "then" of the instruction. Plain data, copied by
// value into a resume frame (it is the part of ReplacementEvent that
// finishDrawTailLocked reads).
type parkedDraw struct {
	actor   uuid.UUID
	pending []pendingDraw
	tail    int
	then    DrawThen
}

// newParkedDraw copies the tail a draw owes out of its event.
func newParkedDraw(ev *ReplacementEvent) *parkedDraw {
	p := &parkedDraw{actor: ev.Actor, tail: ev.drawTail, then: ev.drawThen}
	for _, d := range ev.drawPending {
		d.applied = append([]ReplacementEffectID(nil), d.applied...)
		p.pending = append(p.pending, d)
	}
	return p
}

// event is the frame finishDrawTailLocked reads, rebuilt from the data.
func (p *parkedDraw) event() *ReplacementEvent {
	ev := &ReplacementEvent{Kind: RepEventDraw, Actor: p.actor, drawTail: p.tail, drawThen: p.then}
	for _, d := range p.pending {
		d.applied = append([]ReplacementEffectID(nil), d.applied...)
		ev.drawPending = append(ev.drawPending, d)
	}
	return ev
}

// Park stores the draw's frame on the game and returns its handle, for a
// body that hands `done` to a prompt (put the ID on its Carry, and rebuild
// the DrawDone with DrawDoneByID when the prompt is answered). Parking a
// handle that is already parked returns the same ID.
//
// Caller must hold g.mu.
func (d DrawDone) Park(g *Game) uuid.UUID {
	if d.id != uuid.Nil {
		return d.id
	}
	if d.frame == nil {
		return uuid.Nil
	}
	id := uuid.New()
	if g.parkedDraws == nil {
		g.parkedDraws = map[uuid.UUID]*parkedDraw{}
	}
	g.parkedDraws[id] = cloneParkedDraw(d.frame)
	return id
}

// DrawDoneByID is the handle for a frame Park stored. Running a handle whose
// frame is gone (it was paid already) pays nothing.
func DrawDoneByID(id uuid.UUID) DrawDone { return DrawDone{id: id} }

// Run pays what the replaced draw still owed: the draws still to perform,
// and the "then" of the instruction. A parked frame is forgotten as it is
// paid, so it is paid once.
//
// Caller must hold g.mu.
func (d DrawDone) Run(g *Game) error {
	if d.id != uuid.Nil {
		p := g.parkedDraws[d.id]
		if p == nil {
			return nil
		}
		delete(g.parkedDraws, d.id)
		if len(g.parkedDraws) == 0 {
			g.parkedDraws = nil
		}
		return g.finishDrawTailLocked(p.event())
	}
	if d.frame == nil {
		return nil
	}
	return g.finishDrawTailLocked(d.frame.event())
}

func cloneParkedDraw(p *parkedDraw) *parkedDraw {
	if p == nil {
		return nil
	}
	out := *p
	out.pending = nil
	for _, d := range p.pending {
		d.applied = append([]ReplacementEffectID(nil), d.applied...)
		out.pending = append(out.pending, d)
	}
	return &out
}

// cloneParkedDraws gives an undo clone its own frames.
func cloneParkedDraws(in map[uuid.UUID]*parkedDraw) map[uuid.UUID]*parkedDraw {
	if len(in) == 0 {
		return nil
	}
	out := make(map[uuid.UUID]*parkedDraw, len(in))
	for id, p := range in {
		out[id] = cloneParkedDraw(p)
	}
	return out
}

// parkedDrawSnapshot is a parked frame on disk.
type parkedDrawSnapshot struct {
	ID      uuid.UUID             `json:"id"`
	Actor   uuid.UUID             `json:"actor"`
	Pending []pendingDrawSnapshot `json:"pending,omitempty"`
	Tail    int                   `json:"tail,omitempty"`
	Then    *drawThenSnapshot     `json:"then,omitempty"`
}

// snapshotParkedDraws writes the parked frames, ordered by ID so two captures
// of one state are byte-identical.
func snapshotParkedDraws(in map[uuid.UUID]*parkedDraw) []parkedDrawSnapshot {
	if len(in) == 0 {
		return nil
	}
	out := make([]parkedDrawSnapshot, 0, len(in))
	for id, p := range in {
		s := parkedDrawSnapshot{ID: id, Actor: p.actor, Tail: p.tail}
		for _, d := range p.pending {
			s.Pending = append(s.Pending, pendingDrawSnapshot{Player: d.player, Count: d.count, Applied: append([]ReplacementEffectID(nil), d.applied...)})
		}
		if p.then.Ref.key != "" {
			s.Then = &drawThenSnapshot{
				Key: p.then.Ref.key, Player: uuidPtrOrNil(p.then.Player),
				Source: uuidPtrOrNil(p.then.Source), N: p.then.N,
			}
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID.String() < out[j].ID.String() })
	return out
}

// restoreParkedDraws rebuilds the parked frames. A file written before them
// restores none.
func restoreParkedDraws(in []parkedDrawSnapshot) map[uuid.UUID]*parkedDraw {
	if len(in) == 0 {
		return nil
	}
	out := make(map[uuid.UUID]*parkedDraw, len(in))
	for _, s := range in {
		p := &parkedDraw{actor: s.Actor, tail: s.Tail}
		for _, d := range s.Pending {
			p.pending = append(p.pending, pendingDraw{player: d.Player, count: d.Count, applied: append([]ReplacementEffectID(nil), d.Applied...)})
		}
		if d := s.Then; d != nil {
			p.then = DrawThen{Ref: DrawThenRef{key: d.Key}, Player: uuidOrNil(d.Player), Source: uuidOrNil(d.Source), N: d.N}
		}
		out[s.ID] = p
	}
	return out
}

// unknownParkedDrawKeys is checkEffectKeys' line for a parked frame whose
// "then" this binary does not register: restored without it, the rest of
// "draw, then ..." would never run.
func unknownParkedDrawKeys(in []parkedDrawSnapshot) []string {
	var out []string
	for _, s := range in {
		if d := s.Then; d != nil && !KnownDrawThen(d.Key) {
			out = append(out, "draw-then body "+d.Key)
		}
	}
	return out
}
