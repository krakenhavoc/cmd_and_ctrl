package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- registration --------------------------------------------------

func TestBatch03CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B03CrumblingNecropolisOracle: "Crumbling Necropolis",
		effectstest.B03LandTaxOracle:             "Land Tax",
		effectstest.B03ShamanicRevelationOracle:  "Shamanic Revelation",
		effectstest.B03DimirAqueductOracle:       "Dimir Aqueduct",
		effectstest.B03NomadOutpostOracle:        "Nomad Outpost",
		effectstest.B03GoForTheThroatOracle:      "Go for the Throat",
		effectstest.B03MysticMonasteryOracle:     "Mystic Monastery",
		effectstest.B03RampagingBalothsOracle:    "Rampaging Baloths",
		effectstest.B03OpulentPalaceOracle:       "Opulent Palace",
		effectstest.B03BadlandsOracle:            "Badlands",
		effectstest.B03GreatFurnaceOracle:        "Great Furnace",
		effectstest.B03ScrublandOracle:           "Scrubland",
		effectstest.B03DemolitionFieldOracle:     "Demolition Field",
		effectstest.B03FrontierBivouacOracle:     "Frontier Bivouac",
		effectstest.B03GuttersnipeOracle:         "Guttersnipe",
		effectstest.B03SeasideCitadelOracle:      "Seaside Citadel",
		effectstest.B03OrzhovBasilicaOracle:      "Orzhov Basilica",
		effectstest.B03WarrenSoultraderOracle:    "Warren Soultrader",
		effectstest.B03BayouOracle:               "Bayou",
		effectstest.B03IzzetBoilerworksOracle:    "Izzet Boilerworks",
		effectstest.B03HighMarketOracle:          "High Market",
		effectstest.B03TaigaOracle:               "Taiga",
		effectstest.B03RedElementalBlastOracle:   "Red Elemental Blast",
		effectstest.B03GeierReachOracle:          "Geier Reach Sanitarium",
		effectstest.B03SramOracle:                "Sram, Senior Edificer",
		effectstest.B03PlateauOracle:             "Plateau",
		effectstest.B03EntishRestorationOracle:   "Entish Restoration",
		effectstest.B03UrzasCaveOracle:           "Urza's Cave",
		effectstest.B03GruulTurfOracle:           "Gruul Turf",
		effectstest.B03DispatchOracle:            "Dispatch",
		effectstest.B03SavannahOracle:            "Savannah",
		effectstest.B03ExpeditionMapOracle:       "Expedition Map",
		effectstest.B03SheoldredOracle:           "Sheoldred, the Apocalypse",
		effectstest.B03LivingDeathOracle:         "Living Death",
		effectstest.B03MentalMisstepOracle:       "Mental Misstep",
		effectstest.B03PyroblastOracle:           "Pyroblast",
		effectstest.B03FabricateOracle:           "Fabricate",
	}
	if len(want) != 37 {
		t.Fatalf("the batch is 37 cards, the table lists %d", len(want))
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
	for _, spec := range effects.All() {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PlayLandFromHand(t, g, "Crumbling Necropolis", effectstest.B03CrumblingNecropolisOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Crumbling Necropolis")

	// Untap it by hand and tap for mana: a three-way pick.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 3 {
		t.Fatalf("a tri-land must ask which of three colours, got %+v", pick)
	}
}

func TestOriginalDualEntersUntappedAndTapsForEitherColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Badlands", TypeLine: "Land — Swamp Mountain",
		OracleID: effectstest.B03BadlandsOracle, Owner: me.ID, Controller: me.ID})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play Badlands: %v", err)
	}
	effectstest.Top100AssertEnteredUntapped(t, g, id, "Badlands")
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("a dual must ask B or R, got %+v", pick)
	}
}

// #1337: the bounce is a resolution-time choice (ReturnOneYouControl
// / ChoosePermanents), not a target picked when the trigger goes on
// the stack.
func TestBounceLandEntersTappedAndReturnsAChosenLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	island := effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	handBefore := me.Hand.Size()

	id := effectstest.PlayLandFromHand(t, g, "Dimir Aqueduct", effectstest.B03DimirAqueductOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Dimir Aqueduct")
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	furnace := effectstest.SeedPermanentWithOracle(g, me.ID, "Great Furnace", "Artifact Land", effectstest.B03GreatFurnaceOracle)
	if err := g.ActivateManaAbility(me.ID, furnace, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want [R]", got)
	}
	if effects.B03ArtifactsControlled(g, me.ID) != 1 {
		t.Error("an artifact land is an artifact")
	}
}

