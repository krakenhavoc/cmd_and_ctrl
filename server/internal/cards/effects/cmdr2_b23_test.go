package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	cmdr2SageOfFablesOracle     = "f53da056-97be-41ad-8cfc-0d92e17dcd7b"
	cmdr2RuinousIntrusionOracle = "4f945baa-40a9-4494-b2f9-f1926f3fec84"
	cmdr2MerfolkSkydiverOracle  = "5d2ccaa4-4815-4987-889f-d08c4d1e7a8e"
	cmdr2TenuredInkcasterOracle = "780536da-2923-48e7-916a-61c701382d28"
	cmdr2HeraldSecretOracle     = "366a218e-84d3-4cf9-bbc8-f2f8ecce92a3"
	cmdr2TributaryInstrOracle   = "1c70b9d8-b04e-48f8-a4e0-12ce077d22f2"
	cmdr2SimicAscendancyOracle  = "57c1a98e-c4fb-4144-9e3e-d0fe6789ba26"
	cmdr2PteramanderOracle      = "00187de2-bc48-4137-97d8-a9a0fafc76c1"
	cmdr2ZeganaOracle           = "04260930-ff38-4b35-9ec4-e89196a4f2c7"
	cmdr2BenthicBiomancerOracle = "b9000917-4f09-42f4-ae85-fc1ac7571d93"
)

// cmdr2B23Cast casts a creature with printed P/T so it survives entry.
func cmdr2B23Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, p, tough int) uuid.UUID {
	t.Helper()
	return avB1Cast(t, g, g.Seats[0], game.Card{Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tough}, game.CastSpellParams{})
}

func TestCmdr2SageOfFablesGivesOtherWizardsACounterAndDrawsForOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sage := b12Push(g, me.ID, "Sage of Fables", "Creature — Merfolk Wizard", cmdr2SageOfFablesOracle, 2, 2)
	wiz := cmdr2B23Cast(t, g, "Apprentice", "Creature — Human Wizard", "", 1, 1)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, wiz, "+1/+1"); got != 1 {
		t.Errorf("another Wizard entering: %d counters, want 1", got)
	}
	bear := cmdr2B23Cast(t, g, "Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, "+1/+1"); got != 0 {
		t.Errorf("a non-Wizard: %d counters, want 0", got)
	}
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Wizard", "Creature — Human Wizard", 1, 1)
	if got := counterCount(g, theirs, "+1/+1"); got != 0 {
		t.Errorf("an opponent's creature: %d counters", got)
	}
	// {2}, remove a counter from a creature you control: draw.
	hand := me.Hand.Size()
	b06AddMana(me, "R", "R")
	p7Activate(t, g, me, sage, 0, game.ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{wiz}, CounterCounts: []int{1}})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, got)
	}
	if got := counterCount(g, wiz, "+1/+1"); got != 0 {
		t.Errorf("counter not removed: %d", got)
	}
	// With no counter anywhere the cost can't be paid.
	b06AddMana(me, "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, sage, 0, game.ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{wiz}, CounterCounts: []int{1}}); err == nil {
		t.Error("activated with no counter to remove")
	}
}

func TestCmdr2SageOfFablesDoesNotCounterItself(t *testing.T) {
	g := newCatalogGame(t)
	sage := cmdr2B23Cast(t, g, "Sage of Fables", "Creature — Merfolk Wizard", cmdr2SageOfFablesOracle, 2, 2)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, sage, "+1/+1"); got != 0 {
		t.Errorf("Sage itself: %d counters, want 0 (\"other\")", got)
	}
}

func TestCmdr2RuinousIntrusionExilesAndCountersByManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big Rock", TypeLine: "Artifact",
		ManaCost: "{3}{R}", Owner: g.Seats[1].ID, Controller: g.Seats[1].ID})
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Ruinous Intrusion", "Instant", cmdr2RuinousIntrusionOracle,
		[]game.TargetRef{cardRef(rock), cardRef(bear)})
	passPriorityAroundTable(t, g)
	if z := b12ZoneOf(g, rock); z != game.ZoneExile {
		t.Errorf("artifact is in %v, want exile", z)
	}
	if got := counterCount(g, bear, "+1/+1"); got != 4 {
		t.Errorf("counters = %d, want 4 (mana value of the exiled artifact)", got)
	}
}

func TestCmdr2RuinousIntrusionIllegalCreatureStillExiles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ench := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Aura Thing", TypeLine: "Enchantment",
		ManaCost: "{2}", Owner: g.Seats[1].ID, Controller: g.Seats[1].ID})
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Ruinous Intrusion", "Instant", cmdr2RuinousIntrusionOracle,
		[]game.TargetRef{cardRef(ench), cardRef(bear)})
	b15Destroy(t, g, bear)
	passPriorityAroundTable(t, g)
	if z := b12ZoneOf(g, ench); z != game.ZoneExile {
		t.Errorf("enchantment is in %v, want exile", z)
	}
}

