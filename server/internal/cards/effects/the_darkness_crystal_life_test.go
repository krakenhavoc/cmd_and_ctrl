package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_darkness_crystal_life_test.go — #2950, ADR 0178: "instead exile it
// and you gain 2 life" is one modified event in printed order.

// crystalLifeBoard is the Crystal and Rhox Faithmender under seat 0,
// plus a second, different gain replacement, so the Crystal's gain asks
// a CR 616 question. Seat 1 has a bear to lose.
func crystalLifeBoard(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// The Faithmender goes first: a catalog replacement's ID is its
	// battlefield index, so a permanent BEFORE it leaving while a prompt
	// is open renumbers it (#3482).
	b12Push(g, me.ID, "Rhox Faithmender", "Creature — Rhino Monk", b14RhoxFaithmenderOracle, 1, 5)
	crystal := kbCrystal(g, me.ID)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(game.ReplacementEffect{
			Watches: []game.EventKind{game.EventChangeLife},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
				return ev.Kind == game.RepEventLife && ev.LifeDelta > 0 && ev.LifePlayer == me.ID
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.LifeDelta++
				return nil
			},
			Label: "you gain one more life",
		})
	})
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	return g, me, crystal, bear
}

// The issue's reproduction: the destroying effect's next instruction
// (draw a card) waits for the answer to the life question instead of
// running while it is open. Before the fix the draw happened first and
// the gain was started before the creature had moved.
func TestDarknessCrystalDestroyContinuationWaitsForTheGain(t *testing.T) {
	g, me, _, bear := crystalLifeBoard(t)
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Library Card", Owner: me.ID, Controller: me.ID})
	life, hand := me.Life, me.Hand.Size()
	draws := 0

	g.WithWriteLock(func() {
		err := g.DestroyPermanentsKeyedForEffect([]uuid.UUID{bear}, hookTail(func(g *game.Game, o game.TailOutcome) error {
			draws++
			return g.DrawNForEffect(me.ID, 1)
		}))
		if err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})

	if !g.Exile.Contains(bear) {
		t.Fatal("the exile is the first instruction and lands before the life question")
	}
	if draws != 0 || me.Hand.Size() != hand {
		t.Fatalf("the destroy's draw ran %d times with the life prompt still open", draws)
	}
	if me.Life != life {
		t.Fatalf("life moved to %d before the CR 616 prompt was answered", me.Life)
	}
	answerTheOrderingPrompt(t, g)
	if draws != 1 || me.Hand.Size() != hand+1 {
		t.Fatalf("after the answer the draw ran %d times (hand %d -> %d), want once", draws, hand, me.Hand.Size())
	}
	if gained := me.Life - life; gained != 5 && gained != 6 {
		t.Errorf("gained %d, want 2 doubled and one more, in either order (5 or 6)", gained)
	}
}

// The gain is fixed as the replacement applies: the Crystal leaving
// while the life question is open still gains its controller the life.
func TestDarknessCrystalGainSurvivesTheCrystalLeaving(t *testing.T) {
	g, me, crystal, bear := crystalLifeBoard(t)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, crystal,
	); err != nil {
		t.Fatalf("move the Crystal away: %v", err)
	}
	answerTheOrderingPrompt(t, g)
	if gained := me.Life - life; gained != 5 && gained != 6 {
		t.Errorf("gained %d: the declared gain still happens in full", gained)
	}
	if !g.Exile.Contains(bear) {
		t.Error("the creature stays exiled")
	}
}

// A creature that dies to lethal damage (a state-based action, no
// continuation) is exiled and its gain still lands.
func TestDarknessCrystalGainsOnAStateBasedDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kbCrystal(g, me.ID)
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	life := me.Life
	if err := g.MarkDamage(bear, 2); err != nil {
		t.Fatalf("MarkDamage: %v", err)
	}
	if !g.Exile.Contains(bear) {
		t.Fatal("the creature that died to damage was not exiled")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}
