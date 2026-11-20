package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// c16_5537_b1_cards_test.go — a batch of Commander 2016 cards (#5537).
// Every test plays the real catalog entry through the engine.

const (
	c16SublimeExhalationOracle = "9e0c66e9-7d2b-46bd-890e-886d0c029d1d"
	c16CoastalBreachOracle     = "e42b6bd7-8fdf-4452-85cc-9aed790ff682"
	c16SeedsOfRenewalOracle    = "48d63b86-8ddd-4306-960f-98490d1d80b9"
	c16TanaOracle              = "a7dff624-1f6b-4e00-b13b-0b0d69a48ec2"
	c16StonehoofOracle         = "241d00bf-3301-4385-b559-d50565603dcc"
	c16RavosOracle             = "7880576b-8ad9-4382-97c5-7a631106a35e"
	c16AkiriOracle             = "3e0a0276-b640-41c7-95bd-ba3067f5c911"
	c16KydeleOracle            = "5fca0ce6-e6ba-4319-b3c4-b4d48901e9b4"
	c16GeoscopeOracle          = "be838f57-4e74-44ad-aee1-506e63b0ff37"
	c16GraveUpheavalOracle     = "e5b43523-809a-4baf-b76d-9f59b82846b6"
	c16RunehornOracle          = "a3f2c946-409f-4093-b4db-c954d3480a72"
	c16PartingThoughtsOracle   = "57fd34d3-3e30-4467-bace-3f4dc2a85c80"
	c16EntrapmentOracle        = "2244bef9-1e4e-44ed-9b27-810f04f9e611"
	c16BoompileOracle          = "85b8902f-2515-4ded-b28a-a83334f089f1"
	c16EscalationOracle        = "f9837988-a097-4f00-a94f-4856d5d9894e"
	c16LudevicOracle           = "3b9d587c-39d2-44a2-9802-e4217ed389c0"
	c16CinderhornOracle        = "397e0d30-a3bf-4b33-98b6-b880f8de5cbb"
	c16SaskiaOracle            = "e54d207c-51b8-458e-86a1-2633ac064c8e"
	c16GripOracle              = "4c95b2fe-d107-44f9-b877-8bd4799f5a65"
	c16SpymasterOracle         = "ec08df88-3c13-4940-a090-0c502b9b12f1"

	c16Bear = "Creature — Bear"
)

// c16GraveyardCreature puts a 2/2 creature card into p's graveyard.
func c16GraveyardCreature(p *game.Player, name string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: c16Bear, Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

// --- every card is in the catalog ----------------------------------------

func TestC16B1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		c16SublimeExhalationOracle, c16CoastalBreachOracle, c16SeedsOfRenewalOracle, c16TanaOracle,
		c16StonehoofOracle, c16RavosOracle, c16AkiriOracle, c16KydeleOracle, c16GeoscopeOracle,
		c16GraveUpheavalOracle, c16RunehornOracle, c16PartingThoughtsOracle, c16EntrapmentOracle,
		c16BoompileOracle, c16EscalationOracle, c16LudevicOracle, c16CinderhornOracle,
		c16SaskiaOracle, c16GripOracle, c16SpymasterOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// --- printed keywords -----------------------------------------------------

func TestC16B1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Tana, the Bloodsower", "Legendary Creature — Elf Druid", c16TanaOracle, []string{"trample"}},
		{"Stonehoof Chieftain", "Creature — Centaur Warrior", c16StonehoofOracle, []string{"trample", "indestructible"}},
		{"Ravos, Soultender", "Legendary Creature — Human Cleric", c16RavosOracle, []string{"flying"}},
		{"Akiri, Line-Slinger", "Legendary Creature — Kor Soldier Ally", c16AkiriOracle, []string{"first strike", "vigilance"}},
		{"Runehorn Hellkite", "Creature — Dragon", c16RunehornOracle, []string{"flying"}},
		{"Charging Cinderhorn", "Creature — Elemental Ox", c16CinderhornOracle, []string{"haste"}},
		{"Saskia the Unyielding", "Legendary Creature — Human Soldier", c16SaskiaOracle, []string{"vigilance", "haste"}},
		{"Goblin Spymaster", "Creature — Goblin Rogue", c16SpymasterOracle, []string{"first strike"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 3)
			abilities := effectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !hasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// --- undaunted ---------------------------------------------------------------

func TestC16B1UndauntedCostsOneLessPerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opponents := len(g.Seats) - 1
	for _, tc := range []struct{ oracle, typeLine, cost string }{
		{c16SublimeExhalationOracle, "Sorcery", "{6}{W}"},
		{c16CoastalBreachOracle, "Sorcery", "{6}{U}"},
		{c16SeedsOfRenewalOracle, "Sorcery", "{6}{G}"},
	} {
		if got := selfPricedMV(t, g, me, tc.oracle, tc.typeLine, tc.cost); got != 7-opponents {
			t.Errorf("%s with %d opponents costs %d, want %d", tc.oracle, opponents, got, 7-opponents)
		}
	}
}

// --- the spells -----------------------------------------------------------------

func TestC16B1SublimeExhalationDestroysAllCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", c16Bear, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", c16Bear, 2, 2)
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	castCatalogSpell(t, g, "Sublime Exhalation", "Sorcery", c16SublimeExhalationOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a creature survived")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land was destroyed")
	}
}

func TestC16B1CoastalBreachBouncesEveryNonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", c16Bear, 2, 2)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	castCatalogSpell(t, g, "Coastal Breach", "Sorcery", c16CoastalBreachOracle, nil)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine) {
		t.Error("your creature did not return to your hand")
	}
	if !opp.Hand.Contains(rock) {
		t.Error("the opponent's artifact did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land was returned")
	}
}

func TestC16B1SeedsOfRenewalReturnsTwoAndExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	one := wthGraveyardCard(me, "One", "Instant")
	two := c16GraveyardCreature(me, "Two")
	spell := castCatalogSpell(t, g, "Seeds of Renewal", "Sorcery", c16SeedsOfRenewalOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: one}, {Kind: game.TargetCard, ID: two}})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(one) || !me.Hand.Contains(two) {
		t.Error("the two cards did not return to hand")
	}
	if !g.Exile.Contains(spell) || me.Graveyard.Contains(spell) {
		t.Error("Seeds of Renewal was not exiled")
	}
}

func TestC16B1GraveUpheavalReanimatesFromAnyGraveyardWithHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := wthGraveyardCard(opp, "Their Rock", "Artifact")
	if err := castCatalogSpellErr(t, g, "Grave Upheaval", "Sorcery", c16GraveUpheavalOracle, cardTarget(rock)); err == nil {
		t.Fatal("Grave Upheaval targeted a noncreature card")
	}
	dead := c16GraveyardCreature(opp, "Their Bear")
	castCatalogSpell(t, g, "Grave Upheaval", "Sorcery", c16GraveUpheavalOracle, cardTarget(dead))
	passPriorityAroundTable(t, g)
	c := findBattlefieldCardForTest(g, dead)
	if c == nil {
		t.Fatal("the creature card did not enter the battlefield")
	}
	if c.Controller != me.ID {
		t.Errorf("it entered under %v, want the caster", c.Controller)
	}
	if !hasEffectiveKeyword(t, g, dead, "haste") {
		t.Error("it has no haste")
	}
}

func TestC16B1PartingThoughtsDrawsAndLosesPerCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := apaPush(g, opp.ID, opp.ID, game.Card{
		Name: "Grown Bear", TypeLine: c16Bear, Power: 2, Toughness: 2,
		Counters: map[string]int{game.CounterPlusOne: 2, "charge": 1},
	})
	hand, life := me.Hand.Size(), me.Life
	castCatalogSpell(t, g, "Parting Thoughts", "Sorcery", c16PartingThoughtsOracle, cardTarget(victim))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature was not destroyed")
	}
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3 (one per counter of any kind)", got)
	}
	if me.Life != life-3 {
		t.Errorf("life %d, want %d", me.Life, life-3)
	}
}

func TestC16B1EntrapmentManeuverMakesSoldiersForTheToughness(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	attacker := b12Creature(g, me.ID, "Sturdy Bear", c16Bear, 2, 3)
	brAttack(t, g, attacker)
	id := b13HandCard(me, "Entrapment Maneuver", "Instant", c16EntrapmentOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{playerTarget(me.ID)}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, me.ID, attacker)
	if g.Battlefield.Contains(attacker) {
		t.Fatal("the attacking creature was not sacrificed")
	}
	if n := countOnBattlefieldByName(g, "Soldier", me.ID); n != 3 {
		t.Errorf("%d Soldiers, want 3 (the sacrificed creature's toughness)", n)
	}
}

