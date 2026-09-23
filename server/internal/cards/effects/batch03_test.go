package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch03_test.go — card-level coverage for the card-coverage
// roadmap's batch 03 (#296, `edhrec_rank` 361–471): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land drop or attack. Helper names
// are b03-prefixed; the shared ones come from batch01_test.go and
// the older suites.

const (
	b03CrumblingNecropolisOracle = "7190debf-708b-4f41-9714-0d0a5bd5a74e"
	b03LandTaxOracle             = "d2d9ecea-7925-420e-98b9-2f87f41f387c"
	b03ShamanicRevelationOracle  = "d1d171de-1c6d-4fb9-817a-9c689c709f3d"
	b03DimirAqueductOracle       = "378a1d57-e2f1-4b84-9692-1564602e9e99"
	b03NomadOutpostOracle        = "4619de7e-3d6e-4c6b-8e6e-e24db324839d"
	b03GoForTheThroatOracle      = "2f092562-9e17-43cd-aeb8-d0567f99363e"
	b03MysticMonasteryOracle     = "834b8f71-9a45-42ae-9e99-e749fa6fb45e"
	b03RampagingBalothsOracle    = "2d3e6549-6cc6-434f-a189-ba3b55e64c34"
	b03OpulentPalaceOracle       = "f9e7e855-1e3b-42d3-91b0-64ba8b5b8982"
	b03BadlandsOracle            = "13ff3222-91cb-4796-a34e-899ed817694c"
	b03GreatFurnaceOracle        = "f4819061-b0b5-48ab-af7b-6525c3d2eab7"
	b03ScrublandOracle           = "c8d95ca8-7d12-4072-aeaf-e20f248c7e39"
	b03DemolitionFieldOracle     = "93953926-a644-49bb-9b5a-4c8f19114c7e"
	b03FrontierBivouacOracle     = "e4cf6c2f-0f1e-4980-9ef9-e4eabcae42a9"
	b03GuttersnipeOracle         = "c6bdaf76-6a03-4695-9c4b-f040e73435af"
	b03SeasideCitadelOracle      = "2ae77795-6a80-498b-bf69-6fd612f601e4"
	b03OrzhovBasilicaOracle      = "aa00ae0b-7c0f-427e-8102-ce0e2a6af5df"
	b03WarrenSoultraderOracle    = "ace86e56-efde-4eb7-8815-71456a4c3abe"
	b03BayouOracle               = "b76d1ae6-ad1d-4bac-b4c3-2e03e0e84d9b"
	b03IzzetBoilerworksOracle    = "1cb9d94a-3039-4f2e-8fcc-6996f9a45f74"
	b03HighMarketOracle          = "86fb3749-37d6-48a6-8524-71e996850307"
	b03TaigaOracle               = "22e3cf1d-3559-4ce1-954c-8dc815342979"
	b03RedElementalBlastOracle   = "bb329a5c-b9f9-4973-a53f-090024146325"
	b03GeierReachOracle          = "7b9fafe7-d26a-4ed5-b4c4-ce13763770b5"
	b03SramOracle                = "7e00b0cd-d212-4604-ba07-da21f4fe00b0"
	b03PlateauOracle             = "c7a15ca4-085f-4d92-8387-c3711c04c8fa"
	b03EntishRestorationOracle   = "736017e2-bc33-49e8-812d-1639443fdb51"
	b03UrzasCaveOracle           = "4474ecee-0ec3-409b-90df-738d9313fe3c"
	b03GruulTurfOracle           = "657243dd-e479-4f4b-99d2-09b55d833a35"
	b03DispatchOracle            = "133c99c0-3652-410f-8100-68015a47af9f"
	b03SavannahOracle            = "703243f0-8cb3-420f-958f-5fd4bde30293"
	b03ExpeditionMapOracle       = "8fcf50cd-e6d0-4516-850f-d42ee75dcc3a"
	b03SheoldredOracle           = "34f34409-326d-4994-a0ea-1a69aa278f03"
	b03LivingDeathOracle         = "9e6a3df4-67a3-452e-a6ef-f04dbadb21ef"
	b03MentalMisstepOracle       = "1a0770e6-b093-4439-baff-6889a50ba12e"
	b03PyroblastOracle           = "ecc435e2-deb1-420a-a79f-01dd08747314"
	b03FabricateOracle           = "422e1869-134f-463d-9fa1-86b66a998b3e"
)

