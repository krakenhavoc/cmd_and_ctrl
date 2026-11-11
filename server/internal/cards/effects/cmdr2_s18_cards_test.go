package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	cmdr2S18SteelburrOracle = "d1f3d5ca-cb17-400f-8ed9-3b9a4f151b62"
	cmdr2S18BanditOracle    = "2a14670b-2ab6-4e1d-9b33-610a87bdb7ce"
)

func TestCmdr2S18SteelburrChampionGrowsOnOpponentNoncreatureSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	champ := b12Push(g, me.ID, "Steelburr Champion", "Creature — Mouse Soldier", cmdr2S18SteelburrOracle, 1, 1)
	assertKeywords(t, g, champ, "vigilance")
	advanceToMain(t, g)
	castFree(t, g, me, "My Bear", "Creature — Bear")
	castFree(t, g, me, "My Instant", "Instant")
	if got := plusOnes(g, champ); got != 0 {
		t.Fatalf("a creature spell or my own instant gave %d counters, want 0", got)
	}
	castFree(t, g, opp, "Their Instant", "Instant")
	if got := plusOnes(g, champ); got != 1 {
		t.Errorf("an opponent's instant left %d counters, want 1", got)
	}
}

func TestCmdr2S18SteelburrChampionOffspringMakesAOneOneCopy(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	blbGLCastCreature(t, g, "Steelburr Champion", "Creature — Mouse Soldier", cmdr2S18SteelburrOracle, 1, 1, []int{0})
	for i := 0; i < 4; i++ {
		blbGLSettleTriggerOrder(g, me.ID)
		passPriorityAroundTable(t, g)
	}
	if got := countNamed(g, "Steelburr Champion"); got != 2 {
		t.Fatalf("%d Steelburr Champions, want 2 (offspring paid)", got)
	}
}

func TestCmdr2S18ProsperousBanditMakesThatManyTappedTreasures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bandit := b12Push(g, me.ID, "Prosperous Bandit", "Creature — Raccoon Rogue", cmdr2S18BanditOracle, 2, 2)
	assertKeywords(t, g, bandit, "first strike")
	before := countNamed(g, "Treasure")
	dealCombatDamageToPlayer(g, bandit, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if got := countNamed(g, "Treasure") - before; got != 3 {
		t.Fatalf("created %d Treasures, want 3", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && !c.Tapped {
			t.Errorf("a Treasure entered untapped")
		}
	}
}

const cmdr2S18KillerServiceOracle = "0f30f32f-3c4c-4f73-8b2a-82045a40bda9"

func TestCmdr2S18KillerServiceMakesAFoodPerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	castFromHandForTest(t, g, me, "Killer Service", "Enchantment", "", cmdr2S18KillerServiceOracle, game.CastSpellParams{FromZone: "hand"})
	passPriorityAroundTable(t, g)
	if got := countNamed(g, "Food"); got != 3 {
		t.Errorf("created %d Foods, want 3 (one per opponent)", got)
	}
}

func TestCmdr2S18KillerServicePaysAndSacrificesATokenForARhino(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Killer Service", "Enchantment", cmdr2S18KillerServiceOracle, 0, 0)
	tok := FoodToken()
	tok.InstanceID, tok.Owner, tok.Controller = uuid.New(), me.ID, me.ID
	food := pushBattlefieldCardWithTimestamp(g, tok)
	advanceToEndStepOf(t, g, 0)
	for i := 0; i < 8 && !hasPayUnlessFor(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("the end step offers the optional {2}")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	answerPayUnless(t, g, me.ID, true)
	answerSacrifice(t, g, me.ID, food)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(food) {
		t.Error("the token was not sacrificed")
	}
	if got := countNamed(g, "Rhino Warrior"); got != 1 {
		t.Errorf("%d Rhino Warriors, want 1", got)
	}
}
