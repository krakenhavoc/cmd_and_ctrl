package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Jolene, the Plunder Queen -----------------------------------------

func sncb11Jolene(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushDiesCreatureForTest(g, owner, "Jolene, the Plunder Queen", sncb11JoleneOracle, "Legendary Creature — Human Warrior", 2, 2)
}

func TestS82SncB11JoleneGivesTheAttackerOneTreasurePerAttackOnHerControllersOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	attacker, jolene, victim := g.Seats[0], g.Seats[1], g.Seats[2]
	sncb11Jolene(g, jolene.ID)
	a := effectstest.PushVanillaCreature(g, attacker.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, attacker.ID, "Bear B", 2, 2)
	sncb11Attack(t, g, victim.ID, a, b)
	if got := effectstest.CountTreasures(g, attacker.ID); got != 1 {
		t.Errorf("the attacker has %d Treasures, want 1 (one trigger for two creatures)", got)
	}
	if got := effectstest.CountTreasures(g, jolene.ID); got != 0 {
		t.Errorf("Jolene's controller has %d Treasures, want 0", got)
	}
}

func TestS82SncB11JoleneIgnoresAnAttackOnItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	attacker, jolene := g.Seats[0], g.Seats[1]
	sncb11Jolene(g, jolene.ID)
	a := effectstest.PushVanillaCreature(g, attacker.ID, "Bear A", 2, 2)
	sncb11Attack(t, g, jolene.ID, a)
	if got := effectstest.CountTreasures(g, attacker.ID); got != 0 {
		t.Errorf("an attack on Jolene's controller made %d Treasures", got)
	}
}

func TestS82SncB11JoleneDoublesHerControllersOwnAttackTreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sncb11Jolene(g, me.ID)
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	sncb11Attack(t, g, opp.ID, a, b)
	if got := effectstest.CountTreasures(g, me.ID); got != 2 {
		t.Errorf("Treasures %d, want 2 (the attack Treasure plus the additional one)", got)
	}
}

func TestS82SncB11JoleneAddsATreasureOnlyToItsControllersCreation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sncb11Jolene(g, me.ID)
	sncb11Treasures(t, g, me.ID, 1)
	sncb11Treasures(t, g, opp.ID, 1)
	if got := effectstest.CountTreasures(g, me.ID); got != 2 {
		t.Errorf("Jolene's controller made one and has %d, want 2", got)
	}
	if got := effectstest.CountTreasures(g, opp.ID); got != 1 {
		t.Errorf("an opponent made one and has %d, want 1", got)
	}
	// A second Jolene adds another.
	sncb11Jolene(g, me.ID)
	sncb11Treasures(t, g, me.ID, 1)
	if got := effectstest.CountTreasures(g, me.ID); got != 2+3 {
		t.Errorf("with two Jolenes one Treasure made three: have %d, want 5", got)
	}
}

func TestS82SncB11JoleneSacrificesFiveTreasuresForFiveCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	jolene := sncb11Jolene(g, me.ID)
	// Make Treasures without the replacement: another table seat's, moved
	// over by hand, would be noise, so seed them directly.
	var ids []uuid.UUID
	for i := 0; i < 5; i++ {
		ids = append(ids, effectstest.PushPermanentForTest(g, me.ID, "Treasure", "", "Artifact — Treasure"))
	}
	effectstest.ToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, jolene, 0, game.ActivateAbilityParams{SacrificeIDs: ids[:4]}); err == nil {
		t.Fatal("four Treasures paid a cost of five")
	}
	if err := g.ActivateCatalogAbility(me.ID, jolene, 0, game.ActivateAbilityParams{SacrificeIDs: ids}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, jolene, game.CounterPlusOne); got != 5 {
		t.Errorf("+1/+1 counters %d, want 5", got)
	}
	if got := effectstest.CountTreasures(g, me.ID); got != 0 {
		t.Errorf("%d Treasures left, want 0", got)
	}
}

// --- Vazi, Keen Negotiator ---------------------------------------------

