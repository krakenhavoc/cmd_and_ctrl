package testsi

import (
	"testing"

	cardsmi "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/mi"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kuja_b_midgar_test.go — Midgar, City of Mako // Reactor Raid
// (#3213, slice kuja-b).

func TestReactorRaidSacrificeDrawsTwoThenTheLandIsPlayableFromExile(t *testing.T) {
	g, me, id := effectstest.SeatLandAdventure(t, cardsmi.MidgarOracleID, "Midgar, City of Mako", "Reactor Raid", "{2}{B}", []string{"B"})
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Reactor Raid: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerSacrifice(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
	// The card left the hand (-1) and two were drawn (+2).
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want +1 net", hand, got)
	}
	effectstest.AssertLandPlayableFromExile(t, g, me, id)
}

func TestReactorRaidDecliningDrawsNothing(t *testing.T) {
	g, me, id := effectstest.SeatLandAdventure(t, cardsmi.MidgarOracleID, "Midgar, City of Mako", "Reactor Raid", "{2}{B}", []string{"B"})
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Reactor Raid: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) == false {
		t.Error("declining must not sacrifice")
	}
	if got := me.Hand.Size(); got != hand-1 {
		t.Errorf("hand %d -> %d, want only the cast card gone", hand, got)
	}
}

// With nothing to sacrifice the question is not asked.
func TestReactorRaidWithNothingToSacrificeAsksNothing(t *testing.T) {
	g, me, id := effectstest.SeatLandAdventure(t, cardsmi.MidgarOracleID, "Midgar, City of Mako", "Reactor Raid", "{2}{B}", []string{"B"})
	// The default catalog game has an empty battlefield.
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Reactor Raid: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil {
			t.Fatalf("a %v prompt was queued with nothing to sacrifice", c.Kind)
		}
	}
	if got := me.Hand.Size(); got != hand-1 {
		t.Errorf("hand %d -> %d, want only the cast card gone", hand, got)
	}
}

func TestMidgarIsALandPlayFromHandThatEntersTapped(t *testing.T) {
	g, me, id := effectstest.SeatLandAdventure(t, cardsmi.MidgarOracleID, "Midgar, City of Mako", "Reactor Raid", "{2}{B}", []string{"B"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play Midgar: %v", err)
	}
	if !g.Battlefield.Contains(id) || g.Stack.Contains(id) {
		t.Fatal("Midgar should be a land play, not a spell")
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id && !c.Tapped {
			t.Error("Midgar should enter tapped")
		}
	}
}
