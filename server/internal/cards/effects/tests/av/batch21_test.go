package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b21ExileCard seeds a card owned by `owner` straight into exile.
func b21ExileCard(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Exile.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, Owner: owner, Controller: owner,
	})
	return id
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. The two DMU
// duals are rows in their cycle table, so a transposed row is
// invisible until someone plays that exact card.
func TestBatch21CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B21DemonsDiscipleOracle:       "Demon's Disciple",
		effectstest.B21LlanowarTribeOracle:        "Llanowar Tribe",
		effectstest.B21IdyllicBeachfrontOracle:    "Idyllic Beachfront",
		effectstest.B21SatyrEnchanterOracle:       "Satyr Enchanter",
		effectstest.B21EndTheFestivitiesOracle:    "End the Festivities",
		effectstest.B21ManglehornOracle:           "Manglehorn",
		effectstest.B21RoyalAssassinOracle:        "Royal Assassin",
		effectstest.B21SeraphSanctuaryOracle:      "Seraph Sanctuary",
		effectstest.B21WrathfulRedDragonOracle:    "Wrathful Red Dragon",
		effectstest.B21GutShotOracle:              "Gut Shot",
		effectstest.B21AgentOfTheIronThroneOracle: "Agent of the Iron Throne",
		effectstest.B21SpitefulBanditryOracle:     "Spiteful Banditry",
		effectstest.B21WillOfTheMarduOracle:       "Will of the Mardu",
		effectstest.B21HermitDruidOracle:          "Hermit Druid",
		effectstest.B21ShireTerraceOracle:         "Shire Terrace",
		effectstest.B21MemniteOracle:              "Memnite",
		effectstest.B21TrophyMageOracle:           "Trophy Mage",
		effectstest.B21GeneralKreatOracle:         "General Kreat, the Boltbringer",
		effectstest.B21ElvishVisionaryOracle:      "Elvish Visionary",
		effectstest.B21KokushoOracle:              "Kokusho, the Evening Star",
		effectstest.B21MindsEyeOracle:             "Mind's Eye",
		effectstest.B21LysAlanaHuntmasterOracle:   "Lys Alana Huntmaster",
		effectstest.B21SpitefulVisionsOracle:      "Spiteful Visions",
		effectstest.B21LightningStrikeOracle:      "Lightning Strike",
		effectstest.B21DiregrafCaptainOracle:      "Diregraf Captain",
		effectstest.B21BrilliantRestorationOracle: "Brilliant Restoration",
		effectstest.B21VerduranEnchantressOracle:  "Verduran Enchantress",
		effectstest.B21GethsGrimoireOracle:        "Geth's Grimoire",
		effectstest.B21PrizedStatueOracle:         "Prized Statue",
		effectstest.B21OblivionSowerOracle:        "Oblivion Sower",
		effectstest.B21SanctumSeekerOracle:        "Sanctum Seeker",
		effectstest.B21HazelsBrewmasterOracle:     "Hazel's Brewmaster",
		effectstest.B21GeothermalBogOracle:        "Geothermal Bog",
		// #931 took the last of the four declared skips off the list:
		// every graveyard arrival now opens the CR 614 window, so
		// "if a card or token would be put into a graveyard from
		// anywhere, exile it instead" is finally true of all of them.
		effectstest.B21RestInPeaceOracle: "Rest in Peace",
	}
	if len(want) != 34 {
		t.Fatalf("the batch registers 34 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
	// The declared skip that remains must stay out until its seam
	// lands: a legendary-sorcery cast gate (Elder Gargaroth's mode prompt
	// for a triggered ability has since shipped). Rest in Peace shipped with #931 and is in the
	// table above; Tome of Legends shipped with the Y'shtola deck
	// request (#3657) once its counter-removal cost existed.
	for oracle, name := range map[string]string{
		"f5f0deb0-070a-45b6-9b81-0bc8143f3040": "Jaya's Immolating Inferno",
	} {
		if _, ok := effects.Lookup(oracle); ok {
			t.Errorf("%s is declared skipped on #383 but is registered — update the issue", name)
		}
	}
}

// --- the lands -----------------------------------------------------

func TestB21DominariaDualsEnterTappedAndTapForTheirColours(t *testing.T) {
	for _, tc := range []struct{ name, oracle, produced string }{
		{"Idyllic Beachfront", effectstest.B21IdyllicBeachfrontOracle, "{W|U}"},
		{"Geothermal Bog", effectstest.B21GeothermalBogOracle, "{B|R}"},
	} {
		g := effectstest.NewCatalogGame(t)
		land := effectstest.PlayLandFromHand(t, g, tc.name, tc.oracle)
		effectstest.Top100AssertEnteredTapped(t, g, land, tc.name)
		spec, _ := effects.Lookup(tc.oracle)
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != tc.produced {
			t.Errorf("%s: mana abilities %+v, want one producing %s", tc.name, spec.ManaAbilities, tc.produced)
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: a typed tapland is whole", tc.name)
		}
	}
}

func TestB21SeraphSanctuaryGainsOnEntryAndPerAngel(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := me.Life
	effectstest.PlayLandFromHand(t, g, "Seraph Sanctuary", effectstest.B21SeraphSanctuaryOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Fatalf("entering gains 1: life %d, want %d", me.Life, life+1)
	}
	spec, _ := effects.Lookup(effectstest.B21SeraphSanctuaryOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{C}" {
		t.Errorf("mana abilities %+v, want one producing {C}", spec.ManaAbilities)
	}
	effectstest.CastCatalogSpell(t, g, "Serra Angel", "Creature — Angel", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("an Angel entering gains 1: life %d, want %d", me.Life, life+2)
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Error("a non-Angel gains nothing")
	}
	// An opponent's Angel is not yours.
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.CastCatalogSpell(t, g, "Their Angel", "Creature — Angel", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Error("an opponent's Angel is not an Angel you control")
	}
	_ = opp
}

func TestB21ShireTerraceTapsForColorlessAndFetchesABasicTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	terrace := effectstest.B12Permanent(g, me.ID, "Shire Terrace", "Land")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == terrace {
			g.Battlefield.Cards[i].OracleID = effectstest.B21ShireTerraceOracle
		}
	}
	spec, _ := effects.Lookup(effectstest.B21ShireTerraceOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{C}" {
		t.Errorf("mana abilities %+v, want one producing {C}", spec.ManaAbilities)
	}
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Plains", "Basic Land — Plains"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Command Tower", "Land"),
	)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, terrace, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(terrace) || !me.Graveyard.Contains(terrace) {
		t.Fatal("the Terrace is sacrificed as a cost")
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two basics: the searcher chooses")
	}
	if effectstest.SearchOptionNamed(g, c, "Command Tower") != uuid.Nil {
		t.Error("only basic land cards are offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Island")
	island := effectstest.FindBattlefieldByName(g, "Island")
	if island == uuid.Nil || !effectstest.B16Tapped(t, g, island) {
		t.Error("the basic enters tapped")
	}
}

