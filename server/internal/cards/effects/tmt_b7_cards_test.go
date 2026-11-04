package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmt_b7_cards_test.go — Teenage Mutant Ninja Turtles, batch 7 (#4457).
// Each test plays the real catalog entry and asserts what its text says.

const (
	tmtB7TCRIOracle     = "4139f40a-873b-4842-96a2-b7875de1c20a"
	tmtB7TreatsOracle   = "60826714-b9fd-4c89-b9f2-1a3c3229686b"
	tmtB7TenderOracle   = "a03fa584-ceb5-410c-9b80-f75d00215537"
	tmtB7RatsOracle     = "0a8e5c48-885d-4ea8-9c68-a3e9213d4289"
	tmtB7BlimpOracle    = "bb8c6dd2-1abe-4eb2-be1f-b0c078e5f257"
	tmtB7LairOracle     = "eb002bbc-08df-4bf0-bea3-46494ad261b6"
	tmtB7AllianceOracle = "989b93fe-cad0-4890-a228-8a20ca666f10"
	tmtB7UtromOracle    = "89793c8c-98a3-4621-ad3d-cfc5949c65da"
	tmtB7VenusOracle    = "7b3d364d-6581-43de-a80a-d5b07f5d21ed"
	tmtB7WestWindOracle = "92e45a8f-c582-45e1-9b7f-def219023ebf"
	tmtB7WingnutOracle  = "392a71c7-8f6d-46b8-bcc3-613e086d86ed"
	tmtB7ZogOracle      = "d944d33d-0a95-4d0d-96ce-49d02500227d"
)

func TestTmtB7EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tmtB7TCRIOracle:     "TCRI Building",
		tmtB7TreatsOracle:   "Tainted Treats",
		tmtB7TenderOracle:   "Tenderize",
		tmtB7RatsOracle:     "Tunnel Rats",
		tmtB7BlimpOracle:    "Turtle Blimp",
		tmtB7LairOracle:     "Turtle Lair",
		tmtB7AllianceOracle: "Uneasy Alliance",
		tmtB7UtromOracle:    "Utrom Scientists",
		tmtB7VenusOracle:    "Venus, Torn Between Worlds",
		tmtB7WestWindOracle: "West Wind Avatar",
		tmtB7WingnutOracle:  "Wingnut, Bat on the Belfry",
		tmtB7ZogOracle:      "Zog, Triceraton Castaway",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- TCRI Building -------------------------------------------------------------

func TestTmtB7TCRIBuildingEntersTappedGainsOneAndTapsForUOrR(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Life
	id := playLandFromHand(t, g, "TCRI Building", tmtB7TCRIOracle)
	top100AssertEnteredTapped(t, g, id, "TCRI Building")
	if triggerOnStack(g, id) == nil {
		t.Error("the life is a trigger with a response window")
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d -> %d, want +1", before, me.Life)
	}
	spec, _ := Lookup(tmtB7TCRIOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{U|R}" {
		t.Errorf("mana abilities %+v, want one producing {U|R}", spec.ManaAbilities)
	}
}

// --- Tainted Treats ------------------------------------------------------------

func TestTmtB7TaintedTreatsDestroysAndMakesFoodForASmallTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b27Push(g, opp.ID, "Their Rock", "Artifact", "", "{4}", 0, 0)
	castCatalogSpell(t, g, "Tainted Treats", "Instant", tmtB7TreatsOracle, s58Refs(rock))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if got := tmtB1Count(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food, want 1 for a mana value 4 target", got)
	}
}

func TestTmtB7TaintedTreatsMakesNoFoodForABigTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := b27Push(g, opp.ID, "Their Giant", "Creature — Giant", "", "{3}{G}{G}", 5, 5, "G")
	castCatalogSpell(t, g, "Tainted Treats", "Instant", tmtB7TreatsOracle, s58Refs(big))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the creature survived")
	}
	if got := tmtB1Count(g, me.ID, "Food"); got != 0 {
		t.Errorf("%d Food for a mana value 5 target, want 0", got)
	}
}

func TestTmtB7TaintedTreatsRefusesAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	aura := b27Push(g, opp.ID, "Their Charm", "Enchantment", "", "{1}", 0, 0)
	if err := castCatalogSpellErr(t, g, "Tainted Treats", "Instant", tmtB7TreatsOracle, s58Refs(aura)); err != game.ErrIllegalTarget {
		t.Errorf("an enchantment as the target: err = %v, want ErrIllegalTarget", err)
	}
}

// --- Tenderize -------------------------------------------------------------------