// b03CastRefused pushes a card into the active seat's hand and
// reports whether CastSpell refuses it with the given params.
func b03CastRefused(t *testing.T, g *game.Game, name, typeLine, oracle string, params game.CastSpellParams) bool {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: me.ID, Controller: me.ID})
	return g.CastSpell(me.ID, id, params) != nil
}

// --- registration --------------------------------------------------

func TestBatch03CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b03CrumblingNecropolisOracle: "Crumbling Necropolis",
		b03LandTaxOracle:             "Land Tax",
		b03ShamanicRevelationOracle:  "Shamanic Revelation",
		b03DimirAqueductOracle:       "Dimir Aqueduct",
		b03NomadOutpostOracle:        "Nomad Outpost",
		b03GoForTheThroatOracle:      "Go for the Throat",
		b03MysticMonasteryOracle:     "Mystic Monastery",
		b03RampagingBalothsOracle:    "Rampaging Baloths",
		b03OpulentPalaceOracle:       "Opulent Palace",
		b03BadlandsOracle:            "Badlands",
		b03GreatFurnaceOracle:        "Great Furnace",
		b03ScrublandOracle:           "Scrubland",
		b03DemolitionFieldOracle:     "Demolition Field",
		b03FrontierBivouacOracle:     "Frontier Bivouac",
		b03GuttersnipeOracle:         "Guttersnipe",
		b03SeasideCitadelOracle:      "Seaside Citadel",
		b03OrzhovBasilicaOracle:      "Orzhov Basilica",
		b03WarrenSoultraderOracle:    "Warren Soultrader",
		b03BayouOracle:               "Bayou",
		b03IzzetBoilerworksOracle:    "Izzet Boilerworks",
		b03HighMarketOracle:          "High Market",
		b03TaigaOracle:               "Taiga",
		b03RedElementalBlastOracle:   "Red Elemental Blast",
		b03GeierReachOracle:          "Geier Reach Sanitarium",
		b03SramOracle:                "Sram, Senior Edificer",
		b03PlateauOracle:             "Plateau",
		b03EntishRestorationOracle:   "Entish Restoration",
		b03UrzasCaveOracle:           "Urza's Cave",
		b03GruulTurfOracle:           "Gruul Turf",
		b03DispatchOracle:            "Dispatch",
		b03SavannahOracle:            "Savannah",
		b03ExpeditionMapOracle:       "Expedition Map",
		b03SheoldredOracle:           "Sheoldred, the Apocalypse",
		b03LivingDeathOracle:         "Living Death",
		b03MentalMisstepOracle:       "Mental Misstep",
		b03PyroblastOracle:           "Pyroblast",
		b03FabricateOracle:           "Fabricate",
	}
	if len(want) != 37 {
		t.Fatalf("the batch is 37 cards, the table lists %d", len(want))
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
}

