package gamestats

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// sampler_test.go drives the turn sampler (ADR 0217 §1, §4) with frames
// built by hand, for the cases an engine game is slow to reach.

func land(id string) protocol.CardView {
	return protocol.CardView{InstanceID: id, Name: "Forest", Owner: ana, Controller: ana, TypeLine: "Basic Land — Forest",
		ManaAbilities: []protocol.ManaAbilityView{{Produced: "{G}"}}}
}

func spell(id, name, cost string) protocol.CardView {
	return protocol.CardView{InstanceID: id, Name: name, Owner: ana, Controller: ana, TypeLine: "Sorcery", ManaCost: cost}
}

// frame is Ana's (seat 0) turn at a step, with her hand and her board.
// seq is the newest log Seq the frame carries.
func frame(seq uint64, turn int, stepName string, landsPlayed int, hand, board []protocol.CardView) *protocol.GameView {
	return &protocol.GameView{
		State:       "active",
		Turn:        protocol.TurnView{Seq: turn, ActiveSeat: 0, Step: stepName, PhaseID: 1},
		Battlefield: protocol.ZoneView{Cards: board},
		Log:         []protocol.LogEvent{{Seq: seq}},
		Seats: []protocol.PlayerView{
			{ID: ana, Seat: 0, LandsPlayedThisTurn: landsPlayed, Hand: protocol.ZoneView{Cards: hand}},
			{ID: bo, Seat: 1},
		},
	}
}

func framesOf(fs ...*protocol.GameView) Frames {
	return func(yield func(*protocol.GameView) bool) bool {
		for _, f := range fs {
			if !yield(f) {
				return false
			}
		}
		return len(fs) > 0
	}
}

func playOf(t *testing.T, fs ...*protocol.GameView) *Play {
	t.Helper()
	g := Compute(Input{
		View: protocol.GameView{Seats: []protocol.PlayerView{{ID: ana, Seat: 0}, {ID: bo, Seat: 1}}},
		Log: []protocol.LogEvent{
			{Kind: protocol.LogCast, Turn: 1, Seat: 0, CardID: "x"},
			{Kind: protocol.LogDamage, Turn: 3, Seat: 0, TargetSeat: seatPtr(1), Amount: 4},
		},
		Frames: framesOf(fs...),
	})
	if g.Seats[0].Play == nil {
		t.Fatal("no Play")
	}
	return g.Seats[0].Play
}

func TestSamplerScrewFloodAndColourScrew(t *testing.T) {
	forest := land("f1")
	bear := protocol.CardView{InstanceID: "b1", Name: "Grizzly Bears", Owner: ana, Controller: ana, TypeLine: "Creature — Bear", ManaCost: "{1}{G}"}
	counter := spell("u1", "Counterspell", "{U}{U}")
	p := playOf(t,
		// Turn 1: no land in hand at all, no land played. Screw.
		frame(1, 1, "draw", 0, []protocol.CardView{counter}, nil),
		frame(2, 1, "precombat_main", 0, []protocol.CardView{counter}, nil),
		frame(3, 1, "end", 0, []protocol.CardView{counter}, nil),
		// Turn 3 (Bo had turn 2): a Forest drawn and played; two Forests
		// against the one Counterspell left at the end: flood. With one
		// untapped Forest, Counterspell is castable by amount? No: it
		// costs two. A Bear is not in hand.
		frame(4, 3, "draw", 0, []protocol.CardView{counter, forest, land("f2"), land("f3")}, nil),
		frame(5, 3, "precombat_main", 0, []protocol.CardView{counter, forest, land("f2"), land("f3")}, nil),
		frame(6, 3, "postcombat_main", 1, []protocol.CardView{counter, land("f2"), land("f3")}, []protocol.CardView{forest}),
		// Turn 5: two Forests out, Counterspell castable by amount but
		// not colour: colour screw. A land played.
		frame(7, 5, "precombat_main", 0, []protocol.CardView{counter, land("f3")}, []protocol.CardView{forest, land("f2")}),
		frame(8, 5, "declare_attackers", 1, []protocol.CardView{counter}, []protocol.CardView{forest, land("f2"), land("f3"), bear}),
	)
	if p.OwnTurns != 3 {
		t.Fatalf("own turns %d, want 3 (Bo's turn is not Ana's)", p.OwnTurns)
	}
	r := p.Turns
	if !r[0].Missed || !r[0].Screw || r[0].LandDrops != 0 || r[0].SpellsCast != 1 {
		t.Errorf("turn 1: %+v, want a missed drop with no land in hand (screw) and the one spell cast", r[0])
	}
	if r[1].Missed || !r[1].Flood || r[1].LandDrops != 1 || r[1].LandsInHand != 3 || r[1].DamageToPlayers != 4 {
		t.Errorf("turn 3: %+v, want a drop hit with more lands than spells left (flood), 4 damage", r[1])
	}
	if !r[2].ColourScrew || r[2].ManaSources != 2 || r[2].LandsControlled != 2 || r[2].LandDrops != 2 {
		t.Errorf("turn 5: %+v, want colour screw with two sources", r[2])
	}
	if p.LandDropsHit != 2 || p.LandDropsMissed != 1 || p.ScrewTurns != 1 || p.FloodTurns != 1 || p.ColourScrewTurns != 1 {
		t.Errorf("totals: %+v", p)
	}
	if p.ScrewGame || p.FloodGame || p.ColourScrewGame {
		t.Errorf("one turn of each is below every game threshold: %+v", p)
	}
	var cs SeenCard
	for _, c := range p.Seen {
		if c.Name == "Counterspell" {
			cs = c
		}
	}
	if !cs.InOpeningHand || cs.HeldTurns != 3 || cs.AffordableTurns != 0 || cs.Cast || cs.FirstSeenTurn != 0 {
		t.Errorf("Counterspell: %+v, want held three turns, never affordable, never cast, in the opening hand", cs)
	}
}