func TestTmtB7TenderizeIsOneSided(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Beast", "Creature — Beast", 4, 4)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 4)
	castCatalogSpell(t, g, "Tenderize", "Instant", tmtB7TenderOracle, s58Refs(mine, theirs))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("4 damage should kill the 3/4")
	}
	if damageMarkedOn(g, mine) != 0 {
		t.Error("the Beast deals the damage and takes none")
	}
}

func TestTmtB7TenderizeRefusesYourOwnCreatureAsTheVictim(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := b12Creature(g, me.ID, "My Beast", "Creature — Beast", 4, 4)
	other := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Tenderize", "Instant", tmtB7TenderOracle, s58Refs(mine, other)); err != game.ErrIllegalTarget {
		t.Errorf("your own creature as the victim: err = %v, want ErrIllegalTarget", err)
	}
}

// --- Tunnel Rats -----------------------------------------------------------------

func TestTmtB7TunnelRatsReturnsFromTheGraveyardTapped(t *testing.T) {
	g := newCatalogGame(t)
	id, me := activateFromGraveyard(t, g, "Tunnel Rats", "Creature — Rat", tmtB7RatsOracle, 2, 2, "{B}{B}{B}{B}{B}", game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) || me.Graveyard.Contains(id) {
		t.Fatal("the Rats did not return to the battlefield")
	}
	if c := tmtB1Card(t, g, id); !c.Tapped {
		t.Error("the Rats must enter tapped")
	}
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events: the Rats enter tapped rather than being tapped", n)
	}
}

func TestTmtB7TunnelRatsCostsFourAndBlack(t *testing.T) {
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: tmtB7RatsOracle})
	if len(abilities) != 1 || abilities[0].Cost.Mana != "{4}{B}" {
		t.Fatalf("abilities %+v, want one costing {4}{B}", abilities)
	}
	if !game.AbilityFunctionsFromZone(abilities[0], game.ZoneGraveyard) || game.AbilityFunctionsFromZone(abilities[0], game.ZoneBattlefield) {
		t.Error("the ability works from the graveyard only")
	}
}

// --- Turtle Blimp ----------------------------------------------------------------

func TestTmtB7TurtleBlimpMakesAMutantThatCanCrewIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	blimp := castCatalogSpell(t, g, "Turtle Blimp", "Artifact — Vehicle", tmtB7BlimpOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	mutants := tmtB1Named(g, "Mutant")
	if len(mutants) != 1 || tmtB1Count(g, me.ID, "Mutant") != 1 {
		t.Fatalf("Mutants %v, want one made for the caster", mutants)
	}
	if c := effectiveCardForTest(t, g, blimp); c.IsCreature() || !effectiveHasKeyword(&c, "flying") {
		t.Errorf("an uncrewed Blimp: creature %v, flying %v; want a flying noncreature", c.IsCreature(), effectiveHasKeyword(&c, "flying"))
	}
	crewBy(t, g, me.ID, blimp, mutants[0])
	passPriorityAroundTable(t, g)
	if c := effectiveCardForTest(t, g, blimp); !c.IsCreature() {
		t.Error("crewed by the 2/2 Mutant, the Blimp is not a creature")
	}
}

// --- Turtle Lair -----------------------------------------------------------------

func TestTmtB7TurtleLairManaIsForNinjaAndTurtleSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lair := pushCatalogPermanent(g, me.ID, "Turtle Lair", "Land", tmtB7LairOracle, false)

	if err := g.ActivateManaAbility(me.ID, lair, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility {C}: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "C" || len(me.ManaPool.Units()[0].Restrictions) != 0 {
		t.Fatalf("pool %v, want one unrestricted {C}", got)
	}
	me.ManaPool = game.ManaPool{}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(lair) })

	if err := g.ActivateManaAbility(me.ID, lair, 1, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("ActivateManaAbility any colour: %v", err)
	}
	cost, _ := game.ParseCost("{G}")
	ninja := game.Card{Name: "Test Ninja", TypeLine: "Creature — Human Ninja", ManaCost: "{G}"}
	turtle := game.Card{Name: "Test Turtle", TypeLine: "Creature — Mutant Ninja Turtle", ManaCost: "{G}"}
	bear := game.Card{Name: "Test Bear", TypeLine: "Creature — Bear", ManaCost: "{G}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(ninja)) || !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(turtle)) {
		t.Error("the mana was refused for a Ninja or a Turtle spell")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) {
		t.Error("the mana paid for a Bear spell: stronger than printed")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(ninja)) {
		t.Error("the mana paid for an ability: stronger than printed")
	}
}

