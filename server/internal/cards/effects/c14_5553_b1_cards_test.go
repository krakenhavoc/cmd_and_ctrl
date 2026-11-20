package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// c14_5553_b1_cards_test.go — Commander 2014, batch 1 (#5553). Every test
// plays the real catalog entry through the engine.

const (
	c14DeployToTheFrontOracle     = "135cb4af-f4b6-455b-8fea-84fcef78ee03"
	c14SpoilsOfBloodOracle        = "26516566-c837-428d-bc81-582c74da2c28"
	c14CreeperhulkOracle          = "c6423251-759a-48f1-a9ae-b49b98172c08"
	c14DemonOfWailingAgoniesOracl = "2ee46c98-aee7-459a-80fa-4ed9d8a3acda"
	c14StormsurgeKrakenOracle     = "72bdc7ba-2906-436d-8405-6a8a3335fdae"
	c14WellOfIdeasOracle          = "1ed24fc6-b1d5-47f8-a1d1-3243372e0c9f"
	c14WarmongerHellkiteOracle    = "746d4445-f28d-41af-827a-2b613b9b84ee"
	c14SiegeBehemothOracle        = "cadf7be1-4a74-4163-bb2f-6bfbbb557973"
	c14BreachingLeviathanOracle   = "7d249b16-33ef-463e-9e5d-0b84d434f093"
	c14WolfcallersHowlOracle      = "e4383685-04e2-4516-b089-863c8d309250"
	c14FleshCarverOracle          = "8fcaaf1c-c99b-4f28-bf95-7115a11e803d"
	c14InciteRebellionOracle      = "26bbc473-cf99-4a82-b432-604757a832c4"
	c14MaliciousAfflictionOracle  = "18f22cc4-e10b-4af0-9c1f-d7d15dbcce63"
	c14RavingDeadOracle           = "30e5d70e-6547-483a-a4ab-4c2eed296543"
	c14LoreseekersStoneOracle     = "28a21e56-b97f-487d-9a89-61c248192cec"
	c14DulcetSirensOracle         = "37c3cd4d-8e9a-46d4-a9c2-55f70a321eb6"
	c14FreyaliseOracle            = "659dfdbf-5e9d-4828-a967-348cb3bcf05a"
	c14DarettiScrapSavantOracle   = "ea9f4de3-fc35-435a-8dff-a5657d14a9bd"

	c14BearType = "Creature — Bear"
)

// c14Commander puts a commander the player owns and controls on the
// battlefield, which turns lieutenant on.
func c14Commander(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Commander", TypeLine: "Legendary Creature — Human",
		Power: 2, Toughness: 2, IsCommander: true, Owner: owner, Controller: owner,
	})
}

// c14Named is the first battlefield permanent `controller` controls with
// the given name, or uuid.Nil.
func c14Named(g *game.Game, controller uuid.UUID, name string) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

// c14WalkToStepOf advances the turn to `step` of `seat`'s turn, resolving
// whatever triggers on the way.
func c14WalkToStepOf(t *testing.T, g *game.Game, seat int, step game.Step) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if g.Turn.ActiveSeat == seat && g.Turn.Step == step {
			return
		}
		if !stackFullyEmpty(g) {
			passPriorityAroundTable(t, g)
			continue
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward %s of seat %d: %v", step, seat, err)
		}
	}
	t.Fatalf("never reached %s of seat %d", step, seat)
}

// c14BecomesBlocked announces that the attacker became blocked.
func c14BecomesBlocked(g *game.Game, attacker, defender uuid.UUID) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesBlocked, Actor: defender, CardID: attacker, Source: attacker})
	})
}

// --- every card is in the catalog ----------------------------------------

