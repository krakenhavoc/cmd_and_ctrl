package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	cmdr2B17ChatterstormOracle = "7e6c7b57-bd6a-4ac8-bcab-c3a879f479c2"
	cmdr2B17SquirrelMobOracle  = "451486df-6160-4ff7-b47c-dd3760417a31"
	cmdr2B17AntQueenOracle     = "50de88d4-0beb-45e5-9664-d4f0e280e652"
	cmdr2B17MarrowGnawerOracle = "2b934761-5720-477b-94b7-d1d91dc581a6"
)

func TestCmdr2B17ChatterstormStormCopiesEachMakeASquirrel(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMainOf(t, g, 0)
	stormFiller(t, g)
	stormFiller(t, g)
	castCatalogSpell(t, g, "Chatterstorm", "Sorcery", cmdr2B17ChatterstormOracle, nil)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Squirrel", me.ID); got != 3 {
		t.Errorf("%d Squirrels, want 3 (the spell and two storm copies)", got)
	}
}

func TestCmdr2B17SquirrelMobCountsEveryOtherSquirrel(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mob := b41CatalogCreature(g, me.ID, "Squirrel Mob", "Creature — Squirrel", cmdr2B17SquirrelMobOracle, 2, 2)
	if got := effectivePower(t, g, mob); got != 2 {
		t.Fatalf("alone the Mob is %d power, want 2", got)
	}
	b12Creature(g, me.ID, "Mine", "Creature — Squirrel", 1, 1)
	b12Creature(g, opp.ID, "Theirs", "Creature — Squirrel", 1, 1)
	b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if p, tg := effectivePower(t, g, mob), effectiveToughness(t, g, mob); p != 4 || tg != 4 {
		t.Errorf("Mob is %d/%d, want 4/4 (two other Squirrels, either side, a Bear doesn't count)", p, tg)
	}
	mob2 := b41CatalogCreature(g, me.ID, "Squirrel Mob", "Creature — Squirrel", cmdr2B17SquirrelMobOracle, 2, 2)
	if got := effectivePower(t, g, mob2); got != 5 {
		t.Errorf("second Mob power %d, want 5 (counts the first Mob)", got)
	}
}

func TestCmdr2B17AntQueenMakesInsects(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	q := pushCatalogPermanent(g, me.ID, "Ant Queen", "Creature — Insect", cmdr2B17AntQueenOracle, false)
	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, q, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activating with only {G} in the pool should be refused")
	}
	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, q, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Insect", me.ID); got != 1 {
		t.Errorf("%d Insects, want 1", got)
	}
}

func TestCmdr2B17MarrowGnawerGivesEveryRatFearAndCountsAfterTheSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gn := pushCatalogPermanent(g, me.ID, "Marrow-Gnawer", "Legendary Creature — Rat Rogue", cmdr2B17MarrowGnawerOracle, false)
	mine := b12Creature(g, me.ID, "Rat A", "Creature — Rat", 1, 1)
	theirs := b12Creature(g, opp.ID, "Rat B", "Creature — Rat", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{gn, mine, theirs} {
		if !hasEffectiveKeyword(t, g, id, "fear") {
			t.Errorf("a Rat lacks fear")
		}
	}
	if hasEffectiveKeyword(t, g, bear, "fear") {
		t.Error("a Bear has fear")
	}
	// A non-Rat is not a legal sacrifice.
	if err := g.ActivateCatalogAbility(me.ID, gn, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Error("sacrificing a Bear was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, gn, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{mine}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Rats left under my control: Marrow-Gnawer only (the other Rat is
	// theirs), so X = 1.
	if got := countOnBattlefield(g, "Rat", me.ID); got != 1 {
		t.Errorf("%d Rat tokens, want 1", got)
	}
}
