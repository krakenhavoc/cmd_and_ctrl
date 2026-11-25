package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b096_cards_test.go — Zendikar Rising, batch 096 (#5508). Every test
// plays the real catalog entry through the engine.

const (
	oldB096NissaOracle            = "15e3e226-0dd1-4c9b-a7ec-af67509cbfc5"
	oldB096SquidOracle            = "301d36e7-06ff-48ed-a7e8-c7d3c72e1e2f"
	oldB096PhylathOracle          = "3d79cbe2-cf1b-4313-ab47-0118fcdbfc6c"
	oldB096AlliedAssaultOracle    = "31d8c2c4-3309-4eee-a704-f760f381f080"
	oldB096PlunderOracle          = "d2d5f9b8-7caf-42c8-934b-3f4b81468914"
	oldB096LullmageOracle         = "fdac480b-03b5-4ece-b42a-e6928b609419"
	oldB096AnticognitionOracle    = "4728e320-d29b-4019-8ab5-e45319c77c40"
	oldB096ConcertedDefenseOracle = "77f02ae6-2411-4378-b5e0-cc1cf5964458"
	oldB096RootgrazerOracle       = "ca7964a1-b4ce-412b-bcae-e0dfbb95389a"
	oldB096MindCarverOracle       = "65d6eae4-4318-4dbe-af7a-9f57bb587605"
	oldB096UtilityKnifeOracle     = "b546c78c-5180-46c3-8594-b9d5ad9c3988"
	oldB096RavagersMaceOracle     = "f3e0aaa1-f624-4dc7-a425-7dbf27cba1ce"
)

func TestOldB096EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		oldB096NissaOracle, oldB096SquidOracle, oldB096PhylathOracle, oldB096AlliedAssaultOracle, oldB096PlunderOracle,
		oldB096LullmageOracle, oldB096AnticognitionOracle, oldB096ConcertedDefenseOracle, oldB096RootgrazerOracle,
		oldB096MindCarverOracle, oldB096UtilityKnifeOracle, oldB096RavagersMaceOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// oldB096FillGraveyard leaves p with n cards in their graveyard, the last one
// arriving by a real zone move, which drops the layer cache.
func oldB096FillGraveyard(g *game.Game, p *game.Player, n int) {
	for i := 0; i < n-1; i++ {
		p.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
	effectstest.B27Kill(g, effectstest.B12Creature(g, p.ID, "Doomed Bear", "Creature — Bear", 1, 1))
}

// oldB096Party seats a Cleric and a Rogue for `who` (a party of two).
func oldB096Party(g *game.Game, who uuid.UUID) {
	effectstest.B12Creature(g, who, "Healer", "Creature — Human Cleric", 1, 1)
	effectstest.B12Creature(g, who, "Thief", "Creature — Human Rogue", 1, 1)
}

// --- Skyclave Squid -----------------------------------------------------------

func TestOldB096SkyclaveSquidAttacksOnlyAfterALandEnters(t *testing.T) {
	for _, landfall := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
		squid := effectstest.B12Push(g, me.ID, "Skyclave Squid", "Creature — Squid", oldB096SquidOracle, 3, 2)
		if landfall {
			effectstest.AdvanceToMain(t, g)
			effectstest.B31PlayLand(t, g, "Forest", "Basic Land — Forest", "")
			effectstest.PassPriorityAroundTable(t, g)
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		err := g.DeclareAttacker(squid, opp.ID)
		if landfall && err != nil {
			t.Errorf("after landfall the Squid could not attack: %v", err)
		}
		if !landfall && err == nil {
			t.Error("a Squid with defender attacked with no landfall")
		}
	}
}

// --- Phylath, World Sculptor --------------------------------------------------

func TestOldB096PhylathMakesAPlantPerBasicLandAndGrowsOneOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.B12Permanent(g, me.ID, "Dryad Arbor", "Land Creature — Forest Dryad")
	effectstest.B12Permanent(g, me.ID, "Command Tower", "Land")
	effectstest.CastHoldCreature(t, g, "Phylath, World Sculptor", "Legendary Creature — Elemental", oldB096PhylathOracle, 5, 5)
	if n := effectstest.TmtB1Count(g, me.ID, "Plant Token"); n != 2 {
		if n2 := effectstest.TmtB1Count(g, me.ID, "Plant"); n2 != 2 {
			t.Fatalf("%d Plant tokens, want 2 (one per basic land)", n2)
		}
	}
	effectstest.B31PlayLand(t, g, "Island", "Basic Land — Island", "")
	plant := effectstest.B12Creature(g, me.ID, "Plant Friend", "Creature — Plant", 0, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, plant)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, plant, game.CounterPlusOne); got != 4 {
		t.Errorf("the Plant has %d +1/+1 counters, want 4", got)
	}
}