// The land tables, pinned row by row: a transposed colour in a loop
// is invisible until somebody plays that exact card.
func TestBatch03LandRowsProduceTheirPrintedColours(t *testing.T) {
	tri := map[string]string{
		"Crumbling Necropolis": "{U|B|R}",
		"Nomad Outpost":        "{R|W|B}",
		"Mystic Monastery":     "{U|R|W}",
		"Opulent Palace":       "{B|G|U}",
		"Frontier Bivouac":     "{G|U|R}",
		"Seaside Citadel":      "{G|W|U}",
	}
	dual := map[string]string{
		"Badlands": "{B|R}", "Scrubland": "{W|B}", "Bayou": "{B|G}",
		"Taiga": "{R|G}", "Plateau": "{R|W}", "Savannah": "{G|W}",
	}
	bounce := map[string]string{
		"Dimir Aqueduct": "{U}{B}", "Orzhov Basilica": "{W}{B}",
		"Izzet Boilerworks": "{U}{R}", "Gruul Turf": "{R}{G}",
	}
	found := 0
	for _, spec := range All() {
		if want, ok := tri[spec.Name]; ok {
			found++
			if len(spec.Replacements) != 1 || len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != want {
				t.Errorf("%s: want enters-tapped + %s, got %+v", spec.Name, want, spec.ManaAbilities)
			}
		}
		if want, ok := dual[spec.Name]; ok {
			found++
			if len(spec.Replacements) != 0 || len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != want {
				t.Errorf("%s: want an untapped %s dual, got %+v", spec.Name, want, spec.ManaAbilities)
			}
		}
		if want, ok := bounce[spec.Name]; ok {
			found++
			if len(spec.Replacements) != 1 || len(spec.Triggered) != 1 || len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != want {
				t.Errorf("%s: want enters-tapped + bounce trigger + %s, got %+v", spec.Name, want, spec.ManaAbilities)
			}
		}
	}
	if found != len(tri)+len(dual)+len(bounce) {
		t.Errorf("found %d of %d land rows", found, len(tri)+len(dual)+len(bounce))
	}
}

// --- lands ---------------------------------------------------------

func TestTriLandEntersTappedAndOffersThreeColours(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := playLandFromHand(t, g, "Crumbling Necropolis", b03CrumblingNecropolisOracle)
	top100AssertEnteredTapped(t, g, id, "Crumbling Necropolis")

	// Untap it by hand and tap for mana: a three-way pick.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 3 {
		t.Fatalf("a tri-land must ask which of three colours, got %+v", pick)
	}
}

func TestOriginalDualEntersUntappedAndTapsForEitherColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Badlands", TypeLine: "Land — Swamp Mountain",
		OracleID: b03BadlandsOracle, Owner: me.ID, Controller: me.ID})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play Badlands: %v", err)
	}
	top100AssertEnteredUntapped(t, g, id, "Badlands")
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("a dual must ask B or R, got %+v", pick)
	}
}

// #1337: the bounce is a resolution-time choice (ReturnOneYouControl
// / ChoosePermanents), not a target picked when the trigger goes on
// the stack.
func TestBounceLandEntersTappedAndReturnsAChosenLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	island := seedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	handBefore := me.Hand.Size()

	id := playLandFromHand(t, g, "Dimir Aqueduct", b03DimirAqueductOracle)
	top100AssertEnteredTapped(t, g, id, "Dimir Aqueduct")
	passPriorityAroundTable(t, g)
	p := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("Dimir Aqueduct queued no return pick")
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{island}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}

	if g.Battlefield.Contains(island) || !me.Hand.Contains(island) {
		t.Error("the chosen land should be back in hand")
	}
	// playLandFromHand adds the Aqueduct to the hand and plays it, so
	// the Island coming back is the only net change.
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d -> %d: want +1 (the Island came back)", handBefore, me.Hand.Size())
	}
	if !g.Battlefield.Contains(id) {
		t.Error("the Aqueduct itself should stay")
	}
}

func TestGreatFurnaceTapsForRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	furnace := seedPermanentWithOracle(g, me.ID, "Great Furnace", "Artifact Land", b03GreatFurnaceOracle)
	if err := g.ActivateManaAbility(me.ID, furnace, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want [R]", got)
	}
	if b03ArtifactsControlled(g, me.ID) != 1 {
		t.Error("an artifact land is an artifact")
	}
}

// --- utility lands -------------------------------------------------

func TestHighMarketSacrificesACreatureForALife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	market := pushCatalogPermanent(g, me.ID, "High Market", "Land", b03HighMarketOracle, false)
	bear := seedCreature(g, "Bear", me.ID)
	before := me.Life

	if err := g.ActivateCatalogAbility(me.ID, market, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{bear},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the sacrifice is a cost, paid at announce")
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d -> %d, want +1", before, me.Life)
	}
	card, _ := battlefieldCard(g, market)
	if !card.Tapped {
		t.Error("High Market has a tap cost")
	}
}

func TestGeierReachSanitariumEachPlayerLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sanitarium := pushCatalogPermanent(g, me.ID, "Geier Reach Sanitarium", "Legendary Land", b03GeierReachOracle, false)
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Hand.Size()
	}
	if err := g.ActivateCatalogAbility(me.ID, sanitarium, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Hand.Size() != before[i]+1 {
			t.Errorf("seat %d drew %d, want 1", i, p.Hand.Size()-before[i])
		}
		// One discard prompt per seat, addressed to that seat over
		// their own hand, so each player chooses their own card (the
		// Faithless Looting assertion).
		if got := discardOwed(g, p.ID); got != 1 {
			t.Errorf("seat %d is owed %d discards, want 1 — each player chooses their own", i, got)
		}
	}
}

func TestUrzasCaveFetchesAnyLandTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cave := pushCatalogPermanent(g, me.ID, "Urza's Cave", "Land — Urza's Cave", b03UrzasCaveOracle, false)
	coffers := stapleLibraryCard(me, "Cabal Coffers", "Legendary Land")
	stapleLibraryCard(me, "Forest", "Basic Land — Forest")

	if err := g.ActivateCatalogAbility(me.ID, cave, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, coffers)
	got, ok := battlefieldCard(g, coffers)
	if !ok {
		t.Fatal("the nonbasic land was not fetched")
	}
	if !got.Tapped {
		t.Error("Urza's Cave puts the land onto the battlefield TAPPED")
	}
	if g.Battlefield.Contains(cave) {
		t.Error("the Cave is sacrificed as a cost")
	}
}

func TestDemolitionFieldDestroysANonbasicAndOffersBothSearches(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	field := pushCatalogPermanent(g, me.ID, "Demolition Field", "Land", b03DemolitionFieldOracle, false)
	coffers := seedLandOnBattlefield(g, opp.ID, "Cabal Coffers", "Legendary Land")
	mine := stapleLibraryCard(me, "Forest", "Basic Land — Forest")
	stapleLibraryCard(opp, "Swamp", "Basic Land — Swamp")

	if err := g.ActivateCatalogAbility(me.ID, field, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: coffers}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(coffers) {
		t.Fatal("the nonbasic land survived")
	}
	if searchChoiceFor(g, opp.ID) == nil {
		t.Error("the victim gets a (declinable) search for a basic")
	}
	if searchChoiceFor(g, me.ID) == nil {
		t.Fatal("the controller gets a (declinable) search for a basic")
	}
	answerSearchByID(t, g, me.ID, mine)
	got, ok := battlefieldCard(g, mine)
	if !ok || got.Tapped {
		t.Error("the controller's basic enters UNTAPPED")
	}
}

func TestDemolitionFieldRefusesABasicLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	field := pushCatalogPermanent(g, me.ID, "Demolition Field", "Land", b03DemolitionFieldOracle, false)
	forest := seedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	if err := g.ActivateCatalogAbility(me.ID, field, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: forest}},
	}); err == nil {
		t.Error("a basic land was accepted for 'target nonbasic land'")
	}
}

func TestExpeditionMapTutorsALandToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mapID := pushCatalogPermanent(g, me.ID, "Expedition Map", "Artifact", b03ExpeditionMapOracle, false)
	urborg := stapleLibraryCard(me, "Urborg, Tomb of Yawgmoth", "Legendary Land")
	stapleLibraryCard(me, "Island", "Basic Land — Island")

	if err := g.ActivateCatalogAbility(me.ID, mapID, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, urborg)
	if !me.Hand.Contains(urborg) {
		t.Error("the land did not reach the hand")
	}
	if g.Battlefield.Contains(mapID) {
		t.Error("the Map is sacrificed as a cost")
	}
}

// --- spells --------------------------------------------------------

func TestFabricateTutorsAnArtifactToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ring := stapleLibraryCard(me, "Sol Ring", "Artifact")
	signet := stapleLibraryCard(me, "Arcane Signet", "Artifact")
	stapleLibraryCard(me, "Bear", "Creature — Bear")

	castCatalogSpell(t, g, "Fabricate", "Sorcery", b03FabricateOracle, nil)
	passPriorityAroundTable(t, g)
	// Two artifacts to choose from, so the chooser prompt opens.
	answerSearchByID(t, g, me.ID, ring)
	if !me.Hand.Contains(ring) {
		t.Error("the artifact did not reach the hand")
	}
	if me.Hand.Contains(signet) {
		t.Error("only the chosen artifact should come")
	}
}

func TestGoForTheThroatKillsANonartifactCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := seedCreature(g, "Bear", opp.ID)
	castCatalogSpell(t, g, "Go for the Throat", "Instant", b03GoForTheThroatOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature survived")
	}
}

func TestGoForTheThroatRefusesAnArtifactCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	golem := seedPermanentFor(g, opp.ID, "Golem", "Artifact Creature — Golem")
	if !b03CastRefused(t, g, "Go for the Throat", "Instant", b03GoForTheThroatOracle, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: golem}},
	}) {
		t.Error("an artifact creature was accepted for 'target nonartifact creature'")
	}
}

func TestDispatchTapsAndExilesWithMetalcraft(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		seedPermanentFor(g, me.ID, "Rock", "Artifact")
	}
	bear := seedCreature(g, "Bear", opp.ID)
	castCatalogSpell(t, g, "Dispatch", "Instant", b03DispatchOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) {
		t.Error("with three artifacts Dispatch exiles")
	}
}

func TestDispatchOnlyTapsWithoutMetalcraft(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedPermanentFor(g, me.ID, "Rock", "Artifact")
	seedPermanentFor(g, me.ID, "Rock", "Artifact")
	bear := seedCreature(g, "Bear", opp.ID)
	castCatalogSpell(t, g, "Dispatch", "Instant", b03DispatchOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, bear)
	if !ok {
		t.Fatal("with two artifacts Dispatch must not exile")
	}
	if !card.Tapped {
		t.Error("the creature should be tapped")
	}
}

func TestMentalMisstepCountersOnlyManaValueOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := batch01OpponentCasts(t, g, opp, "One Drop", lightningBoltOracle, "{U}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	castCatalogSpell(t, g, "Mental Misstep", "Instant", b03MentalMisstepOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(spell) {
		t.Fatal("the mana value 1 spell was not countered")
	}

	two := batch01OpponentCasts(t, g, opp, "Two Drop", lightningBoltOracle, "{1}{U}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if !b03CastRefused(t, g, "Mental Misstep", "Instant", b03MentalMisstepOracle, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: two}},
	}) {
		t.Error("a mana value 2 spell was accepted as a target")
	}
}

func TestRedElementalBlastCountersABlueSpellAndRefusesARedOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blue := batch01OpponentCasts(t, g, opp, "Blue Spell", lightningBoltOracle, "{U}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	castModal(t, g, "Red Elemental Blast", "Instant", b03RedElementalBlastOracle, []int{0},
		[]game.TargetRef{{Kind: game.TargetCard, ID: blue}})
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(blue) {
		t.Fatal("the blue spell was not countered")
	}

	red := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if !b03CastRefused(t, g, "Red Elemental Blast", "Instant", b03RedElementalBlastOracle, game.CastSpellParams{
		Modes:   []int{0},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: red}},
	}) {
		t.Error("a red spell was accepted for 'target blue spell'")
	}
}

func TestRedElementalBlastDestroysABluePermanent(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	merfolk := pushCostedPermanentForTest(g, opp.ID, "Merfolk", "Creature — Merfolk", "{U}")
	castModal(t, g, "Red Elemental Blast", "Instant", b03RedElementalBlastOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: merfolk}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(merfolk) {
		t.Error("the blue permanent survived")
	}
}

func TestPyroblastTargetsAnythingButOnlyActsOnBlue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := me.Life
	red := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	// Legal to target the red Bolt; the counter clause does nothing.
	castModal(t, g, "Pyroblast", "Instant", b03PyroblastOracle, []int{0},
		[]game.TargetRef{{Kind: game.TargetCard, ID: red}})
	passPriorityAroundTable(t, g)
	if me.Life != before-3 {
		t.Errorf("the red Bolt should have resolved: life %d -> %d", before, me.Life)
	}

	merfolk := pushCostedPermanentForTest(g, opp.ID, "Merfolk", "Creature — Merfolk", "{U}")
	castModal(t, g, "Pyroblast", "Instant", b03PyroblastOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: merfolk}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(merfolk) {
		t.Error("the blue permanent survived")
	}
}

