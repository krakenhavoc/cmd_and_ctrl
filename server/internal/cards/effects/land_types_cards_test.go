package effects

import (
	"reflect"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// land_types_cards_test.go is ADR 0109 Delivery PR 1 (#1881): a land
// that becomes a basic land type for a while (CR 305.7), the static
// form's CR 205.1a fix, and the cards that use them.

const (
	ltTidalWarriorOracle      = "9a1bc869-08a1-4a70-9c9f-2e02e47cd95b"
	ltReefShamanOracle        = "4ad870b2-133c-4229-84a8-f91e4c62fcd0"
	ltTundraKavuOracle        = "b0b7d88b-694b-4bbe-888d-05245e7d1c3e"
	ltUnstableFrontierOracle  = "495214b5-2eab-4fe4-8879-a30a57a67163"
	ltTerraformerOracle       = "683089a8-d5a9-438a-a463-77cc6148125e"
	ltNightcreepOracle        = "8b4b677e-3b18-454c-b385-b3ffb65f916e"
	ltNavigatorsCompassOracle = "c994e148-1bb5-4113-a9f2-73e6dab8f72d"
	ltEnergybendingOracle     = "c4c1e49d-5f14-486c-acf2-1952512b12b2"
	ltVisionCharmOracle       = "c56d2c5b-c3e3-4236-bc38-286e250aa4f7"
	ltDeepwoodElderOracle     = "ffb62e4a-802d-4f1e-b92e-ec0919baafd5"
	ltGaeasLiegeOracle        = "8d134a60-e1e5-4163-8bdc-36af91567185"
	ltOrcishFarmerOracle      = "c3039d19-8c98-4953-8943-9922b6ab45ef"
	ltTheloniteMonkOracle     = "26b0b742-48f7-4ab5-93de-6bd39a7f61ea"
	ltCyclopeanGiantOracle    = "d02427fa-3ef0-484f-acad-63a1d5218727"
	ltTideShaperOracle        = "cf4be71e-0a9a-47a3-b0ec-43e04c3c07a0"
	ltShimmeringMirageOracle  = "d1720ddc-d6ad-4d78-8dd7-29539bbb73da"
	ltSpreadingSeasOracle     = "af32757c-9d3c-4f6c-a6fa-d5729a440382"
	ltContaminatedOracle      = "55e04860-f4f5-445b-81f2-b500fa9b456a"
	ltConvincingMirageOracle  = "685c4665-e04c-4012-aa5b-888a042c2a21"
	ltIllusionaryOracle       = "17ae63fa-fd63-45e6-81c1-bf30851bf77c"
	ltBloodMoonOracle         = "94fac5fe-97d5-4c12-a80c-8efff9d853ae"
	ltLushGrowthOracle        = "c8f8fbf0-4f51-4a5c-82f9-99917513df6d"
	ltKukemssaOracle          = "49e3ac82-7c22-4dec-b61a-b581148c0419"
	ltZombieTrailblazerOracle = "5cd96cf2-c1c6-4982-8cf8-70c16704e387"
	ltDreamwinderOracle       = "eb70d548-9769-4569-ae79-cddc0e623aef"
	ltMagusOfTheMoonOracle    = "6800d01c-345d-4932-a2a4-8df0fb1902f2"
)

// ltProduced is what a permanent's mana abilities add, sorted.
func ltProduced(t *testing.T, g *game.Game, id uuid.UUID) []string {
	t.Helper()
	c := layeredCard(t, g, id)
	var out []string
	for _, ab := range game.ManaAbilitiesForCard(c) {
		out = append(out, ab.Produced)
	}
	sort.Strings(out)
	return out
}

// ltActivate activates the source's ability `index` for its controller,
// then resolves it.
func ltActivate(t *testing.T, g *game.Game, controller, source uuid.UUID, index int, params game.ActivateAbilityParams) {
	t.Helper()
	if err := g.ActivateCatalogAbility(controller, source, index, params); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func ltCardTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// TestMagusOfTheMoonLeavesDryadArborItsDryad is ADR 0109 §1 decision 3:
// the static form replaces land types only (CR 205.1a), so the Dryad
// stays. On develop the whole subtype list was replaced.
func TestMagusOfTheMoonLeavesDryadArborItsDryad(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	arbor := pushLandFor(g, me.ID, "Dryad Arbor", "Land Creature — Forest Dryad")
	pushCatalogPermanent(g, me.ID, "Magus of the Moon", "Creature — Human Wizard", ltMagusOfTheMoonOracle, false)
	if got := effectiveSubtypes(t, g, arbor); !reflect.DeepEqual(got, []string{"Dryad", "Mountain"}) {
		t.Errorf("Dryad Arbor under Magus of the Moon: subtypes %v, want [Dryad Mountain]", got)
	}
	if got := ltProduced(t, g, arbor); !reflect.DeepEqual(got, []string{"{R}"}) {
		t.Errorf("Dryad Arbor under Magus of the Moon taps for %v, want {R}", got)
	}
}

// TestTidalWarriorMakesALandAnIslandUntilEndOfTurn is the seam's
// headline: CR 305.7 from a resolved ability, gone at cleanup.
func TestTidalWarriorMakesALandAnIslandUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warrior := pushCatalogPermanent(g, me.ID, "Tidal Warrior", "Creature — Merfolk Warrior", ltTidalWarriorOracle, false)
	forest := pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	ltActivate(t, g, me.ID, warrior, 0, game.ActivateAbilityParams{Targets: ltCardTarget(forest)})

	if got := effectiveSubtypes(t, g, forest); !reflect.DeepEqual(got, []string{"Island"}) {
		t.Errorf("subtypes %v, want [Island]", got)
	}
	if got := ltProduced(t, g, forest); !reflect.DeepEqual(got, []string{"{U}"}) {
		t.Errorf("taps for %v, want {U}", got)
	}
	if !layeredCard(t, g, forest).HasSupertype("basic") {
		t.Error("the land stopped being basic: CR 305.7 leaves supertypes alone")
	}
	advanceOneTurnForTest(t, g)
	if got := effectiveSubtypes(t, g, forest); !reflect.DeepEqual(got, []string{"Forest"}) {
		t.Errorf("next turn: subtypes %v, want the Forest back", got)
	}
}

// TestTidalWarriorTakesTheRulesTextAway is CR 305.7's second clause: the
// land loses the abilities its rules text gives it. Unstable Frontier's
// own {C} and its activated ability both go.
func TestTidalWarriorTakesTheRulesTextAway(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	warrior := pushCatalogPermanent(g, me.ID, "Tidal Warrior", "Creature — Merfolk Warrior", ltTidalWarriorOracle, false)
	frontier := pushCatalogPermanent(g, me.ID, "Unstable Frontier", "Land", ltUnstableFrontierOracle, false)
	if got := ltProduced(t, g, frontier); !reflect.DeepEqual(got, []string{"{C}"}) {
		t.Fatalf("setup: Unstable Frontier taps for %v", got)
	}
	ltActivate(t, g, me.ID, warrior, 0, game.ActivateAbilityParams{Targets: ltCardTarget(frontier)})
	if got := ltProduced(t, g, frontier); !reflect.DeepEqual(got, []string{"{U}"}) {
		t.Errorf("an Island Unstable Frontier taps for %v, want only {U}", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, frontier, 0, game.ActivateAbilityParams{Targets: ltCardTarget(frontier)}); err == nil {
		t.Error("the Frontier's own activated ability survived CR 305.7")
	}
}

// TestReefShamanAsksAsItResolves is "the basic land type of your
// choice": the choice is made at resolution, and Dryad Arbor keeps its
// creature type and its card types (CR 205.1a, 305.7).
func TestReefShamanAsksAsItResolves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shaman := pushCatalogPermanent(g, me.ID, "Reef Shaman", "Creature — Merfolk Shaman", ltReefShamanOracle, false)
	arbor := pushLandFor(g, me.ID, "Dryad Arbor", "Land Creature — Forest Dryad")
	ltActivate(t, g, me.ID, shaman, 0, game.ActivateAbilityParams{Targets: ltCardTarget(arbor)})
	pick := latestOptionPickFor(g, me.ID)
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
	answerOptionPick(t, g, me.ID, 2) // Swamp
	c := layeredCard(t, g, arbor)
	if got := c.Effective().Subtypes; !reflect.DeepEqual(got, []string{"Dryad", "Swamp"}) {
		t.Errorf("subtypes %v, want [Dryad Swamp]", got)
	}
	if !c.IsCreature() || !c.IsLand() {
		t.Error("Dryad Arbor lost a card type")
	}
}

func TestTundraKavuOffersPlainsOrIsland(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kavu := pushCatalogPermanent(g, me.ID, "Tundra Kavu", "Creature — Kavu", ltTundraKavuOracle, false)
	swamp := pushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	ltActivate(t, g, me.ID, kavu, 0, game.ActivateAbilityParams{Targets: ltCardTarget(swamp)})
	pick := latestOptionPickFor(g, me.ID)
	if pick == nil || len(pick.PickOptions) != 2 || pick.PickOptions[0].Label != "Plains" || pick.PickOptions[1].Label != "Island" {
		t.Fatalf("pick %+v, want Plains or Island", pick)
	}
	answerOptionPick(t, g, me.ID, 0)
	if got := effectiveSubtypes(t, g, swamp); !reflect.DeepEqual(got, []string{"Plains"}) {
		t.Errorf("subtypes %v, want [Plains]", got)
	}
}

// TestTerraformerFixesItsSetAsItBegins is CR 611.2c: a land you control
// afterwards is not changed.
func TestTerraformerFixesItsSetAsItBegins(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tf := pushCatalogPermanent(g, me.ID, "Terraformer", "Creature — Human Wizard", ltTerraformerOracle, false)
	mine := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	ltActivate(t, g, me.ID, tf, 0, game.ActivateAbilityParams{})
	answerOptionPick(t, g, me.ID, 1) // Island
	later := pushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	if got := effectiveSubtypes(t, g, mine); !reflect.DeepEqual(got, []string{"Island"}) {
		t.Errorf("your land: %v, want [Island]", got)
	}
	if got := effectiveSubtypes(t, g, theirs); !reflect.DeepEqual(got, []string{"Forest"}) {
		t.Errorf("an opponent's land changed: %v", got)
	}
	if got := effectiveSubtypes(t, g, later); !reflect.DeepEqual(got, []string{"Plains"}) {
		t.Errorf("a land that arrived later changed: %v", got)
	}
}

func TestNightcreepMakesCreaturesBlackAndLandsSwamps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushTribalCreature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	island := pushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	_ = me
	castCatalogSpell(t, g, "Nightcreep", "Instant", ltNightcreepOracle, nil)
	passPriorityAroundTable(t, g)
	if got := effectiveOf(t, g, bear).Colors; !reflect.DeepEqual(got, []string{"B"}) {
		t.Errorf("creature colours %v, want [B]", got)
	}
	if got := effectiveSubtypes(t, g, island); !reflect.DeepEqual(got, []string{"Swamp"}) {
		t.Errorf("land subtypes %v, want [Swamp]", got)
	}
}

