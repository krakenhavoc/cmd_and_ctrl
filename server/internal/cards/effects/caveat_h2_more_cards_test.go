package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3172 cav-h2, second file: the cards that closed on existing machinery.

const (
	h2CabbageMerchantOracle = "e31808bb-fb6e-487b-a973-db44961c84ee"
	h2TradingPostOracle     = "63788566-e25a-44bb-bb55-197e1b93b3e8"
	h2VizkopaOracle         = "f19e7c5c-67fa-4ae4-89b8-afa0e08a6c48"
	h2ZurgoOracle           = "6c48d888-9f5d-43f4-adbd-61dbdba09260"
	h2WindingOracle         = "c9404d7d-a026-4082-9fcb-1ab571a136b5"
)

func h2Food(g *game.Game, owner uuid.UUID) uuid.UUID {
	f := FoodToken()
	f.InstanceID, f.Owner, f.Controller = uuid.New(), owner, owner
	return pushBattlefieldCardWithTimestamp(g, f)
}

func TestCabbageMerchantTapsTwoUntappedFoodsForOneManaOfAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	merchant := seedPermanentWithOracle(g, me.ID, "The Cabbage Merchant", "Legendary Creature — Human Citizen", h2CabbageMerchantOracle)
	f1, f2, f3 := h2Food(g, me.ID), h2Food(g, me.ID), h2Food(g, me.ID)

	if err := g.ActivateManaAbility(me.ID, merchant, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{f1}}); err == nil {
		t.Fatal("one Food paid a two-Food cost")
	}
	assertTapped(t, g, false, merchant, f1, f2, f3)

	if err := g.ActivateManaAbility(me.ID, merchant, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{f1, f2}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	assertTapped(t, g, true, f1, f2)
	assertTapped(t, g, false, merchant, f3)
	if got := b10ResolveAllManaPicks(t, g, me.ID, "R"); got != 1 {
		t.Fatalf("mana picks answered = %d, want 1 (one mana of any colour)", got)
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "R" {
		t.Fatalf("mana pool = %v, want one red mana", me.ManaPool)
	}
	if err := g.ActivateManaAbility(me.ID, merchant, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{f2, f3}}); err == nil {
		t.Error("a tapped Food paid the cost")
	}
}

func TestTradingPostDiscardAbilityNeedsTheDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	post := pushCatalogPermanent(g, me.ID, "Trading Post", "Artifact", h2TradingPostOracle, false)
	advanceToMain(t, g)
	b06AddMana(me, "C")
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, post, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with no discard chosen")
	}
	if me.Life != life {
		t.Errorf("a refused activation changed life: %d → %d", life, me.Life)
	}
}

func TestVizkopaGuildmageDrainsOpponentsByTheLifeYouGainThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := b43Catalog(g, me.ID, "Vizkopa Guildmage", "Creature — Human Wizard", h2VizkopaOracle, 2, 2)
	advanceToMain(t, g)
	startOpp := opp.Life
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(mage, me.ID, 2) })
	passPriorityAroundTable(t, g)
	if opp.Life != startOpp {
		t.Fatalf("opponent life %d before the ability, want %d", opp.Life, startOpp)
	}
	b06AddMana(me, "W", "B", "C")
	b16Activate(t, g, me.ID, mage, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(mage, me.ID, 3) })
	passPriorityAroundTable(t, g)
	if opp.Life != startOpp-3 {
		t.Errorf("opponent life %d, want %d (lose what you gained)", opp.Life, startOpp-3)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(mage, me.ID, 1) })
	passPriorityAroundTable(t, g)
	if opp.Life != startOpp-4 {
		t.Errorf("opponent life %d after a second gain, want %d", opp.Life, startOpp-4)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(mage, me.ID, -2) })
	passPriorityAroundTable(t, g)
	if opp.Life != startOpp-4 {
		t.Errorf("losing life drained the opponent: %d", opp.Life)
	}
	advanceToNextSeatsTurn(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(mage, me.ID, 5) })
	passPriorityAroundTable(t, g)
	if opp.Life != startOpp-4 {
		t.Errorf("the effect outlived the turn: opponent at %d", opp.Life)
	}
}

