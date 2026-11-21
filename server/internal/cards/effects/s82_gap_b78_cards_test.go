package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b78_cards_test.go — S82 GAP batch 78 (Ravnica Remastered
// enchantments and instants, #5948). Every test plays the real catalog
// entry through the engine.

const (
	gapB78Repudiate   = "a43ecf8b-5dc9-4ac4-ae1f-b663aa8c9d7b"
	gapB78Thrash      = "342d20ca-178d-433e-9630-6fed33a8d0bc"
	gapB78Angelic     = "507d2f27-08a9-4938-b191-5734cda42ef2"
	gapB78Undercity   = "5b7b1839-2b44-4446-80b4-ddb59e4e835e"
	gapB78Eyes        = "7d0a1ae5-cd93-4789-9666-fc7bbbc6b613"
	gapB78Landing     = "a787b18d-d33f-41ea-a0f8-f05e272d06eb"
	gapB78Kasmina     = "a9be16d1-5e53-4b85-8921-c6fe9473f923"
	gapB78Bedeck      = "b8bb97a8-5f5d-48e9-a4c1-b4c20efe7739"
	gapB78TitanicBraw = "e8434c77-10df-4456-af27-acb95d624189"
)

func TestGapB78EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		gapB78Repudiate:        "Repudiate",
		gapB78Repudiate + "#1": "Replicate",
		gapB78Thrash:           "Thrash",
		gapB78Thrash + "#1":    "Threat",
		gapB78Angelic:          "Angelic Exaltation",
		gapB78Undercity:        "Undercity's Embrace",
		gapB78Eyes:             "Eyes Everywhere",
		gapB78Landing:          "Forced Landing",
		gapB78Kasmina:          "Kasmina's Transmutation",
		gapB78Bedeck:           "Bedeck",
		gapB78Bedeck + "#1":    "Bedazzle",
		gapB78TitanicBraw:      "Titanic Brawl",
	}
	for oracle, name := range want {
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

// gapB78CastSplit casts one half of a split card for the active seat.
func gapB78CastSplit(t *testing.T, g *game.Game, oracle, left, right string, face int, targets []game.TargetRef) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	card := game.Card{
		InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: left, TypeLine: "Instant", ManaCost: "{0}"},
			{Name: right, TypeLine: "Sorcery", ManaCost: "{0}"},
		},
	}
	card.SetFace(0)
	me.Hand.PushTop(card)
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: face, Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", left, err)
	}
}

func TestGapB78RepudiateCountersAnAbility(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := g.Seats[1]
	src := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	fired := false
	ability := taAnnounce(t, g, opp.ID, src, "Their Rock — do a thing", nil, &fired)
	gapB78CastSplit(t, g, gapB78Repudiate, "Repudiate", "Replicate", 0, []game.TargetRef{{Kind: game.TargetCard, ID: ability}})
	passPriorityAroundTable(t, g)
	if fired {
		t.Error("the countered ability resolved")
	}
	if !g.Battlefield.Contains(src) {
		t.Error("the source left the battlefield")
	}
}

func TestGapB78ReplicateCopiesOnlyYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 5, 5)
	advanceToMain(t, g)
	gapB78CastSplit(t, g, gapB78Repudiate, "Repudiate", "Replicate", 1, cardTarget(mine))
	passPriorityAroundTable(t, g)
	if n := len(battlefieldIDsNamed(g, "Mine")); n != 2 {
		t.Errorf("%d creatures named Mine, want 2", n)
	}
	spec, _ := Lookup(gapB78Repudiate + "#1")
	legal := taLegalForSpec(g, me.ID, spec.Targets)
	if legal[theirs] {
		t.Error("Replicate offered an opponent's creature")
	}
}