// --- the spells ----------------------------------------------------

func TestB21LightningStrikeAndGutShotBurnAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.CastCatalogSpell(t, g, "Lightning Strike", "Instant", effectstest.B21LightningStrikeOracle, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 37 {
		t.Errorf("Strike to the face: life %d, want 37", opp.Life)
	}
	effectstest.CastCatalogSpell(t, g, "Gut Shot", "Instant", effectstest.B21GutShotOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DamageMarkedOn(g, bear) != 1 {
		t.Errorf("Gut Shot marks 1 on the creature, marked %d", effectstest.DamageMarkedOn(g, bear))
	}
	effectstest.CastCatalogSpell(t, g, "Gut Shot", "Instant", effectstest.B21GutShotOracle, effectstest.B16TargetPlayer(me.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 39 {
		t.Errorf("Gut Shot to yourself: life %d, want 39", me.Life)
	}
	// #916 gave the cast prompt the stepper that pays {R/P} with 2
	// life, which was the only thing this card declared.
	if spec, _ := effects.Lookup(effectstest.B21GutShotOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the Phyrexian mana gap is closed; Gut Shot is complete")
	}
}

func TestB21EndTheFestivitiesPingsOpponentsAndTheirBoards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := effectstest.B16Creature(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1, "R")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1, "R")
	big := effectstest.B16Creature(g, other.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Jace", TypeLine: "Legendary Planeswalker — Jace", Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{game.CounterLoyalty: 3},
	})
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	before := effectstest.B17Life(g)
	effectstest.CastCatalogSpell(t, g, "End the Festivities", "Sorcery", effectstest.B21EndTheFestivitiesOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := before[i]
		if i != 0 {
			want--
		}
		if p.Life != want {
			t.Errorf("seat %d life %d, want %d", i, p.Life, want)
		}
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("an opponent's 1/1 dies")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your own 1/1 is untouched")
	}
	if !g.Battlefield.Contains(big) || effectstest.DamageMarkedOn(g, big) != 1 {
		t.Error("an opponent's 2/2 takes 1 and lives")
	}
	if effectstest.CounterCount(g, walker, game.CounterLoyalty) != 2 {
		t.Errorf("an opponent's planeswalker loses a loyalty: %d, want 2", effectstest.CounterCount(g, walker, game.CounterLoyalty))
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("a noncreature artifact is not hit")
	}
}

func TestB21WillOfTheMarduMakesWarriorsOrBurnsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B16Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2, "G")
	effectstest.B16Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2, "G")
	effectstest.B16Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Wurm", "Creature — Wurm", 5, 5, "G")
	effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.CastModal(t, g, "Will of the Mardu", "Instant", effectstest.B21WillOfTheMarduOracle, []int{0}, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	warriors := effectstest.BattlefieldIDsNamed(g, "Warrior")
	if len(warriors) != 2 {
		t.Fatalf("target player controls 2 creatures: 2 Warriors, got %d", len(warriors))
	}
	if c := effectstest.CardByID(g, warriors[0]); effectstest.ControllerOf(t, g, warriors[0]) != me.ID || len(c.Colors) != 1 || c.Colors[0] != "R" || effectstest.EffectivePower(t, g, warriors[0]) != 1 {
		t.Error("1/1 red Warriors under your control")
	}
	// Mode two: damage equal to YOUR creature count — three Bears
	// and two Warriors.
	effectstest.CastModal(t, g, "Will of the Mardu", "Instant", effectstest.B21WillOfTheMarduOracle, []int{1}, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("five creatures you control: 5 damage kills the 5/5")
	}
	// Choosing both, with and without a commander, is
	// conditional_mode_count_test.go's (#1590).
}

func TestB21BrilliantRestorationReturnsArtifactsAndEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B17GraveyardCard(me, "Dead Rock", "Artifact", "{2}")
	shrine := effectstest.B17GraveyardCard(me, "Dead Shrine", "Enchantment", "{1}{W}")
	myr := effectstest.B17GraveyardCard(me, "Dead Myr", "Artifact Creature — Myr", "{1}")
	bear := effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	land := effectstest.B17GraveyardCard(me, "Dead Forest", "Basic Land — Forest", "")
	theirs := effectstest.B17GraveyardCard(opp, "Their Dead Rock", "Artifact", "{2}")
	effectstest.CastCatalogSpell(t, g, "Brilliant Restoration", "Sorcery", effectstest.B21BrilliantRestorationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{rock, shrine, myr} {
		if !g.Battlefield.Contains(id) || effectstest.ControllerOf(t, g, id) != me.ID {
			t.Errorf("%s returns to the battlefield under your control", effectstest.CardByID(g, id).Name)
		}
	}
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(land) {
		t.Error("a creature card and a land card stay in the graveyard")
	}
	if g.Battlefield.Contains(theirs) || !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's graveyard is not yours")
	}
	// #3158: an Aura it returns chooses its host as it enters
	// (aura_host_cards_test.go), so nothing is declared.
	if spec, _ := effects.Lookup(effectstest.B21BrilliantRestorationOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("Brilliant Restoration is complete since CR 303.4f's choice")
	}
}

