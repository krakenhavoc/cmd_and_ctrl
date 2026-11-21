package testso

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// otj_b1_cards_test.go — Outlaws of Thunder Junction batch 1 (#4468).
// Each test plays the real catalog entry and asserts what its text says.

const (
	otjB1AmbushGigapedeOracle     = "4faf62c3-3ba7-4540-be26-777305ef184e"
	otjB1AnkleBiterOracle         = "ee8f16ba-cbf0-4286-b26c-d78dd35bee17"
	otjB1ArmoredArmadilloOracle   = "dca25245-523a-4587-a059-a14ee2ef908d"
	otjB1BadlandsRevivalOracle    = "a61267d2-2cc7-4208-a365-581d56b92e1c"
	otjB1BanditsHaulOracle        = "84b08b55-0623-4202-98c6-1756d3c1bb8f"
	otjB1BlacksnagBuzzardOracle   = "99da512c-e222-4e5c-aa64-a276ed9ebe87"
	otjB1BloodHustlerOracle       = "bb02e64c-0c4a-41ae-972e-ce83e6fc4cb9"
	otjB1BoneyardDesecratorOracle = "c3adc698-a0c5-4c82-8347-536745c9e928"
	otjB1ConsumingAshesOracle     = "9ef676f0-53c5-4efe-9341-e960d00dbc69"
	otjB1CunningCoyoteOracle      = "a38eba6a-db90-48ac-bb49-889a7fd90ce9"
)

func TestOtjB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		otjB1AmbushGigapedeOracle:     "Ambush Gigapede",
		otjB1AnkleBiterOracle:         "Ankle Biter",
		otjB1ArmoredArmadilloOracle:   "Armored Armadillo",
		otjB1BadlandsRevivalOracle:    "Badlands Revival",
		otjB1BanditsHaulOracle:        "Bandit's Haul",
		otjB1BlacksnagBuzzardOracle:   "Blacksnag Buzzard",
		otjB1BloodHustlerOracle:       "Blood Hustler",
		otjB1BoneyardDesecratorOracle: "Boneyard Desecrator",
		otjB1ConsumingAshesOracle:     "Consuming Ashes",
		otjB1CunningCoyoteOracle:      "Cunning Coyote",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// otjB1PushCreature seeds a creature with a printed mana cost, so a
// mana-value check has something to read.
func otjB1PushCreature(g *game.Game, owner uuid.UUID, name, typeLine, manaCost string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// otjB1BoltOpponent commits a crime: the active player casts Lightning
// Bolt at an opponent and the table resolves it.
func otjB1BoltOpponent(t *testing.T, g *game.Game, opp uuid.UUID) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(opp))
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Ambush Gigapede ---------------------------------------------------------

func TestOtjB1AmbushGigapedeShrinksAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 3, 3)
	pede := effectstest.CastHoldCreature(t, g, "Ambush Gigapede", "Creature — Insect", otjB1AmbushGigapedeOracle, 6, 2)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enter trigger asked for no target")
	}
	if effectstest.HasID(p.PickTargetCards, mine) || effectstest.HasID(p.PickTargetCards, pede) {
		t.Error("the trigger offers a creature you control")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if pw, tt := effectstest.PtOf(t, g, theirs); pw != 1 || tt != 1 {
		t.Errorf("the opponent's creature is %d/%d, want 1/1", pw, tt)
	}
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, pede), "flash") {
		t.Error("Ambush Gigapede has no flash")
	}
}

func TestOtjB1AmbushGigapedeKillsATwoToughnessCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Ambush Gigapede", "Creature — Insect", otjB1AmbushGigapedeOracle, 6, 2)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("a 2/2 given -2/-2 is still on the battlefield")
	}
}

// --- Ankle Biter -------------------------------------------------------------

func TestOtjB1AnkleBiterHasDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	biter := effectstest.B12Push(g, me.ID, "Ankle Biter", "Creature — Snake", otjB1AnkleBiterOracle, 1, 1)
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, biter), "deathtouch") {
		t.Error("Ankle Biter has no deathtouch")
	}
}

// --- Armored Armadillo -------------------------------------------------------

