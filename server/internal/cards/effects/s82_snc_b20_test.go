package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b20_test.go — S82 SNC batch 20 (#5491).

const falcoSparaOracle = "b9c1ca24-5696-4777-8465-13a8f78e4fb5"

func TestS82SNCB20FalcoSparaIsFull(t *testing.T) {
	spec, ok := Lookup(falcoSparaOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Falco Spara: %+v, want Full with no caveats", spec)
	}
}

// A top-of-library spell is cast by paying its mana cost and removing one
// counter from a creature you control; without a counter, without the offer
// or with a land it is refused.
func TestS82SNCB20FalcoSparaCastsFromTheTopByRemovingACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Falco Spara, Pactweaver", "Legendary Creature — Bird Demon", falcoSparaOracle, false)
	bear := b12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	bearCard := hobB10Card(me, "Top Bear", "Creature — Bear", "{1}")
	bearCard.Power, bearCard.Toughness = 2, 2
	hobB10Stack(me, bearCard)
	floatForTest(g, me, "C")

	if err := g.CastSpell(me.ID, bearCard.InstanceID, game.CastSpellParams{Strict: true, FromZone: "library"}); err == nil {
		t.Fatal("cast from the library without the offer")
	}
	one := game.CastSpellParams{
		Strict: true, FromZone: "library", AlternativeCost: falcoSparaAltCostKey,
		CounterSourceIDs: []uuid.UUID{bear}, CounterCounts: []int{1}, CounterKinds: []string{game.CounterPlusOne},
	}
	if err := g.CastSpell(me.ID, bearCard.InstanceID, one); err == nil {
		t.Fatal("cast with no counter on any creature")
	}
	findBattlefieldCardForTest(g, bear).Counters = map[string]int{game.CounterPlusOne: 2}
	if err := g.CastSpell(me.ID, bearCard.InstanceID, one); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if got := qgCounters(g, bear)[game.CounterPlusOne]; got != 1 {
		t.Errorf("+1/+1 counters left on the Bear: %d, want 1", got)
	}
	if me.ManaPool.Total() != 0 {
		t.Error("the mana cost was not paid")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bearCard.InstanceID) {
		t.Error("the spell did not resolve")
	}
}

// It may remove its own shield counter, and never plays a land.
func TestS82SNCB20FalcoSparaCanSpendItsOwnShieldAndNeverPlaysALand(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	falco := pushCatalogPermanent(g, me.ID, "Falco Spara, Pactweaver", "Legendary Creature — Bird Demon", falcoSparaOracle, false)
	findBattlefieldCardForTest(g, falco).Counters = map[string]int{game.CounterShield: 1}
	forest := hobB10Card(me, "Forest", "Basic Land — Forest", "")
	hobB10Stack(me, forest)
	if err := g.CastSpell(me.ID, forest.InstanceID, game.CastSpellParams{
		Strict: true, FromZone: "library", AlternativeCost: falcoSparaAltCostKey,
		CounterSourceIDs: []uuid.UUID{falco}, CounterCounts: []int{1}, CounterKinds: []string{game.CounterShield},
	}); err == nil {
		t.Fatal("cast a land")
	}
	spell := hobB10Card(me, "Top Rite", "Sorcery", "{1}")
	hobB10Stack(me, spell)
	floatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{
		Strict: true, FromZone: "library", AlternativeCost: falcoSparaAltCostKey,
		CounterSourceIDs: []uuid.UUID{falco}, CounterCounts: []int{1}, CounterKinds: []string{game.CounterShield},
	}); err != nil {
		t.Fatalf("cast spending its own shield counter: %v", err)
	}
	if got := qgCounters(g, falco)[game.CounterShield]; got != 0 {
		t.Errorf("shield counters left: %d, want 0", got)
	}
}

const (
	roccoCabarettiOracle = "41c765a4-a9c9-4169-a9f1-74ededeb793c"
	crypticPursuitOracle = "214cfbc8-369e-4035-9b65-80d0e0024889"
)

func TestS82SNCB20RoccoAndCrypticPursuitAreFull(t *testing.T) {
	for _, oracle := range []string{roccoCabarettiOracle, crypticPursuitOracle} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v, want Full with no caveats", oracle, spec)
		}
	}
}

// Cast with X = 2, Rocco offers only creature cards with mana value 2 or
// less, and the pick enters the battlefield.
func TestS82SNCB20RoccoSearchesForACreatureWithManaValueXOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	small := hobB10Card(me, "Small Bear", "Creature — Bear", "{1}{G}")
	big := hobB10Card(me, "Big Bear", "Creature — Bear", "{2}{G}")
	spell := hobB10Card(me, "Bolt", "Instant", "{R}")
	hobB10Stack(me, small, big, spell)
	floatForTest(g, me, "RGWCC")
	castFromHandForTest(t, g, me, "Rocco, Cabaretti Caterer", "Legendary Creature — Elf Druid", "{X}{R}{G}{W}",
		roccoCabarettiOracle, game.CastSpellParams{Strict: true, XValue: 2})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Big Bear") != uuid.Nil || searchOptionNamed(g, c, "Bolt") != uuid.Nil {
		t.Error("the search offered a card outside 'creature with mana value 2 or less'")
	}
	answerSearchNamed(t, g, me.ID, "Small Bear")
	if _, ok := battlefieldCard(g, small.InstanceID); !ok {
		t.Error("the found creature is not on the battlefield")
	}
}