func TestB21SpitefulBanditrySweepsForXAndTreasuresOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := effectstest.B16Creature(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1, "R")
	small := effectstest.B16Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1, "R")
	alsoSmall := effectstest.B16Creature(g, other.ID, "Their Elf", "Creature — Elf", 1, 1, "G")
	big := effectstest.B16Creature(g, opp.ID, "Their Wurm", "Creature — Wurm", 5, 5, "G")
	banditry := effectstest.CastXSpell(t, g, "Spiteful Banditry", "Enchantment", effectstest.B21SpitefulBanditryOracle, "{X}{R}{R}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(banditry) {
		t.Fatal("the enchantment enters")
	}
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(small) || g.Battlefield.Contains(alsoSmall) {
		t.Error("X=2 kills every creature with toughness 2 or less, yours included")
	}
	if !g.Battlefield.Contains(big) || effectstest.DamageMarkedOn(g, big) != 2 {
		t.Error("the 5/5 takes 2 and lives")
	}
	// Two opponents' creatures died at once: ONE Treasure, and the
	// Banditry was on the battlefield to see it.
	if n := effectstest.B16CountNamed(g, "Treasure"); n != 1 {
		t.Fatalf("one or more opponents' creatures died: 1 Treasure, got %d", n)
	}
	// A later death this turn: nothing (once each turn).
	effectstest.B18Kill(t, g, big)
	if n := effectstest.B16CountNamed(g, "Treasure"); n != 1 {
		t.Errorf("the ability triggers only once each turn: %d Treasures", n)
	}
	// Next turn: an opponent's creature dying triggers again; your
	// own does not.
	effectstest.AdvanceToMainOf(t, g, 1)
	yours := effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	effectstest.B18Kill(t, g, yours)
	if n := effectstest.B16CountNamed(g, "Treasure"); n != 1 {
		t.Errorf("your own creature dying is not an opponent's: %d Treasures", n)
	}
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.B18Kill(t, g, theirs)
	if n := effectstest.B16CountNamed(g, "Treasure"); n != 2 {
		t.Errorf("a new turn: 2 Treasures, got %d", n)
	}
	// #1357: the ETB damage now comes from a real trigger reading
	// CastX(), closing the last declared gap.
	if spec, _ := effects.Lookup(effectstest.B21SpitefulBanditryOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("no caveat should remain now that the ETB trigger reads CastX()")
	}
}

// --- the creatures: vanilla, mana, ETB -----------------------------

func TestB21MemniteIsAWholeVanillaArtifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	memnite := effectstest.CastCatalogSpell(t, g, "Memnite", "Artifact Creature — Construct", effectstest.B21MemniteOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(memnite) {
		t.Fatal("a zero-cost artifact creature resolves to the battlefield")
	}
	spec, _ := effects.Lookup(effectstest.B21MemniteOracle)
	if spec.Completeness != effects.CompletenessFull || len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.PrintedKeywords) != 0 {
		t.Error("no rules text: a whole vanilla spec")
	}
}

func TestB21LlanowarTribeTapsForThreeGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tribe := effectstest.B21Push(g, me.ID, "Llanowar Tribe", "Creature — Elf Druid", effectstest.B21LlanowarTribeOracle, 3, 3, "G")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, tribe, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 3 || got[0] != "G" || got[1] != "G" || got[2] != "G" {
		t.Errorf("pool %v, want GGG", got)
	}
	if !effectstest.B16Tapped(t, g, tribe) {
		t.Error("the Tribe taps")
	}
	// A summoning-sick Tribe cannot tap.
	sick := effectstest.CastCatalogSpell(t, g, "Llanowar Tribe", "Creature — Elf Druid", effectstest.B21LlanowarTribeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateManaAbility(me.ID, sick, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a creature's tap mana ability waits out summoning sickness")
	}
}

func TestB21ElvishVisionaryDrawsOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	visionary := effectstest.CastCatalogSpell(t, g, "Elvish Visionary", "Creature — Elf Shaman", effectstest.B21ElvishVisionaryOracle, nil)
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if effectstest.TriggerOnStack(g, visionary) == nil {
		t.Fatal("the draw is a trigger on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("the seeded card was cast and one was drawn: hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestB21DemonsDiscipleEdictsCreaturesAndPlaneswalkers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other, third := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	bear := effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Jace", TypeLine: "Legendary Planeswalker — Jace", Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{game.CounterLoyalty: 3},
	})
	rock := effectstest.B12Permanent(g, other.ID, "Their Rock", "Artifact")
	disciple := effectstest.CastCatalogSpell(t, g, "Demon's Disciple", "Creature — Human Cleric", effectstest.B21DemonsDiscipleOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	mine := b19SacrificeOptions(g, me.ID)
	if len(mine) != 2 || !effectstest.HasID(mine, bear) || !effectstest.HasID(mine, disciple) {
		t.Errorf("EACH player, the Disciple itself included: %v", mine)
	}
	effectstest.AnswerSacrifice(t, g, me.ID, disciple)
	// #3664: the opponent chooses next, and the Disciple waits for them.
	if !g.Battlefield.Contains(disciple) {
		t.Error("the Disciple was sacrificed before the opponent had chosen")
	}
	opts := b19SacrificeOptions(g, opp.ID)
	if len(opts) != 2 || !effectstest.HasID(opts, theirs) || !effectstest.HasID(opts, walker) {
		t.Errorf("a creature OR a planeswalker: %v", opts)
	}
	if effectstest.SacrificeChoiceFor(g, other.ID) != nil || effectstest.SacrificeChoiceFor(g, third.ID) != nil {
		t.Error("a player with only an artifact, or nothing, is skipped")
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, walker)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(disciple) || g.Battlefield.Contains(walker) {
		t.Error("the chosen permanents are sacrificed")
	}
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(theirs) || !g.Battlefield.Contains(rock) {
		t.Error("and nothing else")
	}
}

