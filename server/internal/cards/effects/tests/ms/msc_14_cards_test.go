package testsms

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_14_cards_test.go — Marvel Super Heroes Commander, fourteenth batch
// (#4455). Each test plays the real catalog entry and asserts what its text
// says.

const (
	msc14AgeOfUltronOracle    = "51042581-4add-4487-a1b6-85446539b975"
	msc14HulkAngryOracle      = "495054c8-d21d-4b89-9853-bb299d30ef63"
	msc14HulkStrongestOracle  = "35daea08-655d-4588-8430-b078769d22b6"
	msc14HulkbusterOracle     = "abc3f55e-5fc6-44b3-a2f5-b8f49d93249f"
	msc14HulklingOracle       = "c92f20c4-838d-43a1-8a71-dae205789d94"
	msc14ImmortusOracle       = "8952a7e6-4ede-42bb-b90c-00ea4f24f8f8"
	msc14ImpossibleManOracle  = "9be1d4b1-064d-4051-aafc-ca48def28e3b"
	msc14IronFistOracle       = "58c8bf6a-1b13-4c90-9e0b-24fb687f1e13"
	msc14IronManArmoredOracle = "c81195d4-19c6-4cac-b312-7f5c33dff962"
	msc14IronManBleedOracle   = "fc6e170c-820c-48bb-95c7-ee79fba634de"
	msc14IronManFuturistOracl = "eabe9a02-ea47-4548-8933-e5d1269fab25"
	msc14IronManModernOracle  = "507fc71d-0dfc-4bfe-bc71-de88162c39a1"
	msc14FrogsOracle          = "36396a14-9d0b-4670-bf8a-5c346fdda197"
	msc14LivingLightningOracl = "b9397684-f226-4f77-a57a-88f97fbd96a5"
)

func TestMsc14EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		msc14AgeOfUltronOracle:    "Age of Ultron",
		msc14HulkAngryOracle:      "Hulk, Always Angry",
		msc14HulkStrongestOracle:  "Hulk, Strongest There Is",
		msc14HulkbusterOracle:     "Hulkbuster Armor",
		msc14HulklingOracle:       "Hulkling, Young Avenger",
		msc14ImmortusOracle:       "Immortus, Master of Eternity",
		msc14ImpossibleManOracle:  "Impossible Man",
		msc14IronFistOracle:       "Iron Fist, Hero for Hire",
		msc14IronManArmoredOracle: "Iron Man, Armored Avenger",
		msc14IronManBleedOracle:   "Iron Man, Bleeding Edge",
		msc14IronManFuturistOracl: "Iron Man, Futurist Paragon",
		msc14IronManModernOracle:  "Iron Man, Modern Marvel",
		msc14FrogsOracle:          "King Solomon's Frogs",
		msc14LivingLightningOracl: "Living Lightning, Charged Up",
	}
	for oracle, name := range want {
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

// msc14Permanent seeds a permanent with a mana cost.
func msc14Permanent(g *game.Game, owner uuid.UUID, name, typeLine, cost string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost,
		Owner: owner, Controller: owner,
	})
}

// msc14WithCounters seeds a creature that already carries +1/+1 counters.
func msc14WithCounters(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough, counters int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, Power: p, Toughness: tough,
		Counters: map[string]int{game.CounterPlusOne: counters}, Owner: owner, Controller: owner,
	})
}

// --- Hulk, Always Angry -------------------------------------------------------

func TestMsc14HulkAlwaysAngryDestroysEveryArtifactWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	golem := effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 3, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	aura := effectstest.B12Permanent(g, opp.ID, "Their Pact", "Enchantment")

	hulk := effectstest.CastHoldCreature(t, g, "Hulk, Always Angry", "Legendary Creature — Gamma Berserker Hero", msc14HulkAngryOracle, 9, 9)
	for _, id := range []uuid.UUID{mine, theirs, golem} {
		if g.Battlefield.Contains(id) {
			t.Errorf("artifact %s survived", id)
		}
	}
	for name, id := range map[string]uuid.UUID{"a creature": bear, "an enchantment": aura, "Hulk": hulk} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was destroyed", name)
		}
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, hulk, "trample") {
		t.Error("Hulk lacks trample")
	}
}

