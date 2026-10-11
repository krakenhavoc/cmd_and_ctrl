package game

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"
)

// scripted_hand_test.go — the practice table's scripted opening hand
// (scripted_hand.go, ADR 0076 §2.2 as amended 2026-10-11, #2992).

// scriptedDeck is a deck with a few named cards in it, so a script has
// something to find: 3 Forests, two Elves and fillers.
func scriptedDeck() []Card {
	deck := buildTestDeck("Marwyn")
	names := []string{"Forest", "Forest", "Forest", "Forest", "Llanowar Elves", "Elvish Mystic"}
	for i, n := range names {
		deck[1+i].Name = n
	}
	return deck
}

func scriptedTable(t *testing.T, seed uint64, script []string) *Game {
	t.Helper()
	g := NewGame()
	p0, err := g.AddPlayer("P1", scriptedDeck())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := g.AddPlayer("P2", scriptedDeck()); err != nil {
		t.Fatal(err)
	}
	if err := g.ScriptOpeningHand(p0.ID, script); err != nil {
		t.Fatalf("ScriptOpeningHand in the lobby: %v", err)
	}
	if err := g.StartWithOpeningRoll(rand.New(rand.NewPCG(seed, 2992))); err != nil {
		t.Fatal(err)
	}
	rollUntilChooser(t, g, reversed)
	if err := g.ChooseStartingPlayer(g.Seats[g.OpeningRoll.Chooser].ID, 0); err != nil {
		t.Fatal(err)
	}
	return g
}

func TestScriptedOpeningHandIsDealtAndTheRestIsShuffled(t *testing.T) {
	script := []string{"Forest", "Forest", "Forest", "Llanowar Elves", "Elvish Mystic"}
	libraries := map[string]bool{}
	for seed := uint64(1); seed <= 8; seed++ {
		g := scriptedTable(t, seed, script)
		me, them := g.Seats[0], g.Seats[1]
		if len(me.Hand.Cards) != OpeningHandSize {
			t.Fatalf("seed %d: scripted hand has %d cards, want %d", seed, len(me.Hand.Cards), OpeningHandSize)
		}
		got := map[string]int{}
		for _, c := range me.Hand.Cards {
			got[c.Name]++
			if !c.IsKnownTo(me.ID) {
				t.Errorf("seed %d: %s in the scripted hand is not known to its owner", seed, c.Name)
			}
		}
		if got["Forest"] != 3 || got["Llanowar Elves"] != 1 || got["Elvish Mystic"] != 1 {
			t.Fatalf("seed %d: hand = %v, want the script's five plus two more", seed, handNames(me))
		}
		if total := len(me.Hand.Cards) + len(me.Library.Cards); total != 99 {
			t.Fatalf("seed %d: hand + library = %d cards, want 99", seed, total)
		}
		if len(them.Hand.Cards) != OpeningHandSize {
			t.Fatalf("seed %d: the unscripted seat was dealt %d cards", seed, len(them.Hand.Cards))
		}
		libraries[fmt.Sprint(libraryOrder(me.Library.Cards))] = true
	}
	// The library stays shuffled: eight seeds, eight different orders.
	if len(libraries) < 8 {
		t.Errorf("the scripted seat's library came out in %d orders over 8 seeds; it should stay shuffled", len(libraries))
	}
}

func libraryOrder(cs []Card) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.Name + c.InstanceID.String()
	}
	return out
}

// A name the library cannot supply is skipped, and the hand still
// fills to seven.
func TestScriptedOpeningHandSkipsWhatTheLibraryLacks(t *testing.T) {
	g := scriptedTable(t, 3, []string{"Forest", "Forest", "Forest", "Forest", "Forest", "Birds of Paradise"})
	me := g.Seats[0]
	if len(me.Hand.Cards) != OpeningHandSize {
		t.Fatalf("hand has %d cards, want %d", len(me.Hand.Cards), OpeningHandSize)
	}
	forests := 0
	for _, c := range me.Hand.Cards {
		if c.Name == "Forest" {
			forests++
		}
	}
	if forests != 4 {
		t.Errorf("hand has %d Forests; the deck holds four, all of which the script asked for", forests)
	}
}

// A mulligan deals a normal random seven: the script is for the deal.
func TestMulliganAfterAScriptedHandIsRandom(t *testing.T) {
	script := []string{"Forest", "Forest", "Forest", "Llanowar Elves", "Elvish Mystic"}
	scriptedEvery := true
	for seed := uint64(1); seed <= 8; seed++ {
		g := scriptedTable(t, seed, script)
		me := g.Seats[0]
		if err := g.Mulligan(me.ID, OpeningHandSize); err != nil {
			t.Fatalf("Mulligan: %v", err)
		}
		n := 0
		for _, c := range me.Hand.Cards {
			if c.Name == "Llanowar Elves" || c.Name == "Elvish Mystic" {
				n++
			}
		}
		if n < 2 {
			scriptedEvery = false
		}
	}
	if scriptedEvery {
		t.Error("every mulligan redealt both scripted Elves; a mulligan should deal a random seven")
	}
}

func TestScriptOpeningHandRefusals(t *testing.T) {
	g := NewGame()
	p, _ := g.AddPlayer("P1", scriptedDeck())
	if _, err := g.AddPlayer("P2", scriptedDeck()); err != nil {
		t.Fatal(err)
	}
	if err := g.ScriptOpeningHand(p.ID, make([]string, OpeningHandSize+1)); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("eight names: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ScriptOpeningHand(g.ID, []string{"Forest"}); !errors.Is(err, ErrPlayerNotFound) {
		t.Errorf("unknown player: err = %v, want ErrPlayerNotFound", err)
	}
	if err := g.StartWithOpeningRoll(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatal(err)
	}
	// Still before the deal: allowed.
	if err := g.ScriptOpeningHand(p.ID, []string{"Forest"}); err != nil {
		t.Errorf("during the opening roll: %v", err)
	}
	rollUntilChooser(t, g, reversed)
	if err := g.ChooseStartingPlayer(g.Seats[g.OpeningRoll.Chooser].ID, 0); err != nil {
		t.Fatal(err)
	}
	if err := g.ScriptOpeningHand(p.ID, []string{"Forest"}); !errors.Is(err, ErrGameAlreadyStarted) {
		t.Errorf("after the deal: err = %v, want ErrGameAlreadyStarted", err)
	}
}
