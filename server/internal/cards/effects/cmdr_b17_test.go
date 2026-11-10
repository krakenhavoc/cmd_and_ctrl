package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b17_test.go — Commander deck cards, batch 17 (#5368).

const (
	cmdrB17ShalaiOracle     = "a0c47ab6-dfb4-46ee-a3f7-9e1521b4bb4b"
	cmdrB17DroverOracle     = "b0d60856-80f9-4c64-9768-28a0a690ecb1"
	cmdrB17GargoyleOracle   = "11284767-7095-4e0e-bc60-c69d58098348"
	cmdrB17ChaplainOracle   = "99e24041-910e-4ae2-a55b-b938d752eeff"
	cmdrB17BlightPileOracle = "1c0fc9b4-177e-4529-827f-a1e17ff7a752"
)

func TestCmdrB17ShalaiGivesHexproofToYouWalkersAndOthers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shalai := pushCatalogPermanent(g, me.ID, "Shalai, Voice of Plenty", "Legendary Creature — Angel", cmdrB17ShalaiOracle, false)
	mine := pushVanillaCreature(g, me.ID, "My Guy", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Guy", 2, 2)
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: me.ID, Controller: me.ID,
	})
	if !hasPlayerAbility(playerAbilities(g, me), KeywordHexproof) {
		t.Error("you have hexproof")
	}
	if hasPlayerAbility(playerAbilities(g, opp), KeywordHexproof) {
		t.Error("opponents don't")
	}
	for _, c := range []struct {
		id   uuid.UUID
		name string
		want bool
	}{{mine, "other creature", true}, {walker, "planeswalker", true}, {theirs, "their creature", false}, {shalai, "Shalai herself", false}} {
		if got := eotHasAbility(effectiveAbilities(t, g, c.id), "hexproof"); got != c.want {
			t.Errorf("%s hexproof = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestCmdrB17ShalaiPutsACounterOnEachCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shalai := pushCatalogPermanent(g, me.ID, "Shalai, Voice of Plenty", "Legendary Creature — Angel", cmdrB17ShalaiOracle, false)
	mine := pushVanillaCreature(g, me.ID, "My Guy", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Guy", 2, 2)
	me.ManaPool.AddMana(game.ManaToken{Color: "G"}, game.ManaToken{Color: "G"},
		game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, shalai, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{shalai: 1, mine: 1, theirs: 0} {
		if got := counterCount(g, id, "+1/+1"); got != want {
			t.Errorf("%s: %d counters, want %d", id, got, want)
		}
	}
}

func TestCmdrB17TwilightDroverGrowsWhenAnyCreatureTokenLeavesAndMakesSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drover := pushCatalogPermanent(g, me.ID, "Twilight Drover", "Creature — Spirit", cmdrB17DroverOracle, false)
	tok := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	nontoken := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	if err := g.DestroyPermanentForEffect(nontoken); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, drover, "+1/+1"); got != 0 {
		t.Fatalf("a nontoken creature leaving gave %d counters", got)
	}
	if err := g.DestroyPermanentForEffect(tok); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, drover, "+1/+1"); got != 1 {
		t.Fatalf("a creature token leaving gave %d counters, want 1", got)
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "W"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, drover, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Spirit"); got != 2 {
		t.Fatalf("want two Spirit tokens, got %d", got)
	}
	// The counter was the cost, and the tokens leaving later feed it again.
	if got := counterCount(g, drover, "+1/+1"); got != 0 {
		t.Errorf("counter should be spent, have %d", got)
	}
}

func TestCmdrB17WakestoneGargoyleLetsDefendersAttackThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gargoyle := pushTestPermanent(g, me.ID, game.Card{Name: "Wakestone Gargoyle", OracleID: cmdrB17GargoyleOracle,
		TypeLine: "Creature — Gargoyle", Power: 3, Toughness: 4})
	wall := pushTestPermanent(g, me.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Keywords: []string{"defender"}})
	theirs := pushTestPermanent(g, opp.ID, game.Card{Name: "Their Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Keywords: []string{"defender"}})
	advanceToMainOf(t, g, 0)
	advanceTo(t, g.Clone(), game.StepDeclareAttackers)
	before := g.Clone()
	advanceTo(t, before, game.StepDeclareAttackers)
	wantDefenderRefused(t, before, gargoyle, opp.ID)
	wantDefenderRefused(t, before, wall, opp.ID)

	me.ManaPool.AddMana(game.ManaToken{Color: "W"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, gargoyle, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	// A defender that arrives after the ability resolved is covered too.
	late := pushTestPermanent(g, me.ID, game.Card{Name: "Late Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Keywords: []string{"defender"}})
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{gargoyle, wall, late} {
		if err := declareResult(g, id, opp.ID); err != nil {
			t.Errorf("attack with %s: %v", id, err)
		}
	}
	if err := declareResult(g, theirs, me.ID); err == nil {
		t.Error("an opponent's defender gets no waiver")
	}
	if !hasEffectiveKeyword(t, g, wall, "defender") {
		t.Error("the keyword stays")
	}
	// Until end of turn only.
	advanceToDeclareAttackersOf(t, g, 1)
	advanceToDeclareAttackersOf(t, g, 0)
	wantDefenderRefused(t, g, wall, opp.ID)
}

func TestCmdrB17WingmantleChaplainCountsDefendersAndTriggersOnOthers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTestPermanent(g, me.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Keywords: []string{"defender"}})
	castCatalogSpell(t, g, "Wingmantle Chaplain", "Creature — Human Cleric", cmdrB17ChaplainOracle, nil)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Bird"); got != 2 {
		t.Fatalf("Birds after the enters trigger = %d, want 2 (the Wall and the Chaplain)", got)
	}
	cmdrB17Enter(t, g, me, game.Card{Name: "Late Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Keywords: []string{"defender"}})
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Bird"); got != 3 {
		t.Fatalf("Birds after another defender entered = %d, want 3", got)
	}
	cmdrB17Enter(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Bird"); got != 3 {
		t.Fatalf("a creature without defender made a Bird: %d", got)
	}
}

func TestCmdrB17BlightPileDrainsByDefenders(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pile := pushTestPermanent(g, me.ID, game.Card{Name: "Blight Pile", OracleID: cmdrB17BlightPileOracle,
		TypeLine: "Creature — Phyrexian", Power: 3, Toughness: 3})
	pushTestPermanent(g, me.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Keywords: []string{"defender"}})
	pushTestPermanent(g, opp.ID, game.Card{Name: "Their Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Keywords: []string{"defender"}})
	before := opp.Life
	me.ManaPool.AddMana(game.ManaToken{Color: "B"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, pile, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 2 {
		t.Errorf("opponent lost %d, want 2 (Blight Pile has defender if the keyword is printed)", got)
	}
}

// cmdrB17Enter puts a card into a player's hand and moves it onto the
// battlefield, so its enters triggers fire as they would in play.
func cmdrB17Enter(t *testing.T, g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	t.Helper()
	c.InstanceID, c.Owner, c.Controller = uuid.New(), p.ID, p.ID
	p.Hand.PushTop(c)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: p.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, c.InstanceID); err != nil {
		t.Fatalf("enter %s: %v", c.Name, err)
	}
	return c.InstanceID
}