func TestC16B1GripOfPhyresisStealsAndEquipsAGerm(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sword := pushCatalogPermanent(g, opp.ID, "Leonin Scimitar", "Artifact — Equipment", bnsB23LeoninScimitar, false)
	castCatalogSpell(t, g, "Grip of Phyresis", "Instant", c16GripOracle, cardTarget(sword))
	passPriorityAroundTable(t, g)
	c := findBattlefieldCardForTest(g, sword)
	if c == nil || c.Controller != me.ID {
		t.Fatal("you do not control the Equipment")
	}
	germ := avB1Named(g, me.ID, "Phyrexian Germ")
	if germ == uuid.Nil {
		t.Fatal("no Germ token")
	}
	if !attachedToNow(t, g, sword, germ) {
		t.Error("the Equipment is not attached to the Germ")
	}
	if p, tough := ptOf(t, g, germ); p != 1 || tough != 1 {
		t.Errorf("the Germ is %d/%d, want 1/1", p, tough)
	}
}

// --- creatures and permanents ------------------------------------------------

func TestC16B1TanaMakesSaprolingsForTheDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tana := b12Push(g, me.ID, "Tana, the Bloodsower", "Legendary Creature — Elf Druid", c16TanaOracle, 2, 2)
	dealNoncombatDamage(g, me.ID, tana, opp.ID)
	passPriorityAroundTable(t, g)
	if n := countOnBattlefieldByName(g, "Saproling", me.ID); n != 0 {
		t.Fatalf("noncombat damage made %d Saprolings", n)
	}
	dealCombatDamage(g, me.ID, tana, opp.ID)
	passPriorityAroundTable(t, g)
	if n := countOnBattlefieldByName(g, "Saproling", me.ID); n != 2 {
		t.Errorf("%d Saprolings, want 2", n)
	}
}

func TestC16B1StonehoofChieftainGrantsToOtherAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chief := b12Push(g, me.ID, "Stonehoof Chieftain", "Creature — Centaur Warrior", c16StonehoofOracle, 8, 8)
	bear := b12Creature(g, me.ID, "Bear", c16Bear, 2, 2)
	idle := b12Creature(g, me.ID, "Idle", c16Bear, 2, 2)
	declareAttack(t, g, opp.ID, chief, bear)
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"trample", "indestructible"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the attacking Bear has no %s", kw)
		}
		if hasEffectiveKeyword(t, g, idle, kw) {
			t.Errorf("a creature that did not attack has %s", kw)
		}
	}
}

func TestC16B1RavosPumpsOthersAndReturnsACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ravos := b12Push(g, me.ID, "Ravos, Soultender", "Legendary Creature — Human Cleric", c16RavosOracle, 2, 2)
	mine := b12Creature(g, me.ID, "Mine", c16Bear, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", c16Bear, 2, 2)
	if p, tough := ptOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("your other creature is %d/%d, want 3/3", p, tough)
	}
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
	if p, tough := ptOf(t, g, ravos); p != 2 || tough != 2 {
		t.Errorf("Ravos is %d/%d, want 2/2", p, tough)
	}
	dead := c16GraveyardCreature(me, "Dead Bear")
	advanceToUpkeepOfSeat(t, g, 0)
	hobAnswerTargets(t, g, me.ID, dead)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the creature card did not return to hand")
	}
}

func TestC16B1AkiriCountsYourArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	akiri := b12Push(g, me.ID, "Akiri, Line-Slinger", "Legendary Creature — Kor Soldier Ally", c16AkiriOracle, 0, 3)
	b12Permanent(g, me.ID, "Rock One", "Artifact")
	b12Permanent(g, me.ID, "Rock Two", "Artifact")
	b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if p, tough := ptOf(t, g, akiri); p != 2 || tough != 3 {
		t.Errorf("Akiri is %d/%d, want 2/3", p, tough)
	}
}