func TestMsc14HulkAlwaysAngryMustAttackEachCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	hulk := effectstest.B12Push(g, g.Seats[0].ID, "Hulk, Always Angry", "Legendary Creature — Gamma Berserker Hero", msc14HulkAngryOracle, 9, 9)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("passing with Hulk at home = %v, want ErrAttackRequirement", err)
	}
	if err := g.DeclareAttacker(hulk, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); errors.Is(err, game.ErrAttackRequirement) {
		t.Errorf("passing after Hulk attacked still refused: %v", err)
	}
}

// --- Hulk, Strongest There Is -------------------------------------------------

func TestMsc14HulkStrongestEntersWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	hulk := effectstest.CastHoldCreature(t, g, "Hulk, Strongest There Is", "Legendary Creature — Gamma Berserker Hero", msc14HulkStrongestOracle, 3, 3)
	if got := effectstest.PlusOnes(g, hulk); got != 1 {
		t.Fatalf("+1/+1 counters on entering: %d, want 1", got)
	}
	if got := effectstest.EffectivePower(t, g, hulk) + effectstest.PlusOnes(g, hulk); got != 4 {
		t.Errorf("power with its counter %d, want 4", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, hulk, "trample") {
		t.Error("Hulk lacks trample")
	}
}

func TestMsc14HulkStrongestDoublesCountersOnEachOfYourGammaCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hulk := msc14WithCounters(g, me.ID, "Hulk", "Legendary Creature — Gamma Berserker Hero", msc14HulkStrongestOracle, 3, 3, 2)
	other := msc14WithCounters(g, me.ID, "Other Gamma", "Creature — Gamma Hero", "", 2, 2, 3)
	bear := msc14WithCounters(g, me.ID, "Bear", "Creature — Bear", "", 2, 2, 2)
	theirs := msc14WithCounters(g, opp.ID, "Their Gamma", "Creature — Gamma Hero", "", 2, 2, 2)
	bare := effectstest.B12Creature(g, me.ID, "Bare Gamma", "Creature — Gamma", 1, 1)

	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)

	for name, c := range map[string]struct {
		id   uuid.UUID
		want int
	}{
		"Hulk (2 doubled)":          {hulk, 4},
		"another Gamma (3 doubled)": {other, 6},
		"a non-Gamma":               {bear, 2},
		"an opponent's Gamma":       {theirs, 2},
		"a Gamma without counters":  {bare, 0},
	} {
		if got := effectstest.PlusOnes(g, c.id); got != c.want {
			t.Errorf("%s has %d +1/+1 counters, want %d", name, got, c.want)
		}
	}
}

// --- Iron Man, Modern Marvel --------------------------------------------------

func TestMsc14IronManModernMarvelPumpsOtherArtifactCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	iron := effectstest.B12Push(g, me.ID, "Iron Man, Modern Marvel", "Legendary Artifact Creature — Human Hero", msc14IronManModernOracle, 3, 3)
	golem := effectstest.B12Creature(g, me.ID, "My Golem", "Artifact Creature — Golem", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")

	for name, c := range map[string]struct {
		id   uuid.UUID
		want int
	}{"another artifact creature": {golem, 3}, "a non-artifact creature": {bear, 2}, "an opponent's artifact creature": {theirs, 2}, "Iron Man himself": {iron, 3}} {
		if got := effectstest.EffectivePower(t, g, c.id); got != c.want {
			t.Errorf("%s has power %d, want %d", name, got, c.want)
		}
	}
	if got := effectstest.EffectiveToughness(t, g, golem); got != 3 {
		t.Errorf("another artifact creature has toughness %d, want 3", got)
	}
	if !g.Battlefield.Contains(rock) || !effectstest.HasKeywordOnBattlefield(t, g, iron, "flying") {
		t.Error("Iron Man lacks flying")
	}
}

