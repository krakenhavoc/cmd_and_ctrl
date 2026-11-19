package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB34ControlOracle   = "755034b1-1859-4f18-9394-c0e1914a5b83"
	gapB34RuinationOracle = "59dbc873-72d9-4e7b-9442-2c7b8e8523e1"
	gapB34LeknifOracle    = "2a4bf27f-cf53-4a10-b909-be2ca696649d"
	gapB34TruceOracle     = "ffe6371c-d137-4062-8c6e-6e9794ab25bc"
)

func TestGapB34EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB34ControlOracle:   "Control of the Court",
		gapB34RuinationOracle: "Ruination",
		gapB34LeknifOracle:    "Wrath of Leknif",
		gapB34TruceOracle:     "Temporary Truce",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func gapB34Perm(g *game.Game, owner uuid.UUID, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = owner, owner
	return pushBattlefieldCardWithTimestamp(g, c)
}

func TestGapB34ControlOfTheCourtDrawsFourAndDiscardsThreeAtRandom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 6; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	hand, yard := me.Hand.Size(), me.Graveyard.Size()
	if err := bnsCast(t, g, "Control of the Court", "Sorcery", gapB34ControlOracle, "{1}{R}", game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	// The spell leaves the hand as it is cast; the hand then grows by four and shrinks by three.
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want +1 (four drawn, three discarded)", hand, got)
	}
	if got := me.Graveyard.Size(); got != yard+3+1 {
		t.Errorf("graveyard %d → %d, want +3 discarded and the spell", yard, got)
	}
}

func TestGapB34RuinationDestroysNonbasicLandsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	basic := gapB34Perm(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	snow := gapB34Perm(g, opp.ID, game.Card{Name: "Snow Swamp", TypeLine: "Basic Snow Land — Swamp"})
	mine := gapB34Perm(g, me.ID, game.Card{Name: "Cradle", TypeLine: "Legendary Land"})
	theirs := gapB34Perm(g, opp.ID, game.Card{Name: "Dual", TypeLine: "Land — Forest Island"})
	solid := gapB34Perm(g, opp.ID, game.Card{Name: "Citadel", TypeLine: "Land", Keywords: []string{"indestructible"}})
	creature := gapB34Perm(g, opp.ID, game.Card{Name: "Dryad Arbor", TypeLine: "Land Creature — Forest Dryad", Power: 1, Toughness: 1})
	other := gapB34Perm(g, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err := bnsCast(t, g, "Ruination", "Sorcery", gapB34RuinationOracle, "{3}{R}", game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"Forest": basic, "Snow Swamp": snow, "Citadel": solid, "Bear": other} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was destroyed", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"Cradle": mine, "Dual": theirs, "Dryad Arbor": creature} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived", name)
		}
	}
}

func TestGapB34WrathOfLeknifDestroysCreaturesThroughRegenAndUntapsYourLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := gapB34Perm(g, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	var myLands []uuid.UUID
	for i := 0; i < 6; i++ {
		myLands = append(myLands, gapB34Perm(g, me.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island", Tapped: true}))
	}
	theirs := gapB34Perm(g, opp.ID, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp", Tapped: true})
	g.WithWriteLock(func() {
		if err := g.RegenerateForEffect(bear); err != nil {
			t.Fatalf("regenerate: %v", err)
		}
	})
	if err := bnsCast(t, g, "Wrath of Leknif", "Sorcery", gapB34LeknifOracle, "{1}{W}{W}{U}", game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a creature survived through its regeneration shield")
	}
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no untap prompt")
	}
	if c.ChooseMax != 4 || c.ChooseMin != 0 {
		t.Errorf("untap prompt %d..%d, want up to four", c.ChooseMin, c.ChooseMax)
	}
	for _, id := range c.ChooseCards {
		if id == theirs {
			t.Error("an opponent's land was offered")
		}
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, append([]uuid.UUID(nil), myLands[:4]...)); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	untapped := 0
	for _, bc := range gapB34Battlefield(g) {
		if bc.Controller == me.ID && bc.IsLand() && !bc.Tapped {
			untapped++
		}
	}
	if untapped != 4 {
		t.Errorf("%d of my lands untapped, want 4", untapped)
	}
}

func TestGapB34TemporaryTruceAsksEachPlayerAndPaysForTheShortfall(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, p := range g.Seats {
		for i := 0; i < 4; i++ {
			p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: p.ID, Controller: p.ID})
		}
	}
	hands := make([]int, len(g.Seats))
	lives := make([]int, len(g.Seats))
	if err := bnsCast(t, g, "Temporary Truce", "Sorcery", gapB34TruceOracle, "{1}{W}", game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	for i, p := range g.Seats {
		hands[i], lives[i] = p.Hand.Size(), p.Life
	}
	passPriorityAroundTable(t, g)
	// Answer 2, 1, 0, 2 in turn order from the active player.
	answers := []int{2, 1, 0, 2}
	for i := 0; i < len(g.Seats); i++ {
		seat := g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)]
		pa := answerNumber(t, g, seat.ID, answers[i])
		if pa.Min != 0 || pa.Max != 2 {
			t.Errorf("seat %d prompt %d..%d, want 0..2", i, pa.Min, pa.Max)
		}
		passPriorityAroundTable(t, g)
	}
	_ = me
	for i := 0; i < len(g.Seats); i++ {
		idx := (g.Turn.ActiveSeat + i) % len(g.Seats)
		p := g.Seats[idx]
		if got := p.Hand.Size() - hands[idx]; got != answers[i] {
			t.Errorf("seat %d drew %d, want %d", i, got, answers[i])
		}
		if got := p.Life - lives[idx]; got != 2*(2-answers[i]) {
			t.Errorf("seat %d gained %d life, want %d", i, got, 2*(2-answers[i]))
		}
	}
}

func gapB34Battlefield(g *game.Game) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() { out = g.BattlefieldCardsForEffect() })
	return out
}
