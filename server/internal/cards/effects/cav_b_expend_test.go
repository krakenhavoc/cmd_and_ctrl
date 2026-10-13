package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cav_b_expend_test.go — expend (#3172) on Trailtracker Scout.

const trailtrackerScoutOracle = "d38646dc-d1ea-473c-aaf1-df8ba327459f"

// castPaid casts a plain instant of the given mana cost by paying
// exactly that much (strict), leaving it on the stack.
func castPaid(t *testing.T, g *game.Game, me *game.Player, name, cost, pool string) {
	t.Helper()
	id := handCardFull(me, name, "Instant", cost, "", nil)
	floatForTest(g, me, pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
}

func setupScout(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushDiesCreatureForTest(g, me.ID, "Trailtracker Scout", trailtrackerScoutOracle, "Creature — Raccoon Scout", 1, 3)
	dead := pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	pushGraveyardCardTyped(me, "Dead Spell", "Instant")
	toMainForCost(t, g)
	return g, me, dead
}

// Expend 8: the eighth total mana spent on spells in a turn, however it
// is split across spells, triggers once and returns a permanent card; a
// nonpermanent card is not a legal target.
func TestTrailtrackerScoutExpendsEightAcrossTwoSpells(t *testing.T) {
	g, me, dead := setupScout(t)
	castPaid(t, g, me, "Four A", "{3}{G}", "CCCG")
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("four mana spent: no expend 8 yet")
	}
	castPaid(t, g, me, "Four B", "{3}{G}", "CCCG")
	pick := latestPickTarget(g, me.ID)
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
	answerTriggerTargets(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the permanent card should be back in hand")
	}

	// Ninth and beyond: no second trigger this turn.
	castPaid(t, g, me, "Four C", "{3}{G}", "CCCG")
	if latestPickTarget(g, me.ID) != nil {
		t.Error("expend 8 triggers once per turn: the running total is already past it")
	}
}

// One spell that carries the total across the threshold triggers it
// (five, then six: the second spell spends mana six through eleven).
func TestTrailtrackerScoutExpendsWhenOneSpellCrossesTheLine(t *testing.T) {
	g, me, _ := setupScout(t)
	castPaid(t, g, me, "Five", "{4}{G}", "CCCCG")
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("five mana: not yet")
	}
	castPaid(t, g, me, "Six", "{5}{G}", "CCCCCG")
	if latestPickTarget(g, me.ID) == nil {
		t.Error("the spell that crossed eight should trigger expend 8")
	}
}

// Seven mana is not eight; a free (unpaid) cast spends nothing.
func TestTrailtrackerScoutSevenManaIsNotEnough(t *testing.T) {
	g, me, _ := setupScout(t)
	castPaid(t, g, me, "Seven", "{6}{G}", "CCCCCCG")
	if latestPickTarget(g, me.ID) != nil {
		t.Error("seven mana must not expend 8")
	}
}