func TestS82SncB11VaziGivesAnOpponentAsManyTreasuresAsYouCreatedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vazi := effectstest.PushDiesCreatureForTest(g, me.ID, "Vazi, Keen Negotiator", sncb11VaziOracle, "Legendary Creature — Human Advisor", 3, 3)
	effectstest.ToMain(t, g)
	sncb11Treasures(t, g, me.ID, 3)
	sncb11Treasures(t, g, opp.ID, 5) // theirs are not mine
	if err := g.ActivateCatalogAbility(me.ID, vazi, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountTreasures(g, opp.ID); got != 5+3 {
		t.Errorf("opponent has %d Treasures, want 8", got)
	}
	if got := effectstest.CountTreasures(g, me.ID); got != 3 {
		t.Errorf("I have %d Treasures, want 3", got)
	}
}

func TestS82SncB11VaziCountsTreasuresYouSpentAndNotOnlyThoseYouStillHave(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vazi := effectstest.PushDiesCreatureForTest(g, me.ID, "Vazi, Keen Negotiator", sncb11VaziOracle, "Legendary Creature — Human Advisor", 3, 3)
	effectstest.ToMain(t, g)
	sncb11Treasures(t, g, me.ID, 2)
	for _, c := range append([]game.Card(nil), g.Battlefield.Cards...) {
		if c.Controller == me.ID && c.Name == "Treasure" {
			if err := g.SacrificePermanent(me.ID, c.InstanceID); err != nil {
				t.Fatalf("sacrifice: %v", err)
			}
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, vazi, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountTreasures(g, opp.ID); got != 2 {
		t.Errorf("opponent has %d Treasures, want 2 (created this turn, then sacrificed)", got)
	}
}

func sncb11CastWith(t *testing.T, g *game.Game, p *game.Player, kinds game.ManaSourceKinds) {
	t.Helper()
	effectstest.ToMain(t, g)
	effectstest.GmrFloatFrom(g, p, "CC", kinds)
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Test Creature", TypeLine: "Creature — Test", ManaCost: "{2}",
			Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID,
		})
	})
	if err := g.CastSpell(p.ID, id, effectstest.Strict); err != nil {
		t.Fatalf("cast: %v", err)
	}
}

func TestS82SncB11VaziTriggersOnAnOpponentsCastPaidWithTreasureMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, vaziOwner := g.Seats[0], g.Seats[1]
	effectstest.PushDiesCreatureForTest(g, vaziOwner.ID, "Vazi, Keen Negotiator", sncb11VaziOracle, "Legendary Creature — Human Advisor", 3, 3)
	bear := effectstest.PushVanillaCreature(g, vaziOwner.ID, "Bear", 2, 2)
	hand := vaziOwner.Hand.Size()
	sncb11CastWith(t, g, me, game.ManaSourceTreasure|game.ManaSourceArtifact)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters %d, want 1", got)
	}
	if got := vaziOwner.Hand.Size() - hand; got != 1 {
		t.Errorf("Vazi's controller drew %d cards, want 1", got)
	}
}

func TestS82SncB11VaziIgnoresOrdinaryManaAndItsOwnControllersSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, vaziOwner := g.Seats[0], g.Seats[1]
	effectstest.PushDiesCreatureForTest(g, vaziOwner.ID, "Vazi, Keen Negotiator", sncb11VaziOracle, "Legendary Creature — Human Advisor", 3, 3)
	bear := effectstest.PushVanillaCreature(g, vaziOwner.ID, "Bear", 2, 2)
	sncb11CastWith(t, g, me, 0)
	if len(g.PendingChoices) != 0 {
		t.Fatal("an opponent's cast with no Treasure mana triggered Vazi")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("counters %d, want 0", got)
	}

	// Vazi's own controller spending Treasure mana is not an opponent.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.PushDiesCreatureForTest(g2, me2.ID, "Vazi, Keen Negotiator", sncb11VaziOracle, "Legendary Creature — Human Advisor", 3, 3)
	bear2 := effectstest.PushVanillaCreature(g2, me2.ID, "Bear", 2, 2)
	sncb11CastWith(t, g2, me2, game.ManaSourceTreasure|game.ManaSourceArtifact)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := effectstest.CounterCount(g2, bear2, game.CounterPlusOne); got != 0 {
		t.Errorf("Vazi's controller's own Treasure cast made %d counters", got)
	}
}