func TestC16B1KydeleAddsColorlessPerCardDrawn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kydele := pushCatalogPermanent(g, me.ID, "Kydele, Chosen of Kruphix", "Legendary Creature — Human Wizard", c16KydeleOracle, false)
	var drawn int
	g.WithWriteLock(func() {
		_ = g.DrawNForEffect(me.ID, 2)
		drawn = g.TurnTallyFor(me.ID).CardsDrawn
		me.ManaPool = nil
	})
	if drawn < 2 {
		t.Fatalf("setup: %d cards drawn this turn", drawn)
	}
	if err := g.ActivateManaAbility(me.ID, kydele, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap Kydele: %v", err)
	}
	if n := me.ManaPool.Total(); n != drawn {
		t.Errorf("pool has %d mana, want one per card drawn this turn (%d)", n, drawn)
	}
}

func TestC16B1PrismaticGeoscopeEntersTappedAndCountsBasicLandTypes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cast := castCatalogSpell(t, g, "Prismatic Geoscope", "Artifact", c16GeoscopeOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cast) || !hmlTapped(t, g, cast) {
		t.Fatal("Prismatic Geoscope did not enter tapped")
	}

	scope := pushCatalogPermanent(g, me.ID, "Prismatic Geoscope", "Artifact", c16GeoscopeOracle, false)
	pushTestPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	pushTestPermanent(g, me.ID, game.Card{Name: "Watery Grave", TypeLine: "Land — Island Swamp"})
	pushTestPermanent(g, me.ID, game.Card{Name: "Another Forest", TypeLine: "Basic Land — Forest"})
	g.WithWriteLock(func() { me.ManaPool = nil })
	if err := g.ActivateManaAbility(me.ID, scope, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Geoscope: %v", err)
	}
	b10ResolveAllManaPicks(t, g, me.ID, "R")
	if n := me.ManaPool.Total(); n != 3 {
		t.Errorf("pool has %d mana, want 3 (Forest, Island, Swamp)", n)
	}
}

func TestC16B1RunehornHellkiteWheelsFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	for _, p := range g.Seats {
		seedLibrary(p, "L1", "L2", "L3", "L4", "L5", "L6", "L7", "L8")
	}
	d2dSetHand(g.Seats[1], 2)
	id, me := activateFromGraveyard(t, g, "Runehorn Hellkite", "Creature — Dragon", c16RunehornOracle, 5, 5, "", game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Error("the Hellkite was not exiled from the graveyard")
	}
	for i, p := range g.Seats {
		if got := p.Hand.Size(); got != 7 {
			t.Errorf("seat %d has %d cards, want 7", i, got)
		}
	}
}

func TestC16B1BoompileDestroysNonlandPermanentsOnAWin(t *testing.T) {
	lost, won := 0, 0
	for table := 0; table < 40 && (lost == 0 || won == 0); table++ {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		pile := pushCatalogPermanent(g, me.ID, "Boompile", "Artifact", c16BoompileOracle, false)
		bear := b12Creature(g, opp.ID, "Bear", c16Bear, 2, 2)
		land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
		if !arcStrongboxFlip(t, g, me, pile, nil) {
			lost++
			if !g.Battlefield.Contains(pile) || !g.Battlefield.Contains(bear) {
				t.Fatal("a lost flip destroyed something")
			}
			continue
		}
		won++
		if g.Battlefield.Contains(pile) || g.Battlefield.Contains(bear) {
			t.Error("a won flip left a nonland permanent")
		}
		if !g.Battlefield.Contains(land) {
			t.Error("a won flip destroyed a land")
		}
	}
	if lost == 0 || won == 0 {
		t.Fatalf("saw %d lost and %d won flips; want both", lost, won)
	}
}

func TestC16B1EvolutionaryEscalationCountersBothTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Evolutionary Escalation", "Enchantment", c16EscalationOracle, false)
	mine := b12Creature(g, me.ID, "Mine", c16Bear, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", c16Bear, 2, 2)
	advanceToUpkeepOfSeat(t, g, 0)
	hobAnswerTargets(t, g, me.ID, mine, theirs)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, mine, game.CounterPlusOne); got != 3 {
		t.Errorf("your creature has %d counters, want 3", got)
	}
	if got := counterCount(g, theirs, game.CounterPlusOne); got != 3 {
		t.Errorf("the opponent's creature has %d counters, want 3", got)
	}
}