func TestMsc14IronManModernMarvelDrawsOnlyWithAnotherArtifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	iron := effectstest.B12Push(g, me.ID, "Iron Man, Modern Marvel", "Legendary Artifact Creature — Human Hero", msc14IronManModernOracle, 3, 3)
	effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 2, 2)
	effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.SeedLibrary(me, "A", "B", "C")
	hand := me.Hand.Size()

	effectstest.DeclareAttack(t, g, opp.ID, iron)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("drew %d cards with no other artifact creature of mine", me.Hand.Size()-hand)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	iron2 := effectstest.B12Push(g2, me2.ID, "Iron Man, Modern Marvel", "Legendary Artifact Creature — Human Hero", msc14IronManModernOracle, 3, 3)
	effectstest.B12Creature(g2, me2.ID, "My Golem", "Artifact Creature — Golem", 2, 2)
	effectstest.SeedLibrary(me2, "A", "B", "C")
	hand2 := me2.Hand.Size()
	effectstest.DeclareAttack(t, g2, opp2.ID, iron2)
	effectstest.PassPriorityAroundTable(t, g2)
	if me2.Hand.Size() != hand2+1 {
		t.Errorf("drew %d cards with another artifact creature, want 1", me2.Hand.Size()-hand2)
	}
}

// --- Living Lightning, Charged Up ---------------------------------------------

func TestMsc14LivingLightningPumpsAndHastesAnotherCreatureAtBeginningOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lightning := effectstest.B12Push(g, me.ID, "Living Lightning, Charged Up", "Legendary Creature — Elemental Hero", msc14LivingLightningOracl, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt at the beginning of combat")
	}
	for _, offered := range p.PickTargetCards {
		if offered == lightning || offered == theirs {
			t.Errorf("offered %s: only another creature I control is legal", offered)
		}
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("the Bear has power %d, want 3", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("the Bear lacks haste")
	}
	if got := effectstest.EffectivePower(t, g, lightning); got != 3 {
		t.Errorf("Living Lightning has power %d, want 3", got)
	}
	for _, kw := range []string{"flying", "haste"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, lightning, kw) {
			t.Errorf("Living Lightning lacks %s", kw)
		}
	}
}

func TestMsc14LivingLightningHasNoTargetWithoutAnotherCreatureOfYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lightning := effectstest.B12Push(g, me.ID, "Living Lightning, Charged Up", "Legendary Creature — Elemental Hero", msc14LivingLightningOracl, 3, 3)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a target prompt with no other creature of mine")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, lightning); got != 3 {
		t.Errorf("Living Lightning pumped himself to %d", got)
	}
}

// --- Hulkbuster Armor ---------------------------------------------------------

func TestMsc14HulkbusterArmorSetsBase9x9AndFlyingOnTopOfCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	hero := msc14WithCounters(g, me.ID, "Test Hero", "Creature — Human Hero", "", 1, 1, 1)
	armor := effectstest.SeedEquipment(g, me.ID, "Hulkbuster Armor", msc14HulkbusterOracle)

	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, armor, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: hero}}, Strict: true,
	}); err != nil {
		t.Fatalf("Equip Hero {3}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, hero), effectstest.EffectiveToughness(t, g, hero); p != 9 || tough != 9 {
		t.Errorf("equipped Hero has base %d/%d, want 9/9", p, tough)
	}
	if got := effectstest.PlusOnes(g, hero); got != 1 {
		t.Errorf("the counter was lost: %d", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, hero, "flying") {
		t.Error("the equipped creature lacks flying")
	}
}

func TestMsc14HulkbusterArmorEquipHeroRefusesANonHeroAndPlainEquipCostsSix(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	armor := effectstest.SeedEquipment(g, me.ID, "Hulkbuster Armor", msc14HulkbusterOracle)

	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, armor, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}, Strict: true,
	}); err == nil {
		t.Fatal("Equip Hero {3} accepted a Bear")
	}
	if err := g.ActivateCatalogAbility(me.ID, armor, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}, Strict: true,
	}); err == nil {
		t.Fatal("Equip {6} was paid with three mana")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, armor, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}, Strict: true,
	}); err != nil {
		t.Fatalf("Equip {6} on a Bear with six mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 9 {
		t.Errorf("the equipped Bear has power %d, want 9", got)
	}
}

// --- Impossible Man -----------------------------------------------------------

