package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch21_test.go — card-level coverage for the card-coverage
// roadmap's batch 21 (#383, `edhrec_rank` 2235–2334): the "no new
// machinery" group, the first slice of the second 2000. One test per
// observable behaviour, driven through a real cast, activation, land
// play, attack or step change. Helpers from the earlier batch test
// files are reused by name; new ones are b21-prefixed.

const (
	b21DemonsDiscipleOracle       = "d5a33091-a348-4b13-8dbd-79ab0ad99afe"
	b21LlanowarTribeOracle        = "cfd0f0e6-1bbf-4450-97d2-c54907abb7e4"
	b21IdyllicBeachfrontOracle    = "0aea7e5d-40d7-46c7-a8a6-479cb8061e49"
	b21SatyrEnchanterOracle       = "aa321138-b1a7-4b8e-a2ca-b9ce65704e92"
	b21EndTheFestivitiesOracle    = "09618da4-6e2d-4047-96d7-0576c0754e7d"
	b21ManglehornOracle           = "b67db32b-30a9-49d2-b4c2-90a9f80c36eb"
	b21RoyalAssassinOracle        = "9ed6f28f-a3db-48c5-9ab0-b90a7fba5f57"
	b21SeraphSanctuaryOracle      = "0b504dc6-61cc-4a72-907c-145fa4c72466"
	b21WrathfulRedDragonOracle    = "17c2a962-994a-49f1-8c4e-ebcc43c3c92a"
	b21GutShotOracle              = "9afb3b6e-4909-4efa-aa79-81c0229411c9"
	b21AgentOfTheIronThroneOracle = "325032d5-c452-4454-8976-82f86fee5ab8"
	b21SpitefulBanditryOracle     = "fde2653f-5270-4b8f-9642-0835dbb076c2"
	b21WillOfTheMarduOracle       = "d8df9813-0376-4d30-8cdc-30451fb67726"
	b21HermitDruidOracle          = "16f6438d-2a29-41cb-bf0c-4d02bd66112b"
	b21ShireTerraceOracle         = "619173f4-0403-49cd-9659-2fedd5028a90"
	b21MemniteOracle              = "7663ac7c-1de3-4250-b96a-fae9dbd66a27"
	b21TrophyMageOracle           = "08afc0d7-192f-4ab6-b6a0-c4265cf5e225"
	b21GeneralKreatOracle         = "da1c1c70-c2fa-4ba7-89fe-d9af3fb353b9"
	b21ElvishVisionaryOracle      = "c6a3a882-a127-4590-93d7-679ef4313efe"
	b21KokushoOracle              = "cf631c93-b7fb-4e4f-b405-3778c15b1117"
	b21MindsEyeOracle             = "63fd2a57-7a47-4e07-947c-f4e9da7ee538"
	b21LysAlanaHuntmasterOracle   = "3f3439e1-75ce-482d-881f-836492dca6e9"
	b21SpitefulVisionsOracle      = "922cf963-2b1b-43ad-819e-6e49133e6aae"
	b21LightningStrikeOracle      = "f34b9bc4-7bfe-47fd-ba23-4eeeb46026eb"
	b21DiregrafCaptainOracle      = "70de24d9-c585-4bf6-ac2c-c5b4b7aa298c"
	b21BrilliantRestorationOracle = "9584a8ae-2aba-42b8-8983-0467d6bd5698"
	b21VerduranEnchantressOracle  = "cd98a31b-cc7e-43f9-982e-109ad9850908"
	b21GethsGrimoireOracle        = "ef809e99-34a2-4471-8269-f56bf8037686"
	b21PrizedStatueOracle         = "681fc668-cb26-4ba4-a915-48ddfa2b9520"
	b21OblivionSowerOracle        = "d39b9f64-dc9b-413f-8d05-e21ef46d6756"
	b21SanctumSeekerOracle        = "afb71560-0fc9-4ea5-9d52-d93c17d72519"
	b21HazelsBrewmasterOracle     = "e8180024-1979-4677-9a0d-e08d4b7c825a"
	b21GeothermalBogOracle        = "e3b67368-1dd6-419b-a95d-7131b1dba23f"
	b21RestInPeaceOracle          = "087f9ad7-e74f-40e2-8102-1ed2925d0418"
)

// b21Push seeds a catalog permanent under `owner` with a type line
// and colours — b12Push with the colours a tribal or colour check
// might read.
func b21Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, Colors: colors,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b21Commander seeds a creature flagged as a commander `owner` owns.
func b21Commander(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, IsCommander: true,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b21ExileCard seeds a card owned by `owner` straight into exile.
func b21ExileCard(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Exile.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, Owner: owner, Controller: owner,
	})
	return id
}