func TestB21TrophyMageTutorsAThreeDropArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Basalt Monolith", TypeLine: "Artifact", ManaCost: "{3}"},
		game.Card{Name: "Crucible of Worlds", TypeLine: "Artifact", ManaCost: "{3}"},
		game.Card{Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}"},
		game.Card{Name: "Solemn", TypeLine: "Artifact Creature — Golem", ManaCost: "{4}"},
		game.Card{Name: "Three-Drop Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{G}"},
	)
	effectstest.CastCatalogSpell(t, g, "Trophy Mage", "Creature — Human Wizard", effectstest.B21TrophyMageOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("you may search: the chooser asks")
	}
	for _, name := range []string{"Basalt Monolith", "Crucible of Worlds"} {
		if effectstest.SearchOptionNamed(g, c, name) == uuid.Nil {
			t.Errorf("%s (an artifact with mana value 3) is offered", name)
		}
	}
	for _, name := range []string{"Mind Stone", "Solemn", "Three-Drop Bear"} {
		if effectstest.SearchOptionNamed(g, c, name) != uuid.Nil {
			t.Errorf("%s is not", name)
		}
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Crucible of Worlds")
	if !effectstest.B02bHandHasNamed(me, "Crucible of Worlds") {
		t.Error("the pick goes to hand")
	}
}

func TestB21HermitDruidRevealsToABasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	druid := effectstest.B21Push(g, me.ID, "Hermit Druid", "Creature — Human Druid", effectstest.B21HermitDruidOracle, 1, 1, "G")
	// seedSearchLibrary lists the library top-first.
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Nonbasic", TypeLine: "Land"},
		game.Card{Name: "First Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Second Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	bolt, nonbasic, first, second, deep := ids[0], ids[1], ids[2], ids[3], ids[4]
	effectstest.AdvanceToMain(t, g)
	hand, yard := me.Hand.Size(), me.Graveyard.Size()
	effectstest.B06AddMana(me, "G")
	effectstest.B16Activate(t, g, me.ID, druid, 0, game.ActivateAbilityParams{})
	if !effectstest.B16Tapped(t, g, druid) {
		t.Error("the Druid taps")
	}
	if !me.Hand.Contains(first) || me.Hand.Size() != hand+1 {
		t.Error("the first basic land from the top goes to hand")
	}
	if !me.Graveyard.Contains(bolt) || !me.Graveyard.Contains(nonbasic) || me.Graveyard.Size() != yard+2 {
		t.Error("the cards above it — a nonbasic land included — go to the graveyard")
	}
	if !me.Library.Contains(second) || !me.Library.Contains(deep) {
		t.Error("the cards below it stay")
	}
	if c := effectstest.CardByID(g, first); !c.IsKnownTo(g.Seats[1].ID) {
		t.Error("the land was revealed")
	}
	// No basic land: the whole library goes.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(druid) })
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Only Sorcery", TypeLine: "Sorcery"},
		game.Card{Name: "Only Instant", TypeLine: "Instant"},
	)
	effectstest.B06AddMana(me, "G")
	effectstest.B16Activate(t, g, me.ID, druid, 0, game.ActivateAbilityParams{})
	if me.Library.Size() != 0 || me.Hand.Size() != hand+1 {
		t.Error("with no basic land every card is milled and none reaches the hand")
	}
}

func TestB21RoyalAssassinDestroysOnlyTappedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	assassin := effectstest.B21Push(g, me.ID, "Royal Assassin", "Creature — Human Assassin", effectstest.B21RoyalAssassinOracle, 1, 1, "B")
	untapped := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	tapped := effectstest.B16Creature(g, opp.ID, "Their Tapped Bear", "Creature — Bear", 2, 2, "G")
	effectstest.B16Tap(g, tapped)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, assassin, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(untapped)}); err == nil {
		t.Fatal("an untapped creature is not a legal target")
	}
	effectstest.B16Activate(t, g, me.ID, assassin, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(tapped)})
	if g.Battlefield.Contains(tapped) || !opp.Graveyard.Contains(tapped) {
		t.Error("the tapped creature is destroyed")
	}
	if !effectstest.B16Tapped(t, g, assassin) {
		t.Error("the Assassin taps as its cost")
	}
	// An attacker is tapped by its declaration: prey.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(assassin) })
	attacker := effectstest.B16Creature(g, opp.ID, "Their Attacker", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.DeclareAttack(t, g, me.ID, attacker)
	if err := g.ActivateCatalogAbility(me.ID, assassin, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(attacker)}); err != nil {
		t.Fatalf("an attacking creature is tapped: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacker is destroyed")
	}
}

