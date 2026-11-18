package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b16_test.go — S82 reprint pass, batch 16 (#5908): Gaea's Might,
// Griffin Protector, Hammerhand, Snarespinner, Alesha, Who Smiles at
// Death, Archelos, Lagoon Mystic, Ashling the Pilgrim, Bad River, Bell
// Borca, Spectral Sergeant, Chromanticore and Day of Destiny.

const (
	rprB16GaeasMightOracle  = "73b26f12-78eb-4d01-9dd6-ee643c7a80a8"
	rprB16GriffinOracle     = "2d91cf05-fc39-4d76-8c95-62738b57e0a0"
	rprB16HammerhandOracle  = "8efa0026-c3f8-44e0-8f55-729f9c30ac93"
	rprB16SnarespinnerOrcl  = "0f63de7b-03f2-49ad-b217-88fe9c50d68a"
	rprB16AleshaOracle      = "6969a7e2-6866-4001-a139-24b3be13deae"
	rprB16ArchelosOracle    = "fd19a009-d1ab-4148-a829-3d4c3af1b0b2"
	rprB16AshlingOracle     = "bc92f0ba-2cf0-453c-a0d3-98a8784fe55b"
	rprB16BadRiverOracle    = "259fc423-d078-4668-9e95-8c1a0d2e49ba"
	rprB16BellBorcaOracle   = "5e29aea1-06e3-49ef-8a22-e3acff1fb679"
	rprB16ChromanticoreOrcl = "695ee4fe-356d-47fb-9442-d171b618151c"
	rprB16DayOfDestinyOrcl  = "64247fcb-0e1b-48c9-bdac-769d7edd8d1d"
)

func TestRprB16EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		rprB16GaeasMightOracle:  "Gaea's Might",
		rprB16GriffinOracle:     "Griffin Protector",
		rprB16HammerhandOracle:  "Hammerhand",
		rprB16SnarespinnerOrcl:  "Snarespinner",
		rprB16AleshaOracle:      "Alesha, Who Smiles at Death",
		rprB16ArchelosOracle:    "Archelos, Lagoon Mystic",
		rprB16AshlingOracle:     "Ashling the Pilgrim",
		rprB16BadRiverOracle:    "Bad River",
		rprB16BellBorcaOracle:   "Bell Borca, Spectral Sergeant",
		rprB16ChromanticoreOrcl: "Chromanticore",
		rprB16DayOfDestinyOrcl:  "Day of Destiny",
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

// --- Gaea's Might -------------------------------------------------------------

func TestRprB16GaeasMightPumpsByBasicLandTypes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	domainLands(g, me, opp, "Forest", "Island") // Forest, Island, Plains = 3
	castCatalogSpell(t, g, "Gaea's Might", "Instant", rprB16GaeasMightOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("bear is %d/%d with domain 3, want 5/5", p, tough)
	}
}

func TestRprB16GaeasMightWithNoLandsIsNothingAndNeedsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Gaea's Might", "Instant", rprB16GaeasMightOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("bear is %d/%d with domain 0, want 2/2", p, tough)
	}
	land := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	if err := castCatalogSpellErr(t, g, "Gaea's Might", "Instant", rprB16GaeasMightOracle, cardTarget(land)); err == nil {
		t.Error("a land is not a legal target")
	}
}

// --- Griffin Protector --------------------------------------------------------