func TestC16B1LudevicOffersADrawOnlyAfterAnotherPlayerLostLife(t *testing.T) {
	t.Run("only you lost life", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		b12Push(g, me.ID, "Ludevic, Necro-Alchemist", "Legendary Creature — Human Wizard", c16LudevicOracle, 1, 4)
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -1) })
		advanceToEndStepOf(t, g, 0)
		passPriorityAroundTable(t, g)
		if c := latestConfirmFor(g, me.ID); c != nil {
			t.Error("offered a draw when only Ludevic's controller lost life")
		}
	})
	t.Run("an opponent lost life", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		b12Push(g, me.ID, "Ludevic, Necro-Alchemist", "Legendary Creature — Human Wizard", c16LudevicOracle, 1, 4)
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1) })
		advanceToEndStepOf(t, g, 0)
		passPriorityAroundTable(t, g)
		hand := me.Hand.Size()
		answerMayChoice(t, g, me.ID, true)
		if got := me.Hand.Size(); got != hand+1 {
			t.Errorf("hand %d, want %d", got, hand+1)
		}
	})
}

func TestC16B1ChargingCinderhornBurnsEachEndStepWithNoAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	horn := b12Push(g, me.ID, "Charging Cinderhorn", "Creature — Elemental Ox", c16CinderhornOracle, 4, 2)
	life := me.Life
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, horn, "fury"); got != 1 {
		t.Fatalf("%d fury counters, want 1", got)
	}
	if me.Life != life-1 {
		t.Errorf("the end step's player is at %d, want %d", me.Life, life-1)
	}
	oppLife := opp.Life
	advanceToEndStepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, horn, "fury"); got != 2 {
		t.Fatalf("%d fury counters, want 2", got)
	}
	if opp.Life != oppLife-2 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, oppLife-2)
	}
}

func TestC16B1ChargingCinderhornDoesNothingAfterAnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	horn := b12Push(g, me.ID, "Charging Cinderhorn", "Creature — Elemental Ox", c16CinderhornOracle, 4, 2)
	bear := b12Creature(g, me.ID, "Bear", c16Bear, 2, 2)
	declareAttack(t, g, opp.ID, bear)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, horn, "fury"); got != 0 {
		t.Errorf("%d fury counters after an attack this turn, want 0", got)
	}
}

func TestC16B1SaskiaSendsCombatDamageToTheChosenPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	castHoldCreature(t, g, "Saskia the Unyielding", "Legendary Creature — Human Soldier", c16SaskiaOracle, 3, 4)
	answerChosenPlayer(t, g, me.ID, other.ID)
	bear := b12Creature(g, me.ID, "Bear", c16Bear, 2, 2)
	life := other.Life
	dealNoncombatDamage(g, me.ID, bear, opp.ID)
	passPriorityAroundTable(t, g)
	if other.Life != life {
		t.Fatalf("noncombat damage was copied: the chosen player is at %d, want %d", other.Life, life)
	}
	dealCombatDamage(g, me.ID, bear, opp.ID)
	passPriorityAroundTable(t, g)
	if other.Life != life-2 {
		t.Errorf("the chosen player is at %d, want %d", other.Life, life-2)
	}
}

func TestC16B1GoblinSpymasterGivesEachOpponentAGoblin(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Goblin Spymaster", "Creature — Goblin Rogue", c16SpymasterOracle, 2, 1)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := len(tokensNamed(g, "Goblin")); n != 0 {
		t.Fatalf("your own end step made %d Goblins", n)
	}
	advanceToEndStepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	toks := tokensNamed(g, "Goblin")
	if len(toks) != 1 || toks[0].Controller != opp.ID {
		t.Fatalf("Goblins after the opponent's end step: %+v, want one the opponent controls", toks)
	}
	theirs := b12Creature(g, opp.ID, "Their Bear", c16Bear, 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", c16Bear, 2, 2)
	var theirReqs, myReqs int
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		theirReqs = len(findBattlefieldCardForTest(g, theirs).Effective().AttackRequirements)
		myReqs = len(findBattlefieldCardForTest(g, mine).Effective().AttackRequirements)
	})
	if theirReqs == 0 {
		t.Error("the token's controller's creature has no attack requirement")
	}
	if myReqs != 0 {
		t.Errorf("your creature has %d attack requirements, want none", myReqs)
	}
}