func TestTmtB7TurtleLairMakesANinjaOrTurtleUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lair := pushCatalogPermanent(g, me.ID, "Turtle Lair", "Land", tmtB7LairOracle, false)
	turtle := b12Creature(g, me.ID, "My Turtle", "Creature — Turtle", 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, lair, 0, game.ActivateAbilityParams{Targets: s58Refs(bear)}); err == nil {
		t.Error("a Bear is not a Ninja or a Turtle")
	}
	b16Activate(t, g, me.ID, lair, 0, game.ActivateAbilityParams{Targets: s58Refs(turtle)})
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, turtle, game.CantBeBlocked)
	assertRestrictions(t, g, bear, 0)
}

// --- Uneasy Alliance -------------------------------------------------------------

func tmtB7AllianceOn(g *game.Game, owner, host uuid.UUID) uuid.UUID {
	aura := pushCatalogPermanent(g, owner, "Uneasy Alliance", "Enchantment — Aura", tmtB7AllianceOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == aura {
				g.Battlefield.Cards[i].AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: host}
			}
		}
	})
	return aura
}

func TestTmtB7UneasyAllianceStopsAttackingAndBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	tmtB7AllianceOn(g, me.ID, victim)
	assertRestrictions(t, g, victim, game.CantAttackOrBlock)
}

func TestTmtB7UneasyAllianceExilesTheCreatureAndMakesANinjaAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	aura := tmtB7AllianceOn(g, me.ID, victim)

	advanceTo(t, g, game.StepBeginCombat)
	floatMana(t, g, me, "{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated outside a main phase: it is a sorcery-speed ability")
	}
	advanceTo(t, g, game.StepPostcombatMain)
	floatMana(t, g, me, "{C}{C}{C}{C}{C}")
	b16Activate(t, g, me.ID, aura, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(victim) {
		t.Error("the enchanted creature was not exiled")
	}
	if got := tmtB1Count(g, me.ID, "Ninja"); got != 1 {
		t.Fatalf("%d Ninja tokens for the activator, want 1", got)
	}
	ninja := tmtB1Card(t, g, tmtB1Named(g, "Ninja")[0])
	if ninja.Power != 1 || ninja.Toughness != 1 || len(ninja.Colors) != 1 || ninja.Colors[0] != "B" {
		t.Errorf("Ninja = %d/%d %v, want a black 1/1", ninja.Power, ninja.Toughness, ninja.Colors)
	}
}

// --- Utrom Scientists ------------------------------------------------------------

func TestTmtB7UtromScientistsTapsAndStunsUpToOneCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	other := b12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Utrom Scientists", "Artifact Creature — Utrom Robot Scientist", tmtB7UtromOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	tc := tmtB1Card(t, g, theirs)
	if !tc.Tapped || tc.Counters[game.CounterStun] != 1 {
		t.Errorf("target tapped=%v stun=%d, want tapped with one stun counter", tc.Tapped, tc.Counters[game.CounterStun])
	}
	if oc := tmtB1Card(t, g, other); oc.Tapped || oc.Counters[game.CounterStun] != 0 {
		t.Error("the other creature was affected")
	}
}

// --- Venus, Torn Between Worlds --------------------------------------------------

func TestTmtB7VenusGrowsByTheDamageSheIsDealt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	venus := b27Push(g, me.ID, "Venus, Torn Between Worlds", "Legendary Creature — Mutant Frog Turtle", tmtB7VenusOracle, "{4}{G}", 5, 5, "G")
	pinger := b12Creature(g, opp.ID, "Pinger", "Creature — Test", 1, 9)
	b27Damage(g, pinger, venus, 3)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, venus); got != 3 {
		t.Errorf("%d +1/+1 counters after 3 damage, want 3", got)
	}
	other := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 4)
	b27Damage(g, pinger, other, 2)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if plusOnes(g, other) != 0 || plusOnes(g, venus) != 3 {
		t.Error("damage to another creature grew something")
	}
}

func TestTmtB7VenusPaysUToDrawForACreatureWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Venus, Torn Between Worlds", "Legendary Creature — Mutant Frog Turtle", tmtB7VenusOracle, "{4}{G}", 5, 5, "G")
	countered := b12Creature(g, me.ID, "Stunned Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(countered, game.CounterStun, 1) })
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(theirs, game.CounterPlusOne, 1) })

	dealCombatDamageToPlayer(g, plain, opp.ID, 2)
	dealCombatDamageToPlayer(g, theirs, me.ID, 3)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Fatal("a creature with no counter, or an opponent's creature, set Venus off")
	}

	before := me.Hand.Size()
	dealCombatDamageToPlayer(g, countered, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("a creature with a stun counter connecting should offer {U} to draw")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	answerPayUnless(t, g, me.ID, true)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("paying {U} drew %d, want 1", got)
	}
}