func TestC14B1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		c14DeployToTheFrontOracle:     "Deploy to the Front",
		c14SpoilsOfBloodOracle:        "Spoils of Blood",
		c14CreeperhulkOracle:          "Creeperhulk",
		c14DemonOfWailingAgoniesOracl: "Demon of Wailing Agonies",
		c14StormsurgeKrakenOracle:     "Stormsurge Kraken",
		c14WellOfIdeasOracle:          "Well of Ideas",
		c14WarmongerHellkiteOracle:    "Warmonger Hellkite",
		c14SiegeBehemothOracle:        "Siege Behemoth",
		c14BreachingLeviathanOracle:   "Breaching Leviathan",
		c14WolfcallersHowlOracle:      "Wolfcaller's Howl",
		c14FleshCarverOracle:          "Flesh Carver",
		c14InciteRebellionOracle:      "Incite Rebellion",
		c14MaliciousAfflictionOracle:  "Malicious Affliction",
		c14RavingDeadOracle:           "Raving Dead",
		c14LoreseekersStoneOracle:     "Loreseeker's Stone",
		c14DulcetSirensOracle:         "Dulcet Sirens",
		c14FreyaliseOracle:            "Freyalise, Llanowar's Fury",
		c14DarettiScrapSavantOracle:   "Daretti, Scrap Savant",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- printed keywords -----------------------------------------------------

func TestC14B1KeywordCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
	}{
		{"Creeperhulk", "Creature — Plant Elemental", c14CreeperhulkOracle, "trample"},
		{"Demon of Wailing Agonies", "Creature — Demon", c14DemonOfWailingAgoniesOracl, "flying"},
		{"Stormsurge Kraken", "Creature — Kraken", c14StormsurgeKrakenOracle, "hexproof"},
		{"Warmonger Hellkite", "Creature — Dragon", c14WarmongerHellkiteOracle, "flying"},
		{"Siege Behemoth", "Creature — Beast", c14SiegeBehemothOracle, "hexproof"},
		{"Flesh Carver", "Creature — Human Wizard", c14FleshCarverOracle, "intimidate"},
		{"Raving Dead", "Creature — Zombie", c14RavingDeadOracle, "deathtouch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			if !hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("%s has no %s", tc.name, tc.keyword)
			}
		})
	}
}

// --- spells ------------------------------------------------------------------

func TestC14B1DeployToTheFrontCountsEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	b12Creature(g, me.ID, "Mine", c14BearType, 2, 2)
	b12Creature(g, opp.ID, "Theirs", c14BearType, 2, 2)
	b12Creature(g, g.Seats[2].ID, "Other", c14BearType, 2, 2)
	b12Permanent(g, opp.ID, "Rock", "Artifact")
	castCatalogSpell(t, g, "Deploy to the Front", "Sorcery", c14DeployToTheFrontOracle, nil)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Soldier"); got != 3 {
		t.Errorf("%d Soldiers, want 3 (one per creature on the battlefield)", got)
	}
}

func TestC14B1SpoilsOfBloodCountsTheDead(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	b27Kill(g, b12Creature(g, me.ID, "Mine", c14BearType, 2, 2))
	b27Kill(g, b12Creature(g, opp.ID, "Theirs", c14BearType, 2, 2))
	g.RunStateChecksForTest()
	castCatalogSpell(t, g, "Spoils of Blood", "Instant", c14SpoilsOfBloodOracle, nil)
	passPriorityAroundTable(t, g)
	horror := c14Named(g, me.ID, "Horror")
	if horror == uuid.Nil {
		t.Fatal("no Horror token")
	}
	if p, tough := ptOf(t, g, horror); p != 2 || tough != 2 {
		t.Errorf("the Horror is %d/%d, want 2/2", p, tough)
	}
}