// --- utility lands -------------------------------------------------

func TestHighMarketSacrificesACreatureForALife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	market := effectstest.PushCatalogPermanent(g, me.ID, "High Market", "Land", effectstest.B03HighMarketOracle, false)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	before := me.Life

	if err := g.ActivateCatalogAbility(me.ID, market, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{bear},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the sacrifice is a cost, paid at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d -> %d, want +1", before, me.Life)
	}
	card, _ := effectstest.BattlefieldCard(g, market)
	if !card.Tapped {
		t.Error("High Market has a tap cost")
	}
}

func TestGeierReachSanitariumEachPlayerLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sanitarium := effectstest.PushCatalogPermanent(g, me.ID, "Geier Reach Sanitarium", "Legendary Land", effectstest.B03GeierReachOracle, false)
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Hand.Size()
	}
	if err := g.ActivateCatalogAbility(me.ID, sanitarium, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Hand.Size() != before[i]+1 {
			t.Errorf("seat %d drew %d, want 1", i, p.Hand.Size()-before[i])
		}
		// One discard prompt per seat, addressed to that seat over
		// their own hand, so each player chooses their own card (the
		// Faithless Looting assertion).
		if got := effectstest.DiscardOwed(g, p.ID); got != 1 {
			t.Errorf("seat %d is owed %d discards, want 1 — each player chooses their own", i, got)
		}
	}
}

func TestUrzasCaveFetchesAnyLandTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cave := effectstest.PushCatalogPermanent(g, me.ID, "Urza's Cave", "Land — Urza's Cave", effectstest.B03UrzasCaveOracle, false)
	coffers := effectstest.StapleLibraryCard(me, "Cabal Coffers", "Legendary Land")
	effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")

	if err := g.ActivateCatalogAbility(me.ID, cave, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, coffers)
	got, ok := effectstest.BattlefieldCard(g, coffers)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	field := effectstest.PushCatalogPermanent(g, me.ID, "Demolition Field", "Land", effectstest.B03DemolitionFieldOracle, false)
	coffers := effectstest.SeedLandOnBattlefield(g, opp.ID, "Cabal Coffers", "Legendary Land")
	mine := effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
	effectstest.StapleLibraryCard(opp, "Swamp", "Basic Land — Swamp")

	if err := g.ActivateCatalogAbility(me.ID, field, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: coffers}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(coffers) {
		t.Fatal("the nonbasic land survived")
	}
	if effectstest.SearchChoiceFor(g, opp.ID) == nil {
		t.Fatal("the victim gets a (declinable) search for a basic")
	}
	// "Its controller may search ...; you may search ..." (CR 608.2c, #4155):
	// yours is asked once theirs is answered.
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Fatal("the controller's search opened before the victim's was answered")
	}
	effectstest.AnswerSearchByID(t, g, opp.ID)
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Fatal("the controller gets a (declinable) search for a basic")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, mine)
	got, ok := effectstest.BattlefieldCard(g, mine)
	if !ok || got.Tapped {
		t.Error("the controller's basic enters UNTAPPED")
	}
}

func TestDemolitionFieldRefusesABasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	field := effectstest.PushCatalogPermanent(g, me.ID, "Demolition Field", "Land", effectstest.B03DemolitionFieldOracle, false)
	forest := effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	if err := g.ActivateCatalogAbility(me.ID, field, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: forest}},
	}); err == nil {
		t.Error("a basic land was accepted for 'target nonbasic land'")
	}
}

func TestExpeditionMapTutorsALandToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mapID := effectstest.PushCatalogPermanent(g, me.ID, "Expedition Map", "Artifact", effectstest.B03ExpeditionMapOracle, false)
	urborg := effectstest.StapleLibraryCard(me, "Urborg, Tomb of Yawgmoth", "Legendary Land")
	effectstest.StapleLibraryCard(me, "Island", "Basic Land — Island")

	if err := g.ActivateCatalogAbility(me.ID, mapID, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, urborg)
	if !me.Hand.Contains(urborg) {
		t.Error("the land did not reach the hand")
	}
	if g.Battlefield.Contains(mapID) {
		t.Error("the Map is sacrificed as a cost")
	}
}

// --- spells --------------------------------------------------------

func TestFabricateTutorsAnArtifactToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ring := effectstest.StapleLibraryCard(me, "Sol Ring", "Artifact")
	signet := effectstest.StapleLibraryCard(me, "Arcane Signet", "Artifact")
	effectstest.StapleLibraryCard(me, "Bear", "Creature — Bear")

	effectstest.CastCatalogSpell(t, g, "Fabricate", "Sorcery", effectstest.B03FabricateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Two artifacts to choose from, so the chooser prompt opens.
	effectstest.AnswerSearchByID(t, g, me.ID, ring)
	if !me.Hand.Contains(ring) {
		t.Error("the artifact did not reach the hand")
	}
	if me.Hand.Contains(signet) {
		t.Error("only the chosen artifact should come")
	}
}

func TestGoForTheThroatKillsANonartifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.SeedCreature(g, "Bear", opp.ID)
	effectstest.CastCatalogSpell(t, g, "Go for the Throat", "Instant", effectstest.B03GoForTheThroatOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature survived")
	}
}

func TestGoForTheThroatRefusesAnArtifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	golem := effectstest.SeedPermanentFor(g, opp.ID, "Golem", "Artifact Creature — Golem")
	if !effectstest.B03CastRefused(t, g, "Go for the Throat", "Instant", effectstest.B03GoForTheThroatOracle, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: golem}},
	}) {
		t.Error("an artifact creature was accepted for 'target nonartifact creature'")
	}
}

func TestDispatchTapsAndExilesWithMetalcraft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		effectstest.SeedPermanentFor(g, me.ID, "Rock", "Artifact")
	}
	bear := effectstest.SeedCreature(g, "Bear", opp.ID)
	effectstest.CastCatalogSpell(t, g, "Dispatch", "Instant", effectstest.B03DispatchOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) {
		t.Error("with three artifacts Dispatch exiles")
	}
}

func TestDispatchOnlyTapsWithoutMetalcraft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedPermanentFor(g, me.ID, "Rock", "Artifact")
	effectstest.SeedPermanentFor(g, me.ID, "Rock", "Artifact")
	bear := effectstest.SeedCreature(g, "Bear", opp.ID)
	effectstest.CastCatalogSpell(t, g, "Dispatch", "Instant", effectstest.B03DispatchOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := effectstest.BattlefieldCard(g, bear)
	if !ok {
		t.Fatal("with two artifacts Dispatch must not exile")
	}
	if !card.Tapped {
		t.Error("the creature should be tapped")
	}
}

func TestMentalMisstepCountersOnlyManaValueOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := effectstest.Batch01OpponentCasts(t, g, opp, "One Drop", effectstest.LightningBoltOracle, "{U}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.CastCatalogSpell(t, g, "Mental Misstep", "Instant", effectstest.B03MentalMisstepOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(spell) {
		t.Fatal("the mana value 1 spell was not countered")
	}

	two := effectstest.Batch01OpponentCasts(t, g, opp, "Two Drop", effectstest.LightningBoltOracle, "{1}{U}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if !effectstest.B03CastRefused(t, g, "Mental Misstep", "Instant", effectstest.B03MentalMisstepOracle, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: two}},
	}) {
		t.Error("a mana value 2 spell was accepted as a target")
	}
}

func TestRedElementalBlastCountersABlueSpellAndRefusesARedOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blue := effectstest.Batch01OpponentCasts(t, g, opp, "Blue Spell", effectstest.LightningBoltOracle, "{U}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.CastModal(t, g, "Red Elemental Blast", "Instant", effectstest.B03RedElementalBlastOracle, []int{0},
		[]game.TargetRef{{Kind: game.TargetCard, ID: blue}})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(blue) {
		t.Fatal("the blue spell was not countered")
	}

	red := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if !effectstest.B03CastRefused(t, g, "Red Elemental Blast", "Instant", effectstest.B03RedElementalBlastOracle, game.CastSpellParams{
		Modes:   []int{0},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: red}},
	}) {
		t.Error("a red spell was accepted for 'target blue spell'")
	}
}

func TestRedElementalBlastDestroysABluePermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	merfolk := effectstest.PushCostedPermanentForTest(g, opp.ID, "Merfolk", "Creature — Merfolk", "{U}")
	effectstest.CastModal(t, g, "Red Elemental Blast", "Instant", effectstest.B03RedElementalBlastOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: merfolk}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(merfolk) {
		t.Error("the blue permanent survived")
	}
}