func TestShamanicRevelationDrawsPerCreatureAndGainsPerBigOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedCreature(g, "Bear A", me.ID)
	seedCreature(g, "Bear B", me.ID)
	pushVanillaCreature(g, me.ID, "Giant", 4, 4)
	seedCreature(g, "Their Bear", g.Seats[1].ID)
	handBefore, lifeBefore := me.Hand.Size(), me.Life

	castCatalogSpell(t, g, "Shamanic Revelation", "Sorcery", b03ShamanicRevelationOracle, nil)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size() - handBefore; got != 3 {
		t.Errorf("drew %d, want 3 (one per creature you control)", got)
	}
	if me.Life != lifeBefore+4 {
		t.Errorf("life %d -> %d, want +4 (one power-4 creature)", lifeBefore, me.Life)
	}
}

func TestEntishRestorationSacrificesALandAndSearchesTwoOrThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	for i := 0; i < 4; i++ {
		stapleLibraryCard(me, "Forest", "Basic Land — Forest")
	}
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Entish Restoration", TypeLine: "Instant",
		OracleID: b03EntishRestorationOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{SacrificeIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land is sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil || c.SearchMax != 2 {
		t.Fatalf("without a power-4 creature the search is for up to TWO, got %+v", c)
	}
	answerSearchFailToFind(t, g, me.ID)

	// With a big creature it is three.
	pushVanillaCreature(g, me.ID, "Giant", 4, 4)
	land2 := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	id2 := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id2, Name: "Entish Restoration", TypeLine: "Instant",
		OracleID: b03EntishRestorationOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id2, game.CastSpellParams{SacrificeIDs: []uuid.UUID{land2}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c == nil || c.SearchMax != 3 {
		t.Fatalf("with a power-4 creature the search is for up to THREE, got %+v", c)
	}
}

func TestLivingDeathSwapsGraveyardsAndBattlefields(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushGraveyardCardForTest(me, "My Dead")
	pushGraveyardCardForTest(opp, "Their Dead")
	batch01GraveyardCard(me, "My Dead Rock", "Artifact")
	alive := seedCreature(g, "My Alive", me.ID)
	theirAlive := seedCreature(g, "Their Alive", opp.ID)
	rock := seedPermanentFor(g, opp.ID, "Rock", "Artifact")

	castCatalogSpell(t, g, "Living Death", "Sorcery", b03LivingDeathOracle, nil)
	passPriorityAroundTable(t, g)

	if countBattlefieldNamed(g, me.ID, "My Dead") != 1 || countBattlefieldNamed(g, opp.ID, "Their Dead") != 1 {
		t.Error("the graveyard creatures should be on the battlefield under their owners")
	}
	if g.Battlefield.Contains(alive) || g.Battlefield.Contains(theirAlive) {
		t.Error("the living creatures should have been sacrificed")
	}
	if !me.Graveyard.Contains(alive) || !opp.Graveyard.Contains(theirAlive) {
		t.Error("the sacrificed creatures stay in the graveyard — they were not exiled 'this way'")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("noncreature permanents are untouched")
	}
	if me.Graveyard.Size() != 3 { // My Dead Rock (not a creature) + My Alive + Living Death itself
		t.Errorf("my graveyard holds %d cards, want 3", me.Graveyard.Size())
	}
}

// --- permanents with triggers --------------------------------------

// The Tax sits under seat 1 so its upkeep is the NEXT one the cursor
// reaches (the game opens on seat 0's turn, past its upkeep) — the
// Phyrexian Arena test's arrangement.
func TestLandTaxSearchesWhenAnOpponentHasMoreLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Land Tax", "Enchantment", b03LandTaxOracle, false)
	seedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	seedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	var basics []uuid.UUID
	for i := 0; i < 4; i++ {
		basics = append(basics, stapleLibraryCard(me, "Plains", "Basic Land — Plains"))
	}

	advanceToUpkeepOf(t, g, 1)
	handBefore := me.Hand.Size()
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil || c.SearchMax != 3 {
		t.Fatalf("want a search for up to three basics, got %+v", c)
	}
	answerSearchByID(t, g, me.ID, basics[0], basics[1], basics[2])
	if got := me.Hand.Size() - handBefore; got != 3 {
		t.Errorf("hand grew by %d, want 3", got)
	}
}