// --- Allied Assault -----------------------------------------------------------

func TestOldB096AlliedAssaultPumpsUpToTwoByTheParty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	oldB096Party(g, me.ID)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Allied Assault", "Instant", oldB096AlliedAssaultOracle, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, tough := effectstest.EffectivePTOf(t, g, id); p != 4 || tough != 4 {
			t.Errorf("a target is %d/%d, want 4/4 (a party of two)", p, tough)
		}
	}
	if p, tough := effectstest.EffectivePTOf(t, g, c); p != 2 || tough != 2 {
		t.Errorf("the untargeted Bear is %d/%d, want 2/2", p, tough)
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Allied Assault", "Instant", oldB096AlliedAssaultOracle, effectstest.CardRefs(a, b, c)); err == nil {
		t.Error("three targets were accepted")
	}
}

// --- Skyclave Plunder ---------------------------------------------------------

func TestOldB096SkyclavePlunderLooksAtThreePlusPartyAndTakesThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	oldB096Party(g, me.ID)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	hand := me.Hand.Size()
	library := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Skyclave Plunder", "Sorcery", oldB096PlunderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no prompt for the look: %+v", g.PendingChoices)
	}
	if len(c.ChooseCards) != 5 {
		t.Errorf("looked at %d cards, want 5 (three plus a party of two)", len(c.ChooseCards))
	}
	if c.ChooseMin != 3 || c.ChooseMax != 3 {
		t.Errorf("must take %d..%d, want exactly 3", c.ChooseMin, c.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, c.ChooseCards[0], c.ChooseCards[1], c.ChooseCards[2])
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d (had %d), want +3", me.Hand.Size(), hand)
	}
	if me.Library.Size() != library-3 {
		t.Errorf("library %d (had %d), want -3 with the other two on the bottom", me.Library.Size(), library)
	}
}

// --- Lullmage's Domination ----------------------------------------------------

func TestOldB096LullmagesDominationStealsACreatureWithManaValueX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	cheap := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Cheap", TypeLine: "Creature — Elf", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	dear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Dear", TypeLine: "Creature — Giant", ManaCost: "{4}{G}", Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U", "U", "U", "C", "C")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Lullmage's Domination", TypeLine: "Sorcery", ManaCost: "{X}{U}{U}{U}", OracleID: oldB096LullmageOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 2, Targets: effectstest.CardRefs(dear)}); err == nil {
		t.Fatal("targeted a creature whose mana value is not X")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 2, Targets: effectstest.CardRefs(cheap)}); err != nil {
		t.Fatalf("cast with X=2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := g.LookupCardForEffect(cheap); !ok || c.Controller != me.ID {
		t.Error("the creature did not change controllers")
	}
}

func TestOldB096LullmagesDominationCostsThreeLessAgainstEightInTheirGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	cheap := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Cheap", TypeLine: "Creature — Elf", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U", "U", "U")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Lullmage's Domination", TypeLine: "Sorcery", ManaCost: "{X}{U}{U}{U}", OracleID: oldB096LullmageOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 2, Targets: effectstest.CardRefs(cheap)}); err == nil {
		t.Fatal("cast X=2 for {U}{U}{U} with a small opposing graveyard")
	}
	oldB096FillGraveyard(g, opp, 8)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 2, Targets: effectstest.CardRefs(cheap)}); err != nil {
		t.Fatalf("cast X=2 for {U}{U}{U} against eight cards: %v", err)
	}
}

// --- counterspells ------------------------------------------------------------

func TestOldB096AnticognitionTaxesWithAFewCardsAndCountersAndScriesWithEight(t *testing.T) {
	for _, eight := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, them := g.Seats[0], g.Seats[1]
		effectstest.ToMainForCost(t, g)
		victim := effectstest.CastCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
		if eight {
			oldB096FillGraveyard(g, me, 8)
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		effectstest.SeedLibrary(them, "A", "B", "C")
		effectstest.CastFromSeat(t, g, them, "Anticognition", "Instant", oldB096AnticognitionOracle, effectstest.CardRefs(victim))
		if !eight {
			effectstest.PassUntilTaxed(t, g, me.ID)
			if p := effectstest.AnswerPayUnless(t, g, me.ID, false); p.PayCost != "{2}" {
				t.Errorf("the tax is %s, want {2}", p.PayCost)
			}
			if g.Stack.Contains(victim) {
				t.Error("declining the tax did not counter the spell")
			}
			continue
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("with eight cards the spell is countered outright, no tax")
		}
		if g.Stack.Contains(victim) || !me.Graveyard.Contains(victim) {
			t.Error("the creature spell was not countered")
		}
		if c := effectstest.ScryChoiceFor(g, them.ID); c == nil || len(c.ScryCards) != 2 {
			t.Errorf("no scry 2 after the counter: %+v", c)
		}
	}
}

