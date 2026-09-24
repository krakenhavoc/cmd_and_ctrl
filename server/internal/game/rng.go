package game

import (
	"crypto/hmac"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// rng.go is the game's ONE source of randomness (ADR 0054 Decisions
// 1-3). Every shuffle, random discard and random-order bottom in
// internal/game draws through randForLocked; nothing else in the
// package touches math/rand/v2 for drawing (TestNoDirectRandomSource
// keeps it that way).
//
// The state is not a stream position but a SECRET KEY plus a map of
// per-stream counters:
//
//	rngKey      [32]byte          minted at Start; never changes
//	rngCounters map[string]uint64 stream name -> draws taken this turn
//	rngTurn     int               the turn index rngCounters belong to
//
// Each random operation (one shuffle, one pick of k cards, and in
// sub-PR 2 one roll of N dice) gets its own ChaCha8 generator seeded
// by HMAC-SHA256(key, stream identity ‖ turn ‖ counter). Three things
// follow, and they are why this is keyed streams rather than one
// copied PCG:
//
//   - UNDO REWINDS. Clone and RestoreFrom carry the counters, so the
//     same action after an undo gets the same draw (ADR 0054 Decision
//     4, the owner's call). Undo can no longer be used to reshuffle.
//   - SPENDING A DRAW ELSEWHERE DOES NOT MOVE YOURS. Streams are
//     independent, so an undo followed by a shuffle somewhere else
//     does not change what a different stream gives next.
//   - OBSERVED OUTPUTS REVEAL NOTHING. Every seed is an HMAC of a key
//     that never leaves the server (it is in the persisted engine
//     snapshot, never in GameView, a replay line or a crash dump).
//
// A zero key means "not minted yet". A game that never called Start
// (many unit tests) mints a crypto-random key on its first draw, so
// there is no nil source and no process-global fallback anywhere.

// Stream kinds. The kind is part of the stream identity, so two
// operations of different kinds never share a counter.
const (
	rngStreamShuffle     = "shuffle"
	rngStreamPick        = "pick"
	rngStreamRandomOrder = "random_order"
	rngStreamRoll        = "roll"
	rngStreamFlip        = "flip"
)

// rngDomain separates this derivation from any other use of HMAC
// with the same key. Changing it, or anything in rngSeed, changes
// what every restored game draws next; TestRNGKnownAnswer pins it.
const rngDomain = "cmdctrl/rng/v1"

// rngStream names one random stream: what kind of operation, whose
// (library / hand / roll), and which object's effect is drawing.
type rngStream struct {
	kind   string
	player uuid.UUID // whose library / who discards; uuid.Nil if nobody
	source uuid.UUID // the object whose effect draws; uuid.Nil if none
}

// randForLocked returns a *rand.Rand for ONE random operation (one
// shuffle, one pick). The returned generator is private to the
// caller; drawing from it any number of times does not affect any
// other operation. Caller must hold g.mu for writing.
//
// The player half of the identity is the player's SEAT when the
// player is seated, not their UUID. Player UUIDs are minted fresh by
// every AddPlayer, so keying by UUID would make a seeded test game
// shuffle differently on every run; seats are permanent once a game
// starts (RemovePlayer is lobby-only). An unseated UUID (or uuid.Nil)
// is keyed by its bytes.
func (g *Game) randForLocked(s rngStream) *rand.Rand {
	if g.rngKey == ([32]byte{}) {
		g.rngKey = mintRNGKey()
	}
	if turn := g.rngTurnIndexLocked(); g.rngTurn != turn {
		g.rngCounters = nil
		g.rngTurn = turn
	}
	if g.rngCounters == nil {
		g.rngCounters = make(map[string]uint64)
	}
	who, label := g.rngPlayerIdentityLocked(s.player)
	source, sourceLabel := g.rngSourceIdentityLocked(s.source)
	name := s.kind + "/" + label + "/" + sourceLabel
	n := g.rngCounters[name]
	g.rngCounters[name] = n + 1
	return rand.New(rand.NewChaCha8(rngSeed(g.rngKey, s.kind, who, uuid.UUID(source), g.rngTurn, n)))
}

// rngSourceIdentityLocked returns the source half of a stream identity.
// uuid.Nil deliberately keeps its old representation, preserving existing
// un-sourced stream counters and their known-answer values.
func (g *Game) rngSourceIdentityLocked(id uuid.UUID) ([16]byte, string) {
	if id == uuid.Nil {
		return [16]byte{}, id.String()
	}
	if ordinal, ok := g.sourceOrdinals[id]; ok {
		var b [16]byte
		copy(b[:11], "cmdctrl-obj")
		b[11] = byte(ordinal >> 32)
		binary.BigEndian.PutUint32(b[12:], uint32(ordinal))
		return b, "obj" + strconv.FormatUint(ordinal>>32, 10) + ":" + strconv.FormatUint(uint64(uint32(ordinal)), 10)
	}
	return [16]byte(id), id.String()
}

func (g *Game) setDeckSourceOrdinalLocked(id uuid.UUID, seat, index int) {
	if id == uuid.Nil {
		return
	}
	if g.sourceOrdinals == nil {
		g.sourceOrdinals = make(map[uuid.UUID]uint64)
	}
	g.sourceOrdinals[id] = uint64(uint32(seat)<<16 | uint32(index))
}

// noteCreatedSourceLocked assigns class 1 at creation time. It is safe to
// call more than once for the same object (for example, from an event path).
func (g *Game) noteCreatedSourceLocked(id uuid.UUID) {
	if id == uuid.Nil {
		return
	}
	if g.sourceOrdinals == nil {
		g.sourceOrdinals = make(map[uuid.UUID]uint64)
	}
	if _, ok := g.sourceOrdinals[id]; ok {
		return
	}
	g.sourceOrdinalNext++
	g.sourceOrdinals[id] = uint64(1)<<32 | g.sourceOrdinalNext
}

// rngTurnIndexLocked is the turn a stream's counters are scoped to.
//
// Turn.Seq changes at every turn boundary, including consecutive turns
// taken by the same seat. That is exactly the boundary at which a new
// deterministic stream must replace anything an undo exposed.
func (g *Game) rngTurnIndexLocked() int {
	return g.Turn.Seq
}

// rngPlayerIdentityLocked returns the 16 bytes that stand for a
// player in the seed, plus the same identity as a readable label for
// the counter map (which a snapshot stores). See randForLocked for
// why a seated player is their seat.
func (g *Game) rngPlayerIdentityLocked(id uuid.UUID) ([16]byte, string) {
	if id != uuid.Nil {
		for _, p := range g.Seats {
			if p != nil && p.ID == id {
				var b [16]byte
				// A v4 UUID has 0x4_ at byte 6; "cmdctrl-seat" puts
				// 'r' there, so a seat identity cannot equal a
				// player UUID.
				copy(b[:12], "cmdctrl-seat")
				binary.BigEndian.PutUint32(b[12:], uint32(p.Seat))
				return b, "seat" + strconv.Itoa(p.Seat)
			}
		}
	}
	return [16]byte(id), id.String()
}

// rngSeed is the derivation, written out so a test can pin it:
//
//	HMAC-SHA256(key, "cmdctrl/rng/v1" ‖ u64(len(kind)) ‖ kind ‖
//	            player[16] ‖ source[16] ‖ u64(turn) ‖ u64(counter))
//
// with every u64 big-endian and turn as its two's-complement bits.
func rngSeed(key [32]byte, kind string, player [16]byte, source uuid.UUID, turn int, counter uint64) [32]byte {
	mac := hmac.New(sha256.New, key[:])
	var u [8]byte
	mac.Write([]byte(rngDomain))
	binary.BigEndian.PutUint64(u[:], uint64(len(kind)))
	mac.Write(u[:])
	mac.Write([]byte(kind))
	mac.Write(player[:])
	mac.Write(source[:])
	binary.BigEndian.PutUint64(u[:], uint64(int64(turn)))
	mac.Write(u[:])
	binary.BigEndian.PutUint64(u[:], counter)
	mac.Write(u[:])
	var out [32]byte
	copy(out[:], mac.Sum(nil))
	return out
}

// pickAtRandomLocked returns k distinct IDs from ids, uniformly at
// random and in a random order; k >= len(ids) returns all of them in
// a random order, k <= 0 returns nil. One call is ONE operation on
// stream s. It is the draw ADR 0054 Decision 5's
// ChooseAtRandomForEffect will expose; today only the random discard
// uses it. Caller must hold g.mu for writing.
func (g *Game) pickAtRandomLocked(s rngStream, ids []uuid.UUID, k int) []uuid.UUID {
	if k <= 0 || len(ids) == 0 {
		return nil
	}
	out := append([]uuid.UUID(nil), ids...)
	g.randForLocked(s).Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	if k < len(out) {
		out = out[:k]
	}
	return out
}

// mintRNGKey returns a fresh key from crypto/rand. It is never zero:
// zero is the "not minted" sentinel.
func mintRNGKey() [32]byte {
	var k [32]byte
	if _, err := crand.Read(k[:]); err != nil {
		// crypto/rand.Read does not fail on any platform we run on.
		// If it ever does, a clock-derived key is still unique per
		// game, which a zero key (identical for every game) is not.
		var t [8]byte
		binary.BigEndian.PutUint64(t[:], uint64(time.Now().UnixNano()))
		k = sha256.Sum256(append([]byte(rngDomain+"/fallback"), t[:]...))
	}
	return nonZeroKey(k)
}

// rngKeyFrom derives a key by reading 32 bytes from a caller-supplied
// source, so a test that starts a game from a seeded *rand.Rand gets
// the same key, and therefore the same draws, on every run.
func rngKeyFrom(src rand.Source) [32]byte {
	r := rand.New(src)
	var k [32]byte
	for i := 0; i < 4; i++ {
		binary.LittleEndian.PutUint64(k[i*8:], r.Uint64())
	}
	return nonZeroKey(k)
}

func nonZeroKey(k [32]byte) [32]byte {
	if k == ([32]byte{}) {
		k[0] = 1
	}
	return k
}

// cloneRNGCounters deep-copies the counter map. Clone copies rather
// than shares it: randForLocked increments it in place, and a shared
// map would let the live game advance the undo entry's counters, which
// is exactly the rewind this file exists for.
func cloneRNGCounters(m map[string]uint64) map[string]uint64 {
	if m == nil {
		return nil
	}
	out := make(map[string]uint64, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// SetRNGKeyForTest installs a fixed key and clears the counters, so a
// test that never calls Start (or wants a specific key) gets
// deterministic, rewindable draws. A later Start(nil) keeps the key;
// Start with a source replaces it. Panics on the all-zero key, which
// is the "not minted" sentinel and would be silently replaced.
func (g *Game) SetRNGKeyForTest(key [32]byte) {
	if key == ([32]byte{}) {
		panic("game: SetRNGKeyForTest with the all-zero key")
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.rngKey = key
	g.rngCounters = nil
	g.rngTurn = g.rngTurnIndexLocked()
}
