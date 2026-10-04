package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrificed_pool_a_test.go — the first pool of spells that read the
// permanent their additional cost sacrificed (ADR 0113 §1, owner
// decision 2, #2072). One test per card; the record itself is
// sacrificed_objects_test.go's.

const (
	thudOracle              = "5d2a9859-1353-4431-9343-f5999450acd1"
	pyrrhicBlastOracle      = "98ec58aa-8776-4a19-bff8-e288a9bd6bac"
	lifesLegacyOracle       = "cb7baa46-7963-4d45-9a6e-e3c34db3c0e6"
	carrionOracle           = "84ca8d74-95ec-4a9c-8793-7a195ad28be9"
	callForBloodOracle      = "f2782c21-ed99-4821-8938-cc539e783ec7"
	ichorExplosionOracle    = "d33aa11b-011b-4d12-85a9-4f956153fb1d"
	finalStrikeOracle       = "4d98aea2-b4ff-4903-ba28-a53fbfaad6b1"
	riteOfConsumptionOracle = "327121a1-f193-44c6-a834-802095abec84"
	tormentedThoughtsOracle = "18aee20d-ab73-4fdb-a65c-101beae5fcf5"
	severedStrandsOracle    = "25f0527c-1340-4218-970f-9e4f84ef96e8"
	worthyCauseOracle       = "b24063f7-157f-47c0-919a-33856d53b902"
	morbidCuriosityOracle   = "59c620fb-a58d-4038-a7e2-0185324b5123"
	reckonersBargainOracle  = "044ea111-ce87-49ab-98f7-ae8447f48ac5"
	nastyEndOracle          = "f73da5d8-fd15-4315-ad3c-c86c28087285"
	forgeArmorOracle        = "7d4d9bfa-7e85-461b-bd62-1de97690a110"
)

// paCast casts a pool spell paying `fodder` and fails the test on a
// refusal.
func paCast(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id, err := castWithTapParams(t, g, name, typeLine, cost, oracle, params)
	if err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// paPermanent is a permanent with a mana cost, for the mana-value
// readers.
func paPermanent(g *game.Game, owner uuid.UUID, name, typeLine, manaCost string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func TestThudDealsTheSacrificedPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := lifeOf(g, opp.ID)
	paCast(t, g, "Thud", "Sorcery", "{R}", thudOracle, game.CastSpellParams{
		Targets: soPlayer(opp.ID), SacrificeIDs: []uuid.UUID{soCreature(g, me.ID, "Beast", 5, 5)},
	})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-5 {
		t.Errorf("opponent's life %d → %d, want 5 damage", life, got)
	}
}

func TestPyrrhicBlastDealsThePowerAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := lifeOf(g, opp.ID)
	paCast(t, g, "Pyrrhic Blast", "Instant", "{3}{R}", pyrrhicBlastOracle, game.CastSpellParams{
		Targets: soPlayer(opp.ID), SacrificeIDs: []uuid.UUID{soCreature(g, me.ID, "Beast", 3, 3)},
	})
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-3 {
		t.Errorf("opponent's life %d → %d, want 3 damage", life, got)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestLifesLegacyDrawsThePower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	paCast(t, g, "Life's Legacy", "Sorcery", "{1}{G}", lifesLegacyOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{soCreature(g, me.ID, "Beast", 4, 4)},
	})
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("drew %d, want 4", got)
	}
}

func TestCarrionMakesAnInsectPerPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	paCast(t, g, "Carrion", "Instant", "{1}{B}{B}", carrionOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{soCreature(g, me.ID, "Beast", 3, 3)},
	})
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Insect"); n != 3 {
		t.Fatalf("%d Insects, want 3", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Insect" && (c.CurrentPower() != 0 || c.CurrentToughness() != 1 || len(c.Colors) != 1 || c.Colors[0] != "B") {
			t.Errorf("Insect is %d/%d %v, want a 0/1 black", c.CurrentPower(), c.CurrentToughness(), c.Colors)
		}
	}
}

func TestCallForBloodShrinksByThePower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := soCreature(g, opp.ID, "Victim", 5, 5)
	paCast(t, g, "Call for Blood", "Instant — Arcane", "{4}{B}", callForBloodOracle, game.CastSpellParams{
		Targets: vsTarget(victim), SacrificeIDs: []uuid.UUID{soCreature(g, me.ID, "Beast", 3, 3)},
	})
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCardByID(g, victim)
	if !ok || c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("Victim is %d/%d, want 2/2 (5/5 with -3/-3)", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestIchorExplosionShrinksEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := soCreature(g, me.ID, "Mine", 3, 3)
	theirs := soCreature(g, opp.ID, "Theirs", 5, 5)
	paCast(t, g, "Ichor Explosion", "Sorcery", "{5}{B}{B}", ichorExplosionOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{soCreature(g, me.ID, "Beast", 2, 2)},
	})
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{mine: 1, theirs: 3} {
		c, ok := battlefieldCardByID(g, id)
		if !ok || c.CurrentPower() != want || c.CurrentToughness() != want {
			t.Errorf("%s is %d/%d, want %d/%d", c.Name, c.CurrentPower(), c.CurrentToughness(), want, want)
		}
	}
}