// TestNavigatorsCompassAddsATypeAndKeepsEverything is CR 205.1b.
func TestNavigatorsCompassAddsATypeAndKeepsEverything(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	compass := pushCatalogPermanent(g, me.ID, "Navigator's Compass", "Artifact", ltNavigatorsCompassOracle, false)
	frontier := pushCatalogPermanent(g, me.ID, "Unstable Frontier", "Land", ltUnstableFrontierOracle, false)
	ltActivate(t, g, me.ID, compass, 0, game.ActivateAbilityParams{Targets: ltCardTarget(frontier)})
	answerOptionPick(t, g, me.ID, 4) // Forest
	if got := ltProduced(t, g, frontier); !reflect.DeepEqual(got, []string{"{C}", "{G}"}) {
		t.Errorf("taps for %v, want its own {C} and the Forest's {G}", got)
	}
}

func TestEnergybendingGivesYourLandsEveryBasicLandType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushLandFor(g, me.ID, "Wastes", "Basic Land")
	theirs := pushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	castCatalogSpell(t, g, "Energybending", "Instant — Lesson", ltEnergybendingOracle, nil)
	passPriorityAroundTable(t, g)
	if got := ltProduced(t, g, mine); len(got) != 5 {
		t.Errorf("your land taps for %v, want all five colours", got)
	}
	if got := effectiveSubtypes(t, g, theirs); !reflect.DeepEqual(got, []string{"Island"}) {
		t.Errorf("an opponent's land changed: %v", got)
	}
}