func TestZurgoGetsACounterWhenACreatureItDamagedDiesThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zurgo := seedPermanentWithOracle(g, me.ID, "Zurgo Helmsmasher", "Legendary Creature — Orc Warrior", h2ZurgoOracle)
	hit := b12Creature(g, opp.ID, "Hit Bear", "Creature — Bear", 2, 3)
	other := b12Creature(g, opp.ID, "Other Bear", "Creature — Bear", 2, 3)
	bystander := b12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 3)
	elsewhere := b12Creature(g, me.ID, "Pinger", "Creature — Wizard", 1, 1)

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(zurgo, hit, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
		if err := g.DealDamageToCreatureForEffect(elsewhere, other, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bystander) })
	passPriorityAroundTable(t, g)
	if n := counterCount(g, zurgo, "+1/+1"); n != 0 {
		t.Fatalf("counters = %d after an unrelated death, want 0", n)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(other) })
	passPriorityAroundTable(t, g)
	if n := counterCount(g, zurgo, "+1/+1"); n != 0 {
		t.Fatalf("counters = %d after a creature another source damaged died, want 0", n)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(hit) })
	passPriorityAroundTable(t, g)
	if n := counterCount(g, zurgo, "+1/+1"); n != 1 {
		t.Errorf("counters = %d after the creature Zurgo damaged died, want 1", n)
	}
}

func TestWindingConstrictorAddsOneToCountersYouGet(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedPermanentWithOracle(g, me.ID, "Winding Constrictor", "Creature — Snake", h2WindingOracle)

	g.WithWriteLock(func() { _ = g.AddPlayerCounterByForEffect(opp.ID, me.ID, game.CounterPoison, 2) })
	if got := playerCounterFor(g, me.ID, game.CounterPoison); got != 3 {
		t.Errorf("poison you get = %d, want 3 (2 plus one)", got)
	}
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(opp.ID, game.CounterPoison, 2) })
	if got := playerCounterFor(g, opp.ID, game.CounterPoison); got != 2 {
		t.Errorf("an opponent's poison = %d, want 2 (not increased)", got)
	}
}

const h2VizierOracle = "4f890d42-e4f1-4eed-aa6c-718865de8927"

func TestVizierOfTheMenagerieLetsAnyManaPayACreatureSpellsColorlessSymbol(t *testing.T) {
	g, me, _ := spendTable(t)
	pushCatalogPermanent(g, me.ID, "Vizier of the Menagerie", "Creature — Snake Cleric", h2VizierOracle, false)
	creature := handSpell(me, "Colourless Seer", "Creature — Eldrazi", "{1}{C}")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("red mana for a creature spell's {C}: %v", err)
	}
}

func TestVizierOfTheMenagerieDoesNotWidenNonCreatureSpells(t *testing.T) {
	g, me, _ := spendTable(t)
	pushCatalogPermanent(g, me.ID, "Vizier of the Menagerie", "Creature — Snake Cleric", h2VizierOracle, false)
	sorcery := handSpell(me, "Colourless Sorcery", "Sorcery", "{C}")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	refusedForMana(t, g.CastSpell(me.ID, sorcery, game.CastSpellParams{Strict: true}), "red for a sorcery's {C}")
}

func TestOathOfNissaStillLeavesColorlessSymbolsAlone(t *testing.T) {
	g, me, _ := spendTable(t)
	pushCatalogPermanent(g, me.ID, "Oath of Nissa", "Legendary Enchantment", oathOfNissaOracle, false)
	walker := handSpell(me, "Colourless Walker", "Legendary Planeswalker — Test", "{1}{C}")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	refusedForMana(t, g.CastSpell(me.ID, walker, game.CastSpellParams{Strict: true}), "red for a planeswalker's {C} under the colour grant")
}