func TestCmdr2RuinousIntrusionRefusesWrongTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact",
		Owner: opp.ID, Controller: opp.ID})
	if err := castCatalogSpellErr(t, g, "Ruinous Intrusion", "Instant", cmdr2RuinousIntrusionOracle,
		[]game.TargetRef{cardRef(mine), cardRef(mine)}); err == nil {
		t.Error("a creature is not an artifact or enchantment")
	}
	if err := castCatalogSpellErr(t, g, "Ruinous Intrusion", "Instant", cmdr2RuinousIntrusionOracle,
		[]game.TargetRef{cardRef(rock), cardRef(theirs)}); err == nil {
		t.Error("the creature must be one you control")
	}
}

func TestCmdr2MerfolkSkydiverCountersATargetAndProliferates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ally := b12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	opp := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	sky := cmdr2B23Cast(t, g, "Merfolk Skydiver", "Creature — Merfolk Mutant", cmdr2MerfolkSkydiverOracle, 1, 1)
	passPriorityAroundTable(t, g)
	eoeABPickIfAsked(t, g, me.ID, ally)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, ally, "+1/+1"); got != 1 {
		t.Fatalf("ally counters = %d, want 1", got)
	}
	_ = opp
	b06AddMana(me, "G", "U", "R", "R", "R")
	p7Activate(t, g, me, sky, 0, game.ActivateAbilityParams{})
	settleAnsweringProliferate(t, g)
	if got := counterCount(g, ally, "+1/+1"); got != 2 {
		t.Errorf("after proliferate ally counters = %d, want 2", got)
	}
}

func TestCmdr2TenuredInkcasterCountersAndDrainsForCounteredAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Tenured Inkcaster", "Creature — Vampire Warlock", cmdr2TenuredInkcasterOracle, 2, 2)
	grown := b12Creature(g, me.ID, "Grown", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	b28AddCounter(t, g, grown, 1)
	mine, theirs := me.Life, opp.Life
	declareAttack(t, g, opp.ID, grown, plain)
	passPriorityAroundTable(t, g)
	if me.Life != mine+1 || opp.Life != theirs-1 {
		t.Errorf("life me %d->%d opp %d->%d, want +1/-1 for the one creature with a counter",
			mine, me.Life, theirs, opp.Life)
	}
}

func TestCmdr2TenuredInkcasterEntersAndCountersATarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ally := b12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	cmdr2B23Cast(t, g, "Tenured Inkcaster", "Creature — Vampire Warlock", cmdr2TenuredInkcasterOracle, 2, 2)
	passPriorityAroundTable(t, g)
	eoeABPickIfAsked(t, g, me.ID, ally)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, ally, "+1/+1"); got != 1 {
		t.Errorf("target counters = %d, want 1", got)
	}
}

func TestCmdr2HeraldOfSecretStreamsMakesCounteredCreaturesUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Herald of Secret Streams", "Creature — Merfolk Warrior", cmdr2HeraldSecretOracle, 2, 3)
	grown := b12Creature(g, me.ID, "Grown", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	b28AddCounter(t, g, grown, 1)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 8)
	brAttack(t, g, grown, plain)
	brRefusal(t, g.DeclareBlocker(blocker, grown), game.BlockReasonCantBeBlocked)
	if err := g.DeclareBlocker(blocker, plain); err != nil {
		t.Errorf("a creature with no counter can be blocked: %v", err)
	}
}

func TestCmdr2TributaryInstructorMentorsAndDrawsOnCounteredDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	inst := b12Push(g, me.ID, "Tributary Instructor", "Creature — Merfolk Shaman", cmdr2TributaryInstrOracle, 4, 4)
	small := b12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, inst, small)
	passPriorityAroundTable(t, g)
	eoeABPickIfAsked(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, small, "+1/+1"); got != 1 {
		t.Fatalf("mentor counters = %d, want 1", got)
	}
	hand := me.Hand.Size()
	b15Destroy(t, g, small)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("a creature with a counter died: hand %d -> %d, want +1", hand, got)
	}
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	hand = me.Hand.Size()
	b15Destroy(t, g, plain)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("a creature with no counter died: drew %d", got-hand)
	}
}

func TestCmdr2SimicAscendancyGrowsAndWinsAtTwenty(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	asc := b12Push(g, me.ID, "Simic Ascendancy", "Enchantment", cmdr2SimicAscendancyOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b06AddMana(me, "R", "G", "U")
	p7Activate(t, g, me, asc, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{cardRef(bear)}})
	if got := counterCount(g, bear, "+1/+1"); got != 1 {
		t.Fatalf("bear counters = %d", got)
	}
	if got := counterCount(g, asc, "growth"); got != 1 {
		t.Errorf("growth counters = %d, want 1", got)
	}
	// Counters put by anyone, in bunches, count that many.
	b28AddCounter(t, g, bear, 3)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, asc, "growth"); got != 4 {
		t.Errorf("growth counters = %d, want 4", got)
	}
	// An opponent's creature does not count.
	theirs := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	b28AddCounter(t, g, theirs, 2)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, asc, "growth"); got != 4 {
		t.Errorf("an opponent's creature grew the enchantment: %d", got)
	}
	// Twenty growth counters: the controller wins at their upkeep only.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(asc, "growth", 16) })
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if g.Outcome != nil {
		t.Fatal("won on another player's upkeep")
	}
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if g.Outcome == nil || g.Outcome.Winner != me.ID {
		t.Errorf("outcome %+v, want a win for seat 0 at twenty growth counters", g.Outcome)
	}
}

