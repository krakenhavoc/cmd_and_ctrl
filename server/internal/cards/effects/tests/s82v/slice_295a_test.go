package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Witch Enchanter ------------------------------------------------

func TestWitchEnchanterDestroysOnlyAnOpponentsArtifactOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushTypedCard(g, opp.ID, "Their Rock", "Artifact", "{1}")
	mine := effectstest.PushTypedCard(g, caster.ID, "My Rock", "Artifact", "{1}")
	theirCreature := effectstest.PushTypedCard(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")

	witchID := effectstest.CastAndResolveCreature(t, g, "Witch Enchanter", "Creature — Human Warlock", effectstest.WitchEnchanterOracle)
	prompt := effectstest.LatestPickTarget(g, caster.ID)
	if prompt == nil {
		t.Fatalf("no pick_target prompt")
	}
	if !effectstest.HasID(prompt.PickTargetCards, theirs) || effectstest.HasID(prompt.PickTargetCards, mine) || effectstest.HasID(prompt.PickTargetCards, theirCreature) {
		t.Errorf("legal set = %v (want only the opponent's artifact/enchantment)", prompt.PickTargetCards)
	}
	effectstest.PickCard(t, g, caster.ID, theirs)
	if effectstest.TriggerOnStack(g, witchID) == nil {
		t.Fatalf("trigger not on the stack after the pick")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Graveyard.Contains(theirs) {
		t.Errorf("Witch Enchanter did not destroy the chosen permanent")
	}
}

// --- Goldspan Dragon --------------------------------------------------

func TestGoldspanDragonCreatesATreasureWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goldspan Dragon", TypeLine: "Creature — Dragon",
		OracleID: effectstest.GoldspanDragonOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	before := effectstest.B16CountNamed(g, "Treasure")
	effectstest.AttackWith(t, g, opp.ID, dragon)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B16CountNamed(g, "Treasure"); got != before+1 {
		t.Errorf("Treasures = %d, want %d after attacking", got, before+1)
	}
}

func TestGoldspanDragonCreatesATreasureWhenTargetedByASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dragon := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goldspan Dragon", TypeLine: "Creature — Dragon",
		OracleID: effectstest.GoldspanDragonOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	before := effectstest.B16CountNamed(g, "Treasure")
	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", "5748ebf1-24e3-499d-ab7c-c2cebd462a24",
		[]game.TargetRef{{Kind: game.TargetCard, ID: dragon}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B16CountNamed(g, "Treasure"); got != before+1 {
		t.Errorf("Treasures = %d, want %d after being targeted by a spell", got, before+1)
	}
}

func TestGoldspanDragonUpgradesTreasuresToTwoMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goldspan Dragon", TypeLine: "Creature — Dragon",
		OracleID: effectstest.GoldspanDragonOracle, Power: 4, Toughness: 4, Owner: me, Controller: me,
	})
	treasure := effectstest.PushBattlefieldCardWithTimestamp(g, effects.TreasureToken())
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == treasure {
				g.Battlefield.Cards[i].Owner = me
				g.Battlefield.Cards[i].Controller = me
			}
		}
	})

	var abs []game.ManaAbilityShape
	var origins game.AbilityOrigins
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(treasure)
		abs, origins = game.ManaAbilitiesWithOrigins(c)
	})
	// The layer-6 grant ADDS an ability (CR 613.1f); it doesn't remove
	// the Treasure's own printed one. Own first, granted last (ADR
	// 0093 Decision 5).
	if len(abs) != 2 || abs[0].Produced != "{W|U|B|R|G}" || abs[1].Produced != effects.OneColorOfAmount(2) {
		t.Fatalf("Treasure mana = %+v, want the printed pick-one plus the upgraded two-of-one-color pipe", abs)
	}
	if err := g.ActivateManaAbility(me, treasure, 1, game.ManaAbilityParams{Ref: origins.Ref(1)}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick == nil || pick.ManaAmounts["G"] != 2 {
		t.Fatalf("pick = %+v, want amounts of two", pick)
	}
}

// --- Kami of Whispered Hopes ------------------------------------------

func TestKamiOfWhisperedHopesAddsOneExtraCounterOnYourPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	kami := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Kami of Whispered Hopes", TypeLine: "Creature — Spirit",
		OracleID: effectstest.KamiOfWhisperedHopesOracle, Power: 1, Toughness: 1, Owner: me, Controller: me,
	})
	bear := effectstest.B16Creature(g, me, "Bear", "Creature — Bear", 2, 2, "G")

	if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if got := effectstest.CountersOn(g, bear, "+1/+1"); got != 2 {
		t.Errorf("counters = %d, want 2 (1 requested + 1 from Kami)", got)
	}

	// A pump on the Kami itself feeds its own dynamic mana ability —
	// Kami started 1/1 and its own replacement doesn't touch a
	// counter nobody tried to place, so its power is still 1.
	if err := g.ActivateManaAbility(me, kami, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick == nil || !effectstest.ContainsString(pick.ColorOptions, "U") {
		t.Fatalf("pick = %+v, want U among the options", pick)
	}
}