func TestRprB16GriffinProtectorGrowsWhenAnotherCreatureEntersUnderYourControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	griffin := b12Push(g, me.ID, "Griffin Protector", "Creature — Griffin", rprB16GriffinOracle, 2, 3)
	if !hasEffectiveKeyword(t, g, griffin, "flying") {
		t.Error("Griffin Protector has flying")
	}
	avB1Cast(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, griffin); p != 3 || tough != 4 {
		t.Errorf("after one creature entered, Griffin is %d/%d, want 3/4", p, tough)
	}
	avB1Cast(t, g, me, game.Card{Name: "Ox", TypeLine: "Creature — Ox", Power: 3, Toughness: 3}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, griffin); p != 4 || tough != 5 {
		t.Errorf("after two creatures entered, Griffin is %d/%d, want 4/5", p, tough)
	}
	// An opponent's creature does not trigger it.
	before, _ := ptOf(t, g, griffin)
	b12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: g.Battlefield.Cards[len(g.Battlefield.Cards)-1].InstanceID})
	})
	passPriorityAroundTable(t, g)
	if after, _ := ptOf(t, g, griffin); after != before {
		t.Errorf("an opponent's creature entering changed Griffin's power from %d to %d", before, after)
	}
}

// --- Snarespinner -------------------------------------------------------------

func TestRprB16SnarespinnerGetsPlusTwoWhenBlockingAFlier(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	snare := b12Push(g, opp.ID, "Snarespinner", "Creature — Spider", rprB16SnarespinnerOrcl, 1, 3)
	if !hasEffectiveKeyword(t, g, snare, "reach") {
		t.Error("Snarespinner has reach")
	}
	bird := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Bird", TypeLine: "Creature — Bird",
		Power: 2, Toughness: 2, Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID,
	})
	attackIntoBlocks(t, g, opp.ID, bird)
	if err := g.DeclareBlocker(snare, bird); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	if p, tough := eoeCDPT(t, g, snare); p != 3 || tough != 3 {
		t.Errorf("blocking a flier, Snarespinner is %d/%d, want 3/3", p, tough)
	}
}

func TestRprB16SnarespinnerIgnoresAGroundCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	snare := b12Push(g, opp.ID, "Snarespinner", "Creature — Spider", rprB16SnarespinnerOrcl, 1, 3)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	attackIntoBlocks(t, g, opp.ID, bear)
	if err := g.DeclareBlocker(snare, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	if p, tough := eoeCDPT(t, g, snare); p != 1 || tough != 3 {
		t.Errorf("blocking a ground creature, Snarespinner is %d/%d, want 1/3", p, tough)
	}
}

// --- Day of Destiny -----------------------------------------------------------

func TestRprB16DayOfDestinyPumpsOnlyYourLegendaryCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	legend := b12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Legend", "Legendary Creature — Human", 2, 2)
	b12Push(g, me.ID, "Day of Destiny", "Legendary Enchantment", rprB16DayOfDestinyOrcl, 0, 0)
	if p, tough := ptOf(t, g, legend); p != 4 || tough != 4 {
		t.Errorf("your legend is %d/%d, want 4/4", p, tough)
	}
	if p, tough := ptOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("a nonlegendary bear is %d/%d, want 2/2", p, tough)
	}
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their legend is %d/%d, want 2/2", p, tough)
	}
}

// --- Hammerhand ---------------------------------------------------------------

func TestRprB16HammerhandPumpsHastesAndStopsAnyBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	other := blbB1Push(g, me.ID, "Ox", "Creature — Ox", "", 3, 3)
	castCatalogSpell(t, g, "Hammerhand", "Enchantment — Aura", rprB16HammerhandOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("the entry trigger should ask for its target")
	}
	// Unlike War Squeak, any creature is legal, including your own.
	answerPickTarget(t, g, other)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the enchanted bear is %d/%d, want 3/3", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("the enchanted bear should have haste")
	}
	assertRestrictions(t, g, other, game.CantBlock)
	assertRestrictions(t, g, bear, 0)
	_ = opp
}

func TestRprB16HammerhandNeedsACreatureToEnchant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	if err := castCatalogSpellErr(t, g, "Hammerhand", "Enchantment — Aura", rprB16HammerhandOracle, cardRefs(land)); err == nil {
		t.Error("a land is not a legal enchant target")
	}
}

// --- Bad River ----------------------------------------------------------------

