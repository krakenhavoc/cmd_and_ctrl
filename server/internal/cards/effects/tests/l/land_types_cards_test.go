package testsl

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ltActivate activates the source's ability `index` for its controller,
// then resolves it.
func ltActivate(t *testing.T, g *game.Game, controller, source uuid.UUID, index int, params game.ActivateAbilityParams) {
	t.Helper()
	if err := g.ActivateCatalogAbility(controller, source, index, params); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// TestMagusOfTheMoonLeavesDryadArborItsDryad is ADR 0109 §1 decision 3:
// the static form replaces land types only (CR 205.1a), so the Dryad
// stays. On develop the whole subtype list was replaced.
func TestMagusOfTheMoonLeavesDryadArborItsDryad(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	arbor := effectstest.PushLandFor(g, me.ID, "Dryad Arbor", "Land Creature — Forest Dryad")
	effectstest.PushCatalogPermanent(g, me.ID, "Magus of the Moon", "Creature — Human Wizard", effectstest.LtMagusOfTheMoonOracle, false)
	if got := effectstest.EffectiveSubtypes(t, g, arbor); !reflect.DeepEqual(got, []string{"Dryad", "Mountain"}) {
		t.Errorf("Dryad Arbor under Magus of the Moon: subtypes %v, want [Dryad Mountain]", got)
	}
	if got := effectstest.LtProduced(t, g, arbor); !reflect.DeepEqual(got, []string{"{R}"}) {
		t.Errorf("Dryad Arbor under Magus of the Moon taps for %v, want {R}", got)
	}
}

// TestTidalWarriorMakesALandAnIslandUntilEndOfTurn is the seam's
// headline: CR 305.7 from a resolved ability, gone at cleanup.
func TestTidalWarriorMakesALandAnIslandUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warrior := effectstest.PushCatalogPermanent(g, me.ID, "Tidal Warrior", "Creature — Merfolk Warrior", effectstest.LtTidalWarriorOracle, false)
	forest := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	ltActivate(t, g, me.ID, warrior, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(forest)})

	if got := effectstest.EffectiveSubtypes(t, g, forest); !reflect.DeepEqual(got, []string{"Island"}) {
		t.Errorf("subtypes %v, want [Island]", got)
	}
	if got := effectstest.LtProduced(t, g, forest); !reflect.DeepEqual(got, []string{"{U}"}) {
		t.Errorf("taps for %v, want {U}", got)
	}
	if !effectstest.LayeredCard(t, g, forest).HasSupertype("basic") {
		t.Error("the land stopped being basic: CR 305.7 leaves supertypes alone")
	}
	effectstest.AdvanceOneTurnForTest(t, g)
	if got := effectstest.EffectiveSubtypes(t, g, forest); !reflect.DeepEqual(got, []string{"Forest"}) {
		t.Errorf("next turn: subtypes %v, want the Forest back", got)
	}
}

// TestTidalWarriorTakesTheRulesTextAway is CR 305.7's second clause: the
// land loses the abilities its rules text gives it. Unstable Frontier's
// own {C} and its activated ability both go.
func TestTidalWarriorTakesTheRulesTextAway(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	warrior := effectstest.PushCatalogPermanent(g, me.ID, "Tidal Warrior", "Creature — Merfolk Warrior", effectstest.LtTidalWarriorOracle, false)
	frontier := effectstest.PushCatalogPermanent(g, me.ID, "Unstable Frontier", "Land", effectstest.LtUnstableFrontierOracle, false)
	if got := effectstest.LtProduced(t, g, frontier); !reflect.DeepEqual(got, []string{"{C}"}) {
		t.Fatalf("setup: Unstable Frontier taps for %v", got)
	}
	ltActivate(t, g, me.ID, warrior, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(frontier)})
	if got := effectstest.LtProduced(t, g, frontier); !reflect.DeepEqual(got, []string{"{U}"}) {
		t.Errorf("an Island Unstable Frontier taps for %v, want only {U}", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, frontier, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(frontier)}); err == nil {
		t.Error("the Frontier's own activated ability survived CR 305.7")
	}
}