// --- Lotho, Corrupt Shirriff --------------------------------------------

func TestLothoTriggersOnlyOnTheSecondSpellEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lotho := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Lotho, Corrupt Shirriff", TypeLine: "Legendary Creature — Halfling Rogue",
		OracleID: effectstest.LothoCorruptShirriffOracle, Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	lifeBefore := me.Life
	treasuresBefore := effectstest.B16CountNamed(g, "Treasure")
	giantGrowthOn := func() {
		effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", "5748ebf1-24e3-499d-ab7c-c2cebd462a24",
			[]game.TargetRef{{Kind: game.TargetCard, ID: lotho}})
		effectstest.PassPriorityAroundTable(t, g)
	}

	// First spell of the turn: no trigger.
	giantGrowthOn()
	if me.Life != lifeBefore || effectstest.B16CountNamed(g, "Treasure") != treasuresBefore {
		t.Fatalf("the first spell must not trigger Lotho")
	}

	// Second spell of the turn: triggers.
	giantGrowthOn()
	if got := me.Life; got != lifeBefore-1 {
		t.Errorf("life %d, want %d after the second spell", got, lifeBefore-1)
	}
	if got := effectstest.B16CountNamed(g, "Treasure"); got != treasuresBefore+1 {
		t.Errorf("Treasures %d, want %d after the second spell", got, treasuresBefore+1)
	}

	// Third spell of the turn: no further trigger.
	giantGrowthOn()
	if got := me.Life; got != lifeBefore-1 {
		t.Errorf("life %d, want %d — a third spell must not trigger again", got, lifeBefore-1)
	}
}

// --- Disciple of Freyalise --------------------------------------------

func TestDiscipleOfFreyaliseSacrificesForLifeAndCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	beast := effectstest.B16Creature(g, me.ID, "Beast", "Creature — Beast", 4, 4, "G")
	discipleID := effectstest.CastAndResolveCreature(t, g, "Disciple of Freyalise", "Creature — Elf Druid", effectstest.DiscipleOfFreyaliseOracle)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil {
		t.Fatal("the sacrifice is asked as the trigger resolves")
	}
	if effectstest.HasID(prompt.ChooseCards, discipleID) || !effectstest.HasID(prompt.ChooseCards, beast) {
		t.Errorf("ANOTHER creature you control: Disciple itself must not be offered")
	}
	lifeBefore := me.Life
	handBefore := me.Hand.Size()
	effectstest.AnswerOwnPermanents(t, g, me.ID, beast)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(beast) {
		t.Fatal("the chosen creature is sacrificed")
	}
	if got := me.Life; got != lifeBefore+4 {
		t.Errorf("life %d, want %d (gained the sacrificed creature's power)", got, lifeBefore+4)
	}
	if got := me.Hand.Size(); got != handBefore+4 {
		t.Errorf("hand %d, want %d (drew the sacrificed creature's power)", got, handBefore+4)
	}
}

func TestDiscipleOfFreyaliseDecliningSacrificesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	lifeBefore := me.Life
	effectstest.CastAndResolveCreature(t, g, "Disciple of Freyalise", "Creature — Elf Druid", effectstest.DiscipleOfFreyaliseOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || me.Life != lifeBefore {
		t.Error("declining sacrifices nothing and gains no life")
	}
}

// The choice is made as the trigger resolves, so a creature with hexproof
// is a legal pick: the sacrifice does not target.
func TestDiscipleOfFreyaliseCanSacrificeAHexproofCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	slippery := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Slippery", TypeLine: "Creature — Beast", Power: 3, Toughness: 3,
		Keywords: []string{"hexproof", "shroud"}, Owner: me.ID, Controller: me.ID,
	})
	lifeBefore := me.Life
	effectstest.CastAndResolveCreature(t, g, "Disciple of Freyalise", "Creature — Elf Druid", effectstest.DiscipleOfFreyaliseOracle)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || !effectstest.HasID(prompt.ChooseCards, slippery) {
		t.Fatal("a hexproof, shroud creature is offered")
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, slippery)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(slippery) || me.Life != lifeBefore+3 {
		t.Errorf("on battlefield %v, life %d, want gone and %d", g.Battlefield.Contains(slippery), me.Life, lifeBefore+3)
	}
}

