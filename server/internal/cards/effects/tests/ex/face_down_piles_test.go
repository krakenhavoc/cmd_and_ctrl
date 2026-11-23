package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestFaceDownPileCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.RansomOracle: "Sauron's Ransom", effectstest.FavorOracle: "Fortune's Favor",
		effectstest.AtrisOracle: "Atris, Oracle of Half-Truths", effectstest.RiddlesOracle: "Riddles in the Dark",
		effectstest.CuratorOfDestiniesOracle: "Curator of Destinies",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name {
			t.Errorf("%s not registered under %s", name, oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: completeness %v", name, spec.Completeness)
		}
	}
}

// ransomTable casts Sauron's Ransom over a four-card seeded library
// and answers the opponent choice, leaving the separator's prompt
// open. Returns the four cards top-first.
func ransomTable(t *testing.T) (g *game.Game, me, opp *game.Player, top []uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	top = effectstest.SeedFactOrFictionLibrary(me, "FD One", "FD Two", "FD Three", "FD Four")
	// The helper leaves the last name on top; top-first order is the
	// reverse of the names, which is what it returns.
	effectstest.CastCatalogSpell(t, g, "Sauron's Ransom", "Instant", effectstest.RansomOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	return g, me, opp, top
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

			split := effectstest.LatestRevealPickFor(g, opp.ID)
			if split == nil || len(split.ChooseCards) != 4 {
				t.Fatalf("the opponent separates four cards: %+v", split)
			}
			down := top[:tc.down]
			up := top[tc.down:]
			// Before the answer only the separator has looked.
			for _, id := range top {
				if !effectstest.KnownBy(g, id, opp.ID) {
					t.Fatalf("the separator looked at %s", id)
				}
				for _, s := range []*game.Player{me, g.Seats[2], g.Seats[3]} {
					if effectstest.KnownBy(g, id, s.ID) {
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
					if !effectstest.KnownBy(g, id, s.ID) {
						t.Errorf("face-up card %s is public, but %s does not know it", id, s.Name)
					}
				}
			}
			for _, id := range down {
				for _, s := range []*game.Player{me, g.Seats[2], g.Seats[3]} {
					if effectstest.KnownBy(g, id, s.ID) {
						t.Errorf("face-down card %s leaked to %s", id, s.Name)
					}
				}
			}

			pick := effectstest.LatestOptionPickFor(g, me.ID)
			if pick == nil || len(pick.PickOptions) != 2 {
				t.Fatalf("the caster picks a pile: %+v", pick)
			}
			effectstest.AnswerOptionPick(t, g, me.ID, tc.take)
			effectstest.PassPriorityAroundTable(t, g)

			if got := me.Hand.Size() - hand; got != tc.wantHand {
				t.Errorf("hand +%d, want +%d", got, tc.wantHand)
			}
			// #3613, CR 608.2n: the spell waits on the stack for the pile
			// choice, so it reaches the graveyard after the piles: the
			// pile not taken plus the spell itself.
			if got := me.Graveyard.Size() - graves; got != 4-tc.wantHand+1 {
				t.Errorf("graveyard +%d, want +%d", got, 4-tc.wantHand+1)
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
			if effectstest.RingCount(g, me.ID) != 1 {
				t.Errorf("the Ring tempted %d times, want 1", effectstest.RingCount(g, me.ID))
			}
		})
	}
}