func TestMsc14ImpossibleManCopiesAnotherPermanentButKeepsHisName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	man := effectstest.DcSeed(g, me.ID, "Impossible Man", "Legendary Creature — Alien Shapeshifter", msc14ImpossibleManOracle, 1, 4)
	giant := effectstest.DcSeed(g, opp.ID, "Hill Giant", "Creature — Giant", "", 3, 3)

	effectstest.DcMana(t, g, me.ID, "{U}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, man, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: giant}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DcCard(t, g, man)
	if c.Name != "Impossible Man" {
		t.Errorf("name %q, want Impossible Man", c.Name)
	}
	if c.Effective().Power != 3 || c.Effective().Toughness != 3 {
		t.Errorf("Impossible Man is %d/%d, want the Giant's 3/3", c.Effective().Power, c.Effective().Toughness)
	}
	if effectstest.HasKeywordOnBattlefield(t, g, man, "flying") {
		t.Error("a copy of a Giant kept flying")
	}
	effectstest.DcCleanup(g)
	c = effectstest.DcCard(t, g, man)
	if c.IsCopy() || c.Effective().Power != 1 || c.Effective().Toughness != 4 {
		t.Errorf("after cleanup he is %d/%d (copy %v), want 1/4", c.Effective().Power, c.Effective().Toughness, c.IsCopy())
	}
}

func TestMsc14ImpossibleManCannotCopyHimself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	man := effectstest.DcSeed(g, me.ID, "Impossible Man", "Legendary Creature — Alien Shapeshifter", msc14ImpossibleManOracle, 1, 4)
	effectstest.DcMana(t, g, me.ID, "{U}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, man, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: man}},
	}); err == nil {
		t.Fatal("Impossible Man targeted himself")
	}
}

// --- Hulkling, Young Avenger --------------------------------------------------

func TestMsc14HulklingBecomesACopyWithItsStatsFlyingAndTheAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hulkling := effectstest.DcSeed(g, me.ID, "Hulkling, Young Avenger", "Legendary Creature — Kree Skrull Hero", msc14HulklingOracle, 4, 4)
	giant := effectstest.DcSeed(g, opp.ID, "Iron Fist, Hero for Hire", "Legendary Creature — Human Mercenary Hero", msc14IronFistOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PickCard(t, g, me.ID, giant)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DcCard(t, g, hulkling)
	if c.Name != "Hulkling, Young Avenger" {
		t.Errorf("name %q, want Hulkling, Young Avenger", c.Name)
	}
	if !c.IsCopy() {
		t.Error("Hulkling did not become a copy")
	}
	if c.Effective().Power != 4 || c.Effective().Toughness != 4 {
		t.Errorf("Hulkling is %d/%d, want 4/4", c.Effective().Power, c.Effective().Toughness)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, hulkling, "flying") {
		t.Error("the copy lacks flying")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, hulkling, "trample") {
		t.Error("the copy lacks the copied creature's trample")
	}

	// The copy has the ability too: another noncreature spell triggers it.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(opp.ID))
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("the copy did not trigger on a second noncreature spell")
	}
}

func TestMsc14HulklingIgnoresCreatureSpellsAndRevertsAtCleanup(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hulkling := effectstest.DcSeed(g, me.ID, "Hulkling, Young Avenger", "Legendary Creature — Kree Skrull Hero", msc14HulklingOracle, 4, 4)
	giant := effectstest.DcSeed(g, opp.ID, "Iron Fist, Hero for Hire", "Legendary Creature — Human Mercenary Hero", msc14IronFistOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)

	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a creature spell triggered Hulkling")
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PickCard(t, g, me.ID, giant)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.DcCard(t, g, hulkling).IsCopy() {
		t.Fatal("setup: Hulkling did not copy")
	}
	effectstest.DcCleanup(g)
	if c := effectstest.DcCard(t, g, hulkling); c.IsCopy() {
		t.Error("Hulkling is still a copy after cleanup")
	}
}

// --- Age of Ultron ------------------------------------------------------------

func msc14CastAgeOfUltron(t *testing.T, g *game.Game, answers ...uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	saga := effectstest.CastCatalogSpell(t, g, "Age of Ultron", "Enchantment — Saga", msc14AgeOfUltronOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.HobAnswerTargets(t, g, me.ID, answers...)
	effectstest.PassPriorityAroundTable(t, g)
	return saga
}

func TestMsc14AgeOfUltronChapterOneDestroysOneNonartifactCreaturePerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	aBear := effectstest.B12Creature(g, a.ID, "A's Bear", "Creature — Bear", 2, 2)
	aSecond := effectstest.B12Creature(g, a.ID, "A's Ox", "Creature — Ox", 3, 3)
	aGolem := effectstest.B12Creature(g, a.ID, "A's Golem", "Artifact Creature — Golem", 2, 2)
	bBear := effectstest.B12Creature(g, b.ID, "B's Bear", "Creature — Bear", 2, 2)

	msc14CastAgeOfUltron(t, g, aBear, bBear, aGolem, mine)
	if g.Battlefield.Contains(aBear) || g.Battlefield.Contains(bBear) {
		t.Error("the chosen nonartifact creatures survived")
	}
	for name, id := range map[string]uuid.UUID{"my creature": mine, "A's second creature": aSecond, "A's artifact creature": aGolem} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was destroyed", name)
		}
	}
}

