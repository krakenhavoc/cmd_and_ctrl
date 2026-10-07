package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// face_down_piles_test.go — #2147. Sauron's Ransom and the four other
// cards that look at the top of a library and separate it into a
// face-down pile and a face-up pile.

const (
	ransomOracle             = "fe04a81b-413c-4bc8-bb83-29978da9a43d"
	favorOracle              = "9baff93e-4ef9-404d-884b-f651a11f3742"
	atrisOracle              = "8b2b00ab-f1c5-4057-9957-d7daac95a847"
	riddlesOracle            = "78bed291-0ec7-466f-bc7a-9dae7d0b56ae"
	curatorOfDestiniesOracle = "9938d178-0ce6-45c0-b317-fd5c54231579"
)

func TestFaceDownPileCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		ransomOracle: "Sauron's Ransom", favorOracle: "Fortune's Favor",
		atrisOracle: "Atris, Oracle of Half-Truths", riddlesOracle: "Riddles in the Dark",
		curatorOfDestiniesOracle: "Curator of Destinies",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name {
			t.Errorf("%s not registered under %s", name, oracle)
			continue
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: completeness %v", name, spec.Completeness)
		}
	}
}

// ransomTable casts Sauron's Ransom over a four-card seeded library
// and answers the opponent choice, leaving the separator's prompt
// open. Returns the four cards top-first.
func ransomTable(t *testing.T) (g *game.Game, me, opp *game.Player, top []uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	top = seedFactOrFictionLibrary(me, "FD One", "FD Two", "FD Three", "FD Four")
	// The helper leaves the last name on top; top-first order is the
	// reverse of the names, which is what it returns.
	castCatalogSpell(t, g, "Sauron's Ransom", "Instant", ransomOracle, nil)
	passPriorityAroundTable(t, g)
	answerChoosePlayer(t, g, me.ID, opp)
	return g, me, opp, top
}

func knownBy(g *game.Game, id, seat uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	return ok && c.KnownBy[seat]
}

// TestSauronsRansomEachPileChoice is the whole card for each outcome:
// the separator sees four, hides some, the caster takes face-up or
// face-down, and the Ring tempts.
func TestSauronsRansomEachPileChoice(t *testing.T) {
	for _, tc := range []struct {
		name     string
		down     int // how many of the four go face down
		take     int // option index: 0 face-up, 1 face-down
		wantHand int
	}{
		{"take face-up of a 1/3 split", 1, 0, 3},
		{"take face-down of a 1/3 split", 1, 1, 1},
		{"take face-up when everything is face down", 4, 0, 0},
		{"take face-down when everything is face down", 4, 1, 4},
		{"take face-down when nothing is", 0, 1, 0},
		{"take face-up when nothing is face down", 0, 0, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp, top := ransomTable(t)
			hand, graves := me.Hand.Size(), me.Graveyard.Size()

			split := latestRevealPickFor(g, opp.ID)
			if split == nil || len(split.ChooseCards) != 4 {
				t.Fatalf("the opponent separates four cards: %+v", split)
			}
			down := top[:tc.down]
			up := top[tc.down:]
			// Before the answer only the separator has looked.
			for _, id := range top {
				if !knownBy(g, id, opp.ID) {
					t.Fatalf("the separator looked at %s", id)
				}
				for _, s := range []*game.Player{me, g.Seats[2], g.Seats[3]} {
					if knownBy(g, id, s.ID) {
						t.Fatalf("seat %s must not know %s before the split", s.Name, id)
					}
				}
			}
			if err := g.ResolveRevealPick(split.ID, opp.ID, down); err != nil {
				t.Fatalf("ResolveRevealPick: %v", err)
			}
			// The face-up pile is public, the face-down pile is not.
			for _, id := range up {
				for _, s := range g.Seats {
					if !knownBy(g, id, s.ID) {
						t.Errorf("face-up card %s is public, but %s does not know it", id, s.Name)
					}
				}
			}
			for _, id := range down {
				for _, s := range []*game.Player{me, g.Seats[2], g.Seats[3]} {
					if knownBy(g, id, s.ID) {
						t.Errorf("face-down card %s leaked to %s", id, s.Name)
					}
				}
			}

			pick := latestOptionPickFor(g, me.ID)
			if pick == nil || len(pick.PickOptions) != 2 {
				t.Fatalf("the caster picks a pile: %+v", pick)
			}
			answerOptionPick(t, g, me.ID, tc.take)
			passPriorityAroundTable(t, g)

			if got := me.Hand.Size() - hand; got != tc.wantHand {
				t.Errorf("hand +%d, want +%d", got, tc.wantHand)
			}
			// The spell is already in the graveyard when the baseline is taken.
			if got := me.Graveyard.Size() - graves; got != 4-tc.wantHand {
				t.Errorf("graveyard +%d, want +%d", got, 4-tc.wantHand)
			}
			taken := up
			if tc.take == 1 {
				taken = down
			}
			for _, id := range taken {
				if !me.Hand.Contains(id) {
					t.Errorf("%s should be in hand", id)
				}
			}
			if ringCount(g, me.ID) != 1 {
				t.Errorf("the Ring tempted %d times, want 1", ringCount(g, me.ID))
			}
		})
	}
}