// --- Faeburrow Elder --------------------------------------------------

func TestFaeburrowElderScalesWithColorsAmongYourPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	elder := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Faeburrow Elder", TypeLine: "Creature — Treefolk Druid",
		OracleID: effectstest.FaeburrowElderOracle, Power: 0, Toughness: 0, Colors: []string{"G", "W"},
		Owner: me, Controller: me,
	})
	// Faeburrow Elder itself contributes G and W; a red permanent adds
	// a third colour.
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Red Rock", TypeLine: "Artifact", Colors: []string{"R"},
		Owner: me, Controller: me,
	})
	if got := effectstest.EffectivePower(t, g, elder); got != 3 {
		t.Errorf("power %d, want 3 (G, W, R)", got)
	}
	if got := effectstest.EffectiveToughness(t, g, elder); got != 3 {
		t.Errorf("toughness %d, want 3", got)
	}
	if err := g.ActivateManaAbility(me, elder, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pool := g.Seats[0].ManaPool.Units()
	if len(pool) != 3 {
		t.Fatalf("pool = %+v, want three tokens (one per colour, no picker)", pool)
	}
}

// --- Pinnacle Monk ------------------------------------------------------

func TestPinnacleMonkReturnsOnlyAnInstantOrSorceryFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	bolt := uuid.New()
	caster.Graveyard.PushTop(game.Card{InstanceID: bolt, Name: "Bolt", TypeLine: "Instant", Owner: caster.ID, Controller: caster.ID})
	creature := uuid.New()
	caster.Graveyard.PushTop(game.Card{InstanceID: creature, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: caster.ID, Controller: caster.ID})
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Bolt", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})

	monkID := effectstest.CastAndResolveCreature(t, g, "Pinnacle Monk", "Creature — Djinn Monk", effectstest.PinnacleMonkOracle)
	prompt := effectstest.LatestPickTarget(g, caster.ID)
	if prompt == nil {
		t.Fatalf("no pick_target prompt")
	}
	if !effectstest.HasID(prompt.PickTargetCards, bolt) || effectstest.HasID(prompt.PickTargetCards, creature) || effectstest.HasID(prompt.PickTargetCards, theirs) {
		t.Errorf("legal set = %v (want only my instant/sorcery)", prompt.PickTargetCards)
	}
	effectstest.PickCard(t, g, caster.ID, bolt)
	if effectstest.TriggerOnStack(g, monkID) == nil {
		t.Fatalf("trigger not on the stack after the pick")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !caster.Hand.Contains(bolt) || caster.Graveyard.Contains(bolt) {
		t.Errorf("Pinnacle Monk did not return the chosen card to hand")
	}
	if ab := effectstest.EffectiveAbilities(t, g, monkID); !effectstest.ContainsString(ab, "prowess") {
		t.Errorf("abilities %v missing prowess", ab)
	}
}

// --- Ignoble Hierarch / exalted -----------------------------------------

func TestIgnobleHierarchExaltedPumpsALoneAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hierarch := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ignoble Hierarch", TypeLine: "Creature — Goblin Shaman",
		OracleID: effectstest.IgnobleHierarchOracle, Power: 0, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.SeedBear(g, me.ID)

	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3 (2 base + 1 exalted)", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 3 {
		t.Errorf("toughness %d, want 3", got)
	}
	_ = hierarch

	if err := g.ActivateManaAbility(me.ID, hierarch, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatalf("no mana_pick prompt")
	}
	for _, c := range []string{"B", "R", "G"} {
		if !effectstest.ContainsString(pick.ColorOptions, c) {
			t.Errorf("options = %v, want B/R/G", pick.ColorOptions)
		}
	}
}

func TestIgnobleHierarchExaltedDoesNotFireAlongsideAnotherAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ignoble Hierarch", TypeLine: "Creature — Goblin Shaman",
		OracleID: effectstest.IgnobleHierarchOracle, Power: 0, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.SeedBear(g, me.ID)
	other := effectstest.SeedBear(g, me.ID)

	effectstest.AttackWith(t, g, opp.ID, bear, other)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("power %d, want 2 — attacking alongside another creature must not trigger exalted", got)
	}
}

// --- Noble Hierarch ------------------------------------------------------

func TestNobleHierarchExaltedAndItsManaPipe(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hierarch := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Noble Hierarch", TypeLine: "Creature — Human Druid",
		OracleID: effectstest.NobleHierarchOracle, Power: 0, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3 (2 base + 1 exalted)", got)
	}

	if err := g.ActivateManaAbility(me.ID, hierarch, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatalf("no mana_pick prompt")
	}
	for _, c := range []string{"G", "W", "U"} {
		if !effectstest.ContainsString(pick.ColorOptions, c) {
			t.Errorf("options = %v, want G/W/U", pick.ColorOptions)
		}
	}
}