func TestMsc14AgeOfUltronChapterTwoMakesARobotVillainPerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	msc14CastAgeOfUltron(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.MshB1Tokens(g, me.ID, "Villain"); got != 3 {
		t.Fatalf("%d Villain tokens after chapter II, want one per opponent (3)", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller != me.ID || !c.IsToken() {
			continue
		}
		if !c.HasSubtype("Robot") || !c.IsArtifact() || !c.IsCreature() {
			t.Errorf("token %q is %q, want an artifact creature Robot Villain", c.Name, c.TypeLine)
		}
		if p, tough := effectstest.EffectivePower(t, g, c.InstanceID), effectstest.EffectiveToughness(t, g, c.InstanceID); p != 2 || tough != 2 {
			t.Errorf("token is %d/%d, want 2/2", p, tough)
		}
	}
}

func TestMsc14AgeOfUltronChapterThreeGivesArtifactCreaturesDeathtouchAndACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	golem := effectstest.B12Creature(g, me.ID, "My Golem", "Artifact Creature — Golem", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 2, 2)
	saga := msc14CastAgeOfUltron(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.PlusOnes(g, golem); got != 1 || !effectstest.HasKeywordOnBattlefield(t, g, golem, "deathtouch") {
		t.Errorf("my artifact creature: %d counters, deathtouch %v", got, effectstest.HasKeywordOnBattlefield(t, g, golem, "deathtouch"))
	}
	robots := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.HasSubtype("Robot") {
			robots++
			if effectstest.PlusOnes(g, c.InstanceID) != 1 || !effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "deathtouch") {
				t.Errorf("a Robot Villain from chapter II was not boosted")
			}
		}
	}
	if robots != 3 {
		t.Errorf("%d Robot Villains, want 3", robots)
	}
	if effectstest.PlusOnes(g, bear) != 0 || effectstest.HasKeywordOnBattlefield(t, g, bear, "deathtouch") {
		t.Error("a non-artifact creature was boosted")
	}
	if effectstest.PlusOnes(g, theirs) != 0 || effectstest.HasKeywordOnBattlefield(t, g, theirs, "deathtouch") {
		t.Error("an opponent's artifact creature was boosted")
	}
	if effectstest.PlusOnes(g, rock) != 0 {
		t.Error("a noncreature artifact got a counter")
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

// --- King Solomon's Frogs -----------------------------------------------------

func TestMsc14KingSolomonsFrogsExilesOnePermanentPerOpponentAndTheyDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	aBig := msc14Permanent(g, a.ID, "A's Big Rock", "Artifact", "{3}")
	aSmall := msc14Permanent(g, a.ID, "A's Small Rock", "Artifact", "{2}")
	cBig := msc14Permanent(g, c.ID, "C's Big Enchantment", "Enchantment", "{4}{W}")
	bSmall := msc14Permanent(g, b.ID, "B's Small Rock", "Artifact", "{1}")
	for _, p := range []*game.Player{me, a, b, c} {
		effectstest.SeedLibrary(p, "X", "Y")
	}
	hands := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		hands[p.ID] = p.Hand.Size()
	}

	effectstest.CastCatalogSpell(t, g, "King Solomon's Frogs", "Legendary Artifact", msc14FrogsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 12; i++ {
		effectstest.HobAnswerTargets(t, g, me.ID, aBig, aSmall, bSmall, cBig)
		if len(g.StackMeta) == 0 && len(g.Stack.Cards) == 0 {
			break
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !g.Exile.Contains(aBig) || !g.Exile.Contains(cBig) {
		t.Fatal("the chosen permanents were not exiled")
	}
	if !g.Battlefield.Contains(aSmall) || !g.Battlefield.Contains(bSmall) {
		t.Error("a permanent with mana value under 3 was exiled")
	}
	for _, p := range []*game.Player{a, c} {
		if got := p.Hand.Size(); got != hands[p.ID]+1 {
			t.Errorf("%s hand %d, want %d: their controller draws a card", p.Name, got, hands[p.ID]+1)
		}
	}
	for _, p := range []*game.Player{me, b} {
		if got := p.Hand.Size(); got != hands[p.ID] {
			t.Errorf("%s drew (hand %d, was %d)", p.Name, got, hands[p.ID])
		}
	}
}

func TestMsc14KingSolomonsFrogsDoesNothingWhenItWasNotCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a := g.Seats[0], g.Seats[1]
	aBig := msc14Permanent(g, a.ID, "A's Big Rock", "Artifact", "{3}")
	card := game.Card{InstanceID: uuid.New(), Name: "King Solomon's Frogs", TypeLine: "Legendary Artifact", OracleID: msc14FrogsOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(card.InstanceID, game.HandEntryOptions{}); err != nil {
			t.Fatalf("put onto the battlefield: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || len(g.PendingTriggers) != 0 {
		t.Error("Frogs that were not cast triggered")
	}
	if !g.Battlefield.Contains(aBig) {
		t.Error("a permanent was exiled by Frogs that were not cast")
	}
}

func TestMsc14KingSolomonsFrogsMakesYouTheMonarchAtTheCostOfItself(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	frogs := effectstest.PushCatalogPermanent(g, me.ID, "King Solomon's Frogs", "Legendary Artifact", msc14FrogsOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, frogs, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the monarch ability was paid with no mana")
	}
	if !g.Battlefield.Contains(frogs) {
		t.Fatal("a refused activation exiled the Frogs")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, frogs, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var monarch bool
	g.WithWriteLock(func() { monarch = effects.YoureTheMonarch(g, me.ID) })
	if !monarch {
		t.Error("you did not become the monarch")
	}
	if g.Battlefield.Contains(frogs) || !g.Exile.Contains(frogs) {
		t.Error("King Solomon's Frogs was not exiled as a cost")
	}
}

// --- Iron Man, Futurist Paragon -----------------------------------------------

func TestMsc14IronManFuturistTurnsAnArtifactIntoAFiveFiveFlyingArtifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Iron Man, Futurist Paragon", "Legendary Artifact Creature — Human Hero", msc14IronManFuturistOracl, 5, 5)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DcCard(t, g, rock)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Fatalf("the rock is %q, want an artifact creature", c.TypeLine)
	}
	if c.Effective().Power != 5 || c.Effective().Toughness != 5 {
		t.Errorf("the rock is %d/%d, want 5/5", c.Effective().Power, c.Effective().Toughness)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, rock, "flying") {
		t.Error("the rock lacks flying")
	}
	effectstest.DcCleanup(g)
	if c := effectstest.DcCard(t, g, rock); !c.IsCreature() || c.Effective().Power != 5 {
		t.Error("the change ended with the turn: it has no printed duration")
	}
}

func TestMsc14IronManFuturistSetsBaseSoCountersStillAdd(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Iron Man, Futurist Paragon", "Legendary Artifact Creature — Human Hero", msc14IronManFuturistOracl, 5, 5)
	theirs := msc14WithCounters(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2, 1)

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, theirs); p != 5 {
		t.Errorf("the Bear has base power %d, want 5", p)
	}
	if got := effectstest.PlusOnes(g, theirs); got != 1 {
		t.Errorf("its +1/+1 counter was lost: %d (so the total would be 6)", got)
	}
	c := effectstest.DcCard(t, g, theirs)
	if !c.IsArtifact() || !c.IsCreature() || !c.HasSubtype("Bear") {
		t.Errorf("type line %q: it should be an artifact creature and still a Bear", c.TypeLine)
	}
}

// --- Iron Man, Armored Avenger ------------------------------------------------

func TestMsc14IronManArmoredAvengerCountersATargetCreatureWhenYouDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Iron Man, Armored Avenger", "Legendary Artifact Creature — Human Hero", msc14IronManArmoredOracle, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.SeedLibrary(opp, "A", "B")

	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("an opponent's draw triggered Iron Man")
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, theirs); got != 1 {
		t.Errorf("the targeted creature has %d counters, want 1", got)
	}
}

