package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// planechase_test.go — ADR 0216 slice 1: the Planechase table option,
// the communal planar deck (CR 901.15), the starting plane (CR 901.5)
// and the planar controller (CR 901.6, 901.15b). The planar cards come
// from a small in-memory source, never the Scryfall dump.

// fakePlanarSource is PlanarCardSource over a fixed list, filtered by
// kind the way deck.Collection.PlanarCards filters the index.
type fakePlanarSource struct{ cards []Card }

func (f fakePlanarSource) PlanarCards(kinds PlanarKinds) []Card {
	var out []Card
	for _, c := range f.cards {
		kind := PlanarPlanes
		if isPhenomenonCard(&c) {
			kind = PlanarPhenomena
		}
		if kind&kinds != 0 {
			out = append(out, c)
		}
	}
	return out
}

// testPlane is a plane template, as deck.toGameCard would build it.
func testPlane(name string) Card {
	c := NewCard(name, uuid.Nil)
	c.OracleID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(name)).String()
	c.TypeLine = "Plane — Testworld"
	c.Layout = "planar"
	return c
}

// testPlanes is n plane templates in oracle-ID order, the order the
// real source returns them in.
func testPlanes(n int) []Card {
	out := make([]Card, n)
	for i := range out {
		out[i] = testPlane(fmt.Sprintf("Plane %02d", i+1))
	}
	slices.SortFunc(out, func(a, b Card) int { return strings.Compare(a.OracleID, b.OracleID) })
	return out
}

// installPlanes installs a source of these planar templates for the
// test, restoring the previous one after it.
func installPlanes(t *testing.T, cards []Card) {
	t.Helper()
	prev := PlanarCards
	PlanarCards = fakePlanarSource{cards: cards}
	t.Cleanup(func() { PlanarCards = prev })
}

// newPlanechaseLobby seats `seats` players at a table with the option
// on, not yet started.
func newPlanechaseLobby(t *testing.T, seats int) *Game {
	t.Helper()
	g := NewGame()
	for i := range seats {
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), buildTestDeck(fmt.Sprintf("Commander %d", i+1))); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	on := true
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{Planechase: &on}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	return g
}

// keepAll keeps every seat's opening hand.
func keepAllPlanechase(t *testing.T, g *Game) {
	t.Helper()
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
}