// Put onto the battlefield without being cast, Rocco does not search.
func TestS82SNCB20RoccoNeedsToHaveBeenCast(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	hobB10Stack(me, hobB10Card(me, "Small Bear", "Creature — Bear", "{1}"))
	pushCatalogPermanent(g, me.ID, "Rocco, Cabaretti Caterer", "Legendary Creature — Elf Druid", roccoCabarettiOracle, false)
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c != nil {
		t.Error("a search prompt opened for a Rocco that was not cast")
	}
}

// Casting an instant from hand manifests the top card; casting a sorcery
// from the graveyard (flashback-style) or a creature does not. A dying
// face-down instant or sorcery is exiled and may be cast until the end of
// the controller's next turn; a face-down creature card stays put.
func TestS82SNCB20CrypticPursuitManifestsOnAHandCast(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Cryptic Pursuit", "Enchantment", crypticPursuitOracle, false)
	top := mlTop(me, "Hidden", "Creature — Bear", "{1}")
	floatForTest(g, me, "C")
	castFromHandForTest(t, g, me, "Shock", "Instant", "{R}", "", game.CastSpellParams{Strict: false})
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, top); !ok {
		t.Fatal("the top card was not manifested")
	}
}

func TestS82SNCB20CrypticPursuitExilesADyingFaceDownInstantForACast(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Cryptic Pursuit", "Enchantment", crypticPursuitOracle, false)
	rite := mlTop(me, "Hidden Rite", "Sorcery", "{1}")
	if _, err := g.ManifestForEffect(me.ID); err != nil {
		t.Fatal(err)
	}
	bear := mlTop(me, "Hidden Bear", "Creature — Bear", "{1}")
	if _, err := g.ManifestForEffect(me.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DestroyPermanentForEffect(bear); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(bear) {
		t.Error("the face-down creature card was exiled")
	}
	if err := g.DestroyPermanentForEffect(rite); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if !inExile(g, rite) {
		t.Fatal("the face-down sorcery was not exiled")
	}
	floatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, rite, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Error("the mana cost was not paid")
	}
}

const (
	xandersPactOracle      = "3df11e5c-19aa-441c-a889-6b40eb3fdabf"
	meetingOfTheFiveOracle = "60fc4c2d-1bf9-4c9a-a240-ab9b1c9c6817"
)

func TestS82SNCB20XandersPactAndMeetingOfTheFiveAreFull(t *testing.T) {
	for _, oracle := range []string{xandersPactOracle, meetingOfTheFiveOracle} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v, want Full with no caveats", oracle, spec)
		}
	}
}

// Each opponent exiles their own top card; the caster may cast them this
// turn paying life equal to the mana value, not mana.
func TestS82SNCB20XandersPactExilesEachOpponentsTopAndCastsThemForLife(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	var tops []game.Card
	for _, p := range g.Seats {
		if p == nil || p.ID == me.ID {
			continue
		}
		c := hobB10Card(p, "Top of "+p.Name, "Sorcery", "{2}{R}")
		hobB10Stack(p, c)
		tops = append(tops, c)
	}
	floatForTest(g, me, "BBCCCC")
	castFromHandForTest(t, g, me, "Xander's Pact", "Sorcery", "{4}{B}{B}", xandersPactOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	for _, c := range tops {
		if !inExile(g, c.InstanceID) {
			t.Fatalf("%s was not exiled", c.Name)
		}
	}
	life := me.Life
	first := tops[0].InstanceID
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Fatal("cast for its mana cost without the offer or the mana")
	}
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: xandersPactAltCostKey}); err != nil {
		t.Fatalf("cast paying life: %v", err)
	}
	if got := life - me.Life; got != 3 {
		t.Errorf("paid %d life, want 3 (the mana value)", got)
	}
}

// Ten cards are exiled, ten restricted mana are added, and the mana pays
// only for a spell with exactly three colors.
func TestS82SNCB20MeetingOfTheFiveRestrictsItsManaToThreeColorSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	three := hobB10Card(me, "Three Colors", "Creature — Bear", "{W}{U}{B}")
	plain := hobB10Card(me, "Plain", "Creature — Bear", "{1}")
	stack := []game.Card{three, plain}
	for i := 0; i < 8; i++ {
		stack = append(stack, hobB10Card(me, "Filler", "Creature — Bear", "{1}"))
	}
	// An eleventh card stays in the library.
	stack = append(stack, hobB10Card(me, "Eleventh", "Creature — Bear", "{1}"))
	hobB10Stack(me, stack...)
	floatForTest(g, me, "WUBRGCCC")
	castFromHandForTest(t, g, me, "Meeting of the Five", "Sorcery", "{3}{W}{U}{B}{R}{G}", meetingOfTheFiveOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 10 {
		t.Fatalf("mana pool holds %d, want 10", got)
	}
	if !inExile(g, three.InstanceID) {
		t.Fatal("the top ten were not exiled")
	}
	two := hobB10Card(me, "Two Colors", "Instant", "{W}{U}")
	putInHand(me, two)
	if err := g.CastSpell(me.ID, two.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the restricted mana paid for a two-color spell")
	}
	if err := g.CastSpell(me.ID, plain.InstanceID, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Fatal("cast a colorless card from exile")
	}
	if err := g.CastSpell(me.ID, three.InstanceID, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast the three-color card from exile: %v", err)
	}
	if got := me.ManaPool.Total(); got != 7 {
		t.Errorf("mana pool holds %d, want 7 after paying {W}{U}{B}", got)
	}
}