// TestVisionCharmAsksForAnyLandTypeThenABasicOne is decision 4's two
// questions, the first over CR 205.3i's seventeen.
func TestVisionCharmAsksForAnyLandTypeThenABasicOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gate := pushLandFor(g, opp.ID, "Azorius Guildgate", "Land — Gate")
	forest := pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	castCatalogSpellWithModes(t, g, "Vision Charm", "Instant", ltVisionCharmOracle, []int{1})
	passPriorityAroundTable(t, g)
	pick := latestOptionPickFor(g, me.ID)
	if pick == nil || len(pick.PickOptions) != len(game.LandTypes) {
		t.Fatalf("first pick %+v, want the seventeen land types", pick)
	}
	answerOptionPick(t, g, me.ID, 3) // Gate
	answerOptionPick(t, g, me.ID, 2) // Swamp
	if got := effectiveSubtypes(t, g, gate); !reflect.DeepEqual(got, []string{"Swamp"}) {
		t.Errorf("Gate: %v, want [Swamp]", got)
	}
	if got := effectiveSubtypes(t, g, forest); !reflect.DeepEqual(got, []string{"Forest"}) {
		t.Errorf("a land of another type changed: %v", got)
	}
}

// TestDeepwoodElderTargetsXLands is TargetSpec.CountFromX on an
// activated ability.
func TestDeepwoodElderTargetsXLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elder := pushCatalogPermanent(g, me.ID, "Deepwood Elder", "Creature — Dryad Spellshaper", ltDeepwoodElderOracle, false)
	a := pushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	b := pushLandFor(g, opp.ID, "Swamp", "Basic Land — Swamp")
	c := pushLandFor(g, opp.ID, "Plains", "Basic Land — Plains")
	fodder := game.NewCard("fodder", me.ID)
	me.Hand.PushTop(fodder)
	ltActivate(t, g, me.ID, elder, 0, game.ActivateAbilityParams{
		XValue: 2, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
		DiscardIDs: []uuid.UUID{fodder.InstanceID},
	})
	for _, id := range []uuid.UUID{a, b} {
		if got := effectiveSubtypes(t, g, id); !reflect.DeepEqual(got, []string{"Forest"}) {
			t.Errorf("target: %v, want [Forest]", got)
		}
	}
	if got := effectiveSubtypes(t, g, c); !reflect.DeepEqual(got, []string{"Plains"}) {
		t.Errorf("an untargeted land changed: %v", got)
	}
}