func TestCmdr2SimicAscendancyDoesNotWinBelowTwenty(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	asc := b12Push(g, me.ID, "Simic Ascendancy", "Enchantment", cmdr2SimicAscendancyOracle, 0, 0)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(asc, "growth", 19) })
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if g.Outcome != nil {
		t.Error("won with 19 growth counters")
	}
}

func TestCmdr2PteramanderAdaptsForLessWithSpellsInGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ptera := b12Push(g, me.ID, "Pteramander", "Creature — Salamander Drake", cmdr2PteramanderOracle, 1, 1)
	// The test table lets a short pool through, so the price is read off
	// what the activation takes from a pool of ten red.
	spent := func() int {
		g.WithWriteLock(func() { me.ManaPool = nil })
		for range 10 {
			b06AddMana(me, "R")
		}
		before := len(me.ManaPool.Units())
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(ptera, game.CounterPlusOne, -99) })
		p7Activate(t, g, me, ptera, 0, game.ActivateAbilityParams{})
		return before - len(me.ManaPool.Units())
	}
	if got := spent(); got != 7 {
		t.Fatalf("no instants or sorceries: generic cost paid = %d, want 7", got)
	}
	if got := counterCount(g, ptera, "+1/+1"); got != 4 {
		t.Errorf("counters = %d, want 4", got)
	}
	for range 3 {
		pushGraveyardCardTyped(me, "Bolt", "Instant")
	}
	pushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	if got := spent(); got != 4 {
		t.Errorf("three instants in the graveyard: generic cost paid = %d, want 4", got)
	}
	// Adapt does nothing while it has counters.
	g.WithWriteLock(func() { me.ManaPool = nil })
	b06AddMana(me, "U", "R", "R", "R", "R")
	p7Activate(t, g, me, ptera, 0, game.ActivateAbilityParams{})
	if got := counterCount(g, ptera, "+1/+1"); got != 4 {
		t.Errorf("a second adapt changed counters: %d", got)
	}
}

func TestCmdr2ZeganaDrawsOnlyWithAnotherCounteredCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	cmdr2B23Cast(t, g, "Zegana, Utopian Speaker", "Legendary Creature — Merfolk Wizard", cmdr2ZeganaOracle, 4, 4)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("no other counter: hand %d -> %d, want unchanged", hand, got)
	}
}

func TestCmdr2ZeganaDrawsWithAnotherCounteredCreatureAndGrantsTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	grown := b12Creature(g, me.ID, "Grown", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	b28AddCounter(t, g, grown, 1)
	hand := me.Hand.Size()
	zeg := cmdr2B23Cast(t, g, "Zegana, Utopian Speaker", "Legendary Creature — Merfolk Wizard", cmdr2ZeganaOracle, 4, 4)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want +1 (cast removes one, draw adds one)", hand, got)
	}
	if !hasEffectiveKeyword(t, g, grown, "trample") {
		t.Error("a creature with a counter has trample")
	}
	if hasEffectiveKeyword(t, g, plain, "trample") {
		t.Error("a creature without a counter has no trample")
	}
	if hasEffectiveKeyword(t, g, zeg, "trample") {
		t.Error("Zegana has no counter yet")
	}
	// Adapt 4.
	b06AddMana(me, "G", "U", "R", "R", "R", "R")
	p7Activate(t, g, me, zeg, 0, game.ActivateAbilityParams{})
	if got := counterCount(g, zeg, "+1/+1"); got != 4 {
		t.Errorf("Zegana counters = %d, want 4", got)
	}
	if !hasEffectiveKeyword(t, g, zeg, "trample") {
		t.Error("Zegana, with counters, has trample")
	}
}

func TestCmdr2BenthicBiomancerAdaptsAndLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bio := b12Push(g, me.ID, "Benthic Biomancer", "Creature — Merfolk Wizard Mutant", cmdr2BenthicBiomancerOracle, 1, 1)
	hand := me.Hand.Size()
	b06AddMana(me, "U", "R")
	p7Activate(t, g, me, bio, 0, game.ActivateAbilityParams{})
	if got := counterCount(g, bio, "+1/+1"); got != 1 {
		t.Fatalf("adapt 1: counters = %d", got)
	}
	// The loot opens a discard prompt after the draw; the draw happened.
	if got := me.Hand.Size(); got < hand {
		t.Errorf("hand shrank before the discard: %d -> %d", hand, got)
	}
}

func TestCmdr2BenthicBiomancerLootsForAnyonesCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bio := b12Push(g, me.ID, "Benthic Biomancer", "Creature — Merfolk Wizard Mutant", cmdr2BenthicBiomancerOracle, 1, 1)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	b28AddCounter(t, g, other, 1)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("a counter on another creature drew: %d -> %d", hand, got)
	}
	b28AddCounter(t, g, bio, 2)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("counters on the Biomancer: hand %d -> %d, want +1 (one trigger for the batch)", hand, got)
	}
}