// --- West Wind Avatar ------------------------------------------------------------

func TestTmtB7WestWindAvatarSacrificesATokenOrLandForLifeThenDisappearDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 black Rat"), 1) })
	rat := findBattlefieldByName(g, "Rat")
	life := me.Life

	castCatalogSpell(t, g, "West Wind Avatar", "Creature — Cat Spirit Avatar", tmtB7WestWindOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	prompt := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 || hasID(prompt.ChooseCards, bear) || !hasID(prompt.ChooseCards, rat) || !hasID(prompt.ChooseCards, land) {
		t.Fatalf("want a may-sacrifice offering the token and the land but not the Bear: %+v", prompt)
	}
	answerOwnPermanents(t, g, me.ID, rat)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rat) {
		t.Error("the chosen token was not sacrificed")
	}
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3", life, me.Life)
	}

	hand := me.Hand.Size()
	advanceToStepOf(t, g, 0, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("disappear drew %d after the token left, want 1", got)
	}
}

func TestTmtB7WestWindAvatarDisappearNeedsAPermanentToLeave(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	avatar := b27Push(g, me.ID, "West Wind Avatar", "Creature — Cat Spirit Avatar", tmtB7WestWindOracle, "{5}{G}{G}", 7, 7, "G")
	hand := me.Hand.Size()
	advanceToStepOf(t, g, 0, game.StepEnd)
	if triggerOnStack(g, avatar) != nil {
		t.Error("disappear triggered with nothing having left the battlefield")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("drew a card with nothing having left the battlefield")
	}
}

// --- Wingnut, Bat on the Belfry --------------------------------------------------

func TestTmtB7WingnutAllianceGainsTheChosenKeyword(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wingnut := b27Push(g, me.ID, "Wingnut, Bat on the Belfry", "Legendary Creature — Bat Mutant", tmtB7WingnutOracle, "{1}{R}", 1, 2, "R")

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TokenCard("1/1 black Rat"), 1) })
	passPriorityAroundTable(t, g)
	if latestOptionPickFor(g, me.ID) != nil {
		t.Fatal("an opponent's creature entering set off alliance")
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 black Rat"), 1) })
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, wingnut, "menace")
	if c := effectiveCardForTest(t, g, wingnut); effectiveHasKeyword(&c, "flying") || effectiveHasKeyword(&c, "haste") {
		t.Error("Wingnut gained a keyword that was not chosen")
	}
}

func TestTmtB7WingnutAttackPumpsEachOtherAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wingnut := b27Push(g, me.ID, "Wingnut, Bat on the Belfry", "Legendary Creature — Bat Mutant", tmtB7WingnutOracle, "{1}{R}", 1, 2, "R")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	idle := b12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{wingnut, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("attacking Bear power %d, want 3", got)
	}
	if got := effectivePower(t, g, wingnut); got != 1 {
		t.Errorf("Wingnut power %d, want 1 (only the others get +1/+0)", got)
	}
	if got := effectivePower(t, g, idle); got != 2 {
		t.Errorf("non-attacking Bear power %d, want 2", got)
	}
}

// --- Zog, Triceraton Castaway ----------------------------------------------------

func TestTmtB7ZogStopsATargetFromBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	other := b12Creature(g, opp.ID, "Other Wall", "Creature — Wall", 0, 4)
	zog := castCatalogSpell(t, g, "Zog, Triceraton Castaway", "Legendary Creature — Dinosaur Soldier", tmtB7ZogOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, wall)
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, zog, "reach", "trample")
	assertRestrictions(t, g, wall, game.CantBlock)
	assertRestrictions(t, g, other, 0)
}

func TestTmtB7ZogMountaincyclesForTwo(t *testing.T) {
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: tmtB7ZogOracle})
	if len(abilities) != 1 || abilities[0].Cost.Mana != "{2}" || !game.AbilityFunctionsFromZone(abilities[0], game.ZoneHand) {
		t.Fatalf("abilities %+v, want one {2} ability from the hand", abilities)
	}
	g := newCatalogGame(t)
	ids := seedSearchLibrary(g.Seats[g.Turn.ActiveSeat],
		searchTestLand("Mountain", "Basic Land — Mountain"),
		searchTestLand("Island", "Basic Land — Island"),
	)
	id, me := cycleFromHand(t, g, "Zog, Triceraton Castaway", "Legendary Creature — Dinosaur Soldier", tmtB7ZogOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled Zog is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != ids[0] {
		t.Fatalf("search %+v, want only the Mountain offered", c)
	}
}