func TestOldB096AnticognitionRefusesANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	them := g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	id := uuid.New()
	them.Hand.PushTop(game.Card{InstanceID: id, Name: "Anticognition", TypeLine: "Instant", OracleID: oldB096AnticognitionOracle, Owner: them.ID, Controller: them.ID})
	if err := g.CastSpell(them.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(victim)}); err == nil {
		t.Error("Anticognition targeted a noncreature spell")
	}
}

func TestOldB096ConcertedDefenseTaxesOnePlusTheParty(t *testing.T) {
	for party := 0; party <= 2; party++ {
		g := effectstest.NewCatalogGame(t)
		me, them := g.Seats[0], g.Seats[1]
		effectstest.ToMainForCost(t, g)
		victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		if party == 2 {
			oldB096Party(g, them.ID)
		}
		if party == 1 {
			effectstest.B12Creature(g, them.ID, "Healer", "Creature — Human Cleric", 1, 1)
		}
		effectstest.CastFromSeat(t, g, them, "Concerted Defense", "Instant", oldB096ConcertedDefenseOracle, effectstest.CardRefs(victim))
		effectstest.PassUntilTaxed(t, g, me.ID)
		want := map[int]string{0: "{1}", 1: "{2}", 2: "{3}"}[party]
		if p := effectstest.AnswerPayUnless(t, g, me.ID, false); p.PayCost != want {
			t.Errorf("with a party of %d the tax is %s, want %s", party, p.PayCost, want)
		}
	}
}

// --- Murasa Rootgrazer --------------------------------------------------------

func TestOldB096MurasaRootgrazerPutsABasicLandFromHandOrBouncesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	rootgrazer := effectstest.PushCatalogPermanent(g, me.ID, "Murasa Rootgrazer", "Creature — Beast", oldB096RootgrazerOracle, false)
	basic := uuid.New()
	nonbasic := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: basic, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	me.Hand.PushTop(game.Card{InstanceID: nonbasic, Name: "Command Tower", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	if !effectstest.HasEffectiveKeyword(t, g, rootgrazer, "vigilance") {
		t.Error("no vigilance")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, rootgrazer, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil || len(c.ChooseCards) != 1 || c.ChooseCards[0] != basic {
		t.Fatalf("want a prompt offering only the basic land: %+v", c)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, basic)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(basic) || !me.Hand.Contains(nonbasic) {
		t.Error("the basic land did not enter, or the other land moved")
	}
	// Untap and bounce the basic it put.
	effectstest.AdvanceTo(t, g, game.StepEnd)
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(rootgrazer) })
	effectstest.B16Activate(t, g, me.ID, rootgrazer, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(basic)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(basic) || !me.Hand.Contains(basic) {
		t.Error("the basic land did not return to its owner's hand")
	}
}

// --- the Equipment ------------------------------------------------------------

func TestOldB096UtilityKnifeAttachesWhenItEntersAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	knife := effectstest.CastHoldCreature(t, g, "Utility Knife", "Artifact — Equipment", oldB096UtilityKnifeOracle, 0, 0)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(knife); !c.IsAttachedTo(bear) {
		t.Fatal("the Knife did not attach as it entered")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3", p, tough)
	}
}

func TestOldB096MindCarverPumpsMoreAgainstEightInAnOpponentsGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	carver := effectstest.CastHoldCreature(t, g, "Mind Carver", "Artifact — Equipment", oldB096MindCarverOracle, 0, 0)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(carver); !c.IsAttachedTo(bear) {
		t.Fatal("the Carver did not attach as it entered")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 3/2 (+1/+0)", p, tough)
	}
	oldB096FillGraveyard(g, opp, 8)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 5 || tough != 3 {
		t.Errorf("the Bear is %d/%d against eight cards, want 5/3 (+3/+1 instead of +1/+0)", p, tough)
	}
}

func TestOldB096RavagersMacePumpsByThePartyAndGrantsMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	cleric := effectstest.B12Creature(g, me.ID, "Healer", "Creature — Human Cleric", 1, 1)
	rogue := effectstest.B12Creature(g, me.ID, "Thief", "Creature — Human Rogue", 2, 2)
	mace := effectstest.CastHoldCreature(t, g, "Ravager's Mace", "Artifact — Equipment", oldB096RavagersMaceOracle, 0, 0)
	effectstest.PickCard(t, g, me.ID, rogue)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(mace); !c.IsAttachedTo(rogue) {
		t.Fatal("the Mace did not attach as it entered")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, rogue); p != 4 {
		t.Errorf("the equipped Rogue has power %d, want 4 (2 + a party of two)", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, rogue, "menace") {
		t.Error("no menace")
	}
	effectstest.B27Kill(g, cleric)
	if p, _ := effectstest.EffectivePTOf(t, g, rogue); p != 3 {
		t.Errorf("after the Cleric died the Rogue has power %d, want 3", p)
	}
}

