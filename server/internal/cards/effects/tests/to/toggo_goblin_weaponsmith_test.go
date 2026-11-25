package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const toggoOracle = "d334717a-f3a8-4ffa-bdc9-44ae7bdcc772"

func toggoLandEnters(t *testing.T, g *game.Game, owner *game.Player) {
	t.Helper()
	c := game.NewCard("Forest", owner.ID)
	c.TypeLine = "Basic Land — Forest"
	owner.Hand.PushTop(c)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, c.InstanceID); err != nil {
		t.Fatalf("land enters: %v", err)
	}
}

func toggoRock(g *game.Game, controller uuid.UUID) uuid.UUID {
	var id uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Rock" && c.Controller == controller {
				id = c.InstanceID
			}
		}
	})
	return id
}

func TestToggoMakesARockWhenYourLandEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B21Push(g, me.ID, "Toggo, Goblin Weaponsmith", "Legendary Creature — Goblin Artificer", toggoOracle, 2, 2)

	toggoLandEnters(t, g, opp)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Rock"); got != 0 {
		t.Fatalf("an opponent's land made %d Rocks, want 0", got)
	}
	toggoLandEnters(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Rock"); got != 1 {
		t.Fatalf("Rocks = %d, want 1", got)
	}
}

func TestToggoRockLetsTheEquippedCreatureSacrificeItToDealTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B21Push(g, me.ID, "Toggo, Goblin Weaponsmith", "Legendary Creature — Goblin Artificer", toggoOracle, 2, 2)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	toggoLandEnters(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	rock := toggoRock(g, me.ID)
	if rock == uuid.Nil {
		t.Fatal("no Rock")
	}
	var err error
	g.WithWriteLock(func() { err = g.AttachForEffect(rock, game.TargetRef{Kind: game.TargetCard, ID: bear}) })
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
	idx, ref := effectstest.GrantedActivatedIndex(t, g, bear)
	if idx < 0 {
		t.Fatal("the equipped creature has no granted ability")
	}
	params := game.ActivateAbilityParams{Ref: ref, Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}
	if err := g.ActivateCatalogAbility(me.ID, bear, idx, params); err == nil {
		t.Fatal("activated with no mana")
	}
	life := opp.Life
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.B16Activate(t, g, me.ID, bear, idx, params)
	if g.Battlefield.Contains(rock) {
		t.Error("the Rock was not sacrificed as a cost")
	}
	if !effectstest.B16Tapped(t, g, bear) {
		t.Error("the creature did not tap to pay")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d → %d, want -2", life, opp.Life)
	}
}