func TestMsc14IronManArmoredAvengerGivesFlyingToOtherAttackingModifiedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	iron := effectstest.B12Push(g, me.ID, "Iron Man, Armored Avenger", "Legendary Artifact Creature — Human Hero", msc14IronManArmoredOracle, 2, 2)
	modified := msc14WithCounters(g, me.ID, "Modified Bear", "Creature — Bear", "", 2, 2, 1)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	homebody := msc14WithCounters(g, me.ID, "Modified Homebody", "Creature — Bear", "", 2, 2, 1)

	effectstest.DeclareAttack(t, g, opp.ID, iron, modified, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, modified, "flying") {
		t.Error("the other attacking modified creature lacks flying")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, plain, "flying") {
		t.Error("an unmodified attacker gained flying")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, homebody, "flying") {
		t.Error("a modified creature that did not attack gained flying")
	}
}

// --- Iron Fist, Hero for Hire -------------------------------------------------

func TestMsc14IronFistKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	fist := effectstest.B12Push(g, g.Seats[0].ID, "Iron Fist, Hero for Hire", "Legendary Creature — Human Mercenary Hero", msc14IronFistOracle, 2, 2)
	for _, kw := range []string{"trample", "prowess"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, fist, kw) {
			t.Errorf("Iron Fist lacks %s", kw)
		}
	}
}

