package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b44_test.go — GAP batch 44, Modern Horizons 3 Commander creatures
// (#5948). Each test plays the real catalog entry.

const (
	gapb44HellkiteOracle   = "3093796d-c429-4c79-a33a-31226e00b12c"
	gapb44MorophonOracle   = "bf418a2f-e639-4e52-9cad-5a0713a1829b"
	gapb44UlalekOracle     = "9f6828e3-39d9-45d9-9bf1-0e3737a0321e"
	gapb44MagmarchOracle   = "f42c0b5b-7a86-4f79-a5d3-859ec25ed870"
	gapb44SireOracle       = "ed6607f4-20ab-4d3f-a8af-bcc148423336"
	gapb44NullifierOracle  = "bd213210-42d1-49af-9959-08188abe692a"
	gapb44LhurgoyfOracle   = "0fe76004-b2f4-4b11-abce-75fd76f68b4d"
	gapb44MagusOracle      = "30e58e88-ac99-4d32-b2fe-6d6cb6288af3"
	gapb44FrogantuaOracle  = "a08393b5-603b-41c4-aa03-2ae9c74884a7"
	gapb44SageOracle       = "d82618e7-494a-46d0-a19e-9cc647f96cfe"
	gapb44RedeemerOracle   = "a7a88bcc-eac2-4f74-804f-b49d16560184"
	gapb44OmoOracle        = "021b91ea-18d8-4b50-97da-dd003380cd68"
	gapb44NaturalizeOracle = "bdb3ca68-ec1f-4e16-81cc-d23f8f52c728"
)

func TestGapB44CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapb44HellkiteOracle:  "Gluttonous Hellkite",
		gapb44MorophonOracle:  "Morophon, the Boundless",
		gapb44UlalekOracle:    "Ulalek, Fused Atrocity",
		gapb44MagmarchOracle:  "Exterminator Magmarch",
		gapb44SireOracle:      "Sire of Stagnation",
		gapb44NullifierOracle: "Ulamog's Nullifier",
		gapb44LhurgoyfOracle:  "Lhurgoyf",
		gapb44MagusOracle:     "Magus of the Candelabra",
		gapb44FrogantuaOracle: "Rampant Frogantua",
		gapb44SageOracle:      "Sage of the Maze",
		gapb44RedeemerOracle:  "Vile Redeemer",
		gapb44OmoOracle:       "Omo, Queen of Vesuva",
	} {
		def, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

func TestGapB44LhurgoyfCountsCreatureCardsInAllGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := b12Push(g, me.ID, "Lhurgoyf", "Creature — Lhurgoyf", gapb44LhurgoyfOracle, 0, 0)
	if effectivePower(t, g, id) != 0 || effectiveToughness(t, g, id) != 1 {
		t.Errorf("with empty graveyards: %d/%d, want 0/1", effectivePower(t, g, id), effectiveToughness(t, g, id))
	}
	pushGraveyardCardWithTypeLine(g, me.ID, "Dead Bear", "Creature — Bear")
	pushGraveyardCardWithTypeLine(g, opp.ID, "Dead Elf", "Creature — Elf")
	pushGraveyardCardWithTypeLine(g, opp.ID, "Old Sorcery", "Sorcery")
	if effectivePower(t, g, id) != 2 || effectiveToughness(t, g, id) != 3 {
		t.Errorf("with two creature cards: %d/%d, want 2/3", effectivePower(t, g, id), effectiveToughness(t, g, id))
	}
}

func TestGapB44MagusOfTheCandelabraUntapsXLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	magus := b12Push(g, me.ID, "Magus of the Candelabra", "Creature — Human Wizard", gapb44MagusOracle, 1, 2)
	a := pushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	b := pushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	c := pushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{a, b, c} {
			_ = g.TapTargetForEffect(id)
		}
	})
	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, magus, 0, game.ActivateAbilityParams{XValue: 2, Targets: cardRefs(a), Strict: true}); err == nil {
		t.Fatal("X=2 with one target was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, magus, 0, game.ActivateAbilityParams{XValue: 2, Targets: cardRefs(a, b), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if isTapped(g, a) || isTapped(g, b) {
		t.Error("a targeted land stayed tapped")
	}
	if !isTapped(g, c) {
		t.Error("an untargeted land was untapped")
	}
}