// TestGaeasLiegeForestLastsWhileItRemains is "until this creature leaves
// the battlefield", and the Liege's CDA counting the new Forest.
func TestGaeasLiegeForestLastsWhileItRemains(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	liege := pushCatalogPermanent(g, me.ID, "Gaea's Liege", "Creature — Avatar", ltGaeasLiegeOracle, false)
	pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	swamp := pushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	if p := effectiveOf(t, g, liege).Power; p != 1 {
		t.Fatalf("setup: Liege power %d, want 1", p)
	}
	ltActivate(t, g, me.ID, liege, 0, game.ActivateAbilityParams{Targets: ltCardTarget(swamp)})
	if got := effectiveSubtypes(t, g, swamp); !reflect.DeepEqual(got, []string{"Forest"}) {
		t.Fatalf("subtypes %v, want [Forest]", got)
	}
	if p := effectiveOf(t, g, liege).Power; p != 2 {
		t.Errorf("Liege power %d, want 2", p)
	}
	advanceOneTurnForTest(t, g)
	if got := effectiveSubtypes(t, g, swamp); !reflect.DeepEqual(got, []string{"Forest"}) {
		t.Errorf("after a turn: %v, want still a Forest", got)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(liege); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if got := effectiveSubtypes(t, g, swamp); !reflect.DeepEqual(got, []string{"Swamp"}) {
		t.Errorf("after the Liege left: %v, want the Swamp back", got)
	}
}

// TestOrcishFarmerLastsUntilTheLandsControllerUntaps is "until its
// controller's next untap step".
func TestOrcishFarmerLastsUntilTheLandsControllerUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[2]
	farmer := pushCatalogPermanent(g, me.ID, "Orcish Farmer", "Creature — Orc", ltOrcishFarmerOracle, false)
	plains := pushLandFor(g, opp.ID, "Plains", "Basic Land — Plains")
	ltActivate(t, g, me.ID, farmer, 0, game.ActivateAbilityParams{Targets: ltCardTarget(plains)})
	advanceOneTurnForTest(t, g) // seat 1's turn
	if got := effectiveSubtypes(t, g, plains); !reflect.DeepEqual(got, []string{"Swamp"}) {
		t.Errorf("on another player's turn: %v, want still a Swamp", got)
	}
	advanceToUpkeepOfSeat(t, g, 2)
	if got := effectiveSubtypes(t, g, plains); !reflect.DeepEqual(got, []string{"Plains"}) {
		t.Errorf("in its controller's turn: %v, want the Plains back", got)
	}
}

