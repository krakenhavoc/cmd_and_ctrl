package effects

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_harness_test.go is the test table for plane cards (ADR 0216 slice
// 4, #4870): a four-seat Commander game with the Planechase option on, its
// communal planar deck built from the plane under test and filler planes
// with no catalog row, so every test drives the real engine walks (the
// layer pass, the harvester, the gates) from the command zone. Every later
// plane batch uses it.

// planeFixture is one plane card as the card database would build it.
type planeFixture struct {
	Oracle, Name, TypeLine string
}

// fakePlanes is game.PlanarCardSource over a fixed list.
type fakePlanes []game.Card

func (f fakePlanes) PlanarCards(game.PlanarKinds) []game.Card { return append([]game.Card(nil), f...) }

func planeTemplate(oracle, name, typeLine string) game.Card {
	c := game.NewCard(name, uuid.Nil)
	c.OracleID = oracle
	c.TypeLine = typeLine
	c.Layout = "planar"
	return c
}

// newPlaneGame is a started four-seat Planechase game, the hands kept and
// the cursor on the starting player's turn-1 draw step (newCatalogGame's
// spot), with `plane` face up in the command zone as the only face-up
// planar card. No "when you planeswalk to" trigger has fired: the plane was
// put there as the starting plane is (CR 901.5).
func newPlaneGame(t *testing.T, plane planeFixture) (*game.Game, game.Card) {
	t.Helper()
	cards := []game.Card{planeTemplate(plane.Oracle, plane.Name, plane.TypeLine)}
	for i := range game.MinCommunalPlanarDeck(4) {
		name := fmt.Sprintf("Filler Plane %02d", i+1)
		cards = append(cards, planeTemplate(uuid.NewSHA1(uuid.NameSpaceOID, []byte(name)).String(), name, "Plane — Testworld"))
	}
	slices.SortFunc(cards, func(a, b game.Card) int { return strings.Compare(a.OracleID, b.OracleID) })
	prev := game.PlanarCards
	game.PlanarCards = fakePlanes(cards)
	t.Cleanup(func() { game.PlanarCards = prev })

	g := game.NewGame()
	for i := range 4 {
		deck := make([]game.Card, 20)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	on := true
	if err := g.UpdateSettings(uuid.Nil, game.SettingsPatch{Planechase: &on}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 22))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	// The seeded shuffle may have put the plane under test on top, to be
	// the starting plane once the hands are kept (CR 901.5). Keep it at
	// the bottom until the cursor is on the draw step, so none of its
	// abilities triggers in the turn-1 steps before it.
	g.WithWriteLock(func() {
		c, err := g.PlanarDeck.Remove(planeID(g, plane.Oracle))
		if err != nil {
			t.Fatalf("%s is not in the planar deck at the start: %v", plane.Name, err)
		}
		g.PlanarDeck.PushBottom(c)
	})
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	for i := 0; g.Turn.Step != game.StepDraw; i++ {
		if i >= 8 {
			t.Fatalf("never reached the turn-1 draw step (stuck at %v)", g.Turn.Step)
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	var up game.Card
	g.WithWriteLock(func() {
		starting, err := g.PlanarFaceUp.PopTop()
		if err != nil {
			t.Fatalf("no starting plane: %v", err)
		}
		c, err := g.PlanarDeck.Remove(planeID(g, plane.Oracle))
		if err != nil {
			t.Fatalf("%s is not in the planar deck: %v", plane.Name, err)
		}
		g.PlanarDeck.PushBottom(starting)
		c.EnteredBattlefieldAt = time.Now().UnixNano()
		g.PlanarFaceUp.PushTop(c)
		up = c
	})
	// The swap is not a planeswalk, so nothing told the layer pass.
	g.BumpLayerVersionForTest()
	if n := itemsFromPlane(g, up.InstanceID); n != 0 {
		t.Fatalf("the plane put in place at the draw step has %d abilities waiting already", n)
	}
	if up.Controller != g.Seats[g.Turn.ActiveSeat].ID {
		t.Fatalf("the plane is controlled by %v, want the active player", up.Controller)
	}
	return g, up
}

// planeID is the instance ID of the planar card with `oracle`, face up or
// in the planar deck.
func planeID(g *game.Game, oracle string) uuid.UUID {
	for _, z := range []*game.Zone{g.PlanarFaceUp, g.PlanarDeck} {
		for _, c := range z.Cards {
			if c.OracleID == oracle {
				return c.InstanceID
			}
		}
	}
	return uuid.Nil
}

// faceUpPlane is the face-up planar card with `oracle`, or ok false.
func faceUpPlane(g *game.Game, oracle string) (game.Card, bool) {
	for _, c := range g.PlanarFaceUpCards() {
		if c.OracleID == oracle {
			return c, true
		}
	}
	return game.Card{}, false
}

// planeswalkToPlane makes the active player planeswalk (CR 901.11) with
// `oracle` on top of the planar deck, so it is the plane they planeswalk to
// and its "when you planeswalk to" ability triggers.
func planeswalkToPlane(t *testing.T, g *game.Game, oracle string) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat].ID
	g.WithWriteLock(func() {
		if up, ok := faceUpPlane(g, oracle); ok {
			// Walk away first, so it is in the deck to be walked to.
			if err := g.PlaneswalkForEffect(me); err != nil {
				t.Fatalf("PlaneswalkForEffect: %v", err)
			}
			_ = up
		}
		c, err := g.PlanarDeck.Remove(planeID(g, oracle))
		if err != nil {
			t.Fatalf("plane %s is not in the planar deck: %v", oracle, err)
		}
		g.PlanarDeck.PushTop(c)
		if err := g.PlaneswalkForEffect(me); err != nil {
			t.Fatalf("PlaneswalkForEffect: %v", err)
		}
	})
	if _, ok := faceUpPlane(g, oracle); !ok {
		t.Fatalf("plane %s is not face up after the planeswalk", oracle)
	}
}

// rollChaos finishes a planar die roll by `roller` showing the chaos face
// (CR 901.9b), as the dice door reports one: an EventRollDie with no number
// and the face.
func rollChaos(t *testing.T, g *game.Game, roller uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventRollDie, Actor: roller, Sides: 6, PlanarFace: game.PlanarFaceChaos})
	})
	// The special action returns to the active player's priority (CR
	// 901.9b), which is when the chaos ability goes on the stack.
	g.RunStateChecksForTest()
}