// b21SettleStack is passPriorityAroundTable for a board where two
// differing triggers from one source arrive together: the CR 603.3b
// order prompt is answered in the offered order whenever it appears.
func b21SettleStack(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if stackFullyEmpty(g) {
			return
		}
		if triggerOrderPrompt(g) != nil {
			answerTriggerOrderInOfferedOrder(t, g)
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatal("stack did not empty after 64 priority passes")
}

// b21DeclineCommandZone answers the CR 903.9 "put it into the
// command zone instead?" prompt for a dying commander with "no", so
// the card goes to the graveyard.
func b21DeclineCommandZone(t *testing.T, g *game.Game, owner uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionalReplacement && c.Chooser == owner {
			if err := g.ResolveOptionalReplacement(c.ID, owner, false); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}
			return
		}
	}
	t.Fatalf("no command-zone prompt for %s", owner)
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. The two DMU
// duals are rows in their cycle table, so a transposed row is
// invisible until someone plays that exact card.
func TestBatch21CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b21DemonsDiscipleOracle:       "Demon's Disciple",
		b21LlanowarTribeOracle:        "Llanowar Tribe",
		b21IdyllicBeachfrontOracle:    "Idyllic Beachfront",
		b21SatyrEnchanterOracle:       "Satyr Enchanter",
		b21EndTheFestivitiesOracle:    "End the Festivities",
		b21ManglehornOracle:           "Manglehorn",
		b21RoyalAssassinOracle:        "Royal Assassin",
		b21SeraphSanctuaryOracle:      "Seraph Sanctuary",
		b21WrathfulRedDragonOracle:    "Wrathful Red Dragon",
		b21GutShotOracle:              "Gut Shot",
		b21AgentOfTheIronThroneOracle: "Agent of the Iron Throne",
		b21SpitefulBanditryOracle:     "Spiteful Banditry",
		b21WillOfTheMarduOracle:       "Will of the Mardu",
		b21HermitDruidOracle:          "Hermit Druid",
		b21ShireTerraceOracle:         "Shire Terrace",
		b21MemniteOracle:              "Memnite",
		b21TrophyMageOracle:           "Trophy Mage",
		b21GeneralKreatOracle:         "General Kreat, the Boltbringer",
		b21ElvishVisionaryOracle:      "Elvish Visionary",
		b21KokushoOracle:              "Kokusho, the Evening Star",
		b21MindsEyeOracle:             "Mind's Eye",
		b21LysAlanaHuntmasterOracle:   "Lys Alana Huntmaster",
		b21SpitefulVisionsOracle:      "Spiteful Visions",
		b21LightningStrikeOracle:      "Lightning Strike",
		b21DiregrafCaptainOracle:      "Diregraf Captain",
		b21BrilliantRestorationOracle: "Brilliant Restoration",
		b21VerduranEnchantressOracle:  "Verduran Enchantress",
		b21GethsGrimoireOracle:        "Geth's Grimoire",
		b21PrizedStatueOracle:         "Prized Statue",
		b21OblivionSowerOracle:        "Oblivion Sower",
		b21SanctumSeekerOracle:        "Sanctum Seeker",
		b21HazelsBrewmasterOracle:     "Hazel's Brewmaster",
		b21GeothermalBogOracle:        "Geothermal Bog",
		// #931 took the last of the four declared skips off the list:
		// every graveyard arrival now opens the CR 614 window, so
		// "if a card or token would be put into a graveyard from
		// anywhere, exile it instead" is finally true of all of them.
		b21RestInPeaceOracle: "Rest in Peace",
	}
	if len(want) != 34 {
		t.Fatalf("the batch registers 34 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
	// The three declared skips that remain must stay out until their
	// seam lands: a legendary-sorcery cast gate, a mode prompt for a
	// triggered ability, and a counter-removal cost component. The
	// fourth, Rest in Peace, shipped with #931 and is in the table
	// above.
	for oracle, name := range map[string]string{
		"f5f0deb0-070a-45b6-9b81-0bc8143f3040": "Jaya's Immolating Inferno",
		"397b9dcf-690a-4a58-8637-bb9baab7cc2f": "Elder Gargaroth",
		"cada2a1c-5db2-4702-9a57-cbe7da1bf208": "Tome of Legends",
	} {
		if _, ok := Lookup(oracle); ok {
			t.Errorf("%s is declared skipped on #383 but is registered — update the issue", name)
		}
	}
}

// --- the lands -----------------------------------------------------

func TestB21DominariaDualsEnterTappedAndTapForTheirColours(t *testing.T) {
	for _, tc := range []struct{ name, oracle, produced string }{
		{"Idyllic Beachfront", b21IdyllicBeachfrontOracle, "{W|U}"},
		{"Geothermal Bog", b21GeothermalBogOracle, "{B|R}"},
	} {
		g := newCatalogGame(t)
		land := playLandFromHand(t, g, tc.name, tc.oracle)
		top100AssertEnteredTapped(t, g, land, tc.name)
		spec, _ := Lookup(tc.oracle)
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != tc.produced {
			t.Errorf("%s: mana abilities %+v, want one producing %s", tc.name, spec.ManaAbilities, tc.produced)
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: a typed tapland is whole", tc.name)
		}
	}
}

func TestB21SeraphSanctuaryGainsOnEntryAndPerAngel(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := me.Life
	playLandFromHand(t, g, "Seraph Sanctuary", b21SeraphSanctuaryOracle)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Fatalf("entering gains 1: life %d, want %d", me.Life, life+1)
	}
	spec, _ := Lookup(b21SeraphSanctuaryOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{C}" {
		t.Errorf("mana abilities %+v, want one producing {C}", spec.ManaAbilities)
	}
	castCatalogSpell(t, g, "Serra Angel", "Creature — Angel", "", nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("an Angel entering gains 1: life %d, want %d", me.Life, life+2)
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Error("a non-Angel gains nothing")
	}
	// An opponent's Angel is not yours.
	advanceToMainOf(t, g, 1)
	castCatalogSpell(t, g, "Their Angel", "Creature — Angel", "", nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Error("an opponent's Angel is not an Angel you control")
	}
	_ = opp
}

func TestB21ShireTerraceTapsForColorlessAndFetchesABasicTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	terrace := b12Permanent(g, me.ID, "Shire Terrace", "Land")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == terrace {
			g.Battlefield.Cards[i].OracleID = b21ShireTerraceOracle
		}
	}
	spec, _ := Lookup(b21ShireTerraceOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{C}" {
		t.Errorf("mana abilities %+v, want one producing {C}", spec.ManaAbilities)
	}
	seedSearchLibrary(me,
		searchTestLand("Plains", "Basic Land — Plains"),
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Command Tower", "Land"),
	)
	advanceToMain(t, g)
	b06AddMana(me, "C")
	b16Activate(t, g, me.ID, terrace, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(terrace) || !me.Graveyard.Contains(terrace) {
		t.Fatal("the Terrace is sacrificed as a cost")
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two basics: the searcher chooses")
	}
	if searchOptionNamed(g, c, "Command Tower") != uuid.Nil {
		t.Error("only basic land cards are offered")
	}
	answerSearchNamed(t, g, me.ID, "Island")
	island := findBattlefieldByName(g, "Island")
	if island == uuid.Nil || !b16Tapped(t, g, island) {
		t.Error("the basic enters tapped")
	}
}

// --- the spells ----------------------------------------------------