func TestB21ManglehornMayDestroyAnArtifactAndTapsTheirs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myRock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirRock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.CastAndResolveCreature(t, g, "Manglehorn", "Creature — Beast", effectstest.B21ManglehornOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, myRock) || !effectstest.HasID(p.PickTargetCards, theirRock) || effectstest.HasID(p.PickTargetCards, bear) {
		t.Error("target ARTIFACT — anyone's, and not a creature")
	}
	effectstest.PickCard(t, g, me.ID, theirRock)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirRock) || !g.Battlefield.Contains(myRock) {
		t.Error("the chosen artifact is destroyed")
	}
	// An opponent's cast artifact enters tapped; yours does not; an
	// opponent's artifact creature is an artifact.
	effectstest.AdvanceToMainOf(t, g, 1)
	signet := effectstest.CastCatalogSpell(t, g, "Their Signet", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Top100AssertEnteredTapped(t, g, signet, "an opponent's artifact")
	myr := effectstest.CastCatalogSpell(t, g, "Their Myr", "Artifact Creature — Myr", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Top100AssertEnteredTapped(t, g, myr, "an opponent's artifact creature")
	theirBear := effectstest.CastCatalogSpell(t, g, "Their Other Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, theirBear) {
		t.Error("a non-artifact is untouched")
	}
	effectstest.AdvanceToMainOf(t, g, 0)
	mine := effectstest.CastCatalogSpell(t, g, "My Signet", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, mine) {
		t.Error("your own artifacts enter untapped")
	}
	// Declined: nothing is destroyed.
	effectstest.CastAndResolveCreature(t, g, "Manglehorn", "Creature — Beast", effectstest.B21ManglehornOracle)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(myRock) || !g.Battlefield.Contains(mine) {
		t.Error("declining destroys nothing")
	}
	// #762: an opponent's artifact TOKEN — a Treasure — enters tapped
	// too, because token creation now runs the entry pipeline.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TreasureToken(), 1) })
	if !effectstest.B16Tapped(t, g, effectstest.FindBattlefieldByName(g, "Treasure")) {
		t.Error("an opponent's artifact TOKEN enters tapped")
	}
	if spec, _ := effects.Lookup(effectstest.B21ManglehornOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the token gap is closed — the caveat must be gone")
	}
}

// --- the creatures: cast, draw and discard triggers ---------------

func TestB21EnchantressesDrawOnEnchantmentSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B21Push(g, me.ID, "Satyr Enchanter", "Creature — Satyr Druid", effectstest.B21SatyrEnchanterOracle, 2, 2, "G", "W")
	effectstest.B21Push(g, me.ID, "Verduran Enchantress", "Creature — Human Druid", effectstest.B21VerduranEnchantressOracle, 0, 2, "G")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Shrine", "Enchantment", "", nil)
	// The Satyr's draw is mandatory; the Enchantress asks as it
	// resolves (CR 603.5). Both go on the stack in one CR 603.3b batch.
	if !effectstest.AnswerTriggerOrderLastQueuedFirst(t, g) {
		t.Error("the Satyr's and the Enchantress's draws were not offered for ordering")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("the seeded card was cast and two were drawn: hand %d, want %d", me.Hand.Size(), hand+2)
	}
	hand = me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Kami", "Enchantment Creature — Spirit", "", nil)
	effectstest.AnswerTriggerOrderLastQueuedFirst(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("an enchantment creature is an enchantment spell; the Enchantress declined: hand %d, want %d", me.Hand.Size(), hand+1)
	}
	hand = me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if effectstest.B19HasTriggerPromptFor(g, me.ID) {
		t.Error("a creature spell is not an enchantment spell")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("no draw")
	}
	// An opponent's enchantment is not yours.
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.CastCatalogSpell(t, g, "Their Shrine", "Enchantment", "", nil)
	if effectstest.B19HasTriggerPromptFor(g, me.ID) || len(g.PendingTriggers) != 0 || effectstest.TriggerOnStack(g, uuid.Nil) != nil {
		t.Error("an opponent casting an enchantment triggers nothing")
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestB21LysAlanaHuntmasterMayMakeAnElfPerElfSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B21Push(g, me.ID, "Lys Alana Huntmaster", "Creature — Elf Warrior", effectstest.B21LysAlanaHuntmasterOracle, 3, 3, "G")
	effectstest.CastCatalogSpell(t, g, "Llanowar Elves", "Creature — Elf Druid", "", nil)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	warriors := effectstest.BattlefieldIDsNamed(g, "Elf Warrior")
	if len(warriors) != 1 {
		t.Fatalf("one Elf Warrior, got %d", len(warriors))
	}
	if c := effectstest.CardByID(g, warriors[0]); effectstest.EffectivePower(t, g, warriors[0]) != 1 || len(c.Colors) != 1 || c.Colors[0] != "G" || !c.HasSubtype("Elf") {
		t.Error("a 1/1 green Elf Warrior")
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if effectstest.B19HasTriggerPromptFor(g, me.ID) {
		t.Error("a non-Elf spell does not trigger")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Elvish Visionary", "Creature — Elf Shaman", effectstest.B21ElvishVisionaryOracle, nil)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16CountNamed(g, "Elf Warrior") != 1 {
		t.Error("declining makes nothing")
	}
}

func TestB21GethsGrimoireMayDrawPerOpponentDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B21Push(g, me.ID, "Geth's Grimoire", "Artifact — Book", effectstest.B21GethsGrimoireOracle, 0, 0)
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 2) })
	if n := effectstest.B18TriggerPromptCount(g, me.ID); n != 2 {
		t.Fatalf("two discards: two prompts, got %d", n)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("one yes, one no: drew %d, want 1", me.Hand.Size()-hand)
	}
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	if effectstest.B19HasTriggerPromptFor(g, me.ID) {
		t.Error("your own discard is not an opponent's")
	}
}

func TestB21MindsEyeMayPayOneToDrawPerOpponentDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B21Push(g, me.ID, "Mind's Eye", "Artifact", effectstest.B21MindsEyeOracle, 0, 0)
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("an opponent drew: you are asked to pay {1}")
	}
	effectstest.B06AddMana(me, "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if me.Hand.Size() != hand+1 {
		t.Errorf("paid: drew %d, want 1", me.Hand.Size()-hand)
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if me.Hand.Size() != hand+1 {
		t.Error("declined: no draw")
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("your own draw is not an opponent's")
	}
}

func TestB21SpitefulVisionsDrawsExtraAndTaxesEveryDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B21Push(g, me.ID, "Spiteful Visions", "Enchantment", effectstest.B21SpitefulVisionsOracle, 0, 0)
	// On to the NEXT seat's draw step — seat 0's own is already
	// behind the cursor (newCatalogGame parks there, #692).
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AdvanceTo(t, g, game.StepDraw)
	active := g.Seats[g.Turn.ActiveSeat]
	if active.ID == me.ID {
		t.Fatal("seat 1's draw step")
	}
	hand, life := active.Hand.Size(), active.Life
	// The turn-based draw already happened and was taxed; its damage
	// trigger and the extra-draw trigger arrive together and the
	// controller orders them.
	effectstest.B21SettleStack(t, g)
	if active.Hand.Size() != hand+1 {
		t.Errorf("the additional draw: hand %d, want %d", active.Hand.Size(), hand+1)
	}
	if active.Life != life-2 {
		t.Errorf("two draws this step, 1 damage each: life %d, want %d", active.Life, life-2)
	}
	// Your own draw is taxed too.
	mine := me.Life
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 2) })
	effectstest.B21SettleStack(t, g)
	if me.Life != mine-2 {
		t.Errorf("a player — you included — draws two: 2 damage, life %d, want %d", me.Life, mine-2)
	}
}