func TestGapB78ThrashBitesAndThreatMakesABeast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 5, 5)
	advanceToMain(t, g)
	gapB78CastSplit(t, g, gapB78Thrash, "Thrash", "Threat", 0, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1}})
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, theirs).DamageMarked; got != 3 {
		t.Errorf("the target took %d, want 3", got)
	}
	if got := findCardAnywhere(t, g, mine).DamageMarked; got != 0 {
		t.Errorf("my creature took %d, want 0 (one-sided)", got)
	}
	gapB78CastSplit(t, g, gapB78Thrash, "Thrash", "Threat", 1, nil)
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Beast", me.ID); n != 1 {
		t.Fatalf("%d Beasts, want 1", n)
	}
	for _, c := range tokensNamed(g, "Beast") {
		if c.Power != 4 || c.Toughness != 4 {
			t.Errorf("Beast is %d/%d, want 4/4", c.Power, c.Toughness)
		}
	}
}

func TestGapB78AngelicExaltationPumpsALoneAttackerByCreatureCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Angelic Exaltation", "Enchantment", gapB78Angelic, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)
	b12Creature(g, me.ID, "My Cat", "Creature — Cat", 1, 1)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if p, to := effectivePTOf(t, g, bear); p != 5 || to != 5 {
		t.Errorf("the lone attacker is %d/%d, want 5/5 (+3/+3 for three creatures)", p, to)
	}
}

func TestGapB78AngelicExaltationDoesNothingForTwoAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Angelic Exaltation", "Enchantment", gapB78Angelic, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	wolf := b12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)
	declareAttack(t, g, opp.ID, bear, wolf)
	passPriorityAroundTable(t, g)
	if p, _ := effectivePTOf(t, g, bear); p != 2 {
		t.Errorf("power %d, want 2", p)
	}
}

func TestGapB78UndercitysEmbraceSacrificesAndGainsOnlyWithAFourPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Small", "Creature — Bear", 3, 3)
	life := me.Life
	castCatalogSpell(t, g, "Undercity's Embrace", "Instant", gapB78Undercity, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, victim) {
		t.Error("the opponent kept their creature")
	}
	if me.Life != life {
		t.Errorf("gained %d with only a 3-power creature", me.Life-life)
	}
	b12Creature(g, me.ID, "Big", "Creature — Giant", 4, 4)
	castCatalogSpell(t, g, "Undercity's Embrace", "Instant", gapB78Undercity, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d, want %d with a 4-power creature", me.Life, life+4)
	}
}

func TestGapB78EyesEverywhereScriesAndExchangesAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	eyes := b12Push(g, me.ID, "Eyes Everywhere", "Enchantment", gapB78Eyes, 0, 0)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	land := b12Permanent(g, opp.ID, "Their Land", "Land")
	if err := g.ActivateCatalogAbility(me.ID, eyes, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(land)}); err == nil {
		t.Fatal("activated with no mana, or targeted a land")
	}
	apaMana(me, "U", "U", "U", "U", "U", "U")
	if err := g.ActivateCatalogAbility(me.ID, eyes, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(land)}); err == nil {
		t.Fatal("targeted a land")
	}
	apaActivate(t, g, me, eyes, 0, game.ActivateAbilityParams{Targets: cardTarget(rock)})
	if c := findBattlefieldCardForTest(g, rock); c == nil || c.Controller != me.ID {
		t.Error("I do not control the artifact")
	}
	if c := findBattlefieldCardForTest(g, eyes); c == nil || c.Controller != opp.ID {
		t.Error("the opponent does not control Eyes Everywhere")
	}
}

func TestGapB78EyesEverywhereRefusesAtInstantSpeed(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	eyes := b12Push(g, me.ID, "Eyes Everywhere", "Enchantment", gapB78Eyes, 0, 0)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	fired := false
	taAnnounce(t, g, opp.ID, rock, "Their Rock — do a thing", nil, &fired)
	apaMana(me, "U", "U", "U", "U", "U", "U")
	if err := g.ActivateCatalogAbility(me.ID, eyes, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(rock)}); err == nil {
		t.Error("activated with something on the stack")
	}
}