// --- Nissa of Shadowed Boughs -------------------------------------------------

func TestOldB096NissaLandfallAddsLoyalty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	nissa := effectstest.B12Push(g, me.ID, "Nissa of Shadowed Boughs", "Legendary Planeswalker — Nissa", oldB096NissaOracle, 0, 0)
	effectstest.SetCounters(g, nissa, map[string]int{game.CounterLoyalty: 4})
	effectstest.AdvanceToMain(t, g)
	effectstest.B31PlayLand(t, g, "Forest", "Basic Land — Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, nissa, game.CounterLoyalty); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestOldB096NissaPlusOneUntapsAndMayAnimateTheLand(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := effectstest.Top100ActiveSeat(g)
		nissa := effectstest.B12Push(g, me.ID, "Nissa of Shadowed Boughs", "Legendary Planeswalker — Nissa", oldB096NissaOracle, 0, 0)
		effectstest.SetCounters(g, nissa, map[string]int{game.CounterLoyalty: 4})
		land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
		g.WithWriteLock(func() { _ = g.TapTargetForEffect(land) })
		effectstest.AdvanceToMain(t, g)
		effectstest.B16Activate(t, g, me.ID, nissa, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(land)})
		effectstest.PassPriorityAroundTable(t, g)
		oldB096AnswerConfirm(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.HmlTapped(t, g, land) {
			t.Error("the land stayed tapped")
		}
		if got := effectstest.CountersOn(g, nissa, game.CounterLoyalty); got != 5 {
			t.Errorf("loyalty %d, want 5", got)
		}
		p, tough := effectstest.EffectivePTOf(t, g, land)
		if yes {
			if p != 3 || tough != 3 || !effectstest.HasEffectiveKeyword(t, g, land, "haste") || !effectstest.HasEffectiveKeyword(t, g, land, "menace") {
				t.Errorf("the land is %d/%d haste=%v menace=%v, want a 3/3 with haste and menace", p, tough,
					effectstest.HasEffectiveKeyword(t, g, land, "haste"), effectstest.HasEffectiveKeyword(t, g, land, "menace"))
			}
		} else if p != 0 && p != 1 {
			t.Errorf("declining still animated the land: %d/%d", p, tough)
		}
	}
}

func TestOldB096NissaMinusFivePutsACreatureFromHandOrGraveyardWithTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	nissa := effectstest.B12Push(g, me.ID, "Nissa of Shadowed Boughs", "Legendary Planeswalker — Nissa", oldB096NissaOracle, 0, 0)
	effectstest.SetCounters(g, nissa, map[string]int{game.CounterLoyalty: 6})
	for i := 0; i < 3; i++ {
		effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	cheapDead := uuid.New()
	dearDead := uuid.New()
	inHand := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: cheapDead, Name: "Cheap Dead", TypeLine: "Creature — Elf", ManaCost: "{2}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: dearDead, Name: "Dear Dead", TypeLine: "Creature — Giant", ManaCost: "{4}{G}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	me.Hand.PushTop(game.Card{InstanceID: inHand, Name: "Hand Elf", TypeLine: "Creature — Elf", ManaCost: "{1}{G}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, nissa, 1, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no prompt: %+v", g.PendingChoices)
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range c.ChooseCards {
		offered[id] = true
	}
	if !offered[cheapDead] || !offered[inHand] || offered[dearDead] {
		t.Errorf("offered %v, want the two cards with mana value 3 or less and not the five-drop", c.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, cheapDead)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheapDead) {
		t.Fatal("the creature card did not enter")
	}
	if got := effectstest.CountersOn(g, cheapDead, game.CounterPlusOne); got != 2 {
		t.Errorf("it has %d +1/+1 counters, want 2", got)
	}
	if got := effectstest.CountersOn(g, nissa, game.CounterLoyalty); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}

// oldB096AnswerConfirm answers the open yes/no prompt of `who`.
func oldB096AnswerConfirm(t *testing.T, g *game.Game, who uuid.UUID, yes bool) {
	t.Helper()
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, who)
	if c == nil {
		t.Fatalf("no confirm prompt for %s: %+v", who, g.PendingChoices)
	}
	if err := g.ResolveConfirm(c.ID, who, yes); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
}