func TestRprB16BadRiverEntersTappedAndFetchesAnIslandOrSwamp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	fID := uuid.New()
	sID := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: sID, Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: fID, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	river := b12PlayFromHand(t, g, "Bad River", "Land", rprB16BadRiverOracle, game.CastSpellParams{})
	c, ok := battlefieldCard(g, river)
	if !ok || !c.Tapped {
		t.Fatalf("Bad River should enter tapped (on battlefield %v, tapped %v)", ok, c.Tapped)
	}
	if err := hobB4Activate(t, g, me.ID, river, 0, "", nil); err == nil {
		t.Fatal("a tapped Bad River can't pay its {T} cost")
	}
	untapForTest(g, river)
	if err := hobB4Activate(t, g, me.ID, river, 0, "", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	if _, ok := battlefieldCard(g, river); ok {
		t.Error("Bad River should have been sacrificed")
	}
	if _, ok := battlefieldCard(g, fID); ok {
		t.Error("a Forest is not an Island or Swamp")
	}
	sw, ok := battlefieldCard(g, sID)
	if !ok {
		t.Fatal("the Swamp should have been put onto the battlefield")
	}
	if sw.Tapped {
		t.Error("the fetched Swamp enters untapped")
	}
}

// --- Chromanticore ------------------------------------------------------------

func TestRprB16ChromanticoreBestowedGivesPlusFourAndFiveKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := bestowCard(t, g, "Chromanticore", "Enchantment Creature — Manticore", rprB16ChromanticoreOrcl, 4, 4)
	if alt := game.AlternativeCostByKey(rprB16ChromanticoreOrcl, game.BestowKey); alt == nil || !alt.Bestow || alt.ManaCost != "{2}{W}{U}{B}{R}{G}" {
		t.Fatalf("bestow offer %+v", alt)
	}
	castBestowed(t, g, id, bear)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("enchanted bear is %d/%d, want 6/6", p, tough)
	}
	for _, kw := range []string{"flying", "first strike", "vigilance", "trample", "lifelink"} {
		if !hasKeywordOnBattlefield(t, g, bear, kw) {
			t.Errorf("enchanted bear should have %s", kw)
		}
	}
}

func TestRprB16ChromanticoreAsACreatureHasTheFiveKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Chromanticore", "Enchantment Creature — Manticore", rprB16ChromanticoreOrcl, 4, 4)
	for _, kw := range []string{"flying", "first strike", "vigilance", "trample", "lifelink"} {
		if !hasKeywordOnBattlefield(t, g, id, kw) {
			t.Errorf("Chromanticore should have %s", kw)
		}
	}
	if p, tough := ptOf(t, g, id); p != 4 || tough != 4 {
		t.Errorf("Chromanticore is %d/%d, want 4/4", p, tough)
	}
}

// --- Ashling the Pilgrim ------------------------------------------------------

