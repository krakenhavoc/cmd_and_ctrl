package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// state_trigger_cards_b_garruk_test.go — Garruk Relentless // Garruk,
// the Veil-Cursed (ADR 0107 §1, #1858): the fight, the CR 603.8
// transform trigger, and the back face's three loyalty abilities.

const stGarrukRelentles = "7cec9021-6f25-4fd8-b40e-adf4ffd3a7b8"

func garrukCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   stGarrukRelentles,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Counters:   map[string]int{game.CounterLoyalty: 3},
		Faces: []game.Face{
			{Name: "Garruk Relentless", TypeLine: "Legendary Planeswalker — Garruk", ManaCost: "{3}{G}",
				Colors: []string{"G"}, StartingLoyalty: 3},
			{Name: "Garruk, the Veil-Cursed", TypeLine: "Legendary Planeswalker — Garruk", Colors: []string{"G"}},
		},
	}
	c.SetFace(0)
	return c
}

// Garruk fights a 2/2: it takes 3 and hits back for 2, Garruk falls to
// one loyalty, and the state trigger turns him over.
func TestGarrukRelentlessFightsThenTransforms(t *testing.T) {
	g := newCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	garruk := pushBattlefieldCardWithTimestamp(g, garrukCard(me.ID))
	bear := apaPush(g, bob.ID, bob.ID, stCard("Bear", "", "Creature — Bear", 2, 2))
	apaActivate(t, g, me, garruk, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	passPriorityAroundTable(t, g)
	if onBattlefield(g, bear) {
		t.Fatal("the bear survived 3 damage")
	}
	c := apaLive(g, garruk)
	if c == nil {
		t.Fatal("Garruk left the battlefield")
	}
	if c.Counters[game.CounterLoyalty] != 1 {
		t.Fatalf("Garruk's loyalty = %d, want 1 (the bear dealt 2)", c.Counters[game.CounterLoyalty])
	}
	if c.ActiveFace != 1 {
		t.Fatal("Garruk at one loyalty did not transform")
	}
}

// Garruk's second 0 makes a 2/2 Wolf and does not transform him.
func TestGarrukRelentlessMakesAWolf(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	garruk := pushBattlefieldCardWithTimestamp(g, garrukCard(me.ID))
	apaActivate(t, g, me, garruk, 1, game.ActivateAbilityParams{})
	if n := onBattlefieldNamed(g, "Wolf"); n != 1 {
		t.Fatalf("%d Wolves, want 1", n)
	}
	if c := apaLive(g, garruk); c.ActiveFace != 0 {
		t.Fatal("Garruk transformed at three loyalty")
	}
}

// The Veil-Cursed: +1 deathtouch Wolf, −3 trample and +X/+X for the
// creature cards in the graveyard.
func TestGarrukTheVeilCursedWolfAndOverrun(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	walker := pushWalkerForTest(g, me.ID, "Garruk, the Veil-Cursed", stGarrukRelentles+"#1", 5)
	apaActivate(t, g, me, walker, 0, game.ActivateAbilityParams{})
	wolves := battlefieldIDsNamed(g, "Wolf")
	if len(wolves) != 1 || !game.HasKeyword(apaLive(g, wolves[0]), "deathtouch") {
		t.Fatalf("+1 made %d Wolves, want one with deathtouch", len(wolves))
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	advanceToMain(t, g2)
	w2 := pushWalkerForTest(g2, me2.ID, "Garruk, the Veil-Cursed", stGarrukRelentles+"#1", 5)
	bear := apaPush(g2, me2.ID, me2.ID, stCard("Bear", "", "Creature — Bear", 2, 2))
	for i := 0; i < 2; i++ {
		me2.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me2.ID})
	}
	apaActivate(t, g2, me2, w2, 2, game.ActivateAbilityParams{})
	c := apaLive(g2, bear)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 || !game.HasKeyword(c, "trample") {
		t.Fatalf("after −3: %d/%d trample=%v, want 4/4 with trample", c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(c, "trample"))
	}
}