func TestOtjB1ArmoredArmadilloGetsPlusItsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dillo := effectstest.B12Push(g, me.ID, "Armored Armadillo", "Creature — Armadillo", otjB1ArmoredArmadilloOracle, 0, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W", "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, dillo, 0, game.ActivateAbilityParams{})
	if pw, tt := effectstest.PtOf(t, g, dillo); pw != 4 || tt != 4 {
		t.Errorf("Armored Armadillo is %d/%d, want 4/4", pw, tt)
	}
	// X is read as it resolves: a +1/+1 counter makes it 5/5, and a
	// second activation adds +5/+0.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(dillo, game.CounterPlusOne, 1) })
	effectstest.B06AddMana(me, "W", "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, dillo, 0, game.ActivateAbilityParams{})
	if pw, tt := effectstest.PtOf(t, g, dillo); pw != 10 || tt != 5 {
		t.Errorf("Armored Armadillo is %d/%d, want 10/5", pw, tt)
	}
}

func TestOtjB1ArmoredArmadilloHasWardOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dillo := effectstest.B12Push(g, me.ID, "Armored Armadillo", "Creature — Armadillo", otjB1ArmoredArmadilloOracle, 0, 4)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := effectstest.CastAtWardedCreature(t, g, opp, dillo)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward {1} payment for the opponent targeting Armored Armadillo")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dillo) || !opp.Graveyard.Contains(blade) {
		t.Error("declining ward did not counter the spell")
	}
}

// --- Badlands Revival --------------------------------------------------------

func TestOtjB1BadlandsRevivalReturnsOneToTheBattlefieldAndOneToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	creature := effectstest.Batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	artifact := effectstest.Batch01GraveyardCard(me, "Broken Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Badlands Revival", "Sorcery", otjB1BadlandsRevivalOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: creature, Slot: 0},
		{Kind: game.TargetCard, ID: artifact, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(creature) {
		t.Error("the creature card did not return to the battlefield")
	}
	if !effectstest.HandHasID(me, artifact) {
		t.Error("the permanent card did not return to your hand")
	}
}

func TestOtjB1BadlandsRevivalRefusesAnInstantForTheHandClause(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	instant := effectstest.Batch01GraveyardCard(me, "Old Shock", "Instant")
	if err := effectstest.CastCatalogSpellErr(t, g, "Badlands Revival", "Sorcery", otjB1BadlandsRevivalOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: instant, Slot: 1},
	}); err == nil {
		t.Error("an instant card was accepted as a permanent card")
	}
}

func TestOtjB1BadlandsRevivalWithOnlyTheHandTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	creature := effectstest.Batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	effectstest.CastCatalogSpell(t, g, "Badlands Revival", "Sorcery", otjB1BadlandsRevivalOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: creature, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, creature) {
		t.Error("the creature card chosen for the hand clause did not go to hand")
	}
}

// --- Bandit's Haul -----------------------------------------------------------

func TestOtjB1BanditsHaulGetsALootCounterOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	haul := effectstest.B12Push(g, me.ID, "Bandit's Haul", "Artifact", otjB1BanditsHaulOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	otjB1BoltOpponent(t, g, opp.ID)
	if got := effectstest.CountersOn(g, haul, effects.OtjB1LootCounter); got != 1 {
		t.Fatalf("%d loot counters after a crime, want 1", got)
	}
	otjB1BoltOpponent(t, g, opp.ID)
	if got := effectstest.CountersOn(g, haul, effects.OtjB1LootCounter); got != 1 {
		t.Errorf("%d loot counters after a second crime this turn, want 1", got)
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AdvanceToMainOf(t, g, 0)
	otjB1BoltOpponent(t, g, opp.ID)
	if got := effectstest.CountersOn(g, haul, effects.OtjB1LootCounter); got != 2 {
		t.Errorf("%d loot counters on a new turn, want 2", got)
	}

	// {2}, {T}, Remove two loot counters: Draw a card.
	hand := me.Hand.Size()
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, haul, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if got := effectstest.CountersOn(g, haul, effects.OtjB1LootCounter); got != 0 {
		t.Errorf("%d loot counters after the draw, want 0", got)
	}
}

func TestOtjB1BanditsHaulCantDrawWithOneLootCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	haul := effectstest.B12Push(g, me.ID, "Bandit's Haul", "Artifact", otjB1BanditsHaulOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	otjB1BoltOpponent(t, g, opp.ID)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, haul, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the draw was activated with one loot counter")
	}
}