// TestReefShamanAsksAsItResolves is "the basic land type of your
// choice": the choice is made at resolution, and Dryad Arbor keeps its
// creature type and its card types (CR 205.1a, 305.7).
func TestReefShamanAsksAsItResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shaman := effectstest.PushCatalogPermanent(g, me.ID, "Reef Shaman", "Creature — Merfolk Shaman", effectstest.LtReefShamanOracle, false)
	arbor := effectstest.PushLandFor(g, me.ID, "Dryad Arbor", "Land Creature — Forest Dryad")
	ltActivate(t, g, me.ID, shaman, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(arbor)})
	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no basic land type was asked for")
	}
	var labels []string
	for _, o := range pick.PickOptions {
		labels = append(labels, o.Label)
	}
	if !reflect.DeepEqual(labels, game.BasicLandTypes) {
		t.Errorf("options %v, want the five basic land types", labels)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 2) // Swamp
	c := effectstest.LayeredCard(t, g, arbor)
	if got := c.Effective().Subtypes; !reflect.DeepEqual(got, []string{"Dryad", "Swamp"}) {
		t.Errorf("subtypes %v, want [Dryad Swamp]", got)
	}
	if !c.IsCreature() || !c.IsLand() {
		t.Error("Dryad Arbor lost a card type")
	}
}

func TestTundraKavuOffersPlainsOrIsland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kavu := effectstest.PushCatalogPermanent(g, me.ID, "Tundra Kavu", "Creature — Kavu", effectstest.LtTundraKavuOracle, false)
	swamp := effectstest.PushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	ltActivate(t, g, me.ID, kavu, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(swamp)})
	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if pick == nil || len(pick.PickOptions) != 2 || pick.PickOptions[0].Label != "Plains" || pick.PickOptions[1].Label != "Island" {
		t.Fatalf("pick %+v, want Plains or Island", pick)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	if got := effectstest.EffectiveSubtypes(t, g, swamp); !reflect.DeepEqual(got, []string{"Plains"}) {
		t.Errorf("subtypes %v, want [Plains]", got)
	}
}

// TestTerraformerFixesItsSetAsItBegins is CR 611.2c: a land you control
// afterwards is not changed.
func TestTerraformerFixesItsSetAsItBegins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tf := effectstest.PushCatalogPermanent(g, me.ID, "Terraformer", "Creature — Human Wizard", effectstest.LtTerraformerOracle, false)
	mine := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	ltActivate(t, g, me.ID, tf, 0, game.ActivateAbilityParams{})
	effectstest.AnswerOptionPick(t, g, me.ID, 1) // Island
	later := effectstest.PushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	if got := effectstest.EffectiveSubtypes(t, g, mine); !reflect.DeepEqual(got, []string{"Island"}) {
		t.Errorf("your land: %v, want [Island]", got)
	}
	if got := effectstest.EffectiveSubtypes(t, g, theirs); !reflect.DeepEqual(got, []string{"Forest"}) {
		t.Errorf("an opponent's land changed: %v", got)
	}
	if got := effectstest.EffectiveSubtypes(t, g, later); !reflect.DeepEqual(got, []string{"Plains"}) {
		t.Errorf("a land that arrived later changed: %v", got)
	}
}

func TestNightcreepMakesCreaturesBlackAndLandsSwamps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushTribalCreature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	island := effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	_ = me
	effectstest.CastCatalogSpell(t, g, "Nightcreep", "Instant", effectstest.LtNightcreepOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectiveOf(t, g, bear).Colors; !reflect.DeepEqual(got, []string{"B"}) {
		t.Errorf("creature colours %v, want [B]", got)
	}
	if got := effectstest.EffectiveSubtypes(t, g, island); !reflect.DeepEqual(got, []string{"Swamp"}) {
		t.Errorf("land subtypes %v, want [Swamp]", got)
	}
}

