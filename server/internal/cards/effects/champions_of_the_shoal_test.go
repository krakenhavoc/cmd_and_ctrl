package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const championsOfTheShoalOracle = "fc76ec95-f001-44df-89b2-b0f1b10c8e11"

func TestChampionsOfTheShoalIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(championsOfTheShoalOracle)
	if !ok || spec.Name != "Champions of the Shoal" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// castChampionsOfTheShoal casts it strictly for {3}{U}, beholding `reveal`.
func castChampionsOfTheShoal(t *testing.T, g *game.Game, reveal ...uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	advanceTo(t, g, game.StepPrecombatMain)
	b06AddMana(g.Seats[g.Turn.ActiveSeat], "U", "C", "C", "C")
	return castWithTapParams(t, g, "Champions of the Shoal", "Creature — Merfolk Soldier", "{3}{U}",
		championsOfTheShoalOracle, game.CastSpellParams{Strict: true, RevealIDs: reveal})
}

func championsOfTheShoalStunned(t *testing.T, g *game.Game, id uuid.UUID) (tapped bool, stun int) {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped, c.Counters[game.CounterStun]
}

// Beholding a Merfolk card from hand exiles it; entering taps and stuns
// one target creature, becoming tapped taps and stuns another, and when
// the Champions leave the exiled card returns to its owner's hand.
func TestChampionsOfTheShoalStunsOnEnterAndOnTapAndReturnsTheBeheldCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	merfolk := handCard(me, "Merrow Card", "Creature — Merfolk Wizard")
	first := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	second := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)

	champ, err := castChampionsOfTheShoal(t, g, merfolk)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !inExile(g, merfolk) || me.Hand.Contains(merfolk) {
		t.Fatal("the beheld Merfolk card was not exiled from the hand")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(champ) {
		t.Fatal("the Champions did not resolve")
	}
	answerPickTarget(t, g, first)
	passPriorityAroundTable(t, g)
	if tapped, stun := championsOfTheShoalStunned(t, g, first); !tapped || stun != 1 {
		t.Errorf("entering: target tapped %v with %d stun counters, want tapped with 1", tapped, stun)
	}

	if err := g.TapCard(champ, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	answerPickTarget(t, g, second)
	passPriorityAroundTable(t, g)
	if tapped, stun := championsOfTheShoalStunned(t, g, second); !tapped || stun != 1 {
		t.Errorf("becoming tapped: target tapped %v with %d stun counters, want tapped with 1", tapped, stun)
	}
	if _, stun := championsOfTheShoalStunned(t, g, first); stun != 1 {
		t.Errorf("the first target has %d stun counters, want still 1", stun)
	}

	b27Kill(g, champ)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(merfolk) {
		t.Error("the exiled Merfolk did not return to its owner's hand")
	}
}

// The behold is mandatory and takes a Merfolk only.
func TestChampionsOfTheShoalRefusesNoBeholdOrANonMerfolk(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	goblin := handCard(me, "Boggart Card", "Creature — Goblin")
	if _, err := castChampionsOfTheShoal(t, g); err == nil {
		t.Error("cast without beholding a Merfolk")
	}
	if _, err := castChampionsOfTheShoal(t, g, goblin); err == nil {
		t.Error("beheld a Goblin")
	}
}
