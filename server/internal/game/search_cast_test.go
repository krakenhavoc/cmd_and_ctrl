package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// search_cast_test.go — #6216: "While you're searching your library, you
// may cast this card from your library" (Panglacial Wurm). CR 701.23a is
// the search, CR 608.2g the cast during a resolution: the card goes on the
// stack, nobody gets priority, and the search goes on.

const wurmOracle = "test-panglacial-wurm"

// wurmLibrary seeds `me`'s library with three Forests, a Wurm and an
// Island (top last) and wires the Wurm's catalog row. It returns the Wurm
// and the first Forest.
func wurmLibrary(t *testing.T, me *Player) (wurm, forest uuid.UUID) {
	t.Helper()
	withCatalogDefs(t, map[string]*CardDef{wurmOracle: {CastWhileSearchingLibrary: true}})
	ids := searchLibraryOf(me, "Forest", "Forest", "Forest")
	forest = ids[0]
	wurm = uuid.New()
	me.Library.PushTop(Card{InstanceID: wurm, Name: "Panglacial Wurm", TypeLine: "Creature — Wurm",
		ManaCost: "{5}{G}{G}", OracleID: wurmOracle, Owner: me.ID, Controller: me.ID, Power: 9, Toughness: 5})
	me.Library.PushTop(Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	return wurm, forest
}

func forestSearch(me *Player) SearchLibrarySpec {
	return SearchLibrarySpec{
		Player:  me.ID,
		Pred:    func(c Card) bool { return c.Name == "Forest" },
		Dest:    ZoneHand,
		Limit:   1,
		Shuffle: true,
	}
}

// openResolution marks a resolution open, as a resolving fetch's effect
// runs inside one (resolution_pause.go).
func openResolution(g *Game) {
	g.mu.Lock()
	g.resolutionOpen = true
	g.mu.Unlock()
}

func addGreen(p *Player, n int) {
	for range n {
		p.ManaPool.AddMana(ManaToken{Color: "G"})
	}
}

// TestCastPanglacialWurmWhileSearching — the acceptance test: while a
// fetch resolves, the searcher casts the Wurm from the library by paying
// its cost, during the opponent's turn and with the fetch still on the
// stack (CR 608.2g waives the creature's timing), then finishes the search.
func TestCastPanglacialWurmWhileSearching(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[1] // not the active player: the cast ignores timing
	wurm, forest := wurmLibrary(t, me)
	addGreen(me, 7)
	openResolution(g)
	runSearch(t, g, forestSearch(me))
	c := openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	var castable []uuid.UUID
	g.ReadSnapshot(func() { castable = g.SearchCastableLocked(me.ID) })
	if !slices.Equal(castable, []uuid.UUID{wurm}) {
		t.Fatalf("castable while searching = %v, want the Wurm", castable)
	}

	if err := g.CastWhileSearching(me.ID, c.ID, wurm, CastSpellParams{}); err != nil {
		t.Fatalf("casting the Wurm mid-search: %v", err)
	}
	if !g.Stack.Contains(wurm) || me.Library.Contains(wurm) {
		t.Fatal("the Wurm is not on the stack")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool after = %d, want the {5}{G}{G} paid", len(me.ManaPool))
	}
	after := openSearch(g, me.ID)
	if after == nil {
		t.Fatal("the search ended with the cast; it goes on (CR 608.2g)")
	}
	if slices.Contains(after.SearchCards, wurm) {
		t.Error("the cast Wurm is still a search candidate")
	}
	if err := g.ResolveSearchLibrary(after.ID, me.ID, []uuid.UUID{forest}); err != nil {
		t.Fatalf("finishing the search: %v", err)
	}
	if !me.Hand.Contains(forest) {
		t.Error("the search did not finish with the Forest in hand")
	}
}

// TestPanglacialWurmOnlyWhileSearching — outside a search the library is
// closed to it, a payment short of its cost is refused with the Wurm left
// in the library, and only its owner's search opens it.
func TestPanglacialWurmOnlyWhileSearching(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wurm, _ := wurmLibrary(t, me)
	addGreen(me, 7)
	if err := g.CastSpell(me.ID, wurm, CastSpellParams{FromZone: "library", Strict: true, AutoTap: true}); err == nil {
		t.Fatal("the Wurm was cast from the library with no search open")
	}

	// The opponent's search does not open my library.
	openResolution(g)
	runSearch(t, g, forestSearch(opp))
	if c := openSearch(g, opp.ID); c != nil {
		if err := g.CastWhileSearching(me.ID, c.ID, wurm, CastSpellParams{}); err == nil {
			t.Fatal("the Wurm was cast during another player's search")
		}
	}

	g2 := newActiveGame(t)
	me2 := g2.Seats[0]
	wurm2, _ := wurmLibrary(t, me2)
	addGreen(me2, 6) // one short
	openResolution(g2)
	runSearch(t, g2, forestSearch(me2))
	c := openSearch(g2, me2.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if err := g2.CastWhileSearching(me2.ID, c.ID, wurm2, CastSpellParams{}); err == nil {
		t.Fatal("the Wurm was cast with six mana for {5}{G}{G}")
	}
	if !me2.Library.Contains(wurm2) {
		t.Error("a refused cast moved the Wurm")
	}
}

// TestASearchTheRulesWouldSettleStillAsksWithAWurmInTheLibrary — an
// unconditional "search for a card" the rules would take unasked (CR
// 701.23d) still opens, so the window to cast the Wurm exists.
func TestASearchTheRulesWouldSettleStillAsksWithAWurmInTheLibrary(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogDefs(t, map[string]*CardDef{wurmOracle: {CastWhileSearchingLibrary: true}})
	me.Library.Cards = nil
	me.Library.PushTop(Card{InstanceID: uuid.New(), Name: "Panglacial Wurm", TypeLine: "Creature — Wurm",
		ManaCost: "{5}{G}{G}", OracleID: wurmOracle, Owner: me.ID, Controller: me.ID})
	openResolution(g)
	runSearch(t, g, SearchLibrarySpec{Player: me.ID, Dest: ZoneHand, Limit: 1, Shuffle: true})
	if openSearch(g, me.ID) == nil {
		t.Fatal("a one-card library with the Wurm was searched unasked")
	}
}