// TestNavigatorsCompassAddsATypeAndKeepsEverything is CR 205.1b.
func TestNavigatorsCompassAddsATypeAndKeepsEverything(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	compass := effectstest.PushCatalogPermanent(g, me.ID, "Navigator's Compass", "Artifact", effectstest.LtNavigatorsCompassOracle, false)
	frontier := effectstest.PushCatalogPermanent(g, me.ID, "Unstable Frontier", "Land", effectstest.LtUnstableFrontierOracle, false)
	ltActivate(t, g, me.ID, compass, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(frontier)})
	effectstest.AnswerOptionPick(t, g, me.ID, 4) // Forest
	if got := effectstest.LtProduced(t, g, frontier); !reflect.DeepEqual(got, []string{"{C}", "{G}"}) {
		t.Errorf("taps for %v, want its own {C} and the Forest's {G}", got)
	}
}

func TestEnergybendingGivesYourLandsEveryBasicLandType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushLandFor(g, me.ID, "Wastes", "Basic Land")
	theirs := effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.CastCatalogSpell(t, g, "Energybending", "Instant — Lesson", effectstest.LtEnergybendingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LtProduced(t, g, mine); len(got) != 5 {
		t.Errorf("your land taps for %v, want all five colours", got)
	}
	if got := effectstest.EffectiveSubtypes(t, g, theirs); !reflect.DeepEqual(got, []string{"Island"}) {
		t.Errorf("an opponent's land changed: %v", got)
	}
}

// TestVisionCharmAsksForAnyLandTypeThenABasicOne is decision 4's two
// questions, the first over CR 205.3i's seventeen.
func TestVisionCharmAsksForAnyLandTypeThenABasicOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gate := effectstest.PushLandFor(g, opp.ID, "Azorius Guildgate", "Land — Gate")
	forest := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpellWithModes(t, g, "Vision Charm", "Instant", effectstest.LtVisionCharmOracle, []int{1})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if pick == nil || len(pick.PickOptions) != len(game.LandTypes) {
		t.Fatalf("first pick %+v, want the seventeen land types", pick)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 3) // Gate
	effectstest.AnswerOptionPick(t, g, me.ID, 2) // Swamp
	if got := effectstest.EffectiveSubtypes(t, g, gate); !reflect.DeepEqual(got, []string{"Swamp"}) {
		t.Errorf("Gate: %v, want [Swamp]", got)
	}
	if got := effectstest.EffectiveSubtypes(t, g, forest); !reflect.DeepEqual(got, []string{"Forest"}) {
		t.Errorf("a land of another type changed: %v", got)
	}
}

// TestDeepwoodElderTargetsXLands is TargetSpec.CountFromX on an
// activated ability.
func TestDeepwoodElderTargetsXLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elder := effectstest.PushCatalogPermanent(g, me.ID, "Deepwood Elder", "Creature — Dryad Spellshaper", effectstest.LtDeepwoodElderOracle, false)
	a := effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	b := effectstest.PushLandFor(g, opp.ID, "Swamp", "Basic Land — Swamp")
	c := effectstest.PushLandFor(g, opp.ID, "Plains", "Basic Land — Plains")
	fodder := game.NewCard("fodder", me.ID)
	me.Hand.PushTop(fodder)
	ltActivate(t, g, me.ID, elder, 0, game.ActivateAbilityParams{
		XValue: 2, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
		DiscardIDs: []uuid.UUID{fodder.InstanceID},
	})
	for _, id := range []uuid.UUID{a, b} {
		if got := effectstest.EffectiveSubtypes(t, g, id); !reflect.DeepEqual(got, []string{"Forest"}) {
			t.Errorf("target: %v, want [Forest]", got)
		}
	}
	if got := effectstest.EffectiveSubtypes(t, g, c); !reflect.DeepEqual(got, []string{"Plains"}) {
		t.Errorf("an untargeted land changed: %v", got)
	}
}