func TestB21LightningStrikeAndGutShotBurnAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	castCatalogSpell(t, g, "Lightning Strike", "Instant", b21LightningStrikeOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != 37 {
		t.Errorf("Strike to the face: life %d, want 37", opp.Life)
	}
	castCatalogSpell(t, g, "Gut Shot", "Instant", b21GutShotOracle, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if damageMarkedOn(g, bear) != 1 {
		t.Errorf("Gut Shot marks 1 on the creature, marked %d", damageMarkedOn(g, bear))
	}
	castCatalogSpell(t, g, "Gut Shot", "Instant", b21GutShotOracle, b16TargetPlayer(me.ID))
	passPriorityAroundTable(t, g)
	if me.Life != 39 {
		t.Errorf("Gut Shot to yourself: life %d, want 39", me.Life)
	}
	// #916 gave the cast prompt the stepper that pays {R/P} with 2
	// life, which was the only thing this card declared.
	if spec, _ := Lookup(b21GutShotOracle); spec.Completeness != CompletenessFull {
		t.Error("the Phyrexian mana gap is closed; Gut Shot is complete")
	}
}

func TestB21EndTheFestivitiesPingsOpponentsAndTheirBoards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := b16Creature(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1, "R")
	theirs := b16Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1, "R")
	big := b16Creature(g, other.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Jace", TypeLine: "Legendary Planeswalker — Jace", Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{game.CounterLoyalty: 3},
	})
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	before := b17Life(g)
	castCatalogSpell(t, g, "End the Festivities", "Sorcery", b21EndTheFestivitiesOracle, nil)
	passPriorityAroundTable(t, g)
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
	if !g.Battlefield.Contains(big) || damageMarkedOn(g, big) != 1 {
		t.Error("an opponent's 2/2 takes 1 and lives")
	}
	if counterCount(g, walker, game.CounterLoyalty) != 2 {
		t.Errorf("an opponent's planeswalker loses a loyalty: %d, want 2", counterCount(g, walker, game.CounterLoyalty))
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("a noncreature artifact is not hit")
	}
}

func TestB21WillOfTheMarduMakesWarriorsOrBurnsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b16Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2, "G")
	b16Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2, "G")
	b16Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Wurm", "Creature — Wurm", 5, 5, "G")
	b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	castModal(t, g, "Will of the Mardu", "Instant", b21WillOfTheMarduOracle, []int{0}, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	warriors := battlefieldIDsNamed(g, "Warrior")
	if len(warriors) != 2 {
		t.Fatalf("target player controls 2 creatures: 2 Warriors, got %d", len(warriors))
	}
	if c := cardByID(g, warriors[0]); controllerOf(t, g, warriors[0]) != me.ID || len(c.Colors) != 1 || c.Colors[0] != "R" || effectivePower(t, g, warriors[0]) != 1 {
		t.Error("1/1 red Warriors under your control")
	}
	// Mode two: damage equal to YOUR creature count — three Bears
	// and two Warriors.
	castModal(t, g, "Will of the Mardu", "Instant", b21WillOfTheMarduOracle, []int{1}, b16TargetCard(theirs))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("five creatures you control: 5 damage kills the 5/5")
	}
	// The declared gap: never both.
	id := handCardFull(me, "Will of the Mardu", "Instant", "{2}{W}", b21WillOfTheMarduOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0, 1}, Targets: b16TargetPlayer(opp.ID)}); err == nil {
		t.Error("choose one — both modes is refused (the commander rider is the declared gap)")
	}
	if spec, _ := Lookup(b21WillOfTheMarduOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the choose-both gap must be declared")
	}
}

func TestB21BrilliantRestorationReturnsArtifactsAndEnchantments(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b17GraveyardCard(me, "Dead Rock", "Artifact", "{2}")
	shrine := b17GraveyardCard(me, "Dead Shrine", "Enchantment", "{1}{W}")
	myr := b17GraveyardCard(me, "Dead Myr", "Artifact Creature — Myr", "{1}")
	bear := b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	land := b17GraveyardCard(me, "Dead Forest", "Basic Land — Forest", "")
	theirs := b17GraveyardCard(opp, "Their Dead Rock", "Artifact", "{2}")
	castCatalogSpell(t, g, "Brilliant Restoration", "Sorcery", b21BrilliantRestorationOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{rock, shrine, myr} {
		if !g.Battlefield.Contains(id) || controllerOf(t, g, id) != me.ID {
			t.Errorf("%s returns to the battlefield under your control", cardByID(g, id).Name)
		}
	}
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(land) {
		t.Error("a creature card and a land card stay in the graveyard")
	}
	if g.Battlefield.Contains(theirs) || !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's graveyard is not yours")
	}
	if spec, _ := Lookup(b21BrilliantRestorationOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the unattached-Aura gap must be declared")
	}
}

func TestB21SpitefulBanditrySweepsForXAndTreasuresOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := b16Creature(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1, "R")
	small := b16Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1, "R")
	alsoSmall := b16Creature(g, other.ID, "Their Elf", "Creature — Elf", 1, 1, "G")
	big := b16Creature(g, opp.ID, "Their Wurm", "Creature — Wurm", 5, 5, "G")
	banditry := castXSpell(t, g, "Spiteful Banditry", "Enchantment", b21SpitefulBanditryOracle, "{X}{R}{R}", 2, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(banditry) {
		t.Fatal("the enchantment enters")
	}
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(small) || g.Battlefield.Contains(alsoSmall) {
		t.Error("X=2 kills every creature with toughness 2 or less, yours included")
	}
	if !g.Battlefield.Contains(big) || damageMarkedOn(g, big) != 2 {
		t.Error("the 5/5 takes 2 and lives")
	}
	// Two opponents' creatures died at once: ONE Treasure, and the
	// Banditry was on the battlefield to see it.
	if n := b16CountNamed(g, "Treasure"); n != 1 {
		t.Fatalf("one or more opponents' creatures died: 1 Treasure, got %d", n)
	}
	// A later death this turn: nothing (once each turn).
	b18Kill(t, g, big)
	if n := b16CountNamed(g, "Treasure"); n != 1 {
		t.Errorf("the ability triggers only once each turn: %d Treasures", n)
	}
	// Next turn: an opponent's creature dying triggers again; your
	// own does not.
	advanceToMainOf(t, g, 1)
	yours := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	b18Kill(t, g, yours)
	if n := b16CountNamed(g, "Treasure"); n != 1 {
		t.Errorf("your own creature dying is not an opponent's: %d Treasures", n)
	}
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	b18Kill(t, g, theirs)
	if n := b16CountNamed(g, "Treasure"); n != 2 {
		t.Errorf("a new turn: 2 Treasures, got %d", n)
	}
	if spec, _ := Lookup(b21SpitefulBanditryOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the X-as-the-spell-resolves gap must be declared")
	}
}