func TestTheloniteMonkMakesAForestForGood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	monk := pushCatalogPermanent(g, me.ID, "Thelonite Monk", "Creature — Insect Monk Cleric", ltTheloniteMonkOracle, false)
	elf := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Llanowar Elves",
		TypeLine: "Creature — Elf Druid", Colors: []string{"G"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	island := pushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	ltActivate(t, g, me.ID, monk, 0, game.ActivateAbilityParams{Targets: ltCardTarget(island), SacrificeIDs: []uuid.UUID{elf}})
	advanceOneTurnForTest(t, g)
	advanceOneTurnForTest(t, g)
	if got := effectiveSubtypes(t, g, island); !reflect.DeepEqual(got, []string{"Forest"}) {
		t.Errorf("two turns later: %v, want still a Forest", got)
	}
}

// TestCyclopeanGiantDiesThenIsExiled is the trigger's two sentences.
func TestCyclopeanGiantDiesThenIsExiled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := pushCatalogPermanent(g, me.ID, "Cyclopean Giant", "Creature — Zombie Giant", ltCyclopeanGiantOracle, false)
	forest := pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(giant); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	answerPickTarget(t, g, forest)
	passPriorityAroundTable(t, g)
	if got := effectiveSubtypes(t, g, forest); !reflect.DeepEqual(got, []string{"Swamp"}) {
		t.Errorf("land: %v, want [Swamp]", got)
	}
	if z := g.FindCardZoneForEffect(giant); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the Giant is in %v, want exile", z)
	}
}

