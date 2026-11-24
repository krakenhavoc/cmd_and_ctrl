package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const halanaKessigRangerOracle = "f5aa7385-12ea-4ecb-86eb-e8622ebb6908"

func halanaKessigRangerEnter(t *testing.T, g *game.Game, owner *game.Player, name, typeLine, oracle string, p, tough int) uuid.UUID {
	t.Helper()
	c := game.NewCard(name, owner.ID)
	c.TypeLine, c.OracleID, c.Power, c.Toughness = typeLine, oracle, p, tough
	owner.Hand.PushTop(c)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, c.InstanceID); err != nil {
		t.Fatalf("enter %s: %v", name, err)
	}
	return c.InstanceID
}

func TestHalanaKessigRangerPaysTwoForPowerDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	halanaKessigRangerEnter(t, g, me, "Halana, Kessig Ranger", "Legendary Creature — Human Archer Ranger", halanaKessigRangerOracle, 3, 4)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Giant", 2, 9)
	halanaKessigRangerEnter(t, g, me, "Big Friend", "Creature — Giant", "", 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B09AddColorless(me, 2)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, victim); got != 3 {
		t.Errorf("damage on their creature = %d, want 3", got)
	}
}

func TestHalanaKessigRangerDecliningPaysNothingAndDealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	halanaKessigRangerEnter(t, g, me, "Halana, Kessig Ranger", "Legendary Creature — Human Archer Ranger", halanaKessigRangerOracle, 3, 4)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Giant", 2, 9)
	halanaKessigRangerEnter(t, g, me, "Big Friend", "Creature — Giant", "", 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B09AddColorless(me, 2)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, victim); got != 0 {
		t.Errorf("damage = %d, want 0", got)
	}
}

func TestHalanaKessigRangerIgnoresItselfEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	halanaKessigRangerEnter(t, g, me, "Halana, Kessig Ranger", "Legendary Creature — Human Archer Ranger", halanaKessigRangerOracle, 3, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("Halana offered a payment for entering herself")
	}
}
