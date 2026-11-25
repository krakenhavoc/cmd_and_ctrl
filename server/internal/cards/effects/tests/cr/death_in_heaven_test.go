package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const deathInHeavenOracle = "f2bf5014-677a-4880-b926-d93457348bdc"

func TestDeathInHeavenIsFullWithNoCaveats(t *testing.T) {
	spec, ok := effects.Lookup(deathInHeavenOracle)
	if !ok || spec.Name != "Death in Heaven" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// TestDeathInHeavenMillsExilesAndReturnsTheCreaturesAsCybermen walks the
// Saga through its three chapters. Chapters I and II mill two and exile the
// target's graveyard linked to the Saga (CR 607.2a); chapter III puts the
// linked creature cards, and only those, onto the battlefield face down
// under my control as 2/2 Cyberman artifact creatures (CR 708.2, CR 400.7).
func TestDeathInHeavenMillsExilesAndReturnsTheCreaturesAsCybermen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	land := effectstest.PutOnTopForTest(opp, "Swamp", "Basic Land — Swamp")
	bear := effectstest.PutOnTopForTest(opp, "Grizzly Bears", "Creature — Bear")
	dead := uuid.New()
	g.WithWriteLock(func() {
		opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Knight", TypeLine: "Creature — Human Knight", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	})
	// A creature card exiled by something else is not "exiled with" the Saga.
	stranger := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: stranger, Name: "Stranger", TypeLine: "Creature — Elf", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Death in Heaven", TypeLine: "Enchantment — Saga", OracleID: deathInHeavenOracle,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLore: 0},
	})

	effectstest.AdvanceToMain(t, g)
	effectstest.PickPlayerTarget(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if len(opp.Graveyard.Cards) != 0 {
		t.Fatalf("chapter I left %d cards in the graveyard", len(opp.Graveyard.Cards))
	}
	for _, id := range []uuid.UUID{land, bear, dead} {
		if !g.Exile.Contains(id) {
			t.Fatalf("card %s was not milled and exiled", id)
		}
	}

	// Chapter II on my next turn exiles two more (filler) cards.
	effectstest.EndTurn(t, g)
	effectstest.AdvanceToMainOf(t, g, seat)
	effectstest.PickPlayerTarget(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.EndTurn(t, g)
	effectstest.AdvanceToMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(land) {
		t.Error("the land card left exile")
	}
	if !g.Exile.Contains(stranger) {
		t.Error("a creature card not exiled with the Saga was put onto the battlefield")
	}
	cybermen := 0
	for _, c := range g.Battlefield.Cards {
		if c.Owner != opp.ID || !c.FaceDown {
			continue
		}
		v := effectstest.LayeredCard(t, g, c.InstanceID)
		if c.Controller != me.ID || !v.IsArtifact() || !v.IsCreature() || !v.HasSubtype("Cyberman") || v.CurrentPower() != 2 || v.CurrentToughness() != 2 {
			t.Errorf("returned %s as %v %v %d/%d under %s, want my 2/2 Cyberman artifact creature",
				c.Name, v.Effective().Types, v.Effective().Subtypes, v.CurrentPower(), v.CurrentToughness(), c.Controller)
		}
		cybermen++
	}
	if cybermen != 2 {
		t.Errorf("%d face-down Cybermen, want 2 (the Bear and the Knight)", cybermen)
	}
	if g.Exile.Contains(bear) || g.Exile.Contains(dead) {
		t.Error("a linked creature card is still in exile")
	}
}