// --- Hydroelectric Specimen ----------------------------------------------

// The redirect itself is pinned_retarget_test.go's (#1743); this is
// the face's printed keyword, and that nothing is left caveated.
func TestHydroelectricSpecimenHasFlashAndIsComplete(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.HydroelectricSpecimenOracle)
	if !ok {
		t.Fatalf("Hydroelectric Specimen is not registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("the redirect is built since #1743; completeness = %v, caveats = %v", spec.Completeness, spec.Caveats)
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hydroelectric Specimen", TypeLine: "Creature — Weird",
		OracleID: effectstest.HydroelectricSpecimenOracle, Power: 1, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	if ab := effectstest.EffectiveAbilities(t, g, id); !effectstest.ContainsString(ab, "flash") {
		t.Errorf("abilities %v missing flash", ab)
	}
}

// --- Boggart Trawler ------------------------------------------------------

func TestBoggartTrawlerExilesTheChosenPlayersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushGraveyardCardForTest(opp, "Their Body")
	mine := effectstest.PushGraveyardCardForTest(caster, "My Body")

	effectstest.CastAndResolveCreature(t, g, "Boggart Trawler", "Creature — Goblin", effectstest.BoggartTrawlerOracle)
	prompt := effectstest.LatestPickTarget(g, caster.ID)
	if prompt == nil {
		t.Fatalf("no pick_target prompt for the player choice")
	}
	effectstest.PickPlayer(t, g, caster.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(theirs) || !g.Exile.Contains(theirs) {
		t.Errorf("Boggart Trawler did not exile the chosen player's graveyard")
	}
	if !caster.Graveyard.Contains(mine) {
		t.Errorf("the caster's own graveyard must be untouched")
	}
}

// --- Sanctum Weaver ------------------------------------------------------

func TestSanctumWeaverScalesWithEnchantmentsYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	weaver := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sanctum Weaver", TypeLine: "Enchantment Creature — Dryad",
		OracleID: effectstest.SanctumWeaverOracle, Power: 0, Toughness: 2, Owner: me, Controller: me,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Some Enchantment", TypeLine: "Enchantment", Owner: me, Controller: me,
	})
	if err := g.ActivateManaAbility(me, weaver, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	// Sanctum Weaver is itself an enchantment (Enchantment Creature),
	// plus the other one: two.
	if pick == nil || pick.ManaAmounts["G"] != 2 {
		t.Fatalf("pick = %+v, want amounts of two (Weaver + one enchantment)", pick)
	}
}

// --- Incubation Druid ------------------------------------------------------

func TestIncubationDruidManaTriplesWithACounterAndAdaptsOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	druid := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Incubation Druid", TypeLine: "Creature — Elf Druid",
		OracleID: effectstest.IncubationDruidOracle, Power: 0, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID,
	})
	untapAndEmptyPool := func() {
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == druid {
					g.Battlefield.Cards[i].Tapped = false
				}
			}
			me.ManaPool.EmptyPool()
		})
	}

	if err := g.ActivateManaAbility(me.ID, druid, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick == nil || !effectstest.ContainsString(pick.ColorOptions, "G") || !effectstest.ContainsString(pick.ColorOptions, "U") {
		t.Fatalf("pick = %+v, want G and U among the options with no counter (amount 1, implicit)", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Fatalf("pool = %+v, want one green token", me.ManaPool)
	}
	untapAndEmptyPool()

	// Adapt: no counters yet, so it puts three on. Seed enough mana to
	// pay {3}{G}{G} directly — the pool is not this test's subject.
	for i := 0; i < 3; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	for i := 0; i < 2; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	}
	if err := g.ActivateCatalogAbility(me.ID, druid, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Adapt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, druid)
	if !ok || c.Counters["+1/+1"] != 3 {
		t.Fatalf("counters = %+v, want three +1/+1 counters", c.Counters)
	}
	untapAndEmptyPool()

	if err := g.ActivateManaAbility(me.ID, druid, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility after adapt: %v", err)
	}
	pick2 := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick2 == nil || pick2.ManaAmounts["G"] != 3 || pick2.ManaAmounts["U"] != 3 {
		t.Fatalf("pick = %+v, want three each of G and U with the counter", pick2)
	}
	if err := g.ResolveManaChoice(pick2.ID, me.ID, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if me.ManaPool.Total() != 3 {
		t.Fatalf("pool = %+v, want three green tokens", me.ManaPool)
	}
	untapAndEmptyPool()

	// Adapt again: already has a +1/+1 counter, so it does nothing.
	for i := 0; i < 3; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	for i := 0; i < 2; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	}
	if err := g.ActivateCatalogAbility(me.ID, druid, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Adapt (second time): %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c2, _ := effectstest.BattlefieldCard(g, druid)
	if c2.Counters["+1/+1"] != 3 {
		t.Errorf("counters = %+v, adapt with an existing counter must be a no-op", c2.Counters)
	}
}

// --- Priest of Titania ------------------------------------------------------

func TestPriestOfTitaniaCountsEveryElfOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	priest := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Priest of Titania", TypeLine: "Creature — Elf Druid",
		OracleID: effectstest.PriestOfTitaniaOracle, Power: 1, Toughness: 1, Owner: me, Controller: me,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Elf", TypeLine: "Creature — Elf Warrior", Power: 1, Toughness: 1, Owner: me, Controller: me,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Elf", TypeLine: "Creature — Elf Warrior", Power: 1, Toughness: 1, Owner: opp, Controller: opp,
	})

	if err := g.ActivateManaAbility(me, priest, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pool := g.Seats[0].ManaPool.Units()
	// Priest herself, "My Elf" and the opponent's Elf: three, straight
	// into the pool with no picker (fixed colour).
	if len(pool) != 3 {
		t.Fatalf("pool = %+v, want three green tokens (every Elf on the battlefield)", pool)
	}
	for _, tok := range pool {
		if tok.Color != "G" {
			t.Errorf("token %+v, want green", tok)
		}
	}
}