func TestLandTaxIsSilentWhenNobodyHasMoreLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Land Tax", "Enchantment", b03LandTaxOracle, false)
	seedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")
	advanceToUpkeepOf(t, g, 1)
	if latestTriggerPrompt(g, me.ID) != nil || triggerOnStack(g, findBattlefieldByName(g, "Land Tax")) != nil {
		t.Error("Land Tax prompted with no opponent ahead on lands")
	}
}

func TestRampagingBalothsMakesAFourFourOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	baloths := pushCatalogPermanent(g, me.ID, "Rampaging Baloths", "Creature — Beast", b03RampagingBalothsOracle, false)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	beast := findBattlefieldByName(g, "Beast")
	if beast == uuid.Nil {
		t.Fatal("no Beast token after a land drop")
	}
	if p := effectivePower(t, g, beast); p != 4 {
		t.Errorf("Beast power = %d, want 4", p)
	}
	if !eotHasAbility(effectiveAbilities(t, g, baloths), "trample") {
		t.Error("printed trample did not reach the effective abilities")
	}
}

func TestGuttersnipeShocksEachOpponentOnAnInstant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Guttersnipe", "Creature — Goblin Shaman", b03GuttersnipeOracle, false)
	before := lifeOfOpponents(g)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d -> %d, want -2", i+1, b, got)
		}
	}
	// A creature spell is not an instant or sorcery.
	mid := lifeOfOpponents(g)
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	for i, b := range mid {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("opponent %d took damage from a creature spell", i+1)
		}
	}
}

func TestWarrenSoultraderPaysLifeAndAnotherCreatureForATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trader := pushCatalogPermanent(g, me.ID, "Warren Soultrader", "Creature — Zombie Goblin Wizard", b03WarrenSoultraderOracle, false)
	bear := seedCreature(g, "Bear", me.ID)
	before := me.Life

	if err := g.ActivateCatalogAbility(me.ID, trader, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{bear},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d -> %d, want -1 paid at announce", before, me.Life)
	}
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Treasure") != 1 {
		t.Error("want one Treasure")
	}
}

func TestWarrenSoultraderCannotEatItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trader := pushCatalogPermanent(g, me.ID, "Warren Soultrader", "Creature — Zombie Goblin Wizard", b03WarrenSoultraderOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, trader, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{trader},
	}); err == nil {
		t.Error("the Soultrader sacrificed itself; the printed cost is ANOTHER creature")
	}
}

func TestSramDrawsOnAuraEquipmentOrVehicleSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Sram, Senior Edificer", "Legendary Creature — Dwarf Advisor", b03SramOracle, false)
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Swiftfoot Boots", "Artifact — Equipment", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("an Equipment spell drew %d, want 1", got)
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("a creature spell changed the hand: now %d over the start", got)
	}
}

func TestSheoldredGainsOnYourDrawAndDrainsOnTheirs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sheoldred := pushCatalogPermanent(g, me.ID, "Sheoldred, the Apocalypse", "Legendary Creature — Phyrexian Praetor", b03SheoldredOracle, false)
	meBefore, oppBefore := me.Life, opp.Life

	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 2) })
	passPriorityAroundTable(t, g)
	if me.Life != meBefore+4 {
		t.Errorf("two of my draws: life %d -> %d, want +4", meBefore, me.Life)
	}

	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	if opp.Life != oppBefore-2 {
		t.Errorf("an opponent's draw: life %d -> %d, want -2", oppBefore, opp.Life)
	}
	if me.Life != meBefore+4 {
		t.Error("an opponent's draw must not gain me life")
	}
	if !eotHasAbility(effectiveAbilities(t, g, sheoldred), "deathtouch") {
		t.Error("printed deathtouch did not reach the effective abilities")
	}
}
