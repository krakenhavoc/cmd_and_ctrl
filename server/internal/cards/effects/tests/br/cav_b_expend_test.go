package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cav_b_expend_test.go — expend (#3172) on Trailtracker Scout.

const trailtrackerScoutOracle = "d38646dc-d1ea-473c-aaf1-df8ba327459f"

func setupScout(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushDiesCreatureForTest(g, me.ID, "Trailtracker Scout", trailtrackerScoutOracle, "Creature — Raccoon Scout", 1, 3)
	dead := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	effectstest.PushGraveyardCardTyped(me, "Dead Spell", "Instant")
	effectstest.ToMainForCost(t, g)
	return g, me, dead
}

// Expend 8: the eighth total mana spent on spells in a turn, however it
// is split across spells, triggers once and returns a permanent card; a
// nonpermanent card is not a legal target.
func TestTrailtrackerScoutExpendsEightAcrossTwoSpells(t *testing.T) {
	g, me, dead := setupScout(t)
	effectstest.CastPaid(t, g, me, "Four A", "{3}{G}", "CCCG")
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("four mana spent: no expend 8 yet")
	}
	effectstest.CastPaid(t, g, me, "Four B", "{3}{G}", "CCCG")
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the eighth mana was spent: the trigger should ask for its target")
	}
	spellInYard := uuid.Nil
	for _, c := range me.Graveyard.Cards {
		if c.Name == "Dead Spell" {
			spellInYard = c.InstanceID
		}
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: spellInYard}}); err == nil {
		t.Error("an instant card is not a permanent card")
	}
	effectstest.AnswerTriggerTargets(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the permanent card should be back in hand")
	}

	// Ninth and beyond: no second trigger this turn.
	effectstest.CastPaid(t, g, me, "Four C", "{3}{G}", "CCCG")
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("expend 8 triggers once per turn: the running total is already past it")
	}
}

// One spell that carries the total across the threshold triggers it
// (five, then six: the second spell spends mana six through eleven).
func TestTrailtrackerScoutExpendsWhenOneSpellCrossesTheLine(t *testing.T) {
	g, me, _ := setupScout(t)
	effectstest.CastPaid(t, g, me, "Five", "{4}{G}", "CCCCG")
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("five mana: not yet")
	}
	effectstest.CastPaid(t, g, me, "Six", "{5}{G}", "CCCCCG")
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Error("the spell that crossed eight should trigger expend 8")
	}
}

// Seven mana is not eight; a free (unpaid) cast spends nothing.
func TestTrailtrackerScoutSevenManaIsNotEnough(t *testing.T) {
	g, me, _ := setupScout(t)
	effectstest.CastPaid(t, g, me, "Seven", "{6}{G}", "CCCCCCG")
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("seven mana must not expend 8")
	}
}