// --- Urza, Lord High Artificer -------------------------------------------

func TestUrzaCreatesAScalingConstructAndTapsArtifactsForMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	urza := effectstest.CastAndResolveCreature(t, g, "Urza, Lord High Artificer", "Legendary Creature — Human Artificer", effectstest.UrzaLordHighArtificerOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Construct"); n != 1 {
		t.Fatalf("Constructs = %d, want 1", n)
	}
	var construct uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Construct" {
			construct = c.InstanceID
		}
	}
	// Two artifacts on the board right now: Urza is a creature, not an
	// artifact, so the Construct itself is the only artifact until
	// another joins it.
	if got := effectstest.EffectivePower(t, g, construct); got != 1 {
		t.Errorf("Construct power %d, want 1 (itself)", got)
	}
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mana Rock", TypeLine: "Artifact", Owner: me, Controller: me,
	})
	if got := effectstest.EffectivePower(t, g, construct); got != 2 {
		t.Errorf("Construct power %d, want 2 (itself + the rock)", got)
	}

	// The mana ability taps ANOTHER untapped artifact, not Urza.
	if err := g.ActivateManaAbility(me, urza, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pool := g.Seats[0].ManaPool.Units()
	if len(pool) != 1 || pool[0].Color != "U" {
		t.Fatalf("pool = %+v, want one blue token", pool)
	}
	rockCard, _ := effectstest.BattlefieldCard(g, rock)
	if !rockCard.Tapped {
		t.Error("the tapped artifact should be the rock, not Urza")
	}
	urzaCard, _ := effectstest.BattlefieldCard(g, urza)
	if urzaCard.Tapped {
		t.Error("Urza itself does not tap to pay this ability")
	}
}

func TestUrzaFreeCastAbilityShufflesExilesAndGrantsPermission(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	urza := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Urza, Lord High Artificer", TypeLine: "Legendary Creature — Human Artificer",
		OracleID: effectstest.UrzaLordHighArtificerOracle, Power: 1, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	for i := 0; i < 5; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	exileBefore := map[uuid.UUID]bool{}
	for _, c := range g.Exile.Cards {
		exileBefore[c.InstanceID] = true
	}

	if err := g.ActivateCatalogAbility(me.ID, urza, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility (the {5} ability): %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.Exile.Cards) != len(exileBefore)+1 {
		t.Fatalf("exile has %d cards, want %d", len(g.Exile.Cards), len(exileBefore)+1)
	}
	var exiled uuid.UUID
	for _, c := range g.Exile.Cards {
		if !exileBefore[c.InstanceID] {
			exiled = c.InstanceID
		}
	}
	if exiled == uuid.Nil {
		t.Fatalf("the shuffled-then-exiled card is not in exile")
	}
	perm := g.CastPermissionOnCardByIDForEffect(exiled)
	if perm == nil || perm.Player != me.ID || perm.Cost != "{0}" {
		t.Errorf("no free-play permission for Urza's controller: %+v", perm)
	}
}