// TestSauronsRansomShortAndEmptyLibraries — a library with fewer than
// four cards looks at what it has; an empty one asks nothing, and the
// Ring tempts either way.
func TestSauronsRansomShortAndEmptyLibraries(t *testing.T) {
	t.Run("two cards", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		me.Library.Cards = nil
		top := seedFactOrFictionLibrary(me, "Short One", "Short Two")
		castCatalogSpell(t, g, "Sauron's Ransom", "Instant", ransomOracle, nil)
		passPriorityAroundTable(t, g)
		answerChoosePlayer(t, g, me.ID, opp)
		split := latestRevealPickFor(g, opp.ID)
		if split == nil || len(split.ChooseCards) != 2 {
			t.Fatalf("two cards to separate: %+v", split)
		}
		if err := g.ResolveRevealPick(split.ID, opp.ID, top[:1]); err != nil {
			t.Fatal(err)
		}
		answerOptionPick(t, g, me.ID, 0)
		passPriorityAroundTable(t, g)
		if !me.Hand.Contains(top[1]) || !me.Graveyard.Contains(top[0]) {
			t.Error("face-up card in hand, face-down card in the graveyard")
		}
		if ringCount(g, me.ID) != 1 {
			t.Error("the Ring tempts")
		}
	})
	t.Run("empty", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		me.Library.Cards = nil
		hand := me.Hand.Size()
		castCatalogSpell(t, g, "Sauron's Ransom", "Instant", ransomOracle, nil)
		passPriorityAroundTable(t, g)
		answerChoosePlayer(t, g, me.ID, opp)
		passPriorityAroundTable(t, g)
		if latestRevealPickFor(g, opp.ID) != nil || latestOptionPickFor(g, me.ID) != nil {
			t.Fatalf("nothing to separate, nothing asked: %+v", g.PendingChoices)
		}
		if me.Hand.Size() != hand {
			t.Error("nothing was taken")
		}
		if ringCount(g, me.ID) != 1 {
			t.Error("the Ring still tempts")
		}
	})
}