// --- the creatures: dies and drain triggers -----------------------

func TestB21KokushoDrainsTheTableOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kokusho := effectstest.B21Push(g, me.ID, "Kokusho, the Evening Star", "Legendary Creature — Dragon Spirit", effectstest.B21KokushoOracle, 5, 5, "B")
	if !effectstest.HasEffectiveKeyword(t, g, kokusho, "flying") {
		t.Error("flying")
	}
	before := effectstest.B17Life(g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(kokusho) })
	if effectstest.TriggerOnStack(g, kokusho) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("the drain is a trigger")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := before[i] - 5
		if i == 0 {
			want = before[0] + 15
		}
		if p.Life != want {
			t.Errorf("seat %d life %d, want %d", i, p.Life, want)
		}
	}
	// A bounce is not a death.
	again := effectstest.B21Push(g, me.ID, "Kokusho, the Evening Star", "Legendary Creature — Dragon Spirit", effectstest.B21KokushoOracle, 5, 5, "B")
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(again) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.Seats[1].Life != before[1]-5 {
		t.Error("bounced, not died: no drain")
	}
}

func TestB21PrizedStatueMakesTreasureOnEntryAndOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	statue := effectstest.CastCatalogSpell(t, g, "Prized Statue", "Artifact", effectstest.B21PrizedStatueOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Treasure"); n != 1 {
		t.Fatalf("entering: 1 Treasure, got %d", n)
	}
	effectstest.B18Kill(t, g, statue)
	if n := effectstest.B16CountNamed(g, "Treasure"); n != 2 {
		t.Fatalf("dying: 2 Treasures, got %d", n)
	}
	for _, id := range effectstest.BattlefieldIDsNamed(g, "Treasure") {
		if effectstest.ControllerOf(t, g, id) != me.ID {
			t.Error("the Treasures are the Statue's controller's")
		}
	}
	// Exiled, not died: nothing.
	other := effectstest.B21Push(g, me.ID, "Prized Statue", "Artifact", effectstest.B21PrizedStatueOracle, 0, 0)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(other) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Treasure"); n != 2 {
		t.Errorf("an exile is not a death: %d Treasures", n)
	}
	_ = opp
}

func TestB21DiregrafCaptainLordsZombiesAndDrainsOnTheirDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	captain := effectstest.B21Push(g, me.ID, "Diregraf Captain", "Creature — Zombie Soldier", effectstest.B21DiregrafCaptainOracle, 2, 2, "U", "B")
	zombie := effectstest.B16Creature(g, me.ID, "Zombie", "Creature — Zombie", 2, 2, "B")
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2, "B")
	if !effectstest.HasEffectiveKeyword(t, g, captain, "deathtouch") {
		t.Error("deathtouch")
	}
	if effectstest.EffectivePower(t, g, zombie) != 3 || effectstest.EffectiveToughness(t, g, zombie) != 3 {
		t.Error("other Zombies you control get +1/+1")
	}
	if effectstest.EffectivePower(t, g, captain) != 2 || effectstest.EffectivePower(t, g, bear) != 2 || effectstest.EffectivePower(t, g, theirs) != 2 {
		t.Error("not itself, not a non-Zombie, not an opponent's")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(zombie) })
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetPlayers, me.ID) || !effectstest.HasID(p.PickTargetPlayers, opp.ID) || !effectstest.HasID(p.PickTargetPlayers, other.ID) {
		t.Error("target OPPONENT")
	}
	effectstest.PickPlayer(t, g, me.ID, other.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if other.Life != 39 || opp.Life != 40 {
		t.Errorf("the chosen opponent loses 1: %d / %d", other.Life, opp.Life)
	}
	// Not a Zombie, not yours, not the Captain itself.
	effectstest.B18Kill(t, g, bear)
	effectstest.B18Kill(t, g, theirs)
	effectstest.B18Kill(t, g, captain)
	if effectstest.LatestPickTarget(g, me.ID) != nil || other.Life != 39 {
		t.Error("a non-Zombie, an opponent's Zombie and the Captain itself drain nothing")
	}
}

func TestB21AgentOfTheIronThroneDrainsWhileYouControlYourCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B21Push(g, me.ID, "Agent of the Iron Throne", "Legendary Enchantment — Background", effectstest.B21AgentOfTheIronThroneOracle, 0, 0, "B")
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	before := effectstest.B17Life(g)
	// No commander creature out: the granted ability does not exist.
	effectstest.B18Kill(t, g, bear)
	if g.Seats[1].Life != before[1] {
		t.Fatal("without a commander creature you own, nothing drains")
	}
	commander := effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human Knight", 3, 3)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	shrine := effectstest.B12Permanent(g, me.ID, "Shrine", "Enchantment")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
	treasure := effectstest.FindBattlefieldByName(g, "Treasure")
	// An artifact, a token: each drains. An enchantment, an
	// opponent's creature: not.
	effectstest.B18Kill(t, g, rock)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(treasure) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B18Kill(t, g, shrine)
	effectstest.B18Kill(t, g, theirs)
	for i := 1; i < 4; i++ {
		if g.Seats[i].Life != before[i]-2 {
			t.Errorf("seat %d life %d, want %d — an artifact and a Treasure drained, a shrine and their creature did not", i, g.Seats[i].Life, before[i]-2)
		}
	}
	if me.Life != before[0] {
		t.Error("you lose nothing")
	}
	// The commander's own death looks back and drains. Since ADR 0115
	// it dies like any other creature, so it drains even when its owner
	// then sends it to the command zone (CR 903.9a).
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(commander) })
	effectstest.AnswerCommanderReturn(t, g, me.ID, true)
	if !me.Command.Contains(commander) {
		t.Error("the commander went home after the yes")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Seats[1].Life != before[1]-3 {
		t.Errorf("the commander dying drains: life %d, want %d", g.Seats[1].Life, before[1]-3)
	}
	// A commander creature an OPPONENT owns is not yours.
	effectstest.B21Commander(g, opp.ID, "Their Commander", "Legendary Creature — Elf", 3, 3)
	mine := effectstest.B16Creature(g, me.ID, "Another Bear", "Creature — Bear", 2, 2, "G")
	effectstest.B18Kill(t, g, mine)
	if g.Seats[1].Life != before[1]-3 {
		t.Error("an opponent's commander does not carry your Background's ability")
	}
	// ADR 0093 made the grant real, so the stolen-commander gap is
	// gone (TestAgentOfTheIronThroneStolenCommanderDrainsForTheThief).
	if spec, _ := effects.Lookup(effectstest.B21AgentOfTheIronThroneOracle); spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Error("nothing is declared any more — CompletenessFull")
	}
}