func TestOtjB1BanditsHaulIgnoresTargetingYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	haul := effectstest.B12Push(g, me.ID, "Bandit's Haul", "Artifact", otjB1BanditsHaulOracle, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, haul, effects.OtjB1LootCounter); got != 0 {
		t.Errorf("%d loot counters for targeting your own creature, want 0", got)
	}
}

// --- Blacksnag Buzzard -------------------------------------------------------

func TestOtjB1BlacksnagBuzzardNoCounterWhenNothingDied(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	buzzard := effectstest.CastHoldCreature(t, g, "Blacksnag Buzzard", "Creature — Bird", otjB1BlacksnagBuzzardOracle, 2, 1)
	if got := effectstest.PlusOnes(g, buzzard); got != 0 {
		t.Errorf("%d +1/+1 counters with no death this turn, want 0", got)
	}
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, buzzard), "flying") {
		t.Error("Blacksnag Buzzard has no flying")
	}
}

func TestOtjB1BlacksnagBuzzardCounterAfterAnyCreatureDied(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the Bolt did not kill the bear")
	}
	buzzard := effectstest.CastHoldCreature(t, g, "Blacksnag Buzzard", "Creature — Bird", otjB1BlacksnagBuzzardOracle, 2, 1)
	if got := effectstest.PlusOnes(g, buzzard); got != 1 {
		t.Errorf("%d +1/+1 counters after a creature died, want 1", got)
	}
}

func TestOtjB1BlacksnagBuzzardPlotsForOneB(t *testing.T) {
	sa := effectstest.PlotOf(otjB1BlacksnagBuzzardOracle)
	if sa == nil || sa.Cost != "{1}{B}" {
		t.Fatalf("plot: %+v, want {1}{B}", sa)
	}
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PlotHandCard(p, "Blacksnag Buzzard", "Creature — Bird", "{2}{B}", otjB1BlacksnagBuzzardOracle, 2, 1)
	effectstest.PlotAndCastLater(t, g, id, game.ManaToken{Color: "C"}, game.ManaToken{Color: "B"})
	if !g.Battlefield.Contains(id) {
		t.Fatal("the plotted Buzzard never reached the battlefield")
	}
}

// --- Blood Hustler -----------------------------------------------------------

func TestOtjB1BloodHustlerGrowsOncePerTurnFromCrimes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hustler := effectstest.B12Push(g, me.ID, "Blood Hustler", "Creature — Vampire Rogue", otjB1BloodHustlerOracle, 1, 1)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, hustler); got != 0 {
		t.Fatalf("%d counters for targeting your own creature, want 0", got)
	}
	otjB1BoltOpponent(t, g, opp.ID)
	if got := effectstest.PlusOnes(g, hustler); got != 1 {
		t.Fatalf("%d counters after a crime, want 1", got)
	}
	otjB1BoltOpponent(t, g, opp.ID)
	if got := effectstest.PlusOnes(g, hustler); got != 1 {
		t.Errorf("%d counters after a second crime this turn, want 1", got)
	}
}

func TestOtjB1BloodHustlerDrainsAndItsDrainIsACrime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hustler := effectstest.B12Push(g, me.ID, "Blood Hustler", "Creature — Vampire Rogue", otjB1BloodHustlerOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	myLife, theirLife := me.Life, opp.Life
	effectstest.B06AddMana(me, "B", "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, hustler, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetPlayer(opp.ID)})
	if me.Life != myLife+1 || opp.Life != theirLife-1 {
		t.Errorf("life %d/%d, want %d/%d", me.Life, opp.Life, myLife+1, theirLife-1)
	}
	if got := effectstest.PlusOnes(g, hustler); got != 1 {
		t.Errorf("%d counters after targeting an opponent with the drain, want 1", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, hustler, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetPlayer(me.ID)}); err == nil {
		t.Error("the drain accepted its controller as the target opponent")
	}
}

// --- Boneyard Desecrator -----------------------------------------------------

func TestOtjB1BoneyardDesecratorTreasureOnlyForAnOutlaw(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		treasure int
	}{
		{"Creature — Bear", 0},
		{"Creature — Human Rogue", 1},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		desecrator := effectstest.B12Push(g, me.ID, "Boneyard Desecrator", "Creature — Zombie Mercenary", otjB1BoneyardDesecratorOracle, 3, 4)
		fodder := effectstest.B12Creature(g, me.ID, "Fodder", tc.typeLine, 1, 1)
		effectstest.AdvanceToMain(t, g)
		effectstest.B06AddMana(me, "B", "C")
		if err := g.ActivateCatalogAbility(me.ID, desecrator, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
			t.Fatalf("%s: activate: %v", tc.typeLine, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(fodder) {
			t.Errorf("%s: the fodder was not sacrificed", tc.typeLine)
		}
		if got := effectstest.PlusOnes(g, desecrator); got != 1 {
			t.Errorf("%s: %d +1/+1 counters, want 1", tc.typeLine, got)
		}
		if got := effectstest.B16CountNamed(g, "Treasure"); got != tc.treasure {
			t.Errorf("%s: %d Treasures, want %d", tc.typeLine, got, tc.treasure)
		}
	}
}

func TestOtjB1BoneyardDesecratorCantSacrificeItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	desecrator := effectstest.B12Push(g, me.ID, "Boneyard Desecrator", "Creature — Zombie Mercenary", otjB1BoneyardDesecratorOracle, 3, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "C")
	if err := g.ActivateCatalogAbility(me.ID, desecrator, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{desecrator}}); err == nil {
		t.Error("the Desecrator sacrificed itself to its \"another creature\" cost")
	}
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, desecrator), "menace") {
		t.Error("Boneyard Desecrator has no menace")
	}
}

