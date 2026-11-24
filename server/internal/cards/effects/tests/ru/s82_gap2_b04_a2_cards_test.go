package testsru

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// S82 GAP2 batch 04 (worker a2): Talon of Pain and Necron Overlord, the
// two cards of #6085's one-announced-X seam (CR 107.3a).

const (
	s82g2a2Talon  = "0e4d2fbc-1b34-4501-9022-8ea800cb2d24"
	s82g2a2Necron = "2c0de95f-4269-439c-a1bc-faa11348b185"
)

func TestS82Gap2B04A2CardsAreFullWithNoCaveats(t *testing.T) {
	for _, oracle := range []string{s82g2a2Talon, s82g2a2Necron} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %v with caveats %v, want full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// s82g2a2Activation is the part of an activate move's params these tests
// read.
type s82g2a2Activation struct {
	XValue        int      `json:"x_value"`
	CounterCounts []int    `json:"counter_counts"`
	TapIDs        []string `json:"tap_ids"`
}

func s82g2a2Moves(t *testing.T, g *game.Game, seat, source uuid.UUID) []s82g2a2Activation {
	t.Helper()
	var out []s82g2a2Activation
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != legal.TypeActivateAbility || m.Source != source {
			continue
		}
		var a s82g2a2Activation
		if err := json.Unmarshal(m.Params, &a); err != nil {
			t.Fatalf("move params: %v", err)
		}
		out = append(out, a)
	}
	return out
}

// Talon of Pain counts a source you control hitting an opponent — not
// itself, not damage to you, not damage to a creature — then pays {X} and
// X charge counters with one X and deals X damage.
func TestS82Gap2B04A2TalonOfPainCountsAndShootsTheOneX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	talon := effectstest.PushCatalogPermanent(g, me.ID, "Talon of Pain", "Artifact", s82g2a2Talon, false)
	mine := effectstest.P2Tough(g, me.ID, "Mine")
	theirs := effectstest.P2Tough(g, opp.ID, "Theirs")

	effectstest.P2Ping(t, g, mine, theirs, 1)
	effectstest.P2Ping(t, g, mine, me.ID, 1)
	effectstest.P2Ping(t, g, talon, opp.ID, 1)
	effectstest.P2Ping(t, g, theirs, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, talon, game.CounterCharge); got != 0 {
		t.Fatalf("%d charge counters from damage that does not count, want 0", got)
	}
	for i := 0; i < 3; i++ {
		effectstest.P2Ping(t, g, mine, opp.ID, 1)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.CounterCount(g, talon, game.CounterCharge); got != 3 {
		t.Fatalf("%d charge counters, want 3", got)
	}

	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "C", 5)
	// The bot is offered the largest X both the pool and the counters
	// pay: three, not five.
	moves := s82g2a2Moves(t, g, me.ID, talon)
	if len(moves) == 0 {
		t.Fatal("no Talon activation offered")
	}
	for _, m := range moves {
		if m.XValue != 3 || (len(m.CounterCounts) > 0 && (len(m.CounterCounts) != 1 || m.CounterCounts[0] != 3)) {
			t.Errorf("offered X=%d removing %v, want X=3 removing 3", m.XValue, m.CounterCounts)
		}
	}

	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	// Four is more than the counters held: refused, nothing paid.
	if err := g.ActivateCatalogAbility(me.ID, talon, 0, game.ActivateAbilityParams{Strict: true, XValue: 4, Targets: target}); !errors.Is(err, game.ErrInsufficientCounters) {
		t.Fatalf("X=4 with three counters: %v, want ErrInsufficientCounters", err)
	}
	// A separate count that disagrees with X is refused.
	if err := g.ActivateCatalogAbility(me.ID, talon, 0, game.ActivateAbilityParams{Strict: true, XValue: 1, CounterCounts: []int{3}, Targets: target}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("X=1 removing three: %v, want ErrInvalidParam", err)
	}
	if got := effectstest.CounterCount(g, talon, game.CounterCharge); got != 3 || me.ManaPool.Total() != 5 {
		t.Fatalf("a refused activation paid: %d counters, %d mana", got, me.ManaPool.Total())
	}

	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, talon, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, Targets: target}); err != nil {
		t.Fatalf("X=2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("opponent at %d, want %d", opp.Life, before-2)
	}
	if got := effectstest.CounterCount(g, talon, game.CounterCharge); got != 1 {
		t.Errorf("%d charge counters left, want 1", got)
	}
	if me.ManaPool.Total() != 3 {
		t.Errorf("%d mana left, want 3 — {X} charged at X=2", me.ManaPool.Total())
	}
	// The Talon's own damage to the opponent adds no counter.
	if got := effectstest.CounterCount(g, talon, game.CounterCharge); got != 1 {
		t.Errorf("the Talon's own damage added a counter: %d", got)
	}
}