// TestTideShaperOnlyWhenKicked is the intervening if, and the land
// going back when the Shaper leaves.
func TestTideShaperOnlyWhenKicked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plains := pushLandFor(g, opp.ID, "Plains", "Basic Land — Plains")
	shaper, err := castWithOptionalCosts(t, g, "Tide Shaper", "Creature — Merfolk Wizard", ltTideShaperOracle, nil, []int{0}, nil)
	if err != nil {
		t.Fatalf("cast kicked: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, plains)
	passPriorityAroundTable(t, g)
	if got := effectiveSubtypes(t, g, plains); !reflect.DeepEqual(got, []string{"Island"}) {
		t.Fatalf("kicked: %v, want [Island]", got)
	}
	// The fixture card carries no printed power, so the +1/+1 is all of
	// it.
	if p := effectiveOf(t, g, shaper).Power; p != 1 {
		t.Errorf("Tide Shaper power %d with an opponent's Island, want 0+1", p)
	}
	_ = me
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(shaper); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if got := effectiveSubtypes(t, g, plains); !reflect.DeepEqual(got, []string{"Plains"}) {
		t.Errorf("after it left: %v, want [Plains]", got)
	}

	// Unkicked, nothing triggers (CR 603.4).
	if _, err := castWithOptionalCosts(t, g, "Tide Shaper", "Creature — Merfolk Wizard", ltTideShaperOracle, nil, nil, nil); err != nil {
		t.Fatalf("cast unkicked: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatal("an unkicked Tide Shaper asked for a target")
		}
	}
}

func TestShimmeringMirageDrawsAfterTheChoice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	castCatalogSpell(t, g, "Shimmering Mirage", "Instant", ltShimmeringMirageOracle, ltCardTarget(forest))
	hand := len(me.Hand.Cards)
	passPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand {
		t.Fatalf("drew before choosing: hand %d, want %d", len(me.Hand.Cards), hand)
	}
	answerOptionPick(t, g, me.ID, 1)
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d after the choice, want %d", len(me.Hand.Cards), hand+1)
	}
	if got := effectiveSubtypes(t, g, forest); !reflect.DeepEqual(got, []string{"Island"}) {
		t.Errorf("%v, want [Island]", got)
	}
}

// TestSpreadingSeasIsTheStaticForm is decision 4: the Aura on a land.
func TestSpreadingSeasIsTheStaticForm(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := g.Seats[0], g.Seats[1]
	arbor := pushLandFor(g, opp.ID, "Dryad Arbor", "Land Creature — Forest Dryad")
	enchant(t, g, "Spreading Seas", ltSpreadingSeasOracle, arbor)
	passPriorityAroundTable(t, g) // the draw trigger
	if got := effectiveSubtypes(t, g, arbor); !reflect.DeepEqual(got, []string{"Dryad", "Island"}) {
		t.Errorf("%v, want [Dryad Island]", got)
	}
}

func TestLushGrowthIsThreeTypes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	island := pushLandFor(g, me.ID, "Island", "Basic Land — Island")
	enchant(t, g, "Lush Growth", ltLushGrowthOracle, island)
	if got := ltProduced(t, g, island); !reflect.DeepEqual(got, []string{"{G}", "{R}", "{W}"}) {
		t.Errorf("taps for %v, want {R}, {G} and {W}", got)
	}
}

func TestContaminatedGroundDrainsWhenTheLandTaps(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	land := pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	enchant(t, g, "Contaminated Ground", ltContaminatedOracle, land)
	life := opp.Life
	if err := g.ActivateManaAbility(opp.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
}

func TestConvincingMirageSetsTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	castCatalogSpell(t, g, "Convincing Mirage", auraTypeLine, ltConvincingMirageOracle, ltCardTarget(forest))
	passPriorityAroundTable(t, g)
	if got := effectiveSubtypes(t, g, forest); !reflect.DeepEqual(got, []string{"Forest"}) {
		t.Errorf("before the choice: %v, want unchanged", got)
	}
	answerOptionPick(t, g, me.ID, 3) // Mountain
	if got := effectiveSubtypes(t, g, forest); !reflect.DeepEqual(got, []string{"Mountain"}) {
		t.Errorf("%v, want [Mountain]", got)
	}
}

func TestIllusionaryTerrainTurnsOneBasicTypeIntoAnother(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	nonbasic := pushLandFor(g, opp.ID, "Tropical Island", "Land — Forest Island")
	castCatalogSpell(t, g, "Illusionary Terrain", "Enchantment", ltIllusionaryOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestOptionPickFor(g, me.ID)
	if pick == nil || len(pick.PickOptions) != 20 {
		t.Fatalf("pick %+v, want twenty ordered pairs", pick)
	}
	idx := -1
	for i, o := range pick.PickOptions {
		if o.Label == "Island, then Swamp" {
			idx = i
		}
	}
	answerOptionPick(t, g, me.ID, idx)
	if got := effectiveSubtypes(t, g, theirs); !reflect.DeepEqual(got, []string{"Swamp"}) {
		t.Errorf("basic Island: %v, want [Swamp]", got)
	}
	if got := effectiveSubtypes(t, g, nonbasic); !reflect.DeepEqual(got, []string{"Forest", "Island"}) {
		t.Errorf("a nonbasic Island changed: %v", got)
	}
}

func TestBloodMoonLeavesBasicsAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Blood Moon", "Enchantment", ltBloodMoonOracle, false)
	dual := pushLandFor(g, me.ID, "Tropical Island", "Land — Forest Island")
	basic := pushLandFor(g, me.ID, "Island", "Basic Land — Island")
	if got := effectiveSubtypes(t, g, dual); !reflect.DeepEqual(got, []string{"Mountain"}) {
		t.Errorf("nonbasic: %v, want [Mountain]", got)
	}
	if got := effectiveSubtypes(t, g, basic); !reflect.DeepEqual(got, []string{"Island"}) {
		t.Errorf("basic: %v, want [Island]", got)
	}
}

