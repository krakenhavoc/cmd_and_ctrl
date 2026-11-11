package draft

import "math/rand/v2"

// PickRequest is what a Picker is asked: the pack the seat holds, what
// it has drafted so far and the pick mode.
type PickRequest struct {
	// Pack is the pack the seat holds. Never empty.
	Pack []Card
	// Picked is every card the seat has drafted, in the order taken.
	Picked []Card
	Mode   Mode
	// Take is how many cards this pick takes: Mode.PerPick(), or fewer
	// when the pack is nearly empty. Always at least 1 and at most
	// len(Pack).
	Take int

	// Where the pick falls. A picker may ignore all of it.
	//
	// Seat is the seat's number from 0 and Seats the size of the draft.
	// Round and Pick count from 1.
	Seat, Seats int
	Round, Pick int
}

// Picker chooses a pick for a seat no person is choosing for: a bot
// seat, a seat whose pick timer ran out, and a seat that is away.
//
// It returns the IDs of exactly Take different cards of Pack. Anything
// else is overruled with the first cards of the pack.
//
// This interface is the whole seam for a draft bot. Slice 5 (#5439)
// implements it with a bot that rates cards; nothing else in this
// package changes for that. A Picker is called with the draft's lock
// held, so it must not block and must not call back into the draft.
type Picker interface {
	Pick(PickRequest) []string
}

// RandomPicker is the placeholder bot: it takes cards of the pack at
// random.
//
// Its choice depends only on the seed and on where the pick falls (the
// seat, the round and the pick number), never on the order picks were
// asked for. So a seeded draft replays the same way whatever the timing
// of its seats.
type RandomPicker struct {
	Seed uint64
}

// Pick takes Take cards of the pack at random.
func (p RandomPicker) Pick(req PickRequest) []string {
	// One PCG stream per pick. The high bit keeps every one clear of
	// packStream, so the packs and the picks never share a sequence.
	stream := 1<<63 | uint64(req.Seat)<<32 | uint64(req.Round)<<16 | uint64(req.Pick)
	rng := rand.New(rand.NewPCG(p.Seed, stream))
	order := rng.Perm(len(req.Pack))
	n := min(max(req.Take, 0), len(order))
	out := make([]string, 0, n)
	for _, i := range order[:n] {
		out = append(out, req.Pack[i].ID)
	}
	return out
}