// TestSauronsRansomShortAndEmptyLibraries — a library with fewer than
// four cards looks at what it has; an empty one asks nothing, and the
// Ring tempts either way.
func TestSauronsRansomShortAndEmptyLibraries(t *testing.T) {
	t.Run("two cards", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		me.Library.Cards = nil
		top := effectstest.SeedFactOrFictionLibrary(me, "Short One", "Short Two")
		effectstest.CastCatalogSpell(t, g, "Sauron's Ransom", "Instant", effectstest.RansomOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
		split := effectstest.LatestRevealPickFor(g, opp.ID)
		if split == nil || len(split.ChooseCards) != 2 {
			t.Fatalf("two cards to separate: %+v", split)
		}
		if err := g.ResolveRevealPick(split.ID, opp.ID, top[:1]); err != nil {
			t.Fatal(err)
		}
		effectstest.AnswerOptionPick(t, g, me.ID, 0)
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(top[1]) || !me.Graveyard.Contains(top[0]) {
			t.Error("face-up card in hand, face-down card in the graveyard")
		}
		if effectstest.RingCount(g, me.ID) != 1 {
			t.Error("the Ring tempts")
		}
	})
	t.Run("empty", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		me.Library.Cards = nil
		hand := me.Hand.Size()
		effectstest.CastCatalogSpell(t, g, "Sauron's Ransom", "Instant", effectstest.RansomOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestRevealPickFor(g, opp.ID) != nil || effectstest.LatestOptionPickFor(g, me.ID) != nil {
			t.Fatalf("nothing to separate, nothing asked: %+v", g.PendingChoices)
		}
		if me.Hand.Size() != hand {
			t.Error("nothing was taken")
		}
		if effectstest.RingCount(g, me.ID) != 1 {
			t.Error("the Ring still tempts")
		}
	})
}

// TestFortunesFavorTargetsTheSeparator — the same split with a target.
func TestFortunesFavorTargetsTheSeparator(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	top := effectstest.SeedFactOrFictionLibrary(me, "FF One", "FF Two", "FF Three", "FF Four")
	effectstest.CastCatalogSpell(t, g, "Fortune's Favor", "Instant", effectstest.FavorOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	split := effectstest.LatestRevealPickFor(g, opp.ID)
	if split == nil {
		t.Fatalf("the targeted opponent separates: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(split.ID, opp.ID, top[:2]); err != nil {
		t.Fatal(err)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	top := effectstest.SeedFactOrFictionLibrary(me, "At One", "At Two", "At Three")
	effectstest.CastCatalogSpell(t, g, "Atris, Oracle of Half-Truths", "Creature — Human Warlock", effectstest.AtrisOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayerTarget(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	split := effectstest.LatestRevealPickFor(g, opp.ID)
	if split == nil || len(split.ChooseCards) != 3 {
		t.Fatalf("three cards to separate: %+v", split)
	}
	if err := g.ResolveRevealPick(split.ID, opp.ID, top[:1]); err != nil {
		t.Fatal(err)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	top := effectstest.SeedFactOrFictionLibrary(me, "Rd One", "Rd Two", "Rd Three", "Rd Four")
	effectstest.CastCatalogSpell(t, g, "Riddles in the Dark", "Instant", effectstest.RiddlesOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)

	split := effectstest.LatestRevealPickFor(g, me.ID)
	if split == nil || len(split.ChooseCards) != 4 {
		t.Fatalf("the caster separates the four they looked at: %+v", split)
	}
	if err := g.ResolveRevealPick(split.ID, me.ID, top[:3]); err != nil {
		t.Fatal(err)
	}
	for _, id := range top[:3] {
		if effectstest.KnownBy(g, id, opp.ID) {
			t.Errorf("the chooser must not know the face-down card %s", id)
		}
	}
	if !effectstest.KnownBy(g, top[3], opp.ID) {
		t.Error("the chooser sees the face-up card")
	}
	effectstest.AnswerOptionPick(t, g, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	top := effectstest.SeedFactOrFictionLibrary(me, "Cu One", "Cu Two", "Cu Three", "Cu Four", "Cu Five")
	effectstest.CastCatalogSpell(t, g, "Curator of Destinies", "Creature — Sphinx", effectstest.CuratorOfDestiniesOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	split := effectstest.LatestRevealPickFor(g, me.ID)
	if split == nil || len(split.ChooseCards) != 5 {
		t.Fatalf("five cards to separate: %+v", split)
	}
	if err := g.ResolveRevealPick(split.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	effectstest.AnswerOptionPick(t, g, opp.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range top {
		if !me.Hand.Contains(id) {
			t.Errorf("an all-face-up split taken face-up gives every card: %s", id)
		}
	}
}