func TestC14B1InciteRebellionDamagesEachPlayerByTheirCreatures(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, other := blbB1Seats(g)
	mine := b12Creature(g, me.ID, "Mine", c14BearType, 5, 5)
	a := b12Creature(g, opp.ID, "A", c14BearType, 5, 5)
	b := b12Creature(g, opp.ID, "B", c14BearType, 5, 5)
	lives := map[uuid.UUID]int{me.ID: me.Life, opp.ID: opp.Life, other.ID: other.Life}
	castCatalogSpell(t, g, "Incite Rebellion", "Sorcery", c14InciteRebellionOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != lives[me.ID]-1 || opp.Life != lives[opp.ID]-2 || other.Life != lives[other.ID] {
		t.Errorf("life me %d opp %d other %d, want %d %d %d",
			me.Life, opp.Life, other.Life, lives[me.ID]-1, lives[opp.ID]-2, lives[other.ID])
	}
	if got := findCardAnywhere(t, g, mine).DamageMarked; got != 1 {
		t.Errorf("your creature has %d damage, want 1", got)
	}
	for _, id := range []uuid.UUID{a, b} {
		if got := findCardAnywhere(t, g, id).DamageMarked; got != 2 {
			t.Errorf("an opponent's creature has %d damage, want 2", got)
		}
	}
}

func TestC14B1MaliciousAfflictionDestroysANonblackCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	black := b12Creature(g, opp.ID, "Black", c14BearType, 2, 2)
	hmlSetColors(g, black, "B")
	green := b12Creature(g, opp.ID, "Green", c14BearType, 2, 2)
	hmlSetColors(g, green, "G")
	if err := castCatalogSpellErr(t, g, "Malicious Affliction", "Instant", c14MaliciousAfflictionOracle, cardTarget(black)); err == nil {
		t.Fatal("Malicious Affliction targeted a black creature")
	}
	castCatalogSpell(t, g, "Malicious Affliction", "Instant", c14MaliciousAfflictionOracle, cardTarget(green))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) {
		t.Error("the nonblack creature survived")
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceTriggerPrompt, top100ActiveSeat(g).ID) != nil {
		t.Error("morbid asked with no creature dead this turn")
	}
}

func TestC14B1MaliciousAfflictionMorbidCopiesItself(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	b27Kill(g, b12Creature(g, opp.ID, "Dead", c14BearType, 2, 2))
	g.RunStateChecksForTest()
	first := b12Creature(g, opp.ID, "First", c14BearType, 2, 2)
	hmlSetColors(g, first, "G")
	second := b12Creature(g, opp.ID, "Second", c14BearType, 2, 2)
	hmlSetColors(g, second, "G")
	castCatalogSpell(t, g, "Malicious Affliction", "Instant", c14MaliciousAfflictionOracle, cardTarget(first))
	answerLatestTriggerPrompt(t, g, me.ID, true)
	if latestRetarget(g, me.ID) == nil {
		t.Fatal("the copy offered no new target")
	}
	answerRetargetWith(t, g, me.ID, cardTarget(second))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(first) || g.Battlefield.Contains(second) {
		t.Error("the spell and its copy did not destroy both creatures")
	}
}

// --- creatures -------------------------------------------------------------

func TestC14B1CreeperhulkMakesYourCreatureAFiveFiveTrampler(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	hulk := b12Push(g, me.ID, "Creeperhulk", "Creature — Plant Elemental", c14CreeperhulkOracle, 5, 5)
	bear := b12Creature(g, me.ID, "Bear", c14BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", c14BearType, 2, 2)
	apaMana(me, "G", "C")
	if err := g.ActivateCatalogAbility(me.ID, hulk, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(theirs)}); err == nil {
		t.Fatal("Creeperhulk targeted a creature you don't control")
	}
	apaActivate(t, g, me, hulk, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if p, tough := ptOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the Bear is %d/%d, want 5/5", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear has no trample")
	}
}

func TestC14B1LieutenantsNeedYourCommander(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tough               int
	}{
		{"Demon of Wailing Agonies", "Creature — Demon", c14DemonOfWailingAgoniesOracl, 4, 4},
		{"Stormsurge Kraken", "Creature — Kraken", c14StormsurgeKrakenOracle, 5, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
			if p, tough := ptOf(t, g, id); p != tc.p || tough != tc.tough {
				t.Fatalf("without a commander it is %d/%d, want %d/%d", p, tough, tc.p, tc.tough)
			}
			c14Commander(g, me.ID)
			g.RunStateChecksForTest()
			if p, tough := ptOf(t, g, id); p != tc.p+2 || tough != tc.tough+2 {
				t.Errorf("with your commander it is %d/%d, want %d/%d", p, tough, tc.p+2, tc.tough+2)
			}
		})
	}
}