func TestRprB16AshlingGrowsTwiceThenBlastsOnTheThirdResolution(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ash := b12Push(g, me.ID, "Ashling the Pilgrim", "Legendary Creature — Elemental Shaman", rprB16AshlingOracle, 1, 1)
	bystander := b12Creature(g, opp.ID, "Bystander", "Creature — Wall", 0, 5)
	small := b12Creature(g, opp.ID, "Small", "Creature — Wall", 0, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	myLife, oppLife := me.Life, opp.Life
	for i := 1; i <= 2; i++ {
		if err := hobB4Activate(t, g, me.ID, ash, 0, "{R}{R}", nil); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
		if p, tough := ptOf(t, g, ash); p != 1+i || tough != 1+i {
			t.Fatalf("after %d resolutions Ashling is %d/%d, want %d/%d", i, p, tough, 1+i, 1+i)
		}
	}
	if me.Life != myLife || opp.Life != oppLife {
		t.Fatal("nothing should have been dealt yet")
	}
	if err := hobB4Activate(t, g, me.ID, ash, 0, "{R}{R}", nil); err != nil {
		t.Fatalf("third activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Three +1/+1 counters were put on and then removed: 3 damage to everything.
	if c, ok := battlefieldCard(g, ash); ok {
		t.Errorf("Ashling (1/1 after losing her counters, 3 damage) should have died; counters %v", c.Counters)
	}
	if _, ok := battlefieldCard(g, small); ok {
		t.Error("the 0/2 should have died to 3 damage")
	}
	if _, ok := battlefieldCard(g, bystander); !ok {
		t.Error("the 0/5 survives 3 damage")
	}
	if me.Life != myLife-3 || opp.Life != oppLife-3 {
		t.Errorf("each player takes 3: you %d→%d, opponent %d→%d", myLife, me.Life, oppLife, opp.Life)
	}
}

func TestRprB16AshlingFourthResolutionOnlyAddsACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ash := b12Push(g, me.ID, "Ashling the Pilgrim", "Legendary Creature — Elemental Shaman", rprB16AshlingOracle, 1, 1)
	eoeCDSet(g, ash, func(c *game.Card) { c.Toughness = 9; c.Power = 1 })
	advanceTo(t, g, game.StepPrecombatMain)
	life := me.Life
	for i := 1; i <= 4; i++ {
		if err := hobB4Activate(t, g, me.ID, ash, 0, "{R}{R}", nil); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
	}
	c, ok := battlefieldCard(g, ash)
	if !ok {
		t.Fatal("Ashling should have survived (toughness 9)")
	}
	if n := c.Counters[game.CounterPlusOne]; n != 1 {
		t.Errorf("after the blast and one more activation Ashling has %d counters, want 1", n)
	}
	if me.Life != life-3 {
		t.Errorf("you took %d, want 3 (the third resolution only)", life-me.Life)
	}
}

// --- Archelos, Lagoon Mystic --------------------------------------------------

func TestRprB16ArchelosTappedMakesOtherPermanentsEnterTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	arch := b12Push(g, me.ID, "Archelos, Lagoon Mystic", "Legendary Creature — Turtle Shaman", rprB16ArchelosOracle, 2, 4)
	eoeCDSet(g, arch, func(c *game.Card) { c.Tapped = true })
	forest := b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	if c, ok := battlefieldCard(g, forest); !ok || !c.Tapped {
		t.Errorf("with Archelos tapped, a land enters tapped (on battlefield %v, tapped %v)", ok, c.Tapped)
	}
	bear := b12PlayFromHand(t, g, "Bear", "Creature — Bear", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, bear); !ok || !c.Tapped {
		t.Errorf("with Archelos tapped, a creature enters tapped (on battlefield %v, tapped %v)", ok, c.Tapped)
	}
	eoeCDSet(g, arch, func(c *game.Card) { c.Tapped = false })
	ox := b12PlayFromHand(t, g, "Ox", "Creature — Ox", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, ox); !ok || c.Tapped {
		t.Errorf("with Archelos untapped, a creature enters untapped (on battlefield %v, tapped %v)", ok, c.Tapped)
	}
}

// Bad River's own "enters tapped" and Archelos's effect both replace the same
// entry, so the land's controller orders them (CR 616.1): whichever applies
// last decides. The land's own self-replacement has the higher id.
func rprB16ArchelosOrderedEntry(t *testing.T, archelosLast bool) game.Card {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Archelos, Lagoon Mystic", "Legendary Creature — Turtle Shaman", rprB16ArchelosOracle, 2, 4)
	river := b12PlayFromHand(t, g, "Bad River", "Land", rprB16BadRiverOracle, game.CastSpellParams{})
	order := b06ReplacementOrderFor(g, me.ID)
	if order == nil || len(order.ReplacementEffectIDs) != 2 {
		t.Fatalf("want a CR 616 ordering prompt between Bad River and Archelos, got %+v", order)
	}
	ids := append([]game.ReplacementEffectID(nil), order.ReplacementEffectIDs...)
	lo, hi := ids[0], ids[1]
	if lo > hi {
		lo, hi = hi, lo
	}
	// lo is Archelos's (a battlefield object's), hi is the land's own.
	ordered := []game.ReplacementEffectID{lo, hi}
	if archelosLast {
		ordered = []game.ReplacementEffectID{hi, lo}
	}
	if err := g.ResolveReplacementOrder(order.ID, me.ID, ordered); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, river)
	if !ok {
		t.Fatal("Bad River should be on the battlefield")
	}
	return c
}