func TestMsc14IronFistPowerUpDividesFiveDamageAndPutsFiveCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PowerUpMain(t, g)
	fist := effectstest.PushPowerUpPermanent(g, me.ID, "Iron Fist, Hero for Hire", "Legendary Creature — Human Mercenary Hero", "{1}{R}", msc14IronFistOracle)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	refs := append(effectstest.CardRefs(bear), game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})

	effectstest.FloatMana(t, g, me, "{R}{C}{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, fist, 0, game.ActivateAbilityParams{
		Targets: refs, Distribution: map[uuid.UUID]int{bear: 2, opp.ID: 2}, Strict: true,
	}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a division of 4: %v, want ErrInvalidParam", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, fist, 0, game.ActivateAbilityParams{
		Targets: refs, Distribution: map[uuid.UUID]int{bear: 2, opp.ID: 3}, Strict: true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("2 damage did not kill the 2/2")
	}
	if opp.Life != life-3 {
		t.Errorf("the opponent lost %d life, want 3", life-opp.Life)
	}
	if got := effectstest.PlusOnes(g, fist); got != 5 {
		t.Errorf("Iron Fist has %d counters, want 5", got)
	}
	effectstest.FloatMana(t, g, me, "{R}{C}{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, fist, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(), Strict: true,
	}); err == nil {
		t.Error("the power-up ability was activated twice")
	}
}

func TestMsc14IronFistPowerUpIsCheaperTheTurnHeEnteredAndRefusesSixTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PowerUpMain(t, g)
	fist := effectstest.CastPowerUpCreature(t, g, me, "Iron Fist, Hero for Hire", "Legendary Creature — Human Mercenary Hero", "{1}{R}", "{R}{C}", msc14IronFistOracle, 2)
	life := opp.Life

	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, fist, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Distribution: map[uuid.UUID]int{opp.ID: 5}, Strict: true,
	}); err == nil {
		t.Fatal("five mana paid a power-up that costs six the turn he entered")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, fist, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Distribution: map[uuid.UUID]int{opp.ID: 5}, Strict: true,
	}); err != nil {
		t.Fatalf("six mana the turn he entered: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-5 {
		t.Errorf("the opponent lost %d, want 5", life-opp.Life)
	}
}

// --- Immortus, Master of Eternity ---------------------------------------------