// TestGaeasLiegeForestLastsWhileItRemains is "until this creature leaves
// the battlefield", and the Liege's CDA counting the new Forest.
func TestGaeasLiegeForestLastsWhileItRemains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	liege := effectstest.PushCatalogPermanent(g, me.ID, "Gaea's Liege", "Creature — Avatar", effectstest.LtGaeasLiegeOracle, false)
	effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	swamp := effectstest.PushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	if p := effectstest.EffectiveOf(t, g, liege).Power; p != 1 {
		t.Fatalf("setup: Liege power %d, want 1", p)
	}
	ltActivate(t, g, me.ID, liege, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(swamp)})
	if got := effectstest.EffectiveSubtypes(t, g, swamp); !reflect.DeepEqual(got, []string{"Forest"}) {
		t.Fatalf("subtypes %v, want [Forest]", got)
	}
	if p := effectstest.EffectiveOf(t, g, liege).Power; p != 2 {
		t.Errorf("Liege power %d, want 2", p)
	}
	effectstest.AdvanceOneTurnForTest(t, g)
	if got := effectstest.EffectiveSubtypes(t, g, swamp); !reflect.DeepEqual(got, []string{"Forest"}) {
		t.Errorf("after a turn: %v, want still a Forest", got)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(liege); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if got := effectstest.EffectiveSubtypes(t, g, swamp); !reflect.DeepEqual(got, []string{"Swamp"}) {
		t.Errorf("after the Liege left: %v, want the Swamp back", got)
	}
}

// TestOrcishFarmerLastsUntilTheLandsControllerUntaps is "until its
// controller's next untap step".
func TestOrcishFarmerLastsUntilTheLandsControllerUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[2]
	farmer := effectstest.PushCatalogPermanent(g, me.ID, "Orcish Farmer", "Creature — Orc", effectstest.LtOrcishFarmerOracle, false)
	plains := effectstest.PushLandFor(g, opp.ID, "Plains", "Basic Land — Plains")
	ltActivate(t, g, me.ID, farmer, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(plains)})
	effectstest.AdvanceOneTurnForTest(t, g) // seat 1's turn
	if got := effectstest.EffectiveSubtypes(t, g, plains); !reflect.DeepEqual(got, []string{"Swamp"}) {
		t.Errorf("on another player's turn: %v, want still a Swamp", got)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 2)
	if got := effectstest.EffectiveSubtypes(t, g, plains); !reflect.DeepEqual(got, []string{"Plains"}) {
		t.Errorf("in its controller's turn: %v, want the Plains back", got)
	}
}

func TestTheloniteMonkMakesAForestForGood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	monk := effectstest.PushCatalogPermanent(g, me.ID, "Thelonite Monk", "Creature — Insect Monk Cleric", effectstest.LtTheloniteMonkOracle, false)
	elf := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Llanowar Elves",
		TypeLine: "Creature — Elf Druid", Colors: []string{"G"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	island := effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	ltActivate(t, g, me.ID, monk, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(island), SacrificeIDs: []uuid.UUID{elf}})
	effectstest.AdvanceOneTurnForTest(t, g)
	effectstest.AdvanceOneTurnForTest(t, g)
	if got := effectstest.EffectiveSubtypes(t, g, island); !reflect.DeepEqual(got, []string{"Forest"}) {
		t.Errorf("two turns later: %v, want still a Forest", got)
	}
}

// TestCyclopeanGiantDiesThenIsExiled is the trigger's two sentences.
func TestCyclopeanGiantDiesThenIsExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := effectstest.PushCatalogPermanent(g, me.ID, "Cyclopean Giant", "Creature — Zombie Giant", effectstest.LtCyclopeanGiantOracle, false)
	forest := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(giant); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	effectstest.AnswerPickTarget(t, g, forest)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectiveSubtypes(t, g, forest); !reflect.DeepEqual(got, []string{"Swamp"}) {
		t.Errorf("land: %v, want [Swamp]", got)
	}
	if z := g.FindCardZoneForEffect(giant); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the Giant is in %v, want exile", z)
	}
}