func TestC14B1DemonOfWailingAgoniesEdictsOnlyWithYourCommander(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	demon := b12Push(g, me.ID, "Demon of Wailing Agonies", "Creature — Demon", c14DemonOfWailingAgoniesOracl, 4, 4)
	victim := b12Creature(g, opp.ID, "Victim", c14BearType, 2, 2)
	dealCombatDamage(g, me.ID, demon, opp.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Fatal("the edict fired without your commander")
	}
	c14Commander(g, me.ID)
	g.RunStateChecksForTest()
	dealCombatDamage(g, me.ID, demon, opp.ID)
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, opp.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the damaged player did not sacrifice a creature")
	}
}

func TestC14B1StormsurgeKrakenMayDrawTwoWhenBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B", "C", "D")
	kraken := b12Push(g, me.ID, "Stormsurge Kraken", "Creature — Kraken", c14StormsurgeKrakenOracle, 5, 5)
	hand := me.Hand.Size()
	c14BecomesBlocked(g, kraken, opp.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Fatalf("drew %d without your commander", got-hand)
	}
	c14Commander(g, me.ID)
	g.RunStateChecksForTest()
	c14BecomesBlocked(g, kraken, opp.ID)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand is %d, want %d", got, hand+2)
	}
}

func TestC14B1WarmongerHellkiteMakesEveryCreatureAttack(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	b12Push(g, me.ID, "Warmonger Hellkite", "Creature — Dragon", c14WarmongerHellkiteOracle, 5, 5)
	b12Creature(g, me.ID, "Bear", c14BearType, 2, 2)
	advanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) {
		t.Fatalf("pass with every creature home = %v, want an attack requirement refusal", err)
	}
}

func TestC14B1WarmongerHellkitePumpsAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kite := b12Push(g, me.ID, "Warmonger Hellkite", "Creature — Dragon", c14WarmongerHellkiteOracle, 5, 5)
	bear := b12Creature(g, me.ID, "Bear", c14BearType, 2, 2)
	home := b12Creature(g, opp.ID, "Home", c14BearType, 2, 2)
	brAttack(t, g, kite, bear)
	b16Activate(t, g, me.ID, kite, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, bear); p != 3 {
		t.Errorf("the attacking Bear has power %d, want 3", p)
	}
	if p, _ := ptOf(t, g, home); p != 2 {
		t.Errorf("a creature that isn't attacking has power %d, want 2", p)
	}
}

func TestC14B1SiegeBehemothLetsYourCreaturesAssignAsUnblockedWhileAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	behemoth := b12Push(g, me.ID, "Siege Behemoth", "Creature — Beast", c14SiegeBehemothOracle, 7, 4)
	bear := b12Creature(g, me.ID, "Bear", c14BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", c14BearType, 2, 2)
	if apaLive(g, bear).Effective().AssignsAsThoughUnblocked {
		t.Fatal("the Bear may assign as though unblocked while the Behemoth is home")
	}
	brAttack(t, g, behemoth)
	if !apaLive(g, bear).Effective().AssignsAsThoughUnblocked || !apaLive(g, behemoth).Effective().AssignsAsThoughUnblocked {
		t.Error("your creatures may not assign as though unblocked while the Behemoth attacks")
	}
	if apaLive(g, theirs).Effective().AssignsAsThoughUnblocked {
		t.Error("an opponent's creature may assign as though unblocked")
	}
}

func TestC14B1BreachingLeviathanTapsNonblueCreaturesWhenCast(t *testing.T) {
	g := newCatalogGame(t)
	opp := top100Opponent(g)
	green := b12Creature(g, opp.ID, "Green", c14BearType, 2, 2)
	hmlSetColors(g, green, "G")
	blue := b12Creature(g, opp.ID, "Blue", "Creature — Merfolk", 2, 2)
	hmlSetColors(g, blue, "U")
	castHoldCreature(t, g, "Breaching Leviathan", "Creature — Leviathan", c14BreachingLeviathanOracle, 9, 9)
	passPriorityAroundTable(t, g)
	if !hmlTapped(t, g, green) {
		t.Error("the nonblue creature is untapped")
	}
	if hmlTapped(t, g, blue) {
		t.Error("the blue creature was tapped")
	}
}

