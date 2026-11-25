package game

import (
	"testing"

	"github.com/google/uuid"
)

// random_piles_test.go — #6248, Whims of the Fates: "Starting with you,
// each player separates all permanents they control into three piles.
// Then each player chooses one of their piles at random and sacrifices
// those permanents." (CR 700.3, 700.3a, 700.3d.)

var (
	randomPilesPicked []uuid.UUID
	randomPilesRan    bool
	randomPilesTold   = RegisterTail("test/random-piles-told", func(_ *Game, o TailOutcome) error {
		randomPilesPicked, randomPilesRan = copyUUIDs(o.IDs), true
		return nil
	})
)

// The issue's acceptance: each player splits three permanents 1/1/1; a
// seeded random picks one pile per player; exactly one permanent of each
// player's is picked, and it is one of theirs.
func TestEachPlayerSeparatesIntoThreePilesAndOneIsPickedAtRandom(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var mine, theirs []uuid.UUID
	for i := 0; i < 3; i++ {
		mine = append(mine, pushCombatant(t, g, me, "Mine", 1, 1))
		theirs = append(theirs, pushCombatant(t, g, opp, "Theirs", 1, 1))
	}
	randomPilesPicked, randomPilesRan = nil, false
	g.WithWriteLock(func() {
		if err := g.SeparateIntoPilesPickOneAtRandomForEffect(RandomPiles{
			Players: []uuid.UUID{me.ID, opp.ID}, Piles: 3, Label: "Whims of the Fates",
			Then: KeyedTail{Ref: randomPilesTold},
		}); err != nil {
			t.Fatal(err)
		}
	})
	// Starting with you: my two piles, then theirs.
	for _, step := range []struct {
		who   *Player
		offer int
		pick  uuid.UUID
	}{{me, 3, mine[0]}, {me, 2, mine[1]}, {opp, 3, theirs[0]}, {opp, 2, theirs[1]}} {
		c := openPromptOfKind(g, PendingChoiceChooseCards)
		if c == nil || c.Chooser != step.who.ID || len(c.ChooseCards) != step.offer {
			t.Fatalf("prompt = %+v, want %s choosing from %d permanents", c, step.who.Name, step.offer)
		}
		answerPile(t, g, step.pick)
	}
	if !randomPilesRan {
		t.Fatal("the tail did not run after the last pile")
	}
	var mineN, theirsN int
	for _, id := range randomPilesPicked {
		for _, m := range mine {
			if m == id {
				mineN++
			}
		}
		for _, o := range theirs {
			if o == id {
				theirsN++
			}
		}
	}
	if mineN != 1 || theirsN != 1 || len(randomPilesPicked) != 2 {
		t.Fatalf("picked %v: want one of my piles (1 permanent) and one of theirs", randomPilesPicked)
	}
}

// Over many seeded tables each of the three piles is picked: the pick is
// drawn, not fixed.
func TestTheRandomPileVaries(t *testing.T) {
	seen := map[int]bool{}
	for i := 0; i < 40 && len(seen) < 3; i++ {
		g := newActiveGame(t)
		me := g.Seats[0]
		perms := []uuid.UUID{pushCombatant(t, g, me, "A", 1, 1), pushCombatant(t, g, me, "B", 1, 1), pushCombatant(t, g, me, "C", 1, 1)}
		randomPilesPicked, randomPilesRan = nil, false
		g.WithWriteLock(func() {
			_ = g.SeparateIntoPilesPickOneAtRandomForEffect(RandomPiles{
				Source: uuid.New(), Players: []uuid.UUID{me.ID}, Piles: 3, Then: KeyedTail{Ref: randomPilesTold},
			})
		})
		answerPile(t, g, perms[0])
		answerPile(t, g, perms[1])
		if len(randomPilesPicked) != 1 {
			t.Fatalf("picked %v, want one permanent", randomPilesPicked)
		}
		for k, p := range perms {
			if p == randomPilesPicked[0] {
				seen[k] = true
			}
		}
	}
	if len(seen) != 3 {
		t.Fatalf("piles picked over the runs: %v, want all three", seen)
	}
}

// Empty piles are legal (CR 700.3d): a player who puts everything in one
// pile is asked no more, and a player with no permanents is not asked.
func TestAllInOnePileAndNothingToSeparate(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	a, b := pushCombatant(t, g, me, "A", 1, 1), pushCombatant(t, g, me, "B", 1, 1)
	randomPilesPicked, randomPilesRan = nil, false
	g.WithWriteLock(func() {
		_ = g.SeparateIntoPilesPickOneAtRandomForEffect(RandomPiles{
			Players: []uuid.UUID{me.ID, g.Seats[1].ID}, Piles: 3, Then: KeyedTail{Ref: randomPilesTold},
		})
	})
	answerPile(t, g, a, b)
	if !randomPilesRan {
		t.Fatal("asked more after every permanent was in pile 1, or asked the player with nothing")
	}
	if n := len(randomPilesPicked); n != 0 && n != 2 {
		t.Fatalf("picked %v: either the full pile or an empty one", randomPilesPicked)
	}
}