// rollBlank is a blank face (CR 901.9a): nothing happens.
func rollBlank(t *testing.T, g *game.Game, roller uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventRollDie, Actor: roller, Sides: 6, PlanarFace: game.PlanarFaceBlank})
	})
}

// itemsFromPlane counts the waiting and stacked items whose source is the
// planar card `id`.
func itemsFromPlane(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, it := range g.PendingTriggers {
		if it != nil && it.SourceCardID == id {
			n++
		}
	}
	for _, it := range g.StackMeta {
		if it != nil && it.SourceCardID == id {
			n++
		}
	}
	return n
}

// planeTriggerControllers lists the controllers of the items from `id`.
func planeTriggerControllers(g *game.Game, id uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, it := range g.PendingTriggers {
		if it != nil && it.SourceCardID == id {
			out = append(out, it.Controller)
		}
	}
	for _, it := range g.StackMeta {
		if it != nil && it.SourceCardID == id {
			out = append(out, it.Controller)
		}
	}
	return out
}

// requirePlaneRegisteredFull is the registration check every plane test
// file starts with: registered under its oracle ID, CompletenessFull, no
// caveats.
func requirePlaneRegisteredFull(t *testing.T, planes ...planeFixture) {
	t.Helper()
	for _, p := range planes {
		spec, ok := Lookup(p.Oracle)
		if !ok {
			t.Errorf("%s is not registered", p.Name)
			continue
		}
		if spec.Name != p.Name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", p.Name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}