// Final Strike hits an opponent; you are not a legal target.
func TestFinalStrikeHitsAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := soCreature(g, me.ID, "Beast", 4, 4)
	if _, err := castWithTapParams(t, g, "Final Strike", "Sorcery", "{2}{B}{B}", finalStrikeOracle,
		game.CastSpellParams{Targets: soPlayer(me.ID), SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Fatal("Final Strike targeted its caster")
	}
	life := lifeOf(g, opp.ID)
	paCast(t, g, "Final Strike", "Sorcery", "{2}{B}{B}", finalStrikeOracle, game.CastSpellParams{
		Targets: soPlayer(opp.ID), SacrificeIDs: []uuid.UUID{fodder},
	})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-4 {
		t.Errorf("opponent's life %d → %d, want 4 damage", life, got)
	}
}

func TestRiteOfConsumptionGainsTheDamageDealt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs, mine := lifeOf(g, opp.ID), lifeOf(g, me.ID)
	paCast(t, g, "Rite of Consumption", "Sorcery", "{1}{B}", riteOfConsumptionOracle, game.CastSpellParams{
		Targets: soPlayer(opp.ID), SacrificeIDs: []uuid.UUID{soCreature(g, me.ID, "Beast", 3, 3)},
	})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != theirs-3 {
		t.Errorf("opponent's life %d → %d, want 3 damage", theirs, got)
	}
	if got := lifeOf(g, me.ID); got != mine+3 {
		t.Errorf("caster's life %d → %d, want +3", mine, got)
	}
}

func TestTormentedThoughtsDiscardsThePower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 4; i++ {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	}
	paCast(t, g, "Tormented Thoughts", "Sorcery", "{2}{B}", tormentedThoughtsOracle, game.CastSpellParams{
		Targets: soPlayer(opp.ID), SacrificeIDs: []uuid.UUID{soCreature(g, me.ID, "Beast", 3, 3)},
	})
	passPriorityAroundTable(t, g)
	c := discardChoiceFor(g, opp.ID)
	if c == nil || c.ChooseMin != 3 || c.ChooseMax != 3 {
		t.Fatalf("discard prompt = %+v, want the opponent to discard 3", c)
	}
	hand := opp.Hand.Size()
	discardFromHand(t, g, opp.ID)
	if got := hand - opp.Hand.Size(); got != 3 {
		t.Errorf("discarded %d, want 3", got)
	}
}

func TestSeveredStrandsGainsToughnessAndDestroys(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := soCreature(g, opp.ID, "Victim", 2, 2)
	life := lifeOf(g, me.ID)
	paCast(t, g, "Severed Strands", "Sorcery", "{1}{B}", severedStrandsOracle, game.CastSpellParams{
		Targets: vsTarget(victim), SacrificeIDs: []uuid.UUID{soCreature(g, me.ID, "Wall", 0, 5)},
	})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID); got != life+5 {
		t.Errorf("life %d → %d, want +5", life, got)
	}
	if onBattlefield(g, victim) {
		t.Error("the opponent's creature survived")
	}
}

// Worthy Cause with buyback returns to hand.
func TestWorthyCauseGainsToughnessWithBuyback(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := lifeOf(g, me.ID)
	id := paCast(t, g, "Worthy Cause", "Instant", "{W}", worthyCauseOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{soCreature(g, me.ID, "Beast", 1, 4)}, OptionalCosts: []int{0},
	})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID); got != life+4 {
		t.Errorf("life %d → %d, want +4", life, got)
	}
	if !me.Hand.Contains(id) {
		t.Error("the bought-back Worthy Cause is not in hand")
	}
}

func TestMorbidCuriosityDrawsTheManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	relic := paPermanent(g, me.ID, "Relic", "Artifact", "{3}{U}", 0, 0)
	paCast(t, g, "Morbid Curiosity", "Sorcery", "{1}{B}{B}", morbidCuriosityOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{relic},
	})
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("drew %d, want 4", got)
	}
}

func TestReckonersBargainGainsTheManaValueAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := lifeOf(g, me.ID)
	paCast(t, g, "Reckoner's Bargain", "Instant", "{1}{B}", reckonersBargainOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{paPermanent(g, me.ID, "Ogre", "Creature — Ogre", "{2}{R}", 3, 3)},
	})
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID); got != life+3 {
		t.Errorf("life %d → %d, want +3", life, got)
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

func TestNastyEndDrawsThreeForALegend(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		want     int
	}{
		{"Creature — Human", 2},
		{"Legendary Creature — Human", 3},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		paCast(t, g, "Nasty End", "Instant", "{1}{B}", nastyEndOracle, game.CastSpellParams{
			SacrificeIDs: []uuid.UUID{paPermanent(g, me.ID, "Someone", tc.typeLine, "{1}{B}", 2, 2)},
		})
		hand := me.Hand.Size()
		passPriorityAroundTable(t, g)
		if got := me.Hand.Size() - hand; got != tc.want {
			t.Errorf("%s: drew %d, want %d", tc.typeLine, got, tc.want)
		}
	}
}

func TestForgeArmorPutsTheManaValueInCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hero := soCreature(g, me.ID, "Hero", 2, 2)
	paCast(t, g, "Forge Armor", "Instant", "{4}{R}", forgeArmorOracle, game.CastSpellParams{
		Targets: vsTarget(hero), SacrificeIDs: []uuid.UUID{paPermanent(g, me.ID, "Relic", "Artifact", "{4}", 0, 0)},
	})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, hero, game.CounterPlusOne); got != 4 {
		t.Errorf("%d +1/+1 counters, want 4", got)
	}
}