// --- the creatures: vanilla, mana, ETB -----------------------------

func TestB21MemniteIsAWholeVanillaArtifactCreature(t *testing.T) {
	g := newCatalogGame(t)
	memnite := castCatalogSpell(t, g, "Memnite", "Artifact Creature — Construct", b21MemniteOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(memnite) {
		t.Fatal("a zero-cost artifact creature resolves to the battlefield")
	}
	spec, _ := Lookup(b21MemniteOracle)
	if spec.Completeness != CompletenessFull || len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.PrintedKeywords) != 0 {
		t.Error("no rules text: a whole vanilla spec")
	}
}

func TestB21LlanowarTribeTapsForThreeGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tribe := b21Push(g, me.ID, "Llanowar Tribe", "Creature — Elf Druid", b21LlanowarTribeOracle, 3, 3, "G")
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, tribe, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 3 || got[0] != "G" || got[1] != "G" || got[2] != "G" {
		t.Errorf("pool %v, want GGG", got)
	}
	if !b16Tapped(t, g, tribe) {
		t.Error("the Tribe taps")
	}
	// A summoning-sick Tribe cannot tap.
	sick := castCatalogSpell(t, g, "Llanowar Tribe", "Creature — Elf Druid", b21LlanowarTribeOracle, nil)
	passPriorityAroundTable(t, g)
	if err := g.ActivateManaAbility(me.ID, sick, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a creature's tap mana ability waits out summoning sickness")
	}
}

func TestB21ElvishVisionaryDrawsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	visionary := castCatalogSpell(t, g, "Elvish Visionary", "Creature — Elf Shaman", b21ElvishVisionaryOracle, nil)
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if triggerOnStack(g, visionary) == nil {
		t.Fatal("the draw is a trigger on the stack")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("the seeded card was cast and one was drawn: hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestB21DemonsDiscipleEdictsCreaturesAndPlaneswalkers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other, third := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	bear := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Jace", TypeLine: "Legendary Planeswalker — Jace", Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{game.CounterLoyalty: 3},
	})
	rock := b12Permanent(g, other.ID, "Their Rock", "Artifact")
	disciple := castCatalogSpell(t, g, "Demon's Disciple", "Creature — Human Cleric", b21DemonsDiscipleOracle, nil)
	passPriorityAroundTable(t, g)
	mine := b19SacrificeOptions(g, me.ID)
	if len(mine) != 2 || !hasID(mine, bear) || !hasID(mine, disciple) {
		t.Errorf("EACH player, the Disciple itself included: %v", mine)
	}
	opts := b19SacrificeOptions(g, opp.ID)
	if len(opts) != 2 || !hasID(opts, theirs) || !hasID(opts, walker) {
		t.Errorf("a creature OR a planeswalker: %v", opts)
	}
	if sacrificeChoiceFor(g, other.ID) != nil || sacrificeChoiceFor(g, third.ID) != nil {
		t.Error("a player with only an artifact, or nothing, is skipped")
	}
	answerSacrifice(t, g, me.ID, disciple)
	answerSacrifice(t, g, opp.ID, walker)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(disciple) || g.Battlefield.Contains(walker) {
		t.Error("the chosen permanents are sacrificed")
	}
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(theirs) || !g.Battlefield.Contains(rock) {
		t.Error("and nothing else")
	}
}

func TestB21TrophyMageTutorsAThreeDropArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Basalt Monolith", TypeLine: "Artifact", ManaCost: "{3}"},
		game.Card{Name: "Crucible of Worlds", TypeLine: "Artifact", ManaCost: "{3}"},
		game.Card{Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}"},
		game.Card{Name: "Solemn", TypeLine: "Artifact Creature — Golem", ManaCost: "{4}"},
		game.Card{Name: "Three-Drop Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{G}"},
	)
	castCatalogSpell(t, g, "Trophy Mage", "Creature — Human Wizard", b21TrophyMageOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("you may search: the chooser asks")
	}
	for _, name := range []string{"Basalt Monolith", "Crucible of Worlds"} {
		if searchOptionNamed(g, c, name) == uuid.Nil {
			t.Errorf("%s (an artifact with mana value 3) is offered", name)
		}
	}
	for _, name := range []string{"Mind Stone", "Solemn", "Three-Drop Bear"} {
		if searchOptionNamed(g, c, name) != uuid.Nil {
			t.Errorf("%s is not", name)
		}
	}
	answerSearchNamed(t, g, me.ID, "Crucible of Worlds")
	if !b02bHandHasNamed(me, "Crucible of Worlds") {
		t.Error("the pick goes to hand")
	}
}

func TestB21HermitDruidRevealsToABasicLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	druid := b21Push(g, me.ID, "Hermit Druid", "Creature — Human Druid", b21HermitDruidOracle, 1, 1, "G")
	// seedSearchLibrary lists the library top-first.
	ids := seedSearchLibrary(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Nonbasic", TypeLine: "Land"},
		game.Card{Name: "First Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Second Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	bolt, nonbasic, first, second, deep := ids[0], ids[1], ids[2], ids[3], ids[4]
	advanceToMain(t, g)
	hand, yard := me.Hand.Size(), me.Graveyard.Size()
	b06AddMana(me, "G")
	b16Activate(t, g, me.ID, druid, 0, game.ActivateAbilityParams{})
	if !b16Tapped(t, g, druid) {
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
	if c := cardByID(g, first); !c.IsKnownTo(g.Seats[1].ID) {
		t.Error("the land was revealed")
	}
	// No basic land: the whole library goes.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(druid) })
	seedSearchLibrary(me,
		game.Card{Name: "Only Sorcery", TypeLine: "Sorcery"},
		game.Card{Name: "Only Instant", TypeLine: "Instant"},
	)
	b06AddMana(me, "G")
	b16Activate(t, g, me.ID, druid, 0, game.ActivateAbilityParams{})
	if me.Library.Size() != 0 || me.Hand.Size() != hand+1 {
		t.Error("with no basic land every card is milled and none reaches the hand")
	}
}

func TestB21RoyalAssassinDestroysOnlyTappedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	assassin := b21Push(g, me.ID, "Royal Assassin", "Creature — Human Assassin", b21RoyalAssassinOracle, 1, 1, "B")
	untapped := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	tapped := b16Creature(g, opp.ID, "Their Tapped Bear", "Creature — Bear", 2, 2, "G")
	b16Tap(g, tapped)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, assassin, 0, game.ActivateAbilityParams{Targets: b16TargetCard(untapped)}); err == nil {
		t.Fatal("an untapped creature is not a legal target")
	}
	b16Activate(t, g, me.ID, assassin, 0, game.ActivateAbilityParams{Targets: b16TargetCard(tapped)})
	if g.Battlefield.Contains(tapped) || !opp.Graveyard.Contains(tapped) {
		t.Error("the tapped creature is destroyed")
	}
	if !b16Tapped(t, g, assassin) {
		t.Error("the Assassin taps as its cost")
	}
	// An attacker is tapped by its declaration: prey.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(assassin) })
	attacker := b16Creature(g, opp.ID, "Their Attacker", "Creature — Bear", 2, 2, "G")
	advanceToMainOf(t, g, 1)
	declareAttack(t, g, me.ID, attacker)
	if err := g.ActivateCatalogAbility(me.ID, assassin, 0, game.ActivateAbilityParams{Targets: b16TargetCard(attacker)}); err != nil {
		t.Fatalf("an attacking creature is tapped: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacker is destroyed")
	}
}

func TestB21ManglehornMayDestroyAnArtifactAndTapsTheirs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myRock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	theirRock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	castAndResolveCreature(t, g, "Manglehorn", "Creature — Beast", b21ManglehornOracle)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, myRock) || !hasID(p.PickTargetCards, theirRock) || hasID(p.PickTargetCards, bear) {
		t.Error("target ARTIFACT — anyone's, and not a creature")
	}
	pickCard(t, g, me.ID, theirRock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirRock) || !g.Battlefield.Contains(myRock) {
		t.Error("the chosen artifact is destroyed")
	}
	// An opponent's cast artifact enters tapped; yours does not; an
	// opponent's artifact creature is an artifact.
	advanceToMainOf(t, g, 1)
	signet := castCatalogSpell(t, g, "Their Signet", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	top100AssertEnteredTapped(t, g, signet, "an opponent's artifact")
	myr := castCatalogSpell(t, g, "Their Myr", "Artifact Creature — Myr", "", nil)
	passPriorityAroundTable(t, g)
	top100AssertEnteredTapped(t, g, myr, "an opponent's artifact creature")
	theirBear := castCatalogSpell(t, g, "Their Other Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, theirBear) {
		t.Error("a non-artifact is untouched")
	}
	advanceToMainOf(t, g, 0)
	mine := castCatalogSpell(t, g, "My Signet", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, mine) {
		t.Error("your own artifacts enter untapped")
	}
	// Declined: nothing is destroyed.
	castAndResolveCreature(t, g, "Manglehorn", "Creature — Beast", b21ManglehornOracle)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(myRock) || !g.Battlefield.Contains(mine) {
		t.Error("declining destroys nothing")
	}
	// #762: an opponent's artifact TOKEN — a Treasure — enters tapped
	// too, because token creation now runs the entry pipeline.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TreasureToken(), 1) })
	if !b16Tapped(t, g, findBattlefieldByName(g, "Treasure")) {
		t.Error("an opponent's artifact TOKEN enters tapped")
	}
	if spec, _ := Lookup(b21ManglehornOracle); spec.Completeness != CompletenessFull {
		t.Error("the token gap is closed — the caveat must be gone")
	}
}

// --- the creatures: cast, draw and discard triggers ---------------

func TestB21EnchantressesDrawOnEnchantmentSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b21Push(g, me.ID, "Satyr Enchanter", "Creature — Satyr Druid", b21SatyrEnchanterOracle, 2, 2, "G", "W")
	b21Push(g, me.ID, "Verduran Enchantress", "Creature — Human Druid", b21VerduranEnchantressOracle, 0, 2, "G")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Shrine", "Enchantment", "", nil)
	// The Satyr's draw is mandatory; the Enchantress asks.
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("the seeded card was cast and two were drawn: hand %d, want %d", me.Hand.Size(), hand+2)
	}
	hand = me.Hand.Size()
	castCatalogSpell(t, g, "Kami", "Enchantment Creature — Spirit", "", nil)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("an enchantment creature is an enchantment spell; the Enchantress declined: hand %d, want %d", me.Hand.Size(), hand+1)
	}
	hand = me.Hand.Size()
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if b19HasTriggerPromptFor(g, me.ID) {
		t.Error("a creature spell is not an enchantment spell")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("no draw")
	}
	// An opponent's enchantment is not yours.
	advanceToMainOf(t, g, 1)
	castCatalogSpell(t, g, "Their Shrine", "Enchantment", "", nil)
	if b19HasTriggerPromptFor(g, me.ID) || len(g.PendingTriggers) != 0 || triggerOnStack(g, uuid.Nil) != nil {
		t.Error("an opponent casting an enchantment triggers nothing")
	}
	passPriorityAroundTable(t, g)
}