// With no counters the Talon is not worth offering (X would be 0), and an
// X above zero is refused.
func TestS82Gap2B04A2TalonOfPainWithNoCountersIsNotOffered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	talon := effectstest.PushCatalogPermanent(g, me.ID, "Talon of Pain", "Artifact", s82g2a2Talon, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "C", 3)
	if moves := s82g2a2Moves(t, g, me.ID, talon); len(moves) != 0 {
		t.Errorf("offered %+v with no charge counters", moves)
	}
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.ActivateCatalogAbility(me.ID, talon, 0, game.ActivateAbilityParams{Strict: true, XValue: 1, Targets: target}); !errors.Is(err, game.ErrInsufficientCounters) {
		t.Fatalf("X=1 with no counters: %v, want ErrInsufficientCounters", err)
	}
}

// Necron Overlord: X=2 pays {2} and taps two other artifacts; the
// Overlord, tapped by its own {T}, is never one of them; too few artifacts
// or a short tap list is refused.
func TestS82Gap2B04A2NecronOverlordDrainsTheOneX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	necron := effectstest.PushCatalogPermanent(g, me.ID, "Necron Overlord", "Artifact Creature — Necron Noble", s82g2a2Necron, false)
	rockA := effectstest.PushCatalogPermanent(g, me.ID, "Rock A", "Artifact", "", false)
	rockB := effectstest.PushCatalogPermanent(g, me.ID, "Rock B", "Artifact", "", false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "C", 2)

	moves := s82g2a2Moves(t, g, me.ID, necron)
	if len(moves) == 0 {
		t.Fatal("no Necron Overlord activation offered")
	}
	for _, m := range moves {
		if m.XValue < 1 || m.XValue != len(m.TapIDs) || m.XValue > 2 {
			t.Errorf("offered X=%d tapping %d, want X = taps, 1..2", m.XValue, len(m.TapIDs))
		}
		for _, id := range m.TapIDs {
			if id == necron.String() {
				t.Error("offered the Overlord as one of its own X artifacts")
			}
		}
	}

	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.ActivateCatalogAbility(me.ID, necron, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, TapIDs: []uuid.UUID{rockA}, Targets: target}); err == nil {
		t.Fatal("X=2 tapping one artifact was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, necron, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, TapIDs: []uuid.UUID{rockA, necron}, Targets: target}); err == nil {
		t.Fatal("the Overlord paid both its {T} and one of the X artifacts")
	}
	if err := g.ActivateCatalogAbility(me.ID, necron, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, TapIDs: []uuid.UUID{rockA, rockB},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Fatal("targeting yourself was accepted")
	}
	if me.ManaPool.Total() != 2 {
		t.Fatalf("a refused activation spent mana: %d left", me.ManaPool.Total())
	}

	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, necron, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, TapIDs: []uuid.UUID{rockA, rockB}, Targets: target}); err != nil {
		t.Fatalf("X=2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("opponent at %d, want %d", opp.Life, before-2)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left, want 0 — {X} charged at X=2", me.ManaPool.Total())
	}
}

// A summoning-sick Overlord cannot pay its {T}.
func TestS82Gap2B04A2NecronOverlordIsSummoningSick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	necron := effectstest.PushCatalogPermanent(g, me.ID, "Necron Overlord", "Artifact Creature — Necron Noble", s82g2a2Necron, true)
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "C", 1)
	if err := g.ActivateCatalogAbility(me.ID, necron, 0, game.ActivateAbilityParams{Strict: true, XValue: 1, TapIDs: []uuid.UUID{rock},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("a summoning-sick Overlord activated its {T} ability")
	}
}