// --- the creatures: attack and damage triggers --------------------

func TestB21SanctumSeekerDrainsPerAttackingVampire(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seeker := effectstest.B21Push(g, me.ID, "Sanctum Seeker", "Creature — Vampire Knight", effectstest.B21SanctumSeekerOracle, 3, 4, "B")
	vampire := effectstest.B16Creature(g, me.ID, "Vampire", "Creature — Vampire", 2, 2, "B")
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	before := effectstest.B17Life(g)
	effectstest.DeclareAttack(t, g, opp.ID, seeker, vampire, bear)
	if n := effectstest.TriggersOnStackFrom(g, seeker); n != 2 {
		t.Fatalf("two attacking Vampires (the Seeker included): 2 triggers, got %d", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i := 1; i < 4; i++ {
		if g.Seats[i].Life != before[i]-2 {
			t.Errorf("seat %d life %d, want %d", i, g.Seats[i].Life, before[i]-2)
		}
	}
	if me.Life != before[0]+2 {
		t.Errorf("you gain 1 per trigger, not per opponent: life %d, want %d", me.Life, before[0]+2)
	}
}

func TestB21GeneralKreatMakesAnAttackingGoblinAndPingsPerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kreat := effectstest.B21Push(g, me.ID, "General Kreat, the Boltbringer", "Legendary Creature — Goblin Soldier", effectstest.B21GeneralKreatOracle, 2, 2, "R")
	goblin := effectstest.B16Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1, "R")
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	before := effectstest.B17Life(g)
	effectstest.DeclareAttack(t, g, opp.ID, kreat, goblin, bear)
	if n := effectstest.TriggersOnStackFrom(g, kreat); n != 1 {
		t.Fatalf("one or more Goblins attack: 1 trigger, got %d", n)
	}
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)
	tokens := effectstest.BattlefieldIDsNamed(g, "Goblin")
	var token uuid.UUID
	for _, id := range tokens {
		if id != goblin {
			token = id
		}
	}
	if token == uuid.Nil {
		t.Fatal("a Goblin token is created")
	}
	tok := effectstest.CardByID(g, token)
	if !tok.Tapped || tok.AttackingTarget != opp.ID || effectstest.EffectivePower(t, g, token) != 1 {
		t.Errorf("a 1/1 tapped and attacking the defender: %+v", tok)
	}
	// The token is another creature entering: 1 to each opponent.
	for i := 1; i < 4; i++ {
		if g.Seats[i].Life != before[i]-1 {
			t.Errorf("seat %d life %d, want %d", i, g.Seats[i].Life, before[i]-1)
		}
	}
	// The token was put onto the battlefield attacking, never
	// declared: no second Goblin-attack trigger.
	if n := effectstest.TriggersOnStackFrom(g, kreat); n != 0 || len(g.PendingTriggers) != 0 {
		t.Error("an attacking token is not a declared attacker")
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before[1]-1-2-1-2-1 {
		t.Errorf("the ping, then Kreat, the Goblin, the Bear and the token connect: life %d, want %d", opp.Life, before[1]-7)
	}
	// A non-Goblin attack alone makes nothing; any creature entering
	// pings.
	effectstest.AdvanceToMainOf(t, g, 0)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Elf", "Creature — Elf", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Error("another creature you control entering pings")
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if effectstest.TriggersOnStackFrom(g, kreat) != 0 {
		t.Error("a Bear attacking is not a Goblin attacking")
	}
	if spec, _ := effects.Lookup(effectstest.B21GeneralKreatOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the controller chooses what the token attacks: nothing is left to declare")
	}
}

func TestB21WrathfulRedDragonReflectsDamageDealtToYourDragons(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	wrathful := effectstest.B21Push(g, me.ID, "Wrathful Red Dragon", "Creature — Dragon", effectstest.B21WrathfulRedDragonOracle, 5, 5, "R")
	dragon := effectstest.B16Creature(g, me.ID, "Dragon", "Creature — Dragon", 4, 4, "R")
	theirDragon := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Dragon", TypeLine: "Creature — Dragon", Colors: []string{"R"},
		Keywords: []string{"flying"}, Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID,
	})
	theirBear := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	if !effectstest.HasEffectiveKeyword(t, g, wrathful, "flying") {
		t.Error("flying")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.B13OpponentCasts(t, g, opp, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, "", effectstest.B16TargetCard(dragon))
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetPlayers, opp.ID) || !effectstest.HasID(p.PickTargetCards, theirBear) {
		t.Error("any target … a player, a non-Dragon creature")
	}
	if effectstest.HasID(p.PickTargetCards, theirDragon) || effectstest.HasID(p.PickTargetCards, wrathful) || effectstest.HasID(p.PickTargetCards, dragon) {
		t.Error("… that isn't a Dragon")
	}
	effectstest.PickPlayer(t, g, me.ID, other.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if other.Life != 37 {
		t.Errorf("the Dragon took 3, so it deals 3: life %d, want 37", other.Life)
	}
	// Combat damage to the Wrathful itself reflects too; an
	// opponent's Dragon is not yours.
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.DeclareAttack(t, g, opp.ID, wrathful)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(theirDragon, wrathful); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirBear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirBear) {
		t.Error("the Wrathful took 4 from the blocker and deals 4 to the chosen Bear")
	}
	if effectstest.LatestPickTarget(g, opp.ID) != nil {
		t.Error("their Dragon taking damage is not a Dragon YOU control")
	}
	if spec, _ := effects.Lookup(effectstest.B21WrathfulRedDragonOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the per-source batching gap is closed (#3172 cav-d)")
	}
}