func TestPyroblastTargetsAnythingButOnlyActsOnBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := me.Life
	red := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	// Legal to target the red Bolt; the counter clause does nothing.
	effectstest.CastModal(t, g, "Pyroblast", "Instant", effectstest.B03PyroblastOracle, []int{0},
		[]game.TargetRef{{Kind: game.TargetCard, ID: red}})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before-3 {
		t.Errorf("the red Bolt should have resolved: life %d -> %d", before, me.Life)
	}

	merfolk := effectstest.PushCostedPermanentForTest(g, opp.ID, "Merfolk", "Creature — Merfolk", "{U}")
	effectstest.CastModal(t, g, "Pyroblast", "Instant", effectstest.B03PyroblastOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: merfolk}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(merfolk) {
		t.Error("the blue permanent survived")
	}
}

func TestShamanicRevelationDrawsPerCreatureAndGainsPerBigOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedCreature(g, "Bear A", me.ID)
	effectstest.SeedCreature(g, "Bear B", me.ID)
	effectstest.PushVanillaCreature(g, me.ID, "Giant", 4, 4)
	effectstest.SeedCreature(g, "Their Bear", g.Seats[1].ID)
	handBefore, lifeBefore := me.Hand.Size(), me.Life

	effectstest.CastCatalogSpell(t, g, "Shamanic Revelation", "Sorcery", effectstest.B03ShamanicRevelationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size() - handBefore; got != 3 {
		t.Errorf("drew %d, want 3 (one per creature you control)", got)
	}
	if me.Life != lifeBefore+4 {
		t.Errorf("life %d -> %d, want +4 (one power-4 creature)", lifeBefore, me.Life)
	}
}

func castEntishRestoration(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Entish Restoration", TypeLine: "Instant",
		OracleID: effectstest.B03EntishRestorationOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	return id
}

// #2876: the sacrifice is part of the effect, chosen as the spell
// resolves, and the search follows it.
func TestEntishRestorationSacrificesALandOnResolutionAndSearchesTwoOrThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	for i := 0; i < 4; i++ {
		effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
	}
	castEntishRestoration(t, g)
	if !g.Battlefield.Contains(land) {
		t.Fatal("nothing is sacrificed at cast")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, me.ID) == nil {
		t.Fatal("the caster is asked to sacrifice a land as the spell resolves")
	}
	effectstest.AnswerSacrifice(t, g, me.ID, land)
	if !me.Graveyard.Contains(land) {
		t.Error("the chosen land is sacrificed")
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil || c.SearchMax != 2 {
		t.Fatalf("without a power-4 creature the search is for up to TWO, got %+v", c)
	}
	effectstest.AnswerSearchFailToFind(t, g, me.ID)

	// With a big creature it is three.
	effectstest.PushVanillaCreature(g, me.ID, "Giant", 4, 4)
	land2 := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	castEntishRestoration(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, land2)
	if c := effectstest.SearchChoiceFor(g, me.ID); c == nil || c.SearchMax != 3 {
		t.Fatalf("with a power-4 creature the search is for up to THREE, got %+v", c)
	}
}

// With no land the spell is still castable and still searches, as
// printed: the sacrifice is not an "if you do".
func TestEntishRestorationWithNoLandStillSearches(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 4; i++ {
		effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
	}
	castEntishRestoration(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Fatal("a caster with no land is not asked to sacrifice")
	}
	if c := effectstest.SearchChoiceFor(g, me.ID); c == nil || c.SearchMax != 2 {
		t.Fatalf("the search still happens, got %+v", c)
	}
}

