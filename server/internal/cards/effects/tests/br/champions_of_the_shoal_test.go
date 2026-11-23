package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const championsOfTheShoalOracle = "fc76ec95-f001-44df-89b2-b0f1b10c8e11"

func TestChampionsOfTheShoalIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(championsOfTheShoalOracle)
	if !ok || spec.Name != "Champions of the Shoal" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// castChampionsOfTheShoal casts it strictly for {3}{U}, beholding `reveal`.
func castChampionsOfTheShoal(t *testing.T, g *game.Game, reveal ...uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.B06AddMana(g.Seats[g.Turn.ActiveSeat], "U", "C", "C", "C")
	return effectstest.CastWithTapParams(t, g, "Champions of the Shoal", "Creature — Merfolk Soldier", "{3}{U}",
		championsOfTheShoalOracle, game.CastSpellParams{Strict: true, RevealIDs: reveal})
}

// Beholding a Merfolk card from hand exiles it; entering taps and stuns
// one target creature, becoming tapped taps and stuns another, and when
// the Champions leave the exiled card returns to its owner's hand.
func TestChampionsOfTheShoalStunsOnEnterAndOnTapAndReturnsTheBeheldCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	merfolk := effectstest.HandCard(me, "Merrow Card", "Creature — Merfolk Wizard")
	first := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	second := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)

	champ, err := castChampionsOfTheShoal(t, g, merfolk)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !effectstest.InExile(g, merfolk) || me.Hand.Contains(merfolk) {
		t.Fatal("the beheld Merfolk card was not exiled from the hand")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(champ) {
		t.Fatal("the Champions did not resolve")
	}
	effectstest.AnswerPickTarget(t, g, first)
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, stun := effectstest.ChampionsOfTheShoalStunned(t, g, first); !tapped || stun != 1 {
		t.Errorf("entering: target tapped %v with %d stun counters, want tapped with 1", tapped, stun)
	}

	if err := g.TapCard(champ, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	effectstest.AnswerPickTarget(t, g, second)
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, stun := effectstest.ChampionsOfTheShoalStunned(t, g, second); !tapped || stun != 1 {
		t.Errorf("becoming tapped: target tapped %v with %d stun counters, want tapped with 1", tapped, stun)
	}
	if _, stun := effectstest.ChampionsOfTheShoalStunned(t, g, first); stun != 1 {
		t.Errorf("the first target has %d stun counters, want still 1", stun)
	}

	effectstest.B27Kill(g, champ)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(merfolk) {
		t.Error("the exiled Merfolk did not return to its owner's hand")
	}
}

// The behold is mandatory and takes a Merfolk only.
func TestChampionsOfTheShoalRefusesNoBeholdOrANonMerfolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	goblin := effectstest.HandCard(me, "Boggart Card", "Creature — Goblin")
	if _, err := castChampionsOfTheShoal(t, g); err == nil {
		t.Error("cast without beholding a Merfolk")
	}
	if _, err := castChampionsOfTheShoal(t, g, goblin); err == nil {
		t.Error("beheld a Goblin")
	}
}