func TestB21LysAlanaHuntmasterMayMakeAnElfPerElfSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b21Push(g, me.ID, "Lys Alana Huntmaster", "Creature — Elf Warrior", b21LysAlanaHuntmasterOracle, 3, 3, "G")
	castCatalogSpell(t, g, "Llanowar Elves", "Creature — Elf Druid", "", nil)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	warriors := battlefieldIDsNamed(g, "Elf Warrior")
	if len(warriors) != 1 {
		t.Fatalf("one Elf Warrior, got %d", len(warriors))
	}
	if c := cardByID(g, warriors[0]); effectivePower(t, g, warriors[0]) != 1 || len(c.Colors) != 1 || c.Colors[0] != "G" || !c.HasSubtype("Elf") {
		t.Error("a 1/1 green Elf Warrior")
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if b19HasTriggerPromptFor(g, me.ID) {
		t.Error("a non-Elf spell does not trigger")
	}
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Elvish Visionary", "Creature — Elf Shaman", b21ElvishVisionaryOracle, nil)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if b16CountNamed(g, "Elf Warrior") != 1 {
		t.Error("declining makes nothing")
	}
}

func TestB21GethsGrimoireMayDrawPerOpponentDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Geth's Grimoire", "Artifact — Book", b21GethsGrimoireOracle, 0, 0)
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 2) })
	if n := b18TriggerPromptCount(g, me.ID); n != 2 {
		t.Fatalf("two discards: two prompts, got %d", n)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("one yes, one no: drew %d, want 1", me.Hand.Size()-hand)
	}
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	if b19HasTriggerPromptFor(g, me.ID) {
		t.Error("your own discard is not an opponent's")
	}
}

func TestB21MindsEyeMayPayOneToDrawPerOpponentDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Mind's Eye", "Artifact", b21MindsEyeOracle, 0, 0)
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("an opponent drew: you are asked to pay {1}")
	}
	b06AddMana(me, "C")
	answerPayUnless(t, g, me.ID, true)
	if me.Hand.Size() != hand+1 {
		t.Errorf("paid: drew %d, want 1", me.Hand.Size()-hand)
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	if me.Hand.Size() != hand+1 {
		t.Error("declined: no draw")
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("your own draw is not an opponent's")
	}
}

func TestB21SpitefulVisionsDrawsExtraAndTaxesEveryDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b21Push(g, me.ID, "Spiteful Visions", "Enchantment", b21SpitefulVisionsOracle, 0, 0)
	// On to the NEXT seat's draw step — seat 0's own is already
	// behind the cursor (newCatalogGame parks there, #692).
	advanceTo(t, g, game.StepEnd)
	advanceTo(t, g, game.StepDraw)
	active := g.Seats[g.Turn.ActiveSeat]
	if active.ID == me.ID {
		t.Fatal("seat 1's draw step")
	}
	hand, life := active.Hand.Size(), active.Life
	// The turn-based draw already happened and was taxed; its damage
	// trigger and the extra-draw trigger arrive together and the
	// controller orders them.
	b21SettleStack(t, g)
	if active.Hand.Size() != hand+1 {
		t.Errorf("the additional draw: hand %d, want %d", active.Hand.Size(), hand+1)
	}
	if active.Life != life-2 {
		t.Errorf("two draws this step, 1 damage each: life %d, want %d", active.Life, life-2)
	}
	// Your own draw is taxed too.
	mine := me.Life
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 2) })
	b21SettleStack(t, g)
	if me.Life != mine-2 {
		t.Errorf("a player — you included — draws two: 2 damage, life %d, want %d", me.Life, mine-2)
	}
}

// --- the creatures: dies and drain triggers -----------------------

func TestB21KokushoDrainsTheTableOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kokusho := b21Push(g, me.ID, "Kokusho, the Evening Star", "Legendary Creature — Dragon Spirit", b21KokushoOracle, 5, 5, "B")
	if !hasEffectiveKeyword(t, g, kokusho, "flying") {
		t.Error("flying")
	}
	before := b17Life(g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(kokusho) })
	if triggerOnStack(g, kokusho) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("the drain is a trigger")
	}
	passPriorityAroundTable(t, g)
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
	again := b21Push(g, me.ID, "Kokusho, the Evening Star", "Legendary Creature — Dragon Spirit", b21KokushoOracle, 5, 5, "B")
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(again) })
	passPriorityAroundTable(t, g)
	if g.Seats[1].Life != before[1]-5 {
		t.Error("bounced, not died: no drain")
	}
}

func TestB21PrizedStatueMakesTreasureOnEntryAndOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	statue := castCatalogSpell(t, g, "Prized Statue", "Artifact", b21PrizedStatueOracle, nil)
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Treasure"); n != 1 {
		t.Fatalf("entering: 1 Treasure, got %d", n)
	}
	b18Kill(t, g, statue)
	if n := b16CountNamed(g, "Treasure"); n != 2 {
		t.Fatalf("dying: 2 Treasures, got %d", n)
	}
	for _, id := range battlefieldIDsNamed(g, "Treasure") {
		if controllerOf(t, g, id) != me.ID {
			t.Error("the Treasures are the Statue's controller's")
		}
	}
	// Exiled, not died: nothing.
	other := b21Push(g, me.ID, "Prized Statue", "Artifact", b21PrizedStatueOracle, 0, 0)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(other) })
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Treasure"); n != 2 {
		t.Errorf("an exile is not a death: %d Treasures", n)
	}
	_ = opp
}

