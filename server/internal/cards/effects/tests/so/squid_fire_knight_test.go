package testsso

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// squid_fire_knight_test.go — #4307: Squid Fire Knight's dice sticker and
// the one dice door (game/dice_roll.go). "The next time target player would
// roll one or more dice this turn, instead they roll that many dice plus
// one, then you choose one of those rolls to ignore."

const componentPouchOracle = "4bb1863e-50f3-4e08-884e-1f58ee55817f"

// sfkScene is the Squid sticker on the seat's Host, activated at `target`.
func sfkScene(t *testing.T, target func(s *effectstest.StkScene) uuid.UUID) *effectstest.StkScene {
	t.Helper()
	s := effectstest.NewStkScene(t, "Squid Fire Knight", "next time target player would roll")
	s.Activate(0, "", game.TargetRef{Kind: game.TargetPlayer, ID: target(s)})
	return s
}

// sfkAnnounced is the EventRollDie amounts emitted after event index from.
func sfkAnnounced(g *game.Game, from int) []int {
	var out []int
	g.WithWriteLock(func() {
		for _, ev := range g.Events[from:] {
			if ev.Kind == game.EventRollDie {
				out = append(out, ev.Amount)
			}
		}
	})
	return out
}

// sfkRoll rolls n d<sides> for player through the engine's door.
func sfkRoll(t *testing.T, g *game.Game, player uuid.UUID, sides, n int) {
	t.Helper()
	var err error
	g.WithWriteLock(func() {
		err = g.RollDiceThenForEffect(game.DiceRoll{Player: player, Source: uuid.New(), Sides: sides, N: n})
	})
	if err != nil {
		t.Fatalf("roll: %v", err)
	}
}

// One die becomes two; the sticker's controller ignores one; only the
// other is rolled as far as the game is concerned. The shield is spent.
func TestSquidFireKnightOneDieBecomesTwoAndTheChosenIsIgnored(t *testing.T) {
	s := sfkScene(t, func(s *effectstest.StkScene) uuid.UUID { return s.Opp.ID })
	from := effectstest.SfkEventCount(s.G)
	sfkRoll(t, s.G, s.Opp.ID, 20, 1)
	if got := sfkAnnounced(s.G, from); len(got) != 0 {
		t.Fatalf("rolls announced before the choice: %v", got)
	}
	c := effectstest.SfkPrompt(s.G, "dice/ignore-roll")
	if c == nil {
		t.Fatal("no roll-to-ignore question")
	}
	rolls := effectstest.SfkRolls(t, c)
	if len(rolls) != 2 {
		t.Fatalf("%d dice rolled, want 2", len(rolls))
	}
	effectstest.SfkAnswer(t, s.G, c, s.Me.ID, 0)
	if got := sfkAnnounced(s.G, from); !effectstest.SameInts(got, rolls[1:]) {
		t.Errorf("announced %v, want %v (the first ignored)", got, rolls[1:])
	}
	// Spent: the next roll is the player's own.
	from = effectstest.SfkEventCount(s.G)
	sfkRoll(t, s.G, s.Opp.ID, 20, 1)
	if effectstest.SfkPrompt(s.G, "dice/ignore-roll") != nil || len(sfkAnnounced(s.G, from)) != 1 {
		t.Error("the shield applied to a second roll")
	}
}

// Several dice: three d6 become four, and the roll at the index chosen is
// the one that never happened, wherever it is.
func TestSquidFireKnightSeveralDice(t *testing.T) {
	for drop := 0; drop < 4; drop++ {
		s := sfkScene(t, func(s *effectstest.StkScene) uuid.UUID { return s.Opp.ID })
		from := effectstest.SfkEventCount(s.G)
		sfkRoll(t, s.G, s.Opp.ID, 6, 3)
		c := effectstest.SfkPrompt(s.G, "dice/ignore-roll")
		if c == nil {
			t.Fatal("no roll-to-ignore question")
		}
		rolls := effectstest.SfkRolls(t, c)
		if len(rolls) != 4 {
			t.Fatalf("%d dice rolled, want 4", len(rolls))
		}
		effectstest.SfkAnswer(t, s.G, c, s.Me.ID, drop)
		want := append(append([]int(nil), rolls[:drop]...), rolls[drop+1:]...)
		if got := sfkAnnounced(s.G, from); !effectstest.SameInts(got, want) {
			t.Errorf("drop %d: announced %v, want %v", drop, got, want)
		}
	}
}

