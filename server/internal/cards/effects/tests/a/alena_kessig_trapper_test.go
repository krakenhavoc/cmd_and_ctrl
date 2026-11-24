package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const alenaKessigTrapperOracle = "ca349ba9-b5ea-48e6-a617-c9b7f2708853"

func alenaKessigTrapperEnter(t *testing.T, g *game.Game, owner *game.Player, name string, p int) uuid.UUID {
	t.Helper()
	c := game.NewCard(name, owner.ID)
	c.TypeLine, c.Power, c.Toughness = "Creature — Bear", p, p
	owner.Hand.PushTop(c)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, c.InstanceID); err != nil {
		t.Fatalf("enter %s: %v", name, err)
	}
	return c.InstanceID
}

func TestAlenaKessigTrapperAddsRedForTheGreatestPowerThatEnteredThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	alena := effectstest.B21Push(g, me.ID, "Alena, Kessig Trapper", "Legendary Creature — Human Scout", alenaKessigTrapperOracle, 4, 3)
	effectstest.Unsick(g, alena)
	// A bigger creature that was already here does not count.
	effectstest.B21Push(g, me.ID, "Old Giant", "Creature — Giant", "", 7, 7)
	alenaKessigTrapperEnter(t, g, me, "Small Newcomer", 2)
	alenaKessigTrapperEnter(t, g, me, "Big Newcomer", 5)

	effectstest.B23TapMana(t, g, me.ID, alena)
	if got := effectstest.PoolSize(g, me.ID); got != 5 {
		t.Errorf("mana added = %d, want 5", got)
	}
}

func TestAlenaKessigTrapperAddsNothingWithNothingNew(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	alena := effectstest.B21Push(g, me.ID, "Alena, Kessig Trapper", "Legendary Creature — Human Scout", alenaKessigTrapperOracle, 4, 3)
	effectstest.Unsick(g, alena)
	effectstest.B21Push(g, me.ID, "Old Giant", "Creature — Giant", "", 7, 7)
	effectstest.B23TapMana(t, g, me.ID, alena)
	if got := effectstest.PoolSize(g, me.ID); got != 0 {
		t.Errorf("mana added = %d, want 0", got)
	}
}