func TestLivingDeathSwapsGraveyardsAndBattlefields(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushGraveyardCardForTest(me, "My Dead")
	effectstest.PushGraveyardCardForTest(opp, "Their Dead")
	effectstest.Batch01GraveyardCard(me, "My Dead Rock", "Artifact")
	alive := effectstest.SeedCreature(g, "My Alive", me.ID)
	theirAlive := effectstest.SeedCreature(g, "Their Alive", opp.ID)
	rock := effectstest.SeedPermanentFor(g, opp.ID, "Rock", "Artifact")

	effectstest.CastCatalogSpell(t, g, "Living Death", "Sorcery", effectstest.B03LivingDeathOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.CountBattlefieldNamed(g, me.ID, "My Dead") != 1 || effectstest.CountBattlefieldNamed(g, opp.ID, "Their Dead") != 1 {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Land Tax", "Enchantment", effectstest.B03LandTaxOracle, false)
	effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	var basics []uuid.UUID
	for i := 0; i < 4; i++ {
		basics = append(basics, effectstest.StapleLibraryCard(me, "Plains", "Basic Land — Plains"))
	}

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	handBefore := me.Hand.Size()
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil || c.SearchMax != 3 {
		t.Fatalf("want a search for up to three basics, got %+v", c)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, basics[0], basics[1], basics[2])
	if got := me.Hand.Size() - handBefore; got != 3 {
		t.Errorf("hand grew by %d, want 3", got)
	}
}

func TestLandTaxIsSilentWhenNobodyHasMoreLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Land Tax", "Enchantment", effectstest.B03LandTaxOracle, false)
	effectstest.SeedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.OwedMays(g, me.ID) > 0 || effectstest.TriggerOnStack(g, effectstest.FindBattlefieldByName(g, "Land Tax")) != nil {
		t.Error("Land Tax prompted with no opponent ahead on lands")
	}
}

func TestRampagingBalothsMakesAFourFourOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	baloths := effectstest.PushCatalogPermanent(g, me.ID, "Rampaging Baloths", "Creature — Beast", effectstest.B03RampagingBalothsOracle, false)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	beast := effectstest.FindBattlefieldByName(g, "Beast")
	if beast == uuid.Nil {
		t.Fatal("no Beast token after a land drop")
	}
	if p := effectstest.EffectivePower(t, g, beast); p != 4 {
		t.Errorf("Beast power = %d, want 4", p)
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, baloths), "trample") {
		t.Error("printed trample did not reach the effective abilities")
	}
}

func TestGuttersnipeShocksEachOpponentOnAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Guttersnipe", "Creature — Goblin Shaman", effectstest.B03GuttersnipeOracle, false)
	before := effectstest.LifeOfOpponents(g)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d -> %d, want -2", i+1, b, got)
		}
	}
	// A creature spell is not an instant or sorcery.
	mid := effectstest.LifeOfOpponents(g)
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range mid {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("opponent %d took damage from a creature spell", i+1)
		}
	}
}

func TestWarrenSoultraderPaysLifeAndAnotherCreatureForATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	trader := effectstest.PushCatalogPermanent(g, me.ID, "Warren Soultrader", "Creature — Zombie Goblin Wizard", effectstest.B03WarrenSoultraderOracle, false)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	before := me.Life

	if err := g.ActivateCatalogAbility(me.ID, trader, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{bear},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d -> %d, want -1 paid at announce", before, me.Life)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Treasure") != 1 {
		t.Error("want one Treasure")
	}
}

func TestWarrenSoultraderCannotEatItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	trader := effectstest.PushCatalogPermanent(g, me.ID, "Warren Soultrader", "Creature — Zombie Goblin Wizard", effectstest.B03WarrenSoultraderOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, trader, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{trader},
	}); err == nil {
		t.Error("the Soultrader sacrificed itself; the printed cost is ANOTHER creature")
	}
}

func TestSramDrawsOnAuraEquipmentOrVehicleSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Sram, Senior Edificer", "Legendary Creature — Dwarf Advisor", effectstest.B03SramOracle, false)
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Swiftfoot Boots", "Artifact — Equipment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("an Equipment spell drew %d, want 1", got)
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("a creature spell changed the hand: now %d over the start", got)
	}
}

func TestSheoldredGainsOnYourDrawAndDrainsOnTheirs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sheoldred := effectstest.PushCatalogPermanent(g, me.ID, "Sheoldred, the Apocalypse", "Legendary Creature — Phyrexian Praetor", effectstest.B03SheoldredOracle, false)
	meBefore, oppBefore := me.Life, opp.Life

	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != meBefore+4 {
		t.Errorf("two of my draws: life %d -> %d, want +4", meBefore, me.Life)
	}

	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppBefore-2 {
		t.Errorf("an opponent's draw: life %d -> %d, want -2", oppBefore, opp.Life)
	}
	if me.Life != meBefore+4 {
		t.Error("an opponent's draw must not gain me life")
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, sheoldred), "deathtouch") {
		t.Error("printed deathtouch did not reach the effective abilities")
	}
}

// #2876: a countered Entish Restoration costs no land, which the old
// cast-time cost could not say.
func TestEntishRestorationCounteredKeepsTheLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	spell := effectstest.Batch01OpponentCasts(t, g, opp, "Entish Restoration", effectstest.B03EntishRestorationOracle, "", nil)

	effectstest.CastCatalogSpell(t, g, "Arcane Denial", "Instant", effectstest.ArcaneDenialOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(spell) {
		t.Fatal("Entish Restoration was not countered")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land was sacrificed even though the spell was countered")
	}
	if effectstest.SacrificeChoiceFor(g, opp.ID) != nil || effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Error("a countered spell asks for no sacrifice")
	}
}