func TestB21HazelsBrewmasterExilesAGraveyardCardAndMakesFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := effectstest.B17GraveyardCard(opp, "Their Dead Bear", "Creature — Bear", "{1}{G}")
	mine := effectstest.B17GraveyardCard(me, "My Dead Bolt", "Instant", "{R}")
	brewmaster := effectstest.CastAndResolveCreature(t, g, "Hazel's Brewmaster", "Creature — Squirrel Warlock", effectstest.B21HazelsBrewmasterOracle)
	if !effectstest.HasEffectiveKeyword(t, g, brewmaster, "menace") {
		t.Error("menace")
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, dead) || !effectstest.HasID(p.PickTargetCards, mine) {
		t.Error("target card from A graveyard — anyone's")
	}
	if p.PickTargetMin != 0 {
		t.Error("up to one")
	}
	effectstest.PickCard(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, dead) {
		t.Error("the chosen card is exiled")
	}
	if n := effectstest.B16CountNamed(g, "Food"); n != 1 {
		t.Fatalf("and a Food is created: %d", n)
	}
	// Attacking triggers again; with no target chosen the Food still
	// comes.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.DeclareAttack(t, g, opp.ID, brewmaster)
	effectstest.B04WaitForPick(t, g, me.ID)
	p = effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
		t.Fatalf("picking no target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(mine) || effectstest.B16CountNamed(g, "Food") != 2 {
		t.Error("no target: nothing exiled, a Food made")
	}
	// Every graveyard empty: no target prompt, and the Food still
	// comes.
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(mine) })
	for _, p := range g.Seats {
		p.Graveyard.Cards = nil
	}
	effectstest.CastAndResolveCreature(t, g, "Hazel's Brewmaster", "Creature — Squirrel Warlock", effectstest.B21HazelsBrewmasterOracle)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("nothing to target: no prompt")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16CountNamed(g, "Food") != 3 {
		t.Error("with every graveyard empty the trigger still makes a Food")
	}
	// The Foods' borrowed abilities are tested with the exile lenders
	// (exile_lender_cards_test.go); the card has no caveat left.
	if spec, _ := effects.Lookup(effectstest.B21HazelsBrewmasterOracle); spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Error("Hazel's Brewmaster is fully automated since ADR 0155 PR 3")
	}
}

func TestB21OblivionSowerExilesFourAndTakesTheirLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// seedSearchLibrary lists the library top-first.
	ids := effectstest.SeedSearchLibrary(opp,
		game.Card{Name: "Their Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Their Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Their Bolt", TypeLine: "Instant"},
		game.Card{Name: "Their Swamp", TypeLine: "Basic Land — Swamp"},
		game.Card{Name: "Fifth Card Forest", TypeLine: "Basic Land — Forest"},
	)
	bear, island, bolt, swamp, fifth := ids[0], ids[1], ids[2], ids[3], ids[4]
	earlier := b21ExileCard(g, opp.ID, "Their Exiled Mountain", "Basic Land — Mountain")
	notTheirs := b21ExileCard(g, g.Seats[2].ID, "Other Exiled Plains", "Basic Land — Plains")
	sower := effectstest.CastCatalogSpell(t, g, "Oblivion Sower", "Creature — Eldrazi", effectstest.B21OblivionSowerOracle, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetPlayers, me.ID) || !effectstest.HasID(p.PickTargetPlayers, opp.ID) {
		t.Error("target OPPONENT, chosen as the cast trigger goes on the stack")
	}
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	if g.Stack.Size() != 1 || effectstest.TriggerOnStack(g, sower) == nil {
		t.Fatal("the trigger sits above the Sower on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, bolt} {
		if !effectstest.InExile(g, id) {
			t.Errorf("%s is exiled", effectstest.CardByID(g, id).Name)
		}
	}
	if !opp.Library.Contains(fifth) {
		t.Error("only the top four")
	}
	// "You may put any number": every land card that player owns in
	// exile is offered, and the Swamp is left behind.
	effectstest.EtChooseCards(t, g, me.ID, 3, island, earlier)
	effectstest.PassPriorityAroundTable(t, g)
	// A returned card is a new object (CR 400.7), so look by name.
	for _, name := range []string{"Their Island", "Their Exiled Mountain"} {
		id := effectstest.FindBattlefieldByName(g, name)
		if id == uuid.Nil || effectstest.ControllerOf(t, g, id) != me.ID {
			t.Errorf("%s — a land that player owns in exile — enters under your control", name)
		}
	}
	for _, id := range []uuid.UUID{island, earlier} {
		if effectstest.InExile(g, id) {
			t.Error("the land left exile")
		}
	}
	if !effectstest.InExile(g, swamp) {
		t.Error("a land that was not chosen stays in exile")
	}
	if !effectstest.InExile(g, notTheirs) {
		t.Error("a land another player owns stays in exile")
	}
	if !g.Battlefield.Contains(sower) {
		t.Error("the Sower resolves after its trigger")
	}
	if spec, _ := effects.Lookup(effectstest.B21OblivionSowerOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the Sower's choice is printed, so it is Full")
	}
}