// TestFortunesFavorTargetsTheSeparator — the same split with a target.
func TestFortunesFavorTargetsTheSeparator(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	top := seedFactOrFictionLibrary(me, "FF One", "FF Two", "FF Three", "FF Four")
	castCatalogSpell(t, g, "Fortune's Favor", "Instant", favorOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	split := latestRevealPickFor(g, opp.ID)
	if split == nil {
		t.Fatalf("the targeted opponent separates: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(split.ID, opp.ID, top[:2]); err != nil {
		t.Fatal(err)
	}
	answerOptionPick(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	for _, id := range top[:2] {
		if !me.Hand.Contains(id) {
			t.Errorf("face-down card %s should be in hand", id)
		}
	}
	for _, id := range top[2:] {
		if !me.Graveyard.Contains(id) {
			t.Errorf("face-up card %s should be in the graveyard", id)
		}
	}
}

// TestAtrisSeparatesThreeOnEntering — the trigger targets an opponent.
func TestAtrisSeparatesThreeOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	top := seedFactOrFictionLibrary(me, "At One", "At Two", "At Three")
	castCatalogSpell(t, g, "Atris, Oracle of Half-Truths", "Creature — Human Warlock", atrisOracle, nil)
	passPriorityAroundTable(t, g)
	pickPlayerTarget(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	split := latestRevealPickFor(g, opp.ID)
	if split == nil || len(split.ChooseCards) != 3 {
		t.Fatalf("three cards to separate: %+v", split)
	}
	if err := g.ResolveRevealPick(split.ID, opp.ID, top[:1]); err != nil {
		t.Fatal(err)
	}
	answerOptionPick(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)
	for _, id := range top[1:] {
		if !me.Hand.Contains(id) {
			t.Errorf("face-up card %s should be in hand", id)
		}
	}
	if !me.Graveyard.Contains(top[0]) {
		t.Error("the face-down card should be in the graveyard")
	}
}

// TestRiddlesInTheDarkLetsAnOpponentChoose — the mirror: the caster
// separates and an opponent chooses; the pile they pick is the
// caster's to keep.
func TestRiddlesInTheDarkLetsAnOpponentChoose(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	top := seedFactOrFictionLibrary(me, "Rd One", "Rd Two", "Rd Three", "Rd Four")
	castCatalogSpell(t, g, "Riddles in the Dark", "Instant", riddlesOracle, nil)
	passPriorityAroundTable(t, g)
	answerChoosePlayer(t, g, me.ID, opp)

	split := latestRevealPickFor(g, me.ID)
	if split == nil || len(split.ChooseCards) != 4 {
		t.Fatalf("the caster separates the four they looked at: %+v", split)
	}
	if err := g.ResolveRevealPick(split.ID, me.ID, top[:3]); err != nil {
		t.Fatal(err)
	}
	for _, id := range top[:3] {
		if knownBy(g, id, opp.ID) {
			t.Errorf("the chooser must not know the face-down card %s", id)
		}
	}
	if !knownBy(g, top[3], opp.ID) {
		t.Error("the chooser sees the face-up card")
	}
	answerOptionPick(t, g, opp.ID, 1)
	passPriorityAroundTable(t, g)
	for _, id := range top[:3] {
		if !me.Hand.Contains(id) {
			t.Errorf("the pile the opponent chose goes to the caster's hand: %s", id)
		}
	}
	if !me.Graveyard.Contains(top[3]) {
		t.Error("the other pile goes to the graveyard")
	}
}

// TestCuratorOfDestiniesLooksAtFive — uncounterable flyer whose
// trigger is Riddles in the Dark over five cards.
func TestCuratorOfDestiniesLooksAtFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	top := seedFactOrFictionLibrary(me, "Cu One", "Cu Two", "Cu Three", "Cu Four", "Cu Five")
	castCatalogSpell(t, g, "Curator of Destinies", "Creature — Sphinx", curatorOfDestiniesOracle, nil)
	passPriorityAroundTable(t, g)
	answerChoosePlayer(t, g, me.ID, opp)
	split := latestRevealPickFor(g, me.ID)
	if split == nil || len(split.ChooseCards) != 5 {
		t.Fatalf("five cards to separate: %+v", split)
	}
	if err := g.ResolveRevealPick(split.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	answerOptionPick(t, g, opp.ID, 0)
	passPriorityAroundTable(t, g)
	for _, id := range top {
		if !me.Hand.Contains(id) {
			t.Errorf("an all-face-up split taken face-up gives every card: %s", id)
		}
	}
}