// newPlanechaseGame is a started Planechase game with the hands kept,
// on a seeded key.
func newPlanechaseGame(t *testing.T, seats int, seed uint64) *Game {
	t.Helper()
	g := newPlanechaseLobby(t, seats)
	if err := g.Start(rand.New(rand.NewPCG(seed, 3))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	keepAllPlanechase(t, g)
	return g
}

func planarOrder(g *Game) []string {
	var out []string
	for _, c := range g.PlanarDeckCards() {
		out = append(out, c.OracleID)
	}
	for _, c := range g.PlanarFaceUpCards() {
		out = append(out, "up:"+c.OracleID)
	}
	return out
}

// A table without the option is unchanged: no planar state, nothing in
// the snapshot, even with a planar source installed.
func TestATableWithoutPlanechaseHasNoPlanarState(t *testing.T) {
	installPlanes(t, testPlanes(40))
	g := newRestorableGame(t)
	if g.Planechase != nil || g.PlanarDeck != nil || g.PlanarFaceUp != nil {
		t.Fatalf("planar state on a plain table: %+v %v %v", g.Planechase, g.PlanarDeck, g.PlanarFaceUp)
	}
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"planechase"`) {
		t.Error("a table without Planechase writes a planechase key into its snapshot")
	}
}

// CR 901.15: one card per planar template, all of them in the command
// zone, outside the card index, and nothing else.
func TestThePlanarDeckHoldsEveryPlanarCardOnce(t *testing.T) {
	planes := testPlanes(25)
	installPlanes(t, planes)
	g := newPlanechaseGame(t, 2, 1)

	if g.Planechase == nil || g.Planechase.Mode != PlanarDeckCommunal {
		t.Fatalf("Planechase = %+v, want the communal mode", g.Planechase)
	}
	all := append(slices.Clone(g.PlanarDeckCards()), g.PlanarFaceUpCards()...)
	if len(all) != len(planes) {
		t.Fatalf("%d planar cards, want %d", len(all), len(planes))
	}
	seen := map[string]bool{}
	ids := map[uuid.UUID]bool{}
	for _, c := range all {
		if seen[c.OracleID] {
			t.Errorf("%s is in the planar deck twice", c.Name)
		}
		seen[c.OracleID] = true
		ids[c.InstanceID] = true
		if c.InstanceID == uuid.Nil {
			t.Errorf("%s was not minted", c.Name)
		}
		if g.findCardByIDLocked(c.InstanceID) != nil {
			t.Errorf("%s is reachable through the card index", c.Name)
		}
	}
	for _, p := range planes {
		if !seen[p.OracleID] {
			t.Errorf("%s is missing from the planar deck", p.Name)
		}
	}
	if len(ids) != len(all) {
		t.Errorf("%d distinct instance IDs for %d cards", len(ids), len(all))
	}
	if g.PlanarDeck.Kind != ZoneCommand || g.PlanarFaceUp.Kind != ZoneCommand {
		t.Errorf("containers are %s and %s, want the command zone (CR 901.4)", g.PlanarDeck.Kind, g.PlanarFaceUp.Kind)
	}
	for _, p := range g.Seats {
		for _, c := range append(slices.Clone(p.Library.Cards), p.Hand.Cards...) {
			if strings.HasPrefix(c.TypeLine, "Plane") {
				t.Errorf("a plane reached %s's library or hand", p.Name)
			}
		}
	}
}

// A seeded game deals the same planar deck on every run, and the
// planar shuffle is its own stream: the libraries are shuffled as they
// would be without the option.
func TestASeededPlanarDeckIsTheSameOnEveryRun(t *testing.T) {
	installPlanes(t, testPlanes(40))
	a := newPlanechaseGame(t, 4, 42)
	b := newPlanechaseGame(t, 4, 42)
	if !slices.Equal(planarOrder(a), planarOrder(b)) {
		t.Fatalf("two runs of one seed built different planar decks:\n%v\n%v", planarOrder(a), planarOrder(b))
	}
	other := newPlanechaseGame(t, 4, 43)
	if slices.Equal(planarOrder(a), planarOrder(other)) {
		t.Error("two seeds built the same planar deck order; the deck is not shuffled")
	}

	plain := NewGame()
	for i := range 4 {
		if _, err := plain.AddPlayer(fmt.Sprintf("P%d", i+1), buildTestDeck(fmt.Sprintf("Commander %d", i+1))); err != nil {
			t.Fatal(err)
		}
	}
	if err := plain.Start(rand.New(rand.NewPCG(42, 3))); err != nil {
		t.Fatal(err)
	}
	keepAllPlanechase(t, plain)
	for i := range a.Seats {
		if !slices.Equal(attrNames(a.Seats[i].Library.Cards), attrNames(plain.Seats[i].Library.Cards)) {
			t.Errorf("seat %d's library moved when the table turned Planechase on", i)
		}
	}
}

// CR 901.15a: at least min(40, 10 × players) cards, at most 2 ×
// players phenomena, every name different. A deck that breaks it, or no
// source at all, refuses Start and changes nothing.
func TestPlanechaseRefusesToStartWithAnIllegalPlanarDeck(t *testing.T) {
	for players, want := range map[int]int{2: 20, 3: 30, 4: 40, 5: 40, 6: 40} {
		if got := MinCommunalPlanarDeck(players); got != want {
			t.Errorf("MinCommunalPlanarDeck(%d) = %d, want %d", players, got, want)
		}
	}
	phenomenon := testPlane("Odd Event")
	phenomenon.TypeLine = "Phenomenon"
	dup := testPlane("Plane 01")
	dup.OracleID = uuid.NewString()

	cases := []struct {
		name    string
		source  PlanarCardSource
		seats   int
		wantErr error
	}{
		{"no card database", nil, 2, ErrNoPlanarCards},
		{"19 planes for two players", fakePlanarSource{testPlanes(19)}, 2, ErrPlanarDeckTooSmall},
		{"39 planes for four players", fakePlanarSource{testPlanes(39)}, 4, ErrPlanarDeckTooSmall},
		{"a repeated name", fakePlanarSource{append(testPlanes(20), dup)}, 2, ErrPlanarDeckIllegal},
		{"20 planes for two players", fakePlanarSource{testPlanes(20)}, 2, nil},
		{"40 planes for four players", fakePlanarSource{testPlanes(40)}, 4, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prev := PlanarCards
			PlanarCards = tc.source
			t.Cleanup(func() { PlanarCards = prev })
			g := newPlanechaseLobby(t, tc.seats)
			err := g.Start(rand.New(rand.NewPCG(1, 1)))
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Start = %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr == nil {
				return
			}
			if g.State != StateLobby || g.Planechase != nil || g.PlanarDeck != nil {
				t.Errorf("a refused Start changed the table: state %s, planechase %+v", g.State, g.Planechase)
			}
		})
	}

	// The phenomenon cap, on the check itself: slice 1's source is
	// asked for planes only, so no table can reach it yet.
	deck := testPlanes(20)
	for i := range 5 {
		p := testPlane(fmt.Sprintf("Event %d", i))
		p.TypeLine = "Phenomenon"
		deck = append(deck, p)
	}
	if err := validateCommunalPlanarDeck(deck, 2); !errors.Is(err, ErrPlanarDeckIllegal) {
		t.Errorf("5 phenomena for two players: %v, want ErrPlanarDeckIllegal", err)
	}
	if err := validateCommunalPlanarDeck(append(testPlanes(20), phenomenon), 2); err != nil {
		t.Errorf("1 phenomenon for two players: %v", err)
	}
}

// ADR 0216 §1: the option is set in the lobby and fixed once the game
// starts.
func TestThePlanechaseSettingLocksAtStart(t *testing.T) {
	installPlanes(t, testPlanes(20))
	g := newPlanechaseLobby(t, 2)
	if !g.Settings.Planechase {
		t.Fatal("the lobby patch did not turn Planechase on")
	}
	var changed bool
	for _, ev := range g.Events {
		if ev.Kind == EventSettingsChanged && ev.Label == SettingPlanechase && ev.SettingNew == "true" {
			changed = true
		}
	}
	if !changed {
		t.Error("turning Planechase on emitted no settings_changed event")
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 1))); err != nil {
		t.Fatal(err)
	}
	off, on := false, true
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{Planechase: &off}); !errors.Is(err, ErrPlanechaseLocked) {
		t.Errorf("turning it off after Start: %v, want ErrPlanechaseLocked", err)
	}
	if !g.Settings.Planechase || g.Planechase == nil {
		t.Error("the refused patch changed the table")
	}
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{Planechase: &on}); err != nil {
		t.Errorf("the same value after Start is not a change: %v", err)
	}

	plain := newRestorableGame(t)
	if err := plain.UpdateSettings(uuid.Nil, SettingsPatch{Planechase: &on}); !errors.Is(err, ErrPlanechaseLocked) {
		t.Errorf("turning it on mid-game: %v, want ErrPlanechaseLocked", err)
	}
}

// CR 901.6 and 901.15b: the planar controller is the active player,
// and owns and controls every planar card; both follow the turn.
func TestThePlanarControllerFollowsTheActivePlayer(t *testing.T) {
	installPlanes(t, testPlanes(30))
	g := newPlanechaseGame(t, 3, 5)
	check := func(when string) {
		t.Helper()
		active := g.Seats[g.Turn.ActiveSeat].ID
		if g.Planechase.Controller != active {
			t.Fatalf("%s: planar controller %v, want the active player %v", when, g.Planechase.Controller, active)
		}
		for _, z := range []*Zone{g.PlanarDeck, g.PlanarFaceUp} {
			if z.Owner != active {
				t.Errorf("%s: a planar container is owned by %v", when, z.Owner)
			}
			for _, c := range z.Cards {
				if c.Owner != active || c.Controller != active {
					t.Errorf("%s: %s is owned by %v and controlled by %v, want %v", when, c.Name, c.Owner, c.Controller, active)
				}
			}
		}
	}
	check("turn 1")
	seen := map[uuid.UUID]bool{g.Planechase.Controller: true}
	for i := range 3 {
		advanceOneTurn(t, g)
		check(fmt.Sprintf("turn %d", i+2))
		seen[g.Planechase.Controller] = true
	}
	if len(seen) != 3 {
		t.Errorf("the designation visited %d players in three turns, want 3", len(seen))
	}
}

// CR 901.5: the starting plane is turned face up once every hand is
// kept, by the starting player, and it is public. It is not a
// planeswalk, so no event names it.
func TestTheStartingPlaneIsTurnedUpWhenTheHandsAreKept(t *testing.T) {
	installPlanes(t, testPlanes(20))
	g := newPlanechaseLobby(t, 2)
	if err := g.Start(rand.New(rand.NewPCG(9, 9))); err != nil {
		t.Fatal(err)
	}
	if n := len(g.PlanarFaceUpCards()); n != 0 {
		t.Fatalf("%d face-up cards while the mulligans are open, want 0", n)
	}
	if g.Planechase.Controller != g.Seats[g.StartingSeat].ID {
		t.Errorf("the planar controller before turn 1 is %v, want the starting player", g.Planechase.Controller)
	}
	top := g.PlanarDeckCards()[len(g.PlanarDeckCards())-1]
	from := len(g.Events)
	keepAllPlanechase(t, g)

	up := g.PlanarFaceUpCards()
	if len(up) != 1 || up[0].InstanceID != top.InstanceID {
		t.Fatalf("face up %v, want the deck's top card %s", attrNames(up), top.Name)
	}
	if len(g.PlanarDeckCards()) != 19 {
		t.Errorf("the deck holds %d, want 19", len(g.PlanarDeckCards()))
	}
	for _, p := range g.Seats {
		if !up[0].IsKnownTo(p.ID) {
			t.Errorf("the starting plane is hidden from %s", p.Name)
		}
	}
	for _, ev := range g.Events[from:] {
		if ev.CardID == top.InstanceID || ev.Source == top.InstanceID || ev.Target == top.InstanceID {
			t.Errorf("an event names the starting plane: %s", ev.Kind)
		}
	}
	// A second settle (a restore, a restart's window) turns nothing
	// more up.
	g.turnUpStartingPlaneLocked()
	if len(g.PlanarFaceUpCards()) != 1 {
		t.Error("a second setup turned up another plane")
	}
}

// CR 901.5's phenomenon rule: a phenomenon on top goes to the bottom
// and the next card is turned up, until a plane is.
func TestTheStartingPlaneSkipsAPhenomenon(t *testing.T) {
	g := NewGame()
	g.Planechase = &PlanechaseState{Mode: PlanarDeckCommunal}
	g.PlanarDeck = newZone(ZoneCommand, uuid.Nil)
	plane := testPlane("Bottom Plane")
	event := testPlane("Top Event")
	event.TypeLine = "Phenomenon"
	g.PlanarDeck.PushTop(plane)
	g.PlanarDeck.PushTop(event)

	g.turnUpStartingPlaneLocked()
	if up := g.PlanarFaceUpCards(); len(up) != 1 || up[0].Name != "Bottom Plane" {
		t.Fatalf("face up %v, want Bottom Plane", attrNames(up))
	}
	if deck := g.PlanarDeckCards(); len(deck) != 1 || deck[0].Name != "Top Event" {
		t.Fatalf("deck %v, want the phenomenon alone on the bottom", attrNames(deck))
	}
}

// ADR 0216 §3: the planar state survives capture → JSON → restore, and
// Clone deep-copies it.
func TestPlanarStateSurvivesASnapshotAndAClone(t *testing.T) {
	installPlanes(t, testPlanes(20))
	g := newPlanechaseGame(t, 2, 11)
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	r, err := snap.Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if r.Planechase == nil || *r.Planechase != *g.Planechase {
		t.Fatalf("restored state %+v, want %+v", r.Planechase, g.Planechase)
	}
	if !slices.Equal(planarOrder(r), planarOrder(g)) {
		t.Fatalf("restored order %v, want %v", planarOrder(r), planarOrder(g))
	}
	if r.PlanarDeck.Kind != ZoneCommand || r.PlanarFaceUp.Kind != ZoneCommand {
		t.Errorf("restored containers are %s and %s", r.PlanarDeck.Kind, r.PlanarFaceUp.Kind)
	}
	if up := r.PlanarFaceUpCards()[0]; up.Owner != g.Planechase.Controller || !up.IsKnownTo(g.Seats[1].ID) {
		t.Errorf("the restored plane lost its owner or its knowers: %+v", up.Owner)
	}

	c := g.Clone()
	before := planarOrder(g)
	if _, err := c.PlanarDeck.PopTop(); err != nil {
		t.Fatal(err)
	}
	c.PlanarFaceUp.Cards[0].Name = "Changed"
	c.Planechase.Controller = uuid.New()
	if !slices.Equal(planarOrder(g), before) || g.PlanarFaceUpCards()[0].Name == "Changed" ||
		g.Planechase.Controller == c.Planechase.Controller {
		t.Error("changing the clone changed the original")
	}
}

// An existing v9 restore point, written before Planechase, restores
// with no planar state and captures again without the key.
func TestAnExistingV9FixtureRestoresWithoutPlanechase(t *testing.T) {
	dir := filepath.Join("testdata", "snapshots", fmt.Sprintf("v%d", SnapshotSchemaVersion))
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no v%d fixtures in %s: %v", SnapshotSchemaVersion, dir, err)
	}
	restored := 0
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(raw), `"planechase"`) {
			continue
		}
		// A corpus file is a restore point: {seq, snapshot}.
		var file struct {
			Snapshot GameSnapshot `json:"snapshot"`
		}
		if err := json.Unmarshal(raw, &file); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		g, err := file.Snapshot.Restore()
		if err != nil {
			// Some boards need the real catalog to restore; the corpus
			// test in internal/cards/effects covers those.
			continue
		}
		restored++
		if g.Planechase != nil || g.PlanarDeck != nil || g.PlanarFaceUp != nil || g.Settings.Planechase {
			t.Errorf("%s restored with planar state", filepath.Base(f))
		}
		again, err := json.Marshal(g.CaptureSnapshot())
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(again), `"planechase"`) {
			t.Errorf("%s captures again with a planechase key", filepath.Base(f))
		}
	}
	if restored == 0 {
		t.Fatal("no fixture restored without the catalog")
	}
}