// The choice is the sticker's controller's, not the roller's.
func TestSquidFireKnightTheStickersControllerChooses(t *testing.T) {
	s := sfkScene(t, func(s *effectstest.StkScene) uuid.UUID { return s.Opp.ID })
	sfkRoll(t, s.G, s.Opp.ID, 20, 2)
	c := effectstest.SfkPrompt(s.G, "dice/ignore-roll")
	if c == nil {
		t.Fatal("no roll-to-ignore question")
	}
	if c.Chooser != s.Me.ID {
		t.Fatalf("the question is the roller's (%v), want the sticker's controller's", c.Chooser)
	}
	if err := s.G.ResolveOptionPick(c.ID, s.Opp.ID, 0); err == nil {
		t.Error("the roller answered the sticker controller's question")
	}
}

// A card's roll waits for the answer, and its tail runs with the kept roll:
// Component Pouch's results table, after a strict restore while the
// question is open (the paused roll and the tail are written as data).
func TestSquidFireKnightComponentPouchWaitsAndRestores(t *testing.T) {
	s := sfkScene(t, func(s *effectstest.StkScene) uuid.UUID { return s.Me.ID })
	pouch := effectstest.P2Push(s.G, s.Me.ID, "Component Pouch", "Artifact", componentPouchOracle, 0, 0)
	if err := effectstest.P2Activate(t, s.G, s.Me, pouch, 0, ""); err != nil {
		t.Fatalf("activate the pouch: %v", err)
	}
	s.Settle()
	if effectstest.SfkPrompt(s.G, "dice/ignore-roll") == nil {
		t.Fatal("no roll-to-ignore question")
	}
	if n := effectstest.CounterOn(s.G, pouch, "component"); n != 0 {
		t.Fatalf("the pouch has %d component counters before the answer", n)
	}

	g := effectstest.RestoreRoundTrip(t, s.G, true)
	c := effectstest.SfkPrompt(g, "dice/ignore-roll")
	if c == nil {
		t.Fatal("the restored table lost the question")
	}
	rolls := effectstest.SfkRolls(t, c)
	effectstest.SfkAnswer(t, g, c, s.Me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	want := 1
	if rolls[0] >= 10 {
		want = 2
	}
	if n := effectstest.CounterOn(g, pouch, "component"); n != want {
		t.Errorf("the pouch has %d component counters, want %d (kept roll %d)", n, want, rolls[0])
	}
}

// Two players' shields on one roll: the roller orders them (CR 616.1), and
// the player whose shield applies first chooses first.
func TestSquidFireKnightTwoShieldsTheRollerOrders(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a, b, roller := g.Seats[0], g.Seats[1], g.Seats[2]
	g.WithWriteLock(func() {
		g.RollExtraDieNextTimeForEffect(uuid.Nil, a.ID, roller.ID, "A's sticker")
		g.RollExtraDieNextTimeForEffect(uuid.Nil, b.ID, roller.ID, "B's sticker")
	})
	from := effectstest.SfkEventCount(g)
	sfkRoll(t, g, roller.ID, 20, 1)
	order := effectstest.SfkPrompt(g, "dice/order-replacements")
	if order == nil || order.Chooser != roller.ID || len(order.PickOptions) != 2 {
		t.Fatalf("the order question is %+v, want the roller's, of two", order)
	}
	effectstest.SfkAnswer(t, g, order, roller.ID, 1) // B's first
	first := effectstest.SfkPrompt(g, "dice/ignore-roll")
	if first == nil || first.Chooser != b.ID || len(first.PickOptions) != 3 {
		t.Fatalf("the first ignore question is %+v, want B's, of three", first)
	}
	rolls := effectstest.SfkRolls(t, first)
	effectstest.SfkAnswer(t, g, first, b.ID, 0)
	second := effectstest.SfkPrompt(g, "dice/ignore-roll")
	if second == nil || second.Chooser != a.ID || len(second.PickOptions) != 2 {
		t.Fatalf("the second ignore question is %+v, want A's, of two", second)
	}
	effectstest.SfkAnswer(t, g, second, a.ID, 1)
	if got := sfkAnnounced(g, from); !effectstest.SameInts(got, rolls[1:2]) {
		t.Errorf("announced %v, want %v", got, rolls[1:2])
	}
}

// No card rolls around the door: every effects file goes through
// RollDice or game.RollDiceThenForEffect, so the sticker reaches each.
func TestNoCardRollsAroundTheDiceDoor(t *testing.T) {
	entries, err := effectstest.FlatReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := effectstest.FlatReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, banned := range []string{"RollDiceForEffect(", "RollDiceIgnoringLowestForEffect("} {
			if strings.Contains(string(src), banned) {
				t.Errorf("%s calls %s, which no replacement sees", name, banned)
			}
		}
	}
}