// --- Consuming Ashes ---------------------------------------------------------

func TestOtjB1ConsumingAshesSurveilsOnlyForManaValueThreeOrLess(t *testing.T) {
	for _, tc := range []struct {
		cost    string
		surveil bool
	}{
		{"{1}{G}{G}", true},
		{"{3}{G}{G}", false},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		victim := otjB1PushCreature(g, opp.ID, "Victim", "Creature — Beast", tc.cost, 3, 3)
		effectstest.CastCatalogSpell(t, g, "Consuming Ashes", "Instant", otjB1ConsumingAshesOracle, effectstest.B16TargetCard(victim))
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Exile.Contains(victim) {
			t.Errorf("%s: the creature was not exiled", tc.cost)
		}
		c := effectstest.SurveilChoiceFor(g, me.ID)
		if tc.surveil && (c == nil || len(c.ScryCards) != 2) {
			t.Errorf("%s: no surveil 2 prompt: %+v", tc.cost, c)
		}
		if !tc.surveil && c != nil {
			t.Errorf("%s: surveilled for a mana value of 5", tc.cost)
		}
	}
}

// --- Cunning Coyote ----------------------------------------------------------

func TestOtjB1CunningCoyotePumpsAndHastesAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	coyote := effectstest.CastHoldCreature(t, g, "Cunning Coyote", "Creature — Coyote", otjB1CunningCoyoteOracle, 2, 2)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enter trigger asked for no target")
	}
	if effectstest.HasID(p.PickTargetCards, coyote) || effectstest.HasID(p.PickTargetCards, theirs) {
		t.Errorf("the trigger offers itself or an opponent's creature: %v", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if pw, tt := effectstest.PtOf(t, g, mine); pw != 3 || tt != 3 {
		t.Errorf("your creature is %d/%d, want 3/3", pw, tt)
	}
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, mine), "haste") {
		t.Error("your creature did not gain haste")
	}
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, coyote), "haste") {
		t.Error("Cunning Coyote has no haste")
	}
}

func TestOtjB1CunningCoyotePlotsForOneR(t *testing.T) {
	sa := effectstest.PlotOf(otjB1CunningCoyoteOracle)
	if sa == nil || sa.Cost != "{1}{R}" {
		t.Fatalf("plot: %+v, want {1}{R}", sa)
	}
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[g.Turn.ActiveSeat]
	mine := effectstest.B12Creature(g, p.ID, "My Bear", "Creature — Bear", 2, 2)
	id := effectstest.PlotHandCard(p, "Cunning Coyote", "Creature — Coyote", "{1}{R}", otjB1CunningCoyoteOracle, 2, 2)
	effectstest.PlotAndCastLater(t, g, id, game.ManaToken{Color: "C"}, game.ManaToken{Color: "R"})
	if !g.Battlefield.Contains(id) {
		t.Fatal("the plotted Coyote never reached the battlefield")
	}
	if effectstest.LatestPickTarget(g, p.ID) != nil {
		effectstest.PickCard(t, g, p.ID, mine)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if pw, _ := effectstest.PtOf(t, g, mine); pw != 3 {
		t.Errorf("your creature has power %d after the free cast, want 3", pw)
	}
}