func TestKukemssaSerpentTargetsOnlyAnOpponentsLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	serpent := pushCatalogPermanent(g, me.ID, "Kukemssa Serpent", "Creature — Serpent", ltKukemssaOracle, false)
	myIsland := pushLandFor(g, me.ID, "Island", "Basic Land — Island")
	otherIsland := pushLandFor(g, me.ID, "Island", "Basic Land — Island")
	mine := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	if err := g.ActivateCatalogAbility(me.ID, serpent, 0, game.ActivateAbilityParams{
		Targets: ltCardTarget(mine), SacrificeIDs: []uuid.UUID{myIsland}}); err == nil {
		t.Error("targeted your own land")
	}
	ltActivate(t, g, me.ID, serpent, 0, game.ActivateAbilityParams{Targets: ltCardTarget(theirs), SacrificeIDs: []uuid.UUID{myIsland}})
	if got := effectiveSubtypes(t, g, theirs); !reflect.DeepEqual(got, []string{"Island"}) {
		t.Errorf("%v, want [Island]", got)
	}
	_ = otherIsland
}

func TestZombieTrailblazerTapsItselfToPay(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := pushCatalogPermanent(g, me.ID, "Zombie Trailblazer", "Creature — Zombie Scout", ltZombieTrailblazerOracle, true)
	forest := pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	ltActivate(t, g, me.ID, zombie, 0, game.ActivateAbilityParams{Targets: ltCardTarget(forest), TapIDs: []uuid.UUID{zombie}})
	if got := effectiveSubtypes(t, g, forest); !reflect.DeepEqual(got, []string{"Swamp"}) {
		t.Errorf("%v, want [Swamp]", got)
	}
	if !layeredCard(t, g, zombie).Tapped {
		t.Error("the Trailblazer paid without tapping")
	}
}

func TestDreamwinderSacrificesAnIsland(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	serpent := pushCatalogPermanent(g, me.ID, "Dreamwinder", "Creature — Serpent", ltDreamwinderOracle, false)
	island := pushLandFor(g, me.ID, "Island", "Basic Land — Island")
	theirs := pushLandFor(g, opp.ID, "Plains", "Basic Land — Plains")
	ltActivate(t, g, me.ID, serpent, 0, game.ActivateAbilityParams{Targets: ltCardTarget(theirs), SacrificeIDs: []uuid.UUID{island}})
	if got := effectiveSubtypes(t, g, theirs); !reflect.DeepEqual(got, []string{"Island"}) {
		t.Errorf("%v, want [Island]", got)
	}
}

// TestALandTypeRecordIsARestorePoint is the snapshot half: a table
// holding a setBasicLandTypes record restores and still applies it.
func TestALandTypeRecordIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warrior := pushCatalogPermanent(g, me.ID, "Tidal Warrior", "Creature — Merfolk Warrior", ltTidalWarriorOracle, false)
	forest := pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	ltActivate(t, g, me.ID, warrior, 0, game.ActivateAbilityParams{Targets: ltCardTarget(forest)})
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("not a restore point: %+v", snap.Continuations)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	if got := effectiveSubtypes(t, restored, forest); !reflect.DeepEqual(got, []string{"Island"}) {
		t.Errorf("restored: %v, want [Island]", got)
	}
}
