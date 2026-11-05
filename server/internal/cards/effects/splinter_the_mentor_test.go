package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// splinter_the_mentor_test.go — Splinter, the Mentor (TMNT Commander,
// #5011), played through the real catalog entry.

const tmcB1SplinterOracle = "fdcb016e-36e7-46b3-9dff-20b9b9cc421d"

func TestTmcB1SplinterIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(tmcB1SplinterOracle)
	if !ok {
		t.Fatal("Splinter, the Mentor is not registered")
	}
	if spec.Name != "Splinter, the Mentor" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

// tmcB1PushSplinter puts Splinter on the battlefield under `owner`
// without an enters event.
func tmcB1PushSplinter(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b27Push(g, owner, "Splinter, the Mentor", "Legendary Creature — Mutant Ninja Rat", tmcB1SplinterOracle, "{1}{B}", 2, 2, "B")
}

// tmcB1Resolve runs state-based actions and resolves what triggered.
func tmcB1Resolve(t *testing.T, g *game.Game) {
	t.Helper()
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
}

// Cast strictly for {1}{B}, he is a 2/2 with menace.
func TestTmcB1SplinterCastsStrictlyWithMenace(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	id, err := msc17Cast(t, g, me, game.Card{
		Name: "Splinter, the Mentor", TypeLine: "Legendary Creature — Mutant Ninja Rat",
		OracleID: tmcB1SplinterOracle, ManaCost: "{1}{B}", Power: 2, Toughness: 2, Colors: []string{"B"},
	}, game.CastSpellParams{Strict: true})
	if err != nil {
		t.Fatalf("strict cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, id, "menace")
	if got := tmtMuMutagens(g, me.ID); got != 0 {
		t.Errorf("%d Mutagens from entering, want 0", got)
	}
}

// With only {B} in the pool, a strict cast is refused.
func TestTmcB1SplinterStrictCastRefusedWithoutTheMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	if _, err := msc17Cast(t, g, me, game.Card{
		Name: "Splinter, the Mentor", TypeLine: "Legendary Creature — Mutant Ninja Rat",
		OracleID: tmcB1SplinterOracle, ManaCost: "{1}{B}", Power: 2, Toughness: 2, Colors: []string{"B"},
	}, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a strict cast with {B} for {1}{B} was accepted")
	}
}

func TestTmcB1SplinterMakesAMutagenWhenHeDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	splinter := tmcB1PushSplinter(g, me.ID)
	b27Kill(g, splinter)
	tmcB1Resolve(t, g)
	if got := tmtMuMutagens(g, me.ID); got != 1 {
		t.Errorf("%d Mutagens after Splinter died, want 1", got)
	}
}

// Any exit counts: a bounce and an exile of another nontoken creature,
// and Splinter himself being bounced.
func TestTmcB1SplinterMakesAMutagenOnAnyExit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	splinter := tmcB1PushSplinter(g, me.ID)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(bear) })
	tmcB1Resolve(t, g)
	if got := tmtMuMutagens(g, me.ID); got != 1 {
		t.Fatalf("%d Mutagens after a Bear was bounced, want 1", got)
	}
	elk := b12Creature(g, me.ID, "Elk", "Creature — Elk", 3, 3)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(elk) })
	tmcB1Resolve(t, g)
	if got := tmtMuMutagens(g, me.ID); got != 2 {
		t.Fatalf("%d Mutagens after an Elk was exiled, want 2", got)
	}
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(splinter) })
	tmcB1Resolve(t, g)
	if got := tmtMuMutagens(g, me.ID); got != 3 {
		t.Errorf("%d Mutagens after Splinter was bounced, want 3", got)
	}
}

// Splinter dying together with two of my creatures sees all three
// (leaves-the-battlefield abilities look back, CR 603.10a).
func TestTmcB1SplinterSeesCreaturesLeavingWithHim(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	splinter := tmcB1PushSplinter(g, me.ID)
	a := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Elk", "Creature — Elk", 3, 3)
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{splinter, a, b}) })
	g.RunStateChecksForTest()
	b27Settle(t, g, me.ID, uuid.Nil)
	if got := tmtMuMutagens(g, me.ID); got != 3 {
		t.Errorf("%d Mutagens after a wrath took Splinter and two creatures, want 3", got)
	}
}

// Not a token creature of mine, not an opponent's creature, not a
// noncreature permanent of mine.
func TestTmcB1SplinterIgnoresTokensOpponentsAndNoncreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tmcB1PushSplinter(g, me.ID)
	token := b12Creature(g, me.ID, "Ninja", "Token Creature — Ninja", 1, 1)
	b27Kill(g, token)
	tmcB1Resolve(t, g)
	theirs := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	b27Kill(g, theirs)
	tmcB1Resolve(t, g)
	rock := b12Permanent(g, me.ID, "Mind Stone", "Artifact")
	b27Kill(g, rock)
	tmcB1Resolve(t, g)
	if got := tmtMuMutagens(g, me.ID); got != 0 {
		t.Errorf("%d Mutagens, want 0", got)
	}
	if got := tmtMuMutagens(g, opp.ID); got != 0 {
		t.Errorf("the opponent got %d Mutagens, want 0", got)
	}
}