func TestC14B1FleshCarverGrowsAndLeavesAHorror(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	carver := b12Push(g, me.ID, "Flesh Carver", "Creature — Human Wizard", c14FleshCarverOracle, 2, 2)
	fodder := b12Creature(g, me.ID, "Fodder", c14BearType, 1, 1)
	apaMana(me, "B", "C")
	if err := g.ActivateCatalogAbility(me.ID, carver, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{carver}}); err == nil {
		t.Fatal("Flesh Carver sacrificed itself to its own ability")
	}
	apaActivate(t, g, me, carver, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	if g.Battlefield.Contains(fodder) {
		t.Error("the other creature was not sacrificed")
	}
	if got := countersOn(g, carver, game.CounterPlusOne); got != 2 {
		t.Fatalf("%d +1/+1 counters, want 2", got)
	}
	b27Kill(g, carver)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	horror := c14Named(g, me.ID, "Horror")
	if horror == uuid.Nil {
		t.Fatal("no Horror token")
	}
	if p, tough := ptOf(t, g, horror); p != 4 || tough != 4 {
		t.Errorf("the Horror is %d/%d, want 4/4", p, tough)
	}
}

func TestC14B1RavingDeadMustAttackARandomOpponent(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	b12Push(g, me.ID, "Raving Dead", "Creature — Zombie", c14RavingDeadOracle, 2, 6)
	c14WalkToStepOf(t, g, seat, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	c14WalkToStepOf(t, g, seat, game.StepDeclareAttackers)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) {
		t.Fatalf("pass with the Raving Dead home = %v, want an attack requirement refusal", err)
	}
}

func TestC14B1RavingDeadHalvesTheDamagedPlayersLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := b12Push(g, me.ID, "Raving Dead", "Creature — Zombie", c14RavingDeadOracle, 2, 6)
	g.WithWriteLock(func() { opp.Life = 25 })
	dealCombatDamage(g, me.ID, dead, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != 13 {
		t.Errorf("life %d, want 13 (25 less half of it, rounded down)", opp.Life)
	}
}

func TestC14B1DulcetSirensForcesAnAttackOnTheTargetOpponent(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], top100Opponent(g)
	sirens := b12Push(g, me.ID, "Dulcet Sirens", "Creature — Siren", c14DulcetSirensOracle, 1, 3)
	bear := b12Creature(g, me.ID, "Bear", c14BearType, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, sirens, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}, playerTarget(me.ID)}}); err == nil {
		t.Fatal("the Sirens targeted their controller as the opponent")
	}
	b16Activate(t, g, me.ID, sirens, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}, playerTarget(opp.ID)}})
	passPriorityAroundTable(t, g)
	advanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != bear {
		t.Fatalf("pass with the Bear home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Bear attacking the target opponent: %v", err)
	}
}

// --- noncreature permanents -------------------------------------------------------

func TestC14B1WellOfIdeasDrawsOnEntryAndEveryDrawStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for _, p := range g.Seats {
		d2dSetHand(p, 0)
		seedLibrary(p, "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12")
	}
	advanceToMain(t, g)
	castHoldCreature(t, g, "Well of Ideas", "Enchantment", c14WellOfIdeasOracle, 0, 0)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 2 {
		t.Fatalf("hand is %d after the Well entered, want 2", got)
	}
	c14WalkToStepOf(t, g, 1, game.StepUpkeep)
	theirs := opp.Hand.Size()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep into the draw step: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - theirs; got != 2 {
		t.Errorf("the opponent drew %d in their draw step, want 2", got)
	}
	c14WalkToStepOf(t, g, 0, game.StepUpkeep)
	mine := me.Hand.Size()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep into the draw step: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - mine; got != 3 {
		t.Errorf("you drew %d in your draw step, want 3", got)
	}
}

func TestC14B1WolfcallersHowlCountsOpponentsWithFourCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Wolfcaller's Howl", "Enchantment", c14WolfcallersHowlOracle, false)
	walkToUpkeepOfSeat0(t, g)
	d2dSetHand(g.Seats[1], 4)
	d2dSetHand(g.Seats[2], 5)
	d2dSetHand(g.Seats[3], 3)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Wolf"); got != 2 {
		t.Errorf("%d Wolves, want 2", got)
	}
}