func TestRprB16ArchelosUntappedOverridesAnEntersTappedLandWhenAppliedLast(t *testing.T) {
	if c := rprB16ArchelosOrderedEntry(t, true); c.Tapped {
		t.Error("with Archelos's effect applied last, Bad River enters untapped")
	}
	if c := rprB16ArchelosOrderedEntry(t, false); !c.Tapped {
		t.Error("with the land's own effect applied last, Bad River enters tapped")
	}
}

func TestRprB16ArchelosDoesNotAffectItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	arch := b12PlayFromHand(t, g, "Archelos, Lagoon Mystic", "Legendary Creature — Turtle Shaman", rprB16ArchelosOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, arch); !ok || c.Tapped {
		t.Errorf("Archelos enters untapped (on battlefield %v)", ok)
	}
	_ = me
}

// --- Alesha, Who Smiles at Death ---------------------------------------------

func rprB16Alesha(t *testing.T, g *game.Game) (alesha uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	alesha = b12Push(g, me.ID, "Alesha, Who Smiles at Death", "Legendary Creature — Human Warrior", rprB16AleshaOracle, 3, 2)
	return alesha
}

func rprB16GraveCreature(g *game.Game, p *game.Player, name string, power int) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Bear", Power: power, Toughness: 2,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

func TestRprB16AleshaPaysToReturnASmallCreatureTappedAndAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	alesha := rprB16Alesha(t, g)
	small := rprB16GraveCreature(g, me, "Small Bear", 2)
	big := rprB16GraveCreature(g, me, "Big Bear", 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(alesha, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	if err := g.ResolvePickTarget(latestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: big}); err == nil {
		t.Error("a power 3 creature card is not a legal target")
	}
	answerPickTarget(t, g, small)
	floatForTest(g, me, "WB")
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, true)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick && c.Chooser == me.ID {
			// "What does it attack?": the first option is an opponent.
			if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			break
		}
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, small)
	if !ok {
		t.Fatal("the power 2 creature should have returned")
	}
	if !c.Tapped {
		t.Error("it should enter tapped")
	}
	if c.AttackingTarget == uuid.Nil {
		t.Error("it should enter attacking")
	}
	if _, ok := battlefieldCard(g, big); ok {
		t.Error("the power 3 creature was not a target and must stay put")
	}
}

func TestRprB16AleshaDecliningThePaymentReturnsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	alesha := rprB16Alesha(t, g)
	small := rprB16GraveCreature(g, me, "Small Bear", 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(alesha, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	answerPickTarget(t, g, small)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		answerPayUnless(t, g, me.ID, false)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, small); ok {
		t.Error("the card must not return without the payment")
	}
}

// --- Bell Borca, Spectral Sergeant --------------------------------------------

func rprB16Exile(t *testing.T, g *game.Game, p *game.Player, name, cost string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Sorcery", ManaCost: cost, Owner: p.ID, Controller: p.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: p.ID}, game.ZoneRef{Kind: game.ZoneExile}, id); err != nil {
		t.Fatalf("exile %s: %v", name, err)
	}
	return id
}