func TestSamplerAttackShare(t *testing.T) {
	bear := func(id string, opts ...func(*protocol.CardView)) protocol.CardView {
		c := protocol.CardView{InstanceID: id, Name: "Grizzly Bears", Owner: ana, Controller: ana, TypeLine: "Creature — Bear"}
		for _, o := range opts {
			o(&c)
		}
		return c
	}
	attacking := func(c *protocol.CardView) { c.AttackingTarget, c.Tapped = bo, true }
	sick := func(c *protocol.CardView) { c.SummoningSick = true }
	tapped := func(c *protocol.CardView) { c.Tapped = true }
	p := playOf(t,
		frame(1, 1, "precombat_main", 1, nil, nil),
		frame(2, 1, "declare_attackers", 1, nil, []protocol.CardView{bear("a"), bear("b"), bear("c", sick), bear("d", tapped)}),
		frame(3, 1, "declare_attackers", 1, nil, []protocol.CardView{bear("a", attacking), bear("b"), bear("c", sick), bear("d", tapped)}),
	)
	// The last frame of the step: one attacker of two that could (the
	// summoning-sick one and the tapped one could not).
	if p.CreaturesAttacked != 1 || p.CreaturesEligible != 2 {
		t.Errorf("attacked %d of %d eligible, want 1 of 2", p.CreaturesAttacked, p.CreaturesEligible)
	}
}

// An undo rewinds the log's Seq: what the undone frames saw is dropped.
func TestSamplerRewindDropsUndoneSamples(t *testing.T) {
	keep := spell("k", "Divination", "{2}{U}")
	drawn := spell("d", "Brainstorm", "{U}")
	p := playOf(t,
		frame(10, 1, "precombat_main", 0, []protocol.CardView{keep}, nil),
		// A draw and a land play, then an undo back to Seq 10.
		frame(11, 1, "precombat_main", 0, []protocol.CardView{keep, drawn, land("f1")}, nil),
		frame(12, 1, "precombat_main", 1, []protocol.CardView{keep, drawn}, []protocol.CardView{land("f1")}),
		frame(10, 1, "precombat_main", 0, []protocol.CardView{keep}, nil),
		frame(13, 1, "end", 0, []protocol.CardView{keep}, nil),
	)
	if r := p.Turns[0]; !r.Missed || !r.Screw {
		t.Errorf("turn 1: %+v; the land and its play were undone, so this is a screw turn", r)
	}
	for _, c := range p.Seen {
		if c.Name == "Brainstorm" || c.Name == "Forest" {
			t.Errorf("%s was only in hand in undone frames, but is seen: %+v", c.Name, c)
		}
	}
	// Without the rewind the undone Forest would have made it no screw.
	p = playOf(t,
		frame(10, 1, "precombat_main", 0, []protocol.CardView{keep}, nil),
		frame(11, 1, "precombat_main", 0, []protocol.CardView{keep, land("f1")}, nil),
		frame(13, 1, "end", 0, []protocol.CardView{keep, land("f1")}, nil),
	)
	if r := p.Turns[0]; !r.Missed || r.Screw {
		t.Errorf("a land kept in hand: %+v, want a missed drop that is not screw", r)
	}
}

func TestSamplerGameThresholds(t *testing.T) {
	var fs []*protocol.GameView
	seq := uint64(0)
	for turn := 1; turn <= 13; turn += 2 {
		seq++
		fs = append(fs, frame(seq, turn, "precombat_main", 0, nil, nil))
	}
	p := playOf(t, fs...)
	if p.OwnTurns != 7 || p.ScrewTurns != 7 || !p.ScrewGame {
		t.Errorf("seven screw turns: %+v, want a screwed game", p)
	}
	var none []*protocol.GameView
	for i, turn := range []int{1, 3, 5, 7, 9, 11, 13, 15} {
		hand := []protocol.CardView{land("x")}
		if i >= 6 {
			hand = nil // screw only after the first six own turns
		}
		none = append(none, frame(uint64(i+1), turn, "precombat_main", 0, hand, nil))
	}
	if p := playOf(t, none...); p.ScrewTurns != 2 || p.ScrewGame {
		t.Errorf("two screw turns after the first six: %+v, want no screwed game", p)
	}
}