func TestB21DiregrafCaptainLordsZombiesAndDrainsOnTheirDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	captain := b21Push(g, me.ID, "Diregraf Captain", "Creature — Zombie Soldier", b21DiregrafCaptainOracle, 2, 2, "U", "B")
	zombie := b16Creature(g, me.ID, "Zombie", "Creature — Zombie", 2, 2, "B")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2, "B")
	if !hasEffectiveKeyword(t, g, captain, "deathtouch") {
		t.Error("deathtouch")
	}
	if effectivePower(t, g, zombie) != 3 || effectiveToughness(t, g, zombie) != 3 {
		t.Error("other Zombies you control get +1/+1")
	}
	if effectivePower(t, g, captain) != 2 || effectivePower(t, g, bear) != 2 || effectivePower(t, g, theirs) != 2 {
		t.Error("not itself, not a non-Zombie, not an opponent's")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(zombie) })
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if hasID(p.PickTargetPlayers, me.ID) || !hasID(p.PickTargetPlayers, opp.ID) || !hasID(p.PickTargetPlayers, other.ID) {
		t.Error("target OPPONENT")
	}
	pickPlayer(t, g, me.ID, other.ID)
	passPriorityAroundTable(t, g)
	if other.Life != 39 || opp.Life != 40 {
		t.Errorf("the chosen opponent loses 1: %d / %d", other.Life, opp.Life)
	}
	// Not a Zombie, not yours, not the Captain itself.
	b18Kill(t, g, bear)
	b18Kill(t, g, theirs)
	b18Kill(t, g, captain)
	if latestPickTarget(g, me.ID) != nil || other.Life != 39 {
		t.Error("a non-Zombie, an opponent's Zombie and the Captain itself drain nothing")
	}
}

func TestB21AgentOfTheIronThroneDrainsWhileYouControlYourCommander(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Agent of the Iron Throne", "Legendary Enchantment — Background", b21AgentOfTheIronThroneOracle, 0, 0, "B")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	before := b17Life(g)
	// No commander creature out: the granted ability does not exist.
	b18Kill(t, g, bear)
	if g.Seats[1].Life != before[1] {
		t.Fatal("without a commander creature you own, nothing drains")
	}
	commander := b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human Knight", 3, 3)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	shrine := b12Permanent(g, me.ID, "Shrine", "Enchantment")
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1) })
	treasure := findBattlefieldByName(g, "Treasure")
	// An artifact, a token: each drains. An enchantment, an
	// opponent's creature: not.
	b18Kill(t, g, rock)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(treasure) })
	passPriorityAroundTable(t, g)
	b18Kill(t, g, shrine)
	b18Kill(t, g, theirs)
	for i := 1; i < 4; i++ {
		if g.Seats[i].Life != before[i]-2 {
			t.Errorf("seat %d life %d, want %d — an artifact and a Treasure drained, a shrine and their creature did not", i, g.Seats[i].Life, before[i]-2)
		}
	}
	if me.Life != before[0] {
		t.Error("you lose nothing")
	}
	// The commander's own death looks back and drains — once its
	// owner declines the command zone, so it really reaches the
	// graveyard.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(commander) })
	b21DeclineCommandZone(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if g.Seats[1].Life != before[1]-3 {
		t.Errorf("the commander dying drains: life %d, want %d", g.Seats[1].Life, before[1]-3)
	}
	// A commander creature an OPPONENT owns is not yours.
	b21Commander(g, opp.ID, "Their Commander", "Legendary Creature — Elf", 3, 3)
	mine := b16Creature(g, me.ID, "Another Bear", "Creature — Bear", 2, 2, "G")
	b18Kill(t, g, mine)
	if g.Seats[1].Life != before[1]-3 {
		t.Error("an opponent's commander does not carry your Background's ability")
	}
	if spec, _ := Lookup(b21AgentOfTheIronThroneOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the stolen-commander gap must be declared")
	}
}

// --- the creatures: attack and damage triggers --------------------

func TestB21SanctumSeekerDrainsPerAttackingVampire(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seeker := b21Push(g, me.ID, "Sanctum Seeker", "Creature — Vampire Knight", b21SanctumSeekerOracle, 3, 4, "B")
	vampire := b16Creature(g, me.ID, "Vampire", "Creature — Vampire", 2, 2, "B")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	before := b17Life(g)
	declareAttack(t, g, opp.ID, seeker, vampire, bear)
	if n := triggersOnStackFrom(g, seeker); n != 2 {
		t.Fatalf("two attacking Vampires (the Seeker included): 2 triggers, got %d", n)
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kreat := b21Push(g, me.ID, "General Kreat, the Boltbringer", "Legendary Creature — Goblin Soldier", b21GeneralKreatOracle, 2, 2, "R")
	goblin := b16Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1, "R")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	before := b17Life(g)
	declareAttack(t, g, opp.ID, kreat, goblin, bear)
	if n := triggersOnStackFrom(g, kreat); n != 1 {
		t.Fatalf("one or more Goblins attack: 1 trigger, got %d", n)
	}
	passPriorityAroundTable(t, g)
	tokens := battlefieldIDsNamed(g, "Goblin")
	var token uuid.UUID
	for _, id := range tokens {
		if id != goblin {
			token = id
		}
	}
	if token == uuid.Nil {
		t.Fatal("a Goblin token is created")
	}
	tok := cardByID(g, token)
	if !tok.Tapped || tok.AttackingTarget != opp.ID || effectivePower(t, g, token) != 1 {
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
	if n := triggersOnStackFrom(g, kreat); n != 0 || len(g.PendingTriggers) != 0 {
		t.Error("an attacking token is not a declared attacker")
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != before[1]-1-2-1-2-1 {
		t.Errorf("the ping, then Kreat, the Goblin, the Bear and the token connect: life %d, want %d", opp.Life, before[1]-7)
	}
	// A non-Goblin attack alone makes nothing; any creature entering
	// pings.
	advanceToMainOf(t, g, 0)
	life := opp.Life
	castCatalogSpell(t, g, "Elf", "Creature — Elf", "", nil)
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Error("another creature you control entering pings")
	}
	declareAttack(t, g, opp.ID, bear)
	if triggersOnStackFrom(g, kreat) != 0 {
		t.Error("a Bear attacking is not a Goblin attacking")
	}
	if spec, _ := Lookup(b21GeneralKreatOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the defender-choice gap must be declared")
	}
}

func TestB21WrathfulRedDragonReflectsDamageDealtToYourDragons(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	wrathful := b21Push(g, me.ID, "Wrathful Red Dragon", "Creature — Dragon", b21WrathfulRedDragonOracle, 5, 5, "R")
	dragon := b16Creature(g, me.ID, "Dragon", "Creature — Dragon", 4, 4, "R")
	theirDragon := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Dragon", TypeLine: "Creature — Dragon", Colors: []string{"R"},
		Keywords: []string{"flying"}, Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID,
	})
	theirBear := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	if !hasEffectiveKeyword(t, g, wrathful, "flying") {
		t.Error("flying")
	}
	advanceToMainOf(t, g, 1)
	b13OpponentCasts(t, g, opp, "Lightning Bolt", "Instant", lightningBoltOracle, "", b16TargetCard(dragon))
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetPlayers, opp.ID) || !hasID(p.PickTargetCards, theirBear) {
		t.Error("any target … a player, a non-Dragon creature")
	}
	if hasID(p.PickTargetCards, theirDragon) || hasID(p.PickTargetCards, wrathful) || hasID(p.PickTargetCards, dragon) {
		t.Error("… that isn't a Dragon")
	}
	pickPlayer(t, g, me.ID, other.ID)
	passPriorityAroundTable(t, g)
	if other.Life != 37 {
		t.Errorf("the Dragon took 3, so it deals 3: life %d, want 37", other.Life)
	}
	// Combat damage to the Wrathful itself reflects too; an
	// opponent's Dragon is not yours.
	advanceToMainOf(t, g, 0)
	declareAttack(t, g, opp.ID, wrathful)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(theirDragon, wrathful); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirBear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirBear) {
		t.Error("the Wrathful took 4 from the blocker and deals 4 to the chosen Bear")
	}
	if latestPickTarget(g, opp.ID) != nil {
		t.Error("their Dragon taking damage is not a Dragon YOU control")
	}
	if spec, _ := Lookup(b21WrathfulRedDragonOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the per-source batching gap must be declared")
	}
}