func TestRprB16BellBorcaPowerIsTheGreatestManaValueExiledThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bell := b12Push(g, me.ID, "Bell Borca, Spectral Sergeant", "Legendary Creature — Spirit Soldier", rprB16BellBorcaOracle, 0, 5)
	if p, tough := ptOf(t, g, bell); p != 0 || tough != 5 {
		t.Fatalf("nothing noted: Bell Borca is %d/%d, want 0/5", p, tough)
	}
	rprB16Exile(t, g, me, "Two", "{1}{R}")
	if p, _ := ptOf(t, g, bell); p != 2 {
		t.Errorf("after exiling a mana value 2 card, power %d, want 2", p)
	}
	rprB16Exile(t, g, opp, "Five", "{3}{R}{R}")
	if p, _ := ptOf(t, g, bell); p != 5 {
		t.Errorf("an opponent's mana value 5 card counts too: power %d, want 5", p)
	}
	rprB16Exile(t, g, me, "One", "{R}")
	if p, tough := ptOf(t, g, bell); p != 5 || tough != 5 {
		t.Errorf("the greatest stays: %d/%d, want 5/5", p, tough)
	}
}

func TestRprB16BellBorcaIgnoresCardsExiledBeforeItEnteredAndResetsNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rprB16Exile(t, g, me, "Early", "{4}")
	bell := b12Push(g, me.ID, "Bell Borca, Spectral Sergeant", "Legendary Creature — Spirit Soldier", rprB16BellBorcaOracle, 0, 5)
	if p, _ := ptOf(t, g, bell); p != 0 {
		t.Errorf("a card exiled before Bell Borca entered was noted: power %d", p)
	}
	rprB16Exile(t, g, me, "Late", "{3}")
	if p, _ := ptOf(t, g, bell); p != 3 {
		t.Fatalf("power %d, want 3", p)
	}
	turn := g.Turn.Seq
	for i := 0; i < 40 && g.Turn.Seq == turn; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.Seq == turn {
		t.Fatal("the turn never advanced")
	}
	if p, _ := ptOf(t, g, bell); p != 0 {
		t.Errorf("on the next turn nothing is noted: power %d, want 0", p)
	}
}

func TestRprB16BellBorcaUpkeepExilesTheTopCardForYouToPlay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Bell Borca, Spectral Sergeant", "Legendary Creature — Spirit Soldier", rprB16BellBorcaOracle, 0, 5)
	top := stackLibrary(me, "Top Card", "Sorcery", "{2}{R}")
	// Walk to the controller's next upkeep.
	for i := 0; i < 80 && !(g.Turn.Step == game.StepUpkeep && g.Seats[g.Turn.ActiveSeat] == me && g.Exile.Contains(top)); i++ {
		if g.Turn.Step == game.StepUpkeep && g.Seats[g.Turn.ActiveSeat] == me {
			passPriorityAroundTable(t, g)
			break
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if !g.Exile.Contains(top) {
		t.Fatal("the top card of the library should have been exiled at upkeep")
	}
	if perm := exiledPermission(g, top); !permissionLive(g, perm, me.ID) {
		t.Error("you may play the exiled card this turn")
	}
	if p, _ := ptOf(t, g, func() uuid.UUID {
		for _, c := range g.Battlefield.Cards {
			if c.OracleID == rprB16BellBorcaOracle {
				return c.InstanceID
			}
		}
		return uuid.Nil
	}()); p != 3 {
		t.Errorf("Bell Borca notes the exiled mana value 3: power %d", p)
	}
}

func TestRprB16ArchelosTappedAlsoTapsTokensAndAnOpponentsPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arch := b12Push(g, me.ID, "Archelos, Lagoon Mystic", "Legendary Creature — Turtle Shaman", rprB16ArchelosOracle, 2, 4)
	eoeCDSet(g, arch, func(c *game.Card) { c.Tapped = true })
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, game.Card{Name: "Elf Token", TypeLine: "Token Creature — Elf", Power: 1, Toughness: 1}, 1)
	})
	passPriorityAroundTable(t, g)
	id := battlefieldIDNamed(g, opp.ID, "Elf Token")
	if c, ok := battlefieldCard(g, id); !ok || !c.Tapped {
		t.Errorf("an opponent's token enters tapped while Archelos is tapped (on battlefield %v, tapped %v)", ok, c.Tapped)
	}
}