func TestS82SncB11VaziTriggersOnAnOpponentsActivationPaidWithTreasureMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, vaziOwner := g.Seats[0], g.Seats[1]
	effectstest.PushDiesCreatureForTest(g, vaziOwner.ID, "Vazi, Keen Negotiator", sncb11VaziOracle, "Legendary Creature — Human Advisor", 3, 3)
	bear := effectstest.PushVanillaCreature(g, vaziOwner.ID, "Bear", 2, 2)
	fixer := effectstest.PushDiesCreatureForTest(g, me.ID, "Jetmir's Fixer", sncb11FixerOracle, "Creature — Cat Warrior", 2, 2)
	hand := vaziOwner.Hand.Size()
	effectstest.ToMain(t, g)
	effectstest.GmrFloatFrom(g, me, "RG", game.ManaSourceTreasure|game.ManaSourceArtifact)
	if err := g.ActivateCatalogAbility(me.ID, fixer, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters %d, want 1", got)
	}
	if got := vaziOwner.Hand.Size() - hand; got != 1 {
		t.Errorf("Vazi's controller drew %d cards, want 1", got)
	}
}

// --- Boxing Ring -------------------------------------------------------

func sncb11CastCreature(t *testing.T, g *game.Game, p *game.Player, cost string, power, toughness int) uuid.UUID {
	t.Helper()
	effectstest.ToMain(t, g)
	effectstest.GmrFloatFrom(g, p, "CCCCC"[:int(cost[1]-'0')], 0)
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Entering Creature", TypeLine: "Creature — Test", ManaCost: cost,
			Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
		})
	})
	if err := g.CastSpell(p.ID, id, effectstest.Strict); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func sncb11Creature(g *game.Game, owner uuid.UUID, cost string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Creature", TypeLine: "Creature — Test", ManaCost: cost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func TestS82SncB11BoxingRingFightsACreatureWithTheSameManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, me.ID, "Boxing Ring", sncb11BoxingRingOracle, "Artifact")
	same := sncb11Creature(g, opp.ID, "{2}", 2, 2)
	other := sncb11Creature(g, opp.ID, "{3}", 2, 2)
	sncb11CastCreature(t, g, me, "{2}", 5, 5)
	// The {3} creature is no legal target.
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: other}); err == nil {
		t.Fatal("a creature with a different mana value was accepted")
	}
	effectstest.AnswerPickTarget(t, g, same)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(same) {
		t.Error("the creature with the same mana value survived a 5-power fight")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("the creature with another mana value was hurt")
	}
}

func TestS82SncB11BoxingRingIgnoresYourOwnCreaturesAndMakesNoFightWithoutATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Boxing Ring", sncb11BoxingRingOracle, "Artifact")
	mine := sncb11Creature(g, me.ID, "{2}", 2, 2)
	id := sncb11CastCreature(t, g, me, "{2}", 5, 5)
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(id) {
		t.Fatal("a creature fought with no legal target")
	}
	if len(g.PendingChoices) != 0 {
		t.Error("the trigger asked for a target with none legal")
	}
}

func TestS82SncB11BoxingRingMakesATreasureOnlyAfterAFight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ring := effectstest.PushPermanentForTest(g, me.ID, "Boxing Ring", sncb11BoxingRingOracle, "Artifact")
	effectstest.ToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, ring, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the Ring made a Treasure with no creature having fought")
	}
	victim := sncb11Creature(g, opp.ID, "{2}", 1, 1)
	sncb11CastCreature(t, g, me, "{2}", 5, 5)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, ring, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("after a fight: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountTreasures(g, me.ID); got != 1 {
		t.Errorf("Treasures %d, want 1", got)
	}
}

func TestS82SncB11BoxingRingAFoughtCreatureThatLeftDoesNotCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ring := effectstest.PushPermanentForTest(g, me.ID, "Boxing Ring", sncb11BoxingRingOracle, "Artifact")
	victim := sncb11Creature(g, opp.ID, "{2}", 1, 1)
	fighter := sncb11CastCreature(t, g, me, "{2}", 5, 5)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.KillCreature(t, g, me.ID, fighter)
	if err := g.ActivateCatalogAbility(me.ID, ring, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the Ring worked although the creature that fought is gone")
	}
}