func TestGapB44SageOfTheMazeAnimatesALandAndUntapsWithAGate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sage := b12Push(g, me.ID, "Sage of the Maze", "Creature — Elf Wizard", gapb44SageOracle, 1, 3)
	forest := pushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	gate1 := pushPermanentForTest(g, me.ID, "Gate One", "", "Land — Gate")
	pushPermanentForTest(g, me.ID, "Gate Two", "", "Land — Gate")
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, sage, 0, game.ActivateAbilityParams{Targets: cardRefs(forest), Strict: true}); err != nil {
		t.Fatalf("animate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, forest), effectiveToughness(t, g, forest); p != 4 || tg != 4 {
		t.Errorf("the land is %d/%d, want 4/4 (twice two Gates)", p, tg)
	}
	if !hasEffectiveKeyword(t, g, forest, "haste") {
		t.Error("the land has no haste")
	}
	if !isTapped(g, sage) {
		t.Fatal("Sage did not tap for its cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, sage, 1, game.ActivateAbilityParams{TapIDs: []uuid.UUID{gate1}, Strict: true}); err != nil {
		t.Fatalf("untap: %v", err)
	}
	passPriorityAroundTable(t, g)
	if isTapped(g, sage) {
		t.Error("Sage did not untap")
	}
	if !isTapped(g, gate1) {
		t.Error("the Gate was not tapped as the cost")
	}
}

func TestGapB44RampantFrogantuaGrowsForEachPlayerWhoLost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Rampant Frogantua", "Creature — Frog", gapb44FrogantuaOracle, 3, 3)
	if effectivePower(t, g, id) != 3 {
		t.Fatalf("power %d, want 3", effectivePower(t, g, id))
	}
	g.WithWriteLock(func() { g.Seats[1].Eliminated = true; g.Seats[2].Eliminated = true; g.BumpLayerVersionForTest() })
	if effectivePower(t, g, id) != 23 || effectiveToughness(t, g, id) != 23 {
		t.Errorf("with two players out: %d/%d, want 23/23", effectivePower(t, g, id), effectiveToughness(t, g, id))
	}
}

func TestGapB44RampantFrogantuaMillsAndPutsLandsOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	frog := b12Push(g, me.ID, "Rampant Frogantua", "Creature — Frog", gapb44FrogantuaOracle, 3, 3)
	forest := pushLibraryTyped(me, "Forest", "Basic Land — Forest")
	pushLibraryTyped(me, "Bear", "Creature — Bear")
	pushLibraryTyped(me, "Bear Two", "Creature — Bear")
	attackWith(t, g, opp.ID, frog)
	passPriorityAroundTable(t, g)
	// 3 damage: a "you may" mill of three.
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm {
			if err := g.ResolveConfirm(c.ID, me.ID, true); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			break
		}
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, forest)
	if !g.Battlefield.Contains(forest) {
		t.Fatal("the milled land was not put onto the battlefield")
	}
	if c, _ := battlefieldCard(g, forest); !c.Tapped {
		t.Error("the land entered untapped")
	}
}

func TestGapB44SireOfStagnationExilesTwoAndDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Sire of Stagnation", "Creature — Eldrazi", gapb44SireOracle, 5, 7)
	pushLibraryTyped(opp, "One", "Creature — Bear")
	pushLibraryTyped(opp, "Two", "Creature — Bear")
	lib, hand := opp.Library.Size(), me.Hand.Size()
	b34LandEntersByEffect(t, g, opp, "Opp Forest", "Basic Land — Forest")
	passPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != lib-2 {
		t.Errorf("their library %d -> %d, want 2 exiled", lib, got)
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d -> %d, want +2", hand, got)
	}
	// Their own lands trigger nothing.
	b34LandEntersByEffect(t, g, me, "My Forest", "Basic Land — Forest")
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand moved for my own land: %d", got)
	}
}

func TestGapB44VileRedeemerMakesScionsForNontokenDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := pushDiesCreatureForTest(g, me.ID, "Dead Bear", "", "Creature — Bear", 2, 2)
	token := b12Creature(g, me.ID, "Dead Token", "Token Creature — Bear", 1, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(dead); _ = g.DestroyPermanentForEffect(token) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Vile Redeemer", TypeLine: "Creature — Eldrazi", OracleID: gapb44RedeemerOracle, ManaCost: "{2}{G}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	floatForTest(g, me, "CCG")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	floatForTest(g, me, "C")
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Eldrazi Scion"); n != 1 {
		t.Errorf("%d Eldrazi Scions, want 1 (one nontoken creature died, a token does not count)", n)
	}
}

func TestGapB44MorophonDiscountsAndPumpsTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushNamedTribePermanent(t, g, me.ID, "Morophon, the Boundless", "Legendary Creature — Shapeshifter", gapb44MorophonOracle, "Bear")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	elf := b12Creature(g, me.ID, "My Elf", "Creature — Elf", 2, 2)
	if effectivePower(t, g, bear) != 3 || effectivePower(t, g, elf) != 2 {
		t.Errorf("bear %d, elf %d, want 3 and 2", effectivePower(t, g, bear), effectivePower(t, g, elf))
	}
	spell := handCardFull(me, "Test Bear Spell", "Creature — Bear", "{2}{W}{U}{B}{R}{G}", "", []string{"W", "U", "B", "R", "G"})
	floatForTest(g, me, "CC")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a Bear spell with {W}{U}{B}{R}{G} less: %v", err)
	}
}

func TestGapB44OmoPutsEverythingCountersAndChangesTypes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := pushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Omo, Queen of Vesuva", "Legendary Creature — Shapeshifter Noble", gapb44OmoOracle, 1, 5)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, forest)
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if counterOn(g, forest, "everything") != 1 || counterOn(g, bear, "everything") != 1 {
		t.Fatalf("counters: land %d, creature %d", counterOn(g, forest, "everything"), counterOn(g, bear, "everything"))
	}
	f, _ := battlefieldCard(g, forest)
	for _, want := range []string{"Gate", "Mountain", "Urza's"} {
		if !f.HasSubtype(want) {
			t.Errorf("the land is not a %s", want)
		}
	}
	b, _ := battlefieldCard(g, bear)
	if !b.HasSubtype("Elf") || !b.HasSubtype("Dragon") {
		t.Error("the creature is not every creature type")
	}
}

func TestGapB44UlamogsNullifierCountersASpellForTwoExiledCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 2; i++ {
		id := uuid.New()
		g.Exile.PushTop(game.Card{InstanceID: id, Name: "Their Exiled Card", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	}
	spell := castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Ulamog's Nullifier", TypeLine: "Creature — Eldrazi Processor", OracleID: gapb44NullifierOracle, Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast (flash): %v", err)
	}
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, spell)
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm {
			if err := g.ResolveConfirm(c.ID, me.ID, true); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			break
		}
	}
	passPriorityAroundTable(t, g)
	pickTwo := chooseCardsChoiceFor(g, me.ID)
	if pickTwo == nil {
		t.Fatal("no choose-cards prompt for the two exiled cards")
	}
	if err := g.ResolveChooseCards(pickTwo.ID, me.ID, pickTwo.ChooseCards[:2]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Stack.Contains(spell) {
		t.Error("the spell was not countered")
	}
	if opp.Graveyard.Size() != 2 {
		t.Errorf("%d cards in their graveyard, want 2", opp.Graveyard.Size())
	}
}

func TestGapB44ExterminatorMagmarchCopiesTheSpellForAnotherOpponentsPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	b12Push(g, me.ID, "Exterminator Magmarch", "Artifact Creature — Phyrexian Construct", gapb44MagmarchOracle, 5, 3)
	first := b12Permanent(g, opp1.ID, "First Relic", "Artifact")
	second := b12Permanent(g, opp2.ID, "Second Relic", "Artifact")
	castCatalogSpell(t, g, "Naturalize", "Instant", gapb44NaturalizeOracle, cardRefs(first))
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(first) || g.Battlefield.Contains(second) {
		t.Errorf("first still there: %v, second still there: %v, want both destroyed by the spell and its copy",
			g.Battlefield.Contains(first), g.Battlefield.Contains(second))
	}
}

func TestGapB44GluttonousHellkiteEntersWithCountersForSacrifices(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "My Fodder", "Creature — Bear", 1, 1)
	b := b12Creature(g, opp.ID, "Their Fodder", "Creature — Bear", 1, 1)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Gluttonous Hellkite", TypeLine: "Creature — Dragon", OracleID: gapb44HellkiteOracle, ManaCost: "{X}{X}{B}{R}{G}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	floatForTest(g, me, "CCBRG")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 1}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, me.ID, a)
	answerSacrifice(t, g, opp.ID, b)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Fatal("a creature was not sacrificed")
	}
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Hellkite did not enter")
	}
	if got := counterOn(g, id, game.CounterPlusOne); got != 4 {
		t.Errorf("%d +1/+1 counters, want 4 (two creatures sacrificed)", got)
	}
}

func TestGapB44UlalekCopiesEverythingForCC(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Ulalek, Fused Atrocity", "Legendary Creature — Eldrazi", gapb44UlalekOracle, 2, 5)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Eldrazi", TypeLine: "Creature — Eldrazi", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	floatForTest(g, me, "CC")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Test Eldrazi" {
			n++
		}
	}
	if n != 2 {
		t.Errorf("%d Test Eldrazi on the battlefield, want the spell and its copy", n)
	}
}