func TestC14B1LoreseekersStoneCostsMorePerCardInHand(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	seedLibrary(me, "A", "B", "C", "D")
	stone := pushCatalogPermanent(g, me.ID, "Loreseeker's Stone", "Artifact", c14LoreseekersStoneOracle, false)
	d2dSetHand(me, 2)
	apaMana(me, "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, stone, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the Stone activated for {4} with two cards in hand")
	}
	apaMana(me, "C")
	apaActivate(t, g, me, stone, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != 5 {
		t.Errorf("hand is %d, want 5", got)
	}
}

// --- planeswalkers ------------------------------------------------------------------

func TestC14B1FreyaliseMakesAManaElf(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	walker := pushWalkerForTest(g, me.ID, "Freyalise, Llanowar's Fury", c14FreyaliseOracle, 3)
	b16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	elf := c14Named(g, me.ID, "Elf Druid")
	if elf == uuid.Nil {
		t.Fatal("no Elf Druid token")
	}
	if len(game.ManaAbilitiesForCard(findCardAnywhere(t, g, elf))) == 0 {
		t.Error("the Elf Druid has no mana ability")
	}
	if got := loyaltyOf(g, walker); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestC14B1FreyaliseDestroysAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	walker := pushWalkerForTest(g, me.ID, "Freyalise, Llanowar's Fury", c14FreyaliseOracle, 3)
	bear := b12Creature(g, opp.ID, "Bear", c14BearType, 2, 2)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	if err := g.ActivateCatalogAbility(me.ID, walker, 1, game.ActivateAbilityParams{Targets: cardTarget(bear)}); err == nil {
		t.Fatal("the −2 targeted a creature")
	}
	b16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{Targets: cardTarget(rock)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if got := loyaltyOf(g, walker); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}

func TestC14B1FreyaliseDrawsForEachGreenCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	seedLibrary(me, "A", "B", "C", "D")
	walker := pushWalkerForTest(g, me.ID, "Freyalise, Llanowar's Fury", c14FreyaliseOracle, 7)
	for _, c := range []struct {
		owner uuid.UUID
		color string
	}{{me.ID, "G"}, {me.ID, "G"}, {me.ID, "R"}, {opp.ID, "G"}} {
		hmlSetColors(g, b12Creature(g, c.owner, "Beast", c14BearType, 2, 2), c.color)
	}
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

func TestC14B1DarettiRummagesUpToTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	walker := pushWalkerForTest(g, me.ID, "Daretti, Scrap Savant", c14DarettiScrapSavantOracle, 3)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	a := handCardForTest(me, "A", "Instant", "")
	b := handCardForTest(me, "B", "Instant", "")
	handCardForTest(me, "C", "Instant", "")
	seedLibrary(me, "Drawn 1", "Drawn 2", "Drawn 3")
	b16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil {
		t.Fatal("no discard prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("discard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(a) || !me.Graveyard.Contains(b) {
		t.Error("the chosen cards were not discarded")
	}
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("hand is %d, want 3 (one kept, two drawn)", got)
	}
	if got := loyaltyOf(g, walker); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestC14B1DarettiSacrificesAnArtifactToReturnOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	walker := pushWalkerForTest(g, me.ID, "Daretti, Scrap Savant", c14DarettiScrapSavantOracle, 3)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	old := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: old, Name: "Old Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	b16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{Targets: cardTarget(old)})
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("no artifact was sacrificed")
	}
	if !g.Battlefield.Contains(old) {
		t.Error("the target artifact card did not return to the battlefield")
	}
	if got := loyaltyOf(g, walker); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}

func TestC14B1DarettiEmblemReturnsYourArtifactsAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	walker := pushWalkerForTest(g, me.ID, "Daretti, Scrap Savant", c14DarettiScrapSavantOracle, 10)
	b16Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if n := len(emblemsOf(g, me.ID)); n != 1 {
		t.Fatalf("%d emblems, want 1", n)
	}
	mine := b12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	b27Kill(g, mine)
	b27Kill(g, theirs)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Fatal("the artifact came back before the end step")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Error("your artifact did not return at the end step")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("an opponent's artifact returned")
	}
}