func TestGapB78ForcedLandingTucksAFlyer(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	flyer := pushPermanent(g, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErrStrict(t, g, "Forced Landing", "Instant", gapB78Landing, cardTarget(bear)); err == nil {
		t.Fatal("targeted a creature without flying")
	}
	castCatalogSpell(t, g, "Forced Landing", "Instant", gapB78Landing, cardTarget(flyer))
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, flyer) {
		t.Error("the flyer stayed")
	}
	cards := opp.Library.Cards
	if len(cards) == 0 || cards[0].InstanceID != flyer {
		t.Error("the flyer is not on the bottom of its owner's library")
	}
}

func TestGapB78KasminasTransmutationMakesItAOneOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := pushPermanent(g, opp.ID, game.Card{Name: "Host", TypeLine: "Creature — Dragon", Power: 5, Toughness: 5, Keywords: []string{"flying"}})
	if !hasEffectiveKeyword(t, g, host, "flying") {
		t.Fatal("setup: the host has no flying")
	}
	cmdr2B19Attach(g, me.ID, host, "Kasmina's Transmutation", gapB78Kasmina)
	if p, to := effectivePTOf(t, g, host); p != 1 || to != 1 {
		t.Errorf("host is %d/%d, want 1/1", p, to)
	}
	if hasEffectiveKeyword(t, g, host, "flying") {
		t.Error("the host kept flying")
	}
}

func TestGapB78BedeckAndBedazzle(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 4)
	advanceToMain(t, g)
	gapB78CastSplit(t, g, gapB78Bedeck, "Bedeck", "Bedazzle", 0, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, to := effectivePTOf(t, g, bear); p != 5 || to != 1 {
		t.Errorf("Bedeck left it %d/%d, want 5/1", p, to)
	}
	basic := pushCatalogPermanent(g, opp.ID, "Plains", "Basic Land — Plains", "", false)
	nonbasic := pushCatalogPermanent(g, opp.ID, "Breeding Pool", "Land — Forest Island", "", false)
	targets := func(land uuid.UUID) []game.TargetRef {
		return []game.TargetRef{{Kind: game.TargetCard, ID: land, Slot: 0}, {Kind: game.TargetPlayer, ID: opp.ID, Slot: 1}}
	}
	card := game.Card{InstanceID: uuid.New(), OracleID: gapB78Bedeck, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{{Name: "Bedeck", TypeLine: "Instant", ManaCost: "{0}"}, {Name: "Bedazzle", TypeLine: "Instant", ManaCost: "{0}"}}}
	card.SetFace(0)
	me.Hand.PushTop(card)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Strict: true, Targets: targets(basic)}); err == nil {
		t.Fatal("Bedazzle targeted a basic land")
	}
	life := opp.Life
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: targets(nonbasic)}); err != nil {
		t.Fatalf("cast Bedazzle: %v", err)
	}
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, nonbasic) {
		t.Error("the nonbasic land survived")
	}
	if !battlefieldHas(g, basic) {
		t.Error("the basic land was destroyed")
	}
	if opp.Life != life-2 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-2)
	}
}

func TestGapB78TitanicBrawlFightsAndCostsLessWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Titanic Brawl", "Instant", gapB78TitanicBraw, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1}})
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, theirs) {
		t.Error("their creature survived the fight")
	}
	if got := findCardAnywhere(t, g, mine).DamageMarked; got != 2 {
		t.Errorf("my creature took %d, want 2", got)
	}
}

func TestGapB78TitanicBrawlCostsOneLessAtACreatureWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plain := pushVanillaCreature(g, me.ID, "Plain", 2, 2)
	counted := pushVanillaCreature(g, me.ID, "Counted", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	if err := g.AddCounterForEffect(counted, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	refs := func(mine uuid.UUID) []game.TargetRef {
		return []game.TargetRef{{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1}}
	}
	if got := selfPriced(t, g, me, gapB78TitanicBraw, "Instant", "{1}{G}", game.ZoneHand, refs(counted)).ManaValue(); got != 1 {
		t.Errorf("at a creature with a counter: %d, want 1", got)
	}
	if got := selfPriced(t, g, me, gapB78TitanicBraw, "Instant", "{1}{G}", game.ZoneHand, refs(plain)).ManaValue(); got != 2 {
		t.Errorf("at a plain creature: %d, want 2", got)
	}
}