func TestMsc14ImmortusAddsOneBluePerCardDrawnAndOnlyForNoncreatureSpells(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	immortus := effectstest.PushCatalogPermanent(g, me.ID, "Immortus, Master of Eternity", "Legendary Creature — Human Hero Villain", msc14ImmortusOracle, false)
	effectstest.SeedLibrary(me, "A", "B", "C")
	var drawn int
	g.WithWriteLock(func() {
		_ = g.DrawNForEffect(me.ID, 2)
		drawn = g.TurnTallyFor(me.ID).CardsDrawn
	})
	if drawn < 2 {
		t.Fatalf("setup: %d cards drawn this turn", drawn)
	}
	if err := g.ActivateManaAbility(me.ID, immortus, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap Immortus: %v", err)
	}
	if n := me.ManaPool.Total(); n != drawn {
		t.Fatalf("pool has %d mana, want one {U} per card drawn this turn (%d)", n, drawn)
	}
	for _, tl := range []string{"Creature — Merfolk", "Artifact Creature — Construct"} {
		id := effectstest.HandSpell(me, "A Creature", tl, "{U}")
		effectstest.RefusedForMana(t, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}), tl)
	}
	id := effectstest.HandSpell(me, "A Spell", "Instant", "{U}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("Immortus's {U} for an Instant: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestMsc14ImmortusPowerUpShufflesHandsAndGraveyardsAndEveryoneDrawsSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PowerUpMain(t, g)
	immortus := effectstest.PushPowerUpPermanent(g, me.ID, "Immortus, Master of Eternity", "Legendary Creature — Human Hero Villain", "{2}{U}{U}", msc14ImmortusOracle)
	for _, p := range g.Seats {
		g.WithWriteLock(func() { p.Hand.Cards = nil; p.Graveyard.Cards = nil })
		for i := 0; i < 12; i++ {
			effectstest.SeedLibrary(p, "Library")
		}
	}
	for i := 0; i < 3; i++ {
		effectstest.HandCardForTest(me, "My Hand", "Sorcery", "")
	}
	effectstest.PushGraveyardPermanent(me, "My Dead", "Artifact", "{1}")
	effectstest.HandCardForTest(opp, "Their Hand", "Sorcery", "")
	effectstest.PushGraveyardPermanent(opp, "Their Dead", "Artifact", "{1}")

	effectstest.FloatMana(t, g, me, "{U}{U}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, immortus, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if got := p.Hand.Size(); got != 7 {
			t.Errorf("%s has %d cards in hand, want 7", p.Name, got)
		}
		if got := p.Graveyard.Size(); got != 0 {
			t.Errorf("%s has %d cards in the graveyard, want 0 (shuffled in)", p.Name, got)
		}
	}
	if got := effectstest.PlusOnes(g, immortus); got != 1 {
		t.Errorf("Immortus has %d counters, want 1", got)
	}
	effectstest.FloatMana(t, g, me, "{U}{U}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, immortus, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the power-up ability was activated twice")
	}
}

// --- Iron Man, Bleeding Edge --------------------------------------------------

func msc14BleedingEdgeTable(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Iron Man, Bleeding Edge", "Legendary Artifact Creature — Human Hero", msc14IronManBleedOracle, 3, 5)
	return g, me
}

func msc14CastLegendaryRock(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Legend Rock", "Legendary Artifact", "", nil)
}

func TestMsc14BleedingEdgeCopiesAnArtifactSpellAsANonlegendaryToken(t *testing.T) {
	g, me := msc14BleedingEdgeTable(t)
	msc14CastLegendaryRock(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	var originals, copies int
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Legend Rock" {
			continue
		}
		if c.IsToken() {
			copies++
			if c.HasSupertype("Legendary") {
				t.Errorf("the copy is legendary: %q", c.TypeLine)
			}
		} else {
			originals++
		}
	}
	if originals != 1 || copies != 1 {
		t.Fatalf("%d Legend Rocks and %d token copies, want 1 and 1", originals, copies)
	}
}

func TestMsc14BleedingEdgeCopiesOnlyOnceEachTurnButDecliningDoesNotUseItUp(t *testing.T) {
	g, me := msc14BleedingEdgeTable(t)
	msc14CastLegendaryRock(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldNamed(g, me.ID, "Legend Rock") != 1 {
		t.Fatal("declining still copied")
	}

	effectstest.CastCatalogSpell(t, g, "Second Rock", "Artifact", "", nil)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldNamed(g, me.ID, "Second Rock") != 2 {
		t.Fatal("the second artifact spell was not copied after the first was declined")
	}

	effectstest.CastCatalogSpell(t, g, "Third Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil || effectstest.BattlefieldNamed(g, me.ID, "Third Rock") != 1 {
		t.Error("a third artifact spell in the same turn was copied")
	}
}

func TestMsc14BleedingEdgeIgnoresNonartifactSpells(t *testing.T) {
	g, me := msc14BleedingEdgeTable(t)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(g.Seats[1].ID))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Error("an Instant offered a copy")
	}
	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Error("a nonartifact creature offered a copy")
	}
}