// TestTideShaperOnlyWhenKicked is the intervening if, and the land
// going back when the Shaper leaves.
func TestTideShaperOnlyWhenKicked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plains := effectstest.PushLandFor(g, opp.ID, "Plains", "Basic Land — Plains")
	shaper, err := effectstest.CastWithOptionalCosts(t, g, "Tide Shaper", "Creature — Merfolk Wizard", effectstest.LtTideShaperOracle, nil, []int{0}, nil)
	if err != nil {
		t.Fatalf("cast kicked: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, plains)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectiveSubtypes(t, g, plains); !reflect.DeepEqual(got, []string{"Island"}) {
		t.Fatalf("kicked: %v, want [Island]", got)
	}
	// The fixture card carries no printed power, so the +1/+1 is all of
	// it.
	if p := effectstest.EffectiveOf(t, g, shaper).Power; p != 1 {
		t.Errorf("Tide Shaper power %d with an opponent's Island, want 0+1", p)
	}
	_ = me
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(shaper); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if got := effectstest.EffectiveSubtypes(t, g, plains); !reflect.DeepEqual(got, []string{"Plains"}) {
		t.Errorf("after it left: %v, want [Plains]", got)
	}

	// Unkicked, nothing triggers (CR 603.4).
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Tide Shaper", "Creature — Merfolk Wizard", effectstest.LtTideShaperOracle, nil, nil, nil); err != nil {
		t.Fatalf("cast unkicked: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatal("an unkicked Tide Shaper asked for a target")
		}
	}
}

func TestShimmeringMirageDrawsAfterTheChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Shimmering Mirage", "Instant", effectstest.LtShimmeringMirageOracle, effectstest.LtCardTarget(forest))
	hand := len(me.Hand.Cards)
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand {
		t.Fatalf("drew before choosing: hand %d, want %d", len(me.Hand.Cards), hand)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d after the choice, want %d", len(me.Hand.Cards), hand+1)
	}
	if got := effectstest.EffectiveSubtypes(t, g, forest); !reflect.DeepEqual(got, []string{"Island"}) {
		t.Errorf("%v, want [Island]", got)
	}
}

// TestSpreadingSeasIsTheStaticForm is decision 4: the Aura on a land.
func TestSpreadingSeasIsTheStaticForm(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := g.Seats[0], g.Seats[1]
	arbor := effectstest.PushLandFor(g, opp.ID, "Dryad Arbor", "Land Creature — Forest Dryad")
	effectstest.Enchant(t, g, "Spreading Seas", effectstest.LtSpreadingSeasOracle, arbor)
	effectstest.PassPriorityAroundTable(t, g) // the draw trigger
	if got := effectstest.EffectiveSubtypes(t, g, arbor); !reflect.DeepEqual(got, []string{"Dryad", "Island"}) {
		t.Errorf("%v, want [Dryad Island]", got)
	}
}

func TestLushGrowthIsThreeTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	island := effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	effectstest.Enchant(t, g, "Lush Growth", effectstest.LtLushGrowthOracle, island)
	if got := effectstest.LtProduced(t, g, island); !reflect.DeepEqual(got, []string{"{G}", "{R}", "{W}"}) {
		t.Errorf("taps for %v, want {R}, {G} and {W}", got)
	}
}

func TestContaminatedGroundDrainsWhenTheLandTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.Enchant(t, g, "Contaminated Ground", effectstest.LtContaminatedOracle, land)
	life := opp.Life
	if err := g.ActivateManaAbility(opp.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
}

func TestConvincingMirageSetsTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Convincing Mirage", effectstest.AuraTypeLine, effectstest.LtConvincingMirageOracle, effectstest.LtCardTarget(forest))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectiveSubtypes(t, g, forest); !reflect.DeepEqual(got, []string{"Forest"}) {
		t.Errorf("before the choice: %v, want unchanged", got)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 3) // Mountain
	if got := effectstest.EffectiveSubtypes(t, g, forest); !reflect.DeepEqual(got, []string{"Mountain"}) {
		t.Errorf("%v, want [Mountain]", got)
	}
}

func TestIllusionaryTerrainTurnsOneBasicTypeIntoAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	nonbasic := effectstest.PushLandFor(g, opp.ID, "Tropical Island", "Land — Forest Island")
	effectstest.CastCatalogSpell(t, g, "Illusionary Terrain", "Enchantment", effectstest.LtIllusionaryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if pick == nil || len(pick.PickOptions) != 20 {
		t.Fatalf("pick %+v, want twenty ordered pairs", pick)
	}
	idx := -1
	for i, o := range pick.PickOptions {
		if o.Label == "Island, then Swamp" {
			idx = i
		}
	}
	effectstest.AnswerOptionPick(t, g, me.ID, idx)
	if got := effectstest.EffectiveSubtypes(t, g, theirs); !reflect.DeepEqual(got, []string{"Swamp"}) {
		t.Errorf("basic Island: %v, want [Swamp]", got)
	}
	if got := effectstest.EffectiveSubtypes(t, g, nonbasic); !reflect.DeepEqual(got, []string{"Forest", "Island"}) {
		t.Errorf("a nonbasic Island changed: %v", got)
	}
}

func TestBloodMoonLeavesBasicsAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Blood Moon", "Enchantment", effectstest.LtBloodMoonOracle, false)
	dual := effectstest.PushLandFor(g, me.ID, "Tropical Island", "Land — Forest Island")
	basic := effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	if got := effectstest.EffectiveSubtypes(t, g, dual); !reflect.DeepEqual(got, []string{"Mountain"}) {
		t.Errorf("nonbasic: %v, want [Mountain]", got)
	}
	if got := effectstest.EffectiveSubtypes(t, g, basic); !reflect.DeepEqual(got, []string{"Island"}) {
		t.Errorf("basic: %v, want [Island]", got)
	}
}

func TestKukemssaSerpentTargetsOnlyAnOpponentsLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	serpent := effectstest.PushCatalogPermanent(g, me.ID, "Kukemssa Serpent", "Creature — Serpent", effectstest.LtKukemssaOracle, false)
	myIsland := effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	otherIsland := effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	mine := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	if err := g.ActivateCatalogAbility(me.ID, serpent, 0, game.ActivateAbilityParams{
		Targets: effectstest.LtCardTarget(mine), SacrificeIDs: []uuid.UUID{myIsland}}); err == nil {
		t.Error("targeted your own land")
	}
	ltActivate(t, g, me.ID, serpent, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(theirs), SacrificeIDs: []uuid.UUID{myIsland}})
	if got := effectstest.EffectiveSubtypes(t, g, theirs); !reflect.DeepEqual(got, []string{"Island"}) {
		t.Errorf("%v, want [Island]", got)
	}
	_ = otherIsland
}

func TestZombieTrailblazerTapsItselfToPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := effectstest.PushCatalogPermanent(g, me.ID, "Zombie Trailblazer", "Creature — Zombie Scout", effectstest.LtZombieTrailblazerOracle, true)
	forest := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	ltActivate(t, g, me.ID, zombie, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(forest), TapIDs: []uuid.UUID{zombie}})
	if got := effectstest.EffectiveSubtypes(t, g, forest); !reflect.DeepEqual(got, []string{"Swamp"}) {
		t.Errorf("%v, want [Swamp]", got)
	}
	if !effectstest.LayeredCard(t, g, zombie).Tapped {
		t.Error("the Trailblazer paid without tapping")
	}
}

func TestDreamwinderSacrificesAnIsland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	serpent := effectstest.PushCatalogPermanent(g, me.ID, "Dreamwinder", "Creature — Serpent", effectstest.LtDreamwinderOracle, false)
	island := effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	theirs := effectstest.PushLandFor(g, opp.ID, "Plains", "Basic Land — Plains")
	ltActivate(t, g, me.ID, serpent, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(theirs), SacrificeIDs: []uuid.UUID{island}})
	if got := effectstest.EffectiveSubtypes(t, g, theirs); !reflect.DeepEqual(got, []string{"Island"}) {
		t.Errorf("%v, want [Island]", got)
	}
}

// TestALandTypeRecordIsARestorePoint is the snapshot half: a table
// holding a setBasicLandTypes record restores and still applies it.
func TestALandTypeRecordIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warrior := effectstest.PushCatalogPermanent(g, me.ID, "Tidal Warrior", "Creature — Merfolk Warrior", effectstest.LtTidalWarriorOracle, false)
	forest := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	ltActivate(t, g, me.ID, warrior, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(forest)})
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("not a restore point: %+v", snap.Continuations)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	if got := effectstest.EffectiveSubtypes(t, restored, forest); !reflect.DeepEqual(got, []string{"Island"}) {
		t.Errorf("restored: %v, want [Island]", got)
	}
}