func TestB21HazelsBrewmasterExilesAGraveyardCardAndMakesFood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := b17GraveyardCard(opp, "Their Dead Bear", "Creature — Bear", "{1}{G}")
	mine := b17GraveyardCard(me, "My Dead Bolt", "Instant", "{R}")
	brewmaster := castAndResolveCreature(t, g, "Hazel's Brewmaster", "Creature — Squirrel Warlock", b21HazelsBrewmasterOracle)
	if !hasEffectiveKeyword(t, g, brewmaster, "menace") {
		t.Error("menace")
	}
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, dead) || !hasID(p.PickTargetCards, mine) {
		t.Error("target card from A graveyard — anyone's")
	}
	if p.PickTargetMin != 0 {
		t.Error("up to one")
	}
	pickCard(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	if !inExile(g, dead) {
		t.Error("the chosen card is exiled")
	}
	if n := b16CountNamed(g, "Food"); n != 1 {
		t.Fatalf("and a Food is created: %d", n)
	}
	// Attacking triggers again; with no target chosen the Food still
	// comes.
	advanceToPrecombatMainOf(t, g, 0)
	declareAttack(t, g, opp.ID, brewmaster)
	b04WaitForPick(t, g, me.ID)
	p = latestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
		t.Fatalf("picking no target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(mine) || b16CountNamed(g, "Food") != 2 {
		t.Error("no target: nothing exiled, a Food made")
	}
	// Every graveyard empty: no target prompt, and the Food still
	// comes.
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(mine) })
	for _, p := range g.Seats {
		p.Graveyard.Cards = nil
	}
	castAndResolveCreature(t, g, "Hazel's Brewmaster", "Creature — Squirrel Warlock", b21HazelsBrewmasterOracle)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("nothing to target: no prompt")
	}
	passPriorityAroundTable(t, g)
	if b16CountNamed(g, "Food") != 3 {
		t.Error("with every graveyard empty the trigger still makes a Food")
	}
	if spec, _ := Lookup(b21HazelsBrewmasterOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the Food-abilities gap must be declared")
	}
}

func TestB21OblivionSowerExilesFourAndTakesTheirLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// seedSearchLibrary lists the library top-first.
	ids := seedSearchLibrary(opp,
		game.Card{Name: "Their Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Their Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Their Bolt", TypeLine: "Instant"},
		game.Card{Name: "Their Swamp", TypeLine: "Basic Land — Swamp"},
		game.Card{Name: "Fifth Card Forest", TypeLine: "Basic Land — Forest"},
	)
	bear, island, bolt, swamp, fifth := ids[0], ids[1], ids[2], ids[3], ids[4]
	earlier := b21ExileCard(g, opp.ID, "Their Exiled Mountain", "Basic Land — Mountain")
	notTheirs := b21ExileCard(g, g.Seats[2].ID, "Other Exiled Plains", "Basic Land — Plains")
	sower := castCatalogSpell(t, g, "Oblivion Sower", "Creature — Eldrazi", b21OblivionSowerOracle, nil)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if hasID(p.PickTargetPlayers, me.ID) || !hasID(p.PickTargetPlayers, opp.ID) {
		t.Error("target OPPONENT, chosen as the cast trigger goes on the stack")
	}
	pickPlayer(t, g, me.ID, opp.ID)
	if g.Stack.Size() != 1 || triggerOnStack(g, sower) == nil {
		t.Fatal("the trigger sits above the Sower on the stack")
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, bolt} {
		if !inExile(g, id) {
			t.Errorf("%s is exiled", cardByID(g, id).Name)
		}
	}
	if !opp.Library.Contains(fifth) {
		t.Error("only the top four")
	}
	// A returned card is a new object (CR 400.7), so look by name.
	for _, name := range []string{"Their Island", "Their Swamp", "Their Exiled Mountain"} {
		id := findBattlefieldByName(g, name)
		if id == uuid.Nil || controllerOf(t, g, id) != me.ID {
			t.Errorf("%s — a land that player owns in exile — enters under your control", name)
		}
	}
	for _, id := range []uuid.UUID{island, swamp, earlier} {
		if inExile(g, id) {
			t.Error("the land left exile")
		}
	}
	if !inExile(g, notTheirs) {
		t.Error("a land another player owns stays in exile")
	}
	if !g.Battlefield.Contains(sower) {
		t.Error("the Sower resolves after its trigger")
	}
	if spec, _ := Lookup(b21OblivionSowerOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the no-choice gap must be declared")
	}
}
