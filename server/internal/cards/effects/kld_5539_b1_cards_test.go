package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kld_5539_b1_cards_test.go — Kaladesh cards, batch 1 (#5539). Every
// test plays the real catalog entry through the engine.

const (
	kldB1CowlProwlerOracle              = "67419470-3418-466f-b3c1-4ded1d0ee34b"
	kldB1CurioVendorOracle              = "7db99e21-0faf-408d-a770-189726179c66"
	kldB1PrakhataClubSecurityOracle     = "9c5f130d-5972-449e-83db-41fb8cf76004"
	kldB1TasseledDromedaryOracle        = "3121eecd-5f8a-408f-8b2d-75e13aa8a1ea"
	kldB1TerrainElementalOracle         = "3a9b9fe0-1846-41fc-b7c3-ed9abef27a6a"
	kldB1TerrorOfTheFairgroundsOracle   = "75a20db5-5ce8-4e5d-a815-e69bc3dd8d4c"
	kldB1AerialResponderOracle          = "69b1b0ec-9db0-48d1-a7b5-71281aca16fe"
	kldB1ConsulateSkygateOracle         = "671db0d7-b103-4c13-9914-a444f2ecf9d6"
	kldB1SkyswirlHarrierOracle          = "0feeeb19-7c90-4009-9a71-226a443b5c3f"
	kldB1SnareThopterOracle             = "a3be0e16-6063-4cd0-87cb-606591a958c2"
	kldB1WaywardGiantOracle             = "124cf15a-587a-4f1f-a943-f462072f2098"
	kldB1VedalkenBlademasterOracle      = "7a8405f5-ef37-460e-8e0b-55378f3b407b"
	kldB1AccomplishedAutomatonOracle    = "f5af0d1b-531e-4053-848a-5e0fa1022f9a"
	kldB1AmbitiousAetherbornOracle      = "31f54f88-85a6-43f4-a6de-fb6ef5f0c38d"
	kldB1GlintSleeveArtisanOracle       = "b5f35577-3d94-4835-bfca-3090cb0c63ff"
	kldB1VisionaryAugmenterOracle       = "1782bfd7-7d8f-4ed2-9d45-ba9adb41e996"
	kldB1WeaponcraftEnthusiastOracle    = "7c27136a-bf9d-4f42-9c37-abe347fad1bb"
	kldB1HighspireArtisanOracle         = "fc8c54c4-a436-43cc-888b-f4ba3e751c16"
	kldB1IronLeagueSteedOracle          = "bafeab74-918b-4fcd-8057-55922dbc1784"
	kldB1MaulfistSquadOracle            = "2a8ec988-b513-4e66-b7f5-b2196b4bbff2"
	kldB1PeemaOutriderOracle            = "7747e94c-35a7-4e3b-956b-339a16634502"
	kldB1PropellerPioneerOracle         = "d17a9750-a89f-4ad1-9c15-7ccffc554ee5"
	kldB1ElegantEdgecraftersOracle      = "eeb3d282-97df-4df2-8bea-9432e21b6d62"
	kldB1BastionMastodonOracle          = "48a01671-da1f-4875-bcd5-ca99566ff399"
	kldB1DukharaPeafowlOracle           = "abc5ac3e-f196-49df-8a9d-16df3114ffad"
	kldB1NarnamCobraOracle              = "2ac0d584-997b-4d11-aecf-f1494eebc94c"
	kldB1PrakhataPillarBugOracle        = "1ec95969-4a6f-49f3-af23-1680276e7da2"
	kldB1WeldfastMonitorOracle          = "5d604aff-ba60-4064-b025-b1f3d1c0254e"
	kldB1WilyBandarOracle               = "2b0ba5a5-649a-4d7a-8136-6bcf692337f5"
	kldB1InventorsApprenticeOracle      = "ed8d54fa-ec02-45d6-b70e-4c7e1c66efd6"
	kldB1FoundryScreecherOracle         = "0e0bef57-d977-48d1-be78-9a4f0aacd9c0"
	kldB1LongFinnedSkywhaleOracle       = "a8287efd-23ca-498b-90f1-c76099919fa4"
	kldB1ArborbackStomperOracle         = "a3f35e0f-bf7a-42bf-b612-ac3af3a33727"
	kldB1NimbleInnovatorOracle          = "30929713-47dc-40f0-883c-4bd41aa399d5"
	kldB1AppetiteForTheUnnaturalOracle  = "f041b37e-25af-4aff-b7e3-f07a4d5c6f9a"
	kldB1EssenceExtractionOracle        = "14363df6-ac0f-476a-bbd8-ce351de0babf"
	kldB1FragmentizeOracle              = "f025c1da-734e-4cbb-99d8-9d700d6edbde"
	kldB1LargerThanLifeOracle           = "96e8210a-a547-431d-8146-7399ec20c591"
	kldB1ServoExhibitionOracle          = "76a6ddc2-f894-4331-89a9-87ca16efd0c0"
	kldB1SkywhalersShotOracle           = "f3d4e48e-5f98-44a6-ad5c-a1c068aa01eb"
	kldB1SelectForInspectionOracle      = "cfee4ba9-3cf2-4e8b-a4a7-21974828cd54"
	kldB1RefurbishOracle                = "e4e61fe3-27a4-4149-bdfb-67dcfc8eb32c"
	kldB1CeremoniousRejectionOracle     = "6aa5aa3d-fb46-47e5-bcf1-51a28e0f9974"
	kldB1MakeObsoleteOracle             = "3ce755f5-855f-4b88-aa10-22f4ee76b28a"
	kldB1AcrobaticManeuverOracle        = "1aa1cf50-ee79-4bef-826b-5f1a9e1aa1a5"
	kldB1TorchGauntletOracle            = "07fdc0b0-7131-411d-b0a8-f667165a64b8"
	kldB1SkySkiffOracle                 = "f2c43b85-8c42-4b23-84c2-bec8d6230039"
	kldB1OvalchaseDragsterOracle        = "345de988-4d26-4400-8202-225968db370f"
	kldB1AradaraExpressOracle           = "be96b9e1-10ef-4955-8d23-c39384def86d"
	kldB1DemolitionStomperOracle        = "282e5e5f-365a-4e11-aeac-ead71d1cf68a"
	kldB1BearType                       = "Creature — Bear"
	kldB1VehicleType                    = "Artifact — Vehicle"
	kldB1ArtifactCreatureConstructType  = "Artifact Creature — Construct"
	kldB1CreatureAetherbornArtificer    = "Creature — Aetherborn Artificer"
	kldB1CreatureDwarfArtificerTypeLine = "Creature — Dwarf Artificer"
)

// kldB1Names is every card in the batch, by oracle ID.
var kldB1Names = map[string]string{
	kldB1CowlProwlerOracle:             "Cowl Prowler",
	kldB1CurioVendorOracle:             "Curio Vendor",
	kldB1PrakhataClubSecurityOracle:    "Prakhata Club Security",
	kldB1TasseledDromedaryOracle:       "Tasseled Dromedary",
	kldB1TerrainElementalOracle:        "Terrain Elemental",
	kldB1TerrorOfTheFairgroundsOracle:  "Terror of the Fairgrounds",
	kldB1AerialResponderOracle:         "Aerial Responder",
	kldB1ConsulateSkygateOracle:        "Consulate Skygate",
	kldB1SkyswirlHarrierOracle:         "Skyswirl Harrier",
	kldB1SnareThopterOracle:            "Snare Thopter",
	kldB1WaywardGiantOracle:            "Wayward Giant",
	kldB1VedalkenBlademasterOracle:     "Vedalken Blademaster",
	kldB1AccomplishedAutomatonOracle:   "Accomplished Automaton",
	kldB1AmbitiousAetherbornOracle:     "Ambitious Aetherborn",
	kldB1GlintSleeveArtisanOracle:      "Glint-Sleeve Artisan",
	kldB1VisionaryAugmenterOracle:      "Visionary Augmenter",
	kldB1WeaponcraftEnthusiastOracle:   "Weaponcraft Enthusiast",
	kldB1HighspireArtisanOracle:        "Highspire Artisan",
	kldB1IronLeagueSteedOracle:         "Iron League Steed",
	kldB1MaulfistSquadOracle:           "Maulfist Squad",
	kldB1PeemaOutriderOracle:           "Peema Outrider",
	kldB1PropellerPioneerOracle:        "Propeller Pioneer",
	kldB1ElegantEdgecraftersOracle:     "Elegant Edgecrafters",
	kldB1BastionMastodonOracle:         "Bastion Mastodon",
	kldB1DukharaPeafowlOracle:          "Dukhara Peafowl",
	kldB1NarnamCobraOracle:             "Narnam Cobra",
	kldB1PrakhataPillarBugOracle:       "Prakhata Pillar-Bug",
	kldB1WeldfastMonitorOracle:         "Weldfast Monitor",
	kldB1WilyBandarOracle:              "Wily Bandar",
	kldB1InventorsApprenticeOracle:     "Inventor's Apprentice",
	kldB1FoundryScreecherOracle:        "Foundry Screecher",
	kldB1LongFinnedSkywhaleOracle:      "Long-Finned Skywhale",
	kldB1ArborbackStomperOracle:        "Arborback Stomper",
	kldB1NimbleInnovatorOracle:         "Nimble Innovator",
	kldB1AppetiteForTheUnnaturalOracle: "Appetite for the Unnatural",
	kldB1EssenceExtractionOracle:       "Essence Extraction",
	kldB1FragmentizeOracle:             "Fragmentize",
	kldB1LargerThanLifeOracle:          "Larger Than Life",
	kldB1ServoExhibitionOracle:         "Servo Exhibition",
	kldB1SkywhalersShotOracle:          "Skywhaler's Shot",
	kldB1SelectForInspectionOracle:     "Select for Inspection",
	kldB1RefurbishOracle:               "Refurbish",
	kldB1CeremoniousRejectionOracle:    "Ceremonious Rejection",
	kldB1MakeObsoleteOracle:            "Make Obsolete",
	kldB1AcrobaticManeuverOracle:       "Acrobatic Maneuver",
	kldB1TorchGauntletOracle:           "Torch Gauntlet",
	kldB1SkySkiffOracle:                "Sky Skiff",
	kldB1OvalchaseDragsterOracle:       "Ovalchase Dragster",
	kldB1AradaraExpressOracle:          "Aradara Express",
	kldB1DemolitionStomperOracle:       "Demolition Stomper",
}

// kldB1Creature seeds a non-catalog creature with an exact body, tapped
// or not.
func kldB1Creature(g *game.Game, owner uuid.UUID, name string, power, toughness int, tapped bool) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: kldB1BearType,
		Power: power, Toughness: toughness, Tapped: tapped, Owner: owner, Controller: owner,
	})
}

// kldB1OpponentCastsInstant has opp cast a non-catalog instant with the
// given cost and colours, and returns its ID on the stack.
func kldB1OpponentCastsInstant(t *testing.T, g *game.Game, opp *game.Player, name, manaCost string, colors ...string) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	id := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Instant", ManaCost: manaCost, Colors: colors,
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent CastSpell %s: %v", name, err)
	}
	return id
}

// kldB1Crew crews a vehicle with one creature of the given power.
func kldB1Crew(t *testing.T, g *game.Game, me *game.Player, vehicle uuid.UUID, power int) {
	t.Helper()
	crewer := b12Creature(g, me.ID, "Crewer", kldB1BearType, power, power)
	if err := g.ActivateCatalogAbility(me.ID, vehicle, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crewer}}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// --- every card is in the catalog ----------------------------------------

func TestKld5539B1EveryCardIsRegisteredFull(t *testing.T) {
	if len(kldB1Names) != 50 {
		t.Fatalf("the batch lists %d cards, want 50", len(kldB1Names))
	}
	for oracle, name := range kldB1Names {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v; want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- vanilla bodies ------------------------------------------------------------

func TestKld5539B1VanillaBodiesHaveNoAbilities(t *testing.T) {
	for _, oracle := range []string{
		kldB1CowlProwlerOracle, kldB1CurioVendorOracle, kldB1PrakhataClubSecurityOracle,
		kldB1TasseledDromedaryOracle, kldB1TerrainElementalOracle, kldB1TerrorOfTheFairgroundsOracle,
	} {
		spec, _ := Lookup(oracle)
		if len(spec.Triggered) != 0 || len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 ||
			len(spec.BlockRules) != 0 || len(spec.Static) != 0 || spec.OnResolve != nil {
			t.Errorf("%s registers an ability; it is a vanilla creature", kldB1Names[oracle])
		}
	}
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wurm := b12Push(g, me.ID, "Cowl Prowler", "Creature — Wurm", kldB1CowlProwlerOracle, 6, 6)
	camel := b12Push(g, opp.ID, "Tasseled Dromedary", "Creature — Camel", kldB1TasseledDromedaryOracle, 0, 4)
	if err := porBlock(t, g, wurm, opp.ID, camel); err != nil {
		t.Errorf("Tasseled Dromedary couldn't block Cowl Prowler: %v", err)
	}
}

// --- keyword bodies ------------------------------------------------------------

func TestKld5539B1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Aerial Responder", "Creature — Dwarf Soldier", kldB1AerialResponderOracle, []string{"flying", "vigilance", "lifelink"}},
		{"Consulate Skygate", "Artifact Creature — Wall", kldB1ConsulateSkygateOracle, []string{"defender", "reach"}},
		{"Skyswirl Harrier", "Creature — Bird", kldB1SkyswirlHarrierOracle, []string{"flying"}},
		{"Snare Thopter", "Artifact Creature — Thopter", kldB1SnareThopterOracle, []string{"flying", "haste"}},
		{"Wayward Giant", "Creature — Giant", kldB1WaywardGiantOracle, []string{"menace"}},
		{"Vedalken Blademaster", "Creature — Vedalken Soldier", kldB1VedalkenBlademasterOracle, []string{"prowess"}},
		{"Highspire Artisan", "Creature — Elf Artificer", kldB1HighspireArtisanOracle, []string{"reach"}},
		{"Iron League Steed", kldB1ArtifactCreatureConstructType, kldB1IronLeagueSteedOracle, []string{"haste"}},
		{"Maulfist Squad", "Creature — Human Artificer", kldB1MaulfistSquadOracle, []string{"menace"}},
		{"Peema Outrider", "Creature — Elf Artificer", kldB1PeemaOutriderOracle, []string{"trample"}},
		{"Propeller Pioneer", "Creature — Human Artificer", kldB1PropellerPioneerOracle, []string{"flying"}},
		{"Foundry Screecher", "Creature — Bat", kldB1FoundryScreecherOracle, []string{"flying"}},
		{"Long-Finned Skywhale", "Creature — Whale", kldB1LongFinnedSkywhaleOracle, []string{"flying"}},
		{"Arborback Stomper", "Creature — Beast", kldB1ArborbackStomperOracle, []string{"trample"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			for _, kw := range tc.keywords {
				if !hasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// Vedalken Blademaster's prowess: a noncreature spell pumps it.
func TestKld5539B1VedalkenBlademasterProwess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	blade := b12Push(g, me.ID, "Vedalken Blademaster", "Creature — Vedalken Soldier", kldB1VedalkenBlademasterOracle, 2, 3)
	castCatalogSpell(t, g, "Servo Exhibition", "Sorcery", kldB1ServoExhibitionOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, blade); p != 3 || tough != 4 {
		t.Errorf("the Blademaster is %d/%d after a noncreature spell, want 3/4", p, tough)
	}
}

// --- fabricate -------------------------------------------------------------------

func TestKld5539B1FabricateEitherWay(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		power, tough, n        int
	}{
		{"Accomplished Automaton", kldB1ArtifactCreatureConstructType, kldB1AccomplishedAutomatonOracle, 5, 7, 1},
		{"Ambitious Aetherborn", kldB1CreatureAetherbornArtificer, kldB1AmbitiousAetherbornOracle, 4, 3, 1},
		{"Glint-Sleeve Artisan", kldB1CreatureDwarfArtificerTypeLine, kldB1GlintSleeveArtisanOracle, 2, 2, 1},
		{"Visionary Augmenter", kldB1CreatureDwarfArtificerTypeLine, kldB1VisionaryAugmenterOracle, 2, 1, 2},
		{"Weaponcraft Enthusiast", kldB1CreatureAetherbornArtificer, kldB1WeaponcraftEnthusiastOracle, 0, 1, 2},
		{"Highspire Artisan", "Creature — Elf Artificer", kldB1HighspireArtisanOracle, 0, 3, 1},
		{"Iron League Steed", kldB1ArtifactCreatureConstructType, kldB1IronLeagueSteedOracle, 2, 2, 1},
		{"Maulfist Squad", "Creature — Human Artificer", kldB1MaulfistSquadOracle, 3, 1, 1},
		{"Peema Outrider", "Creature — Elf Artificer", kldB1PeemaOutriderOracle, 3, 3, 1},
		{"Propeller Pioneer", "Creature — Human Artificer", kldB1PropellerPioneerOracle, 2, 1, 1},
		{"Elegant Edgecrafters", "Creature — Elf Artificer", kldB1ElegantEdgecraftersOracle, 3, 4, 2},
	}
	for _, tc := range cases {
		for _, counters := range []bool{true, false} {
			t.Run(tc.name, func(t *testing.T) {
				g := newCatalogGame(t)
				me := g.Seats[g.Turn.ActiveSeat]
				id := castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
				answerMayChoice(t, g, me.ID, counters)
				passPriorityAroundTable(t, g)
				wantCounters, wantServos := tc.n, 0
				if !counters {
					wantCounters, wantServos = 0, tc.n
				}
				if got := plusOnes(g, id); got != wantCounters {
					t.Errorf("%d +1/+1 counters, want %d", got, wantCounters)
				}
				if got := cardsNamed(g, "Servo", me.ID); got != wantServos {
					t.Errorf("%d Servos, want %d", got, wantServos)
				}
			})
		}
	}
}

// --- block restrictions --------------------------------------------------------

func TestKld5539B1EdgecraftersEvadePowerTwoOrLess(t *testing.T) {
	for _, tc := range []struct {
		power int
		want  bool
	}{{2, false}, {1, false}, {3, true}} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		elf := b12Push(g, me.ID, "Elegant Edgecrafters", "Creature — Elf Artificer", kldB1ElegantEdgecraftersOracle, 3, 4)
		blocker := b12Creature(g, opp.ID, "Blocker", kldB1BearType, tc.power, 4)
		err := porBlock(t, g, elf, opp.ID, blocker)
		if tc.want && err != nil {
			t.Errorf("a %d-power creature couldn't block the Edgecrafters: %v", tc.power, err)
		}
		if !tc.want && err == nil {
			t.Errorf("a %d-power creature blocked the Edgecrafters", tc.power)
		}
	}
}

func TestKld5539B1DemolitionStomperEvadesPowerTwoOrLess(t *testing.T) {
	for _, tc := range []struct {
		power int
		want  bool
	}{{2, false}, {3, true}} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		stomper := b12Push(g, me.ID, "Demolition Stomper", kldB1VehicleType, kldB1DemolitionStomperOracle, 10, 7)
		kldB1Crew(t, g, me, stomper, 5)
		blocker := b12Creature(g, opp.ID, "Blocker", kldB1BearType, tc.power, 4)
		err := porBlock(t, g, stomper, opp.ID, blocker)
		if tc.want && err != nil {
			t.Errorf("a %d-power creature couldn't block the Stomper: %v", tc.power, err)
		}
		if !tc.want && err == nil {
			t.Errorf("a %d-power creature blocked the Stomper", tc.power)
		}
	}
}

func TestKld5539B1LongFinnedSkywhaleBlocksOnlyFlyers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := b12Creature(g, me.ID, "Walker", kldB1BearType, 2, 2)
	whale := b12Push(g, opp.ID, "Long-Finned Skywhale", "Creature — Whale", kldB1LongFinnedSkywhaleOracle, 4, 3)
	if err := porBlock(t, g, walker, opp.ID, whale); err == nil {
		t.Error("the Skywhale blocked a creature without flying")
	}
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	bird := porFlyer(g, me.ID, "Bird")
	whale = b12Push(g, opp.ID, "Long-Finned Skywhale", "Creature — Whale", kldB1LongFinnedSkywhaleOracle, 4, 3)
	if err := porBlock(t, g, bird, opp.ID, whale); err != nil {
		t.Errorf("the Skywhale couldn't block a flyer: %v", err)
	}
}

// --- self keyword grants ------------------------------------------------------------

func TestKld5539B1SelfKeywordGrants(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
	}{
		{"Bastion Mastodon", "Artifact Creature — Elephant", kldB1BastionMastodonOracle, "vigilance"},
		{"Dukhara Peafowl", "Artifact Creature — Bird", kldB1DukharaPeafowlOracle, "flying"},
		{"Narnam Cobra", "Artifact Creature — Snake", kldB1NarnamCobraOracle, "deathtouch"},
		{"Prakhata Pillar-Bug", "Artifact Creature — Insect", kldB1PrakhataPillarBugOracle, "lifelink"},
		{"Weldfast Monitor", "Artifact Creature — Lizard", kldB1WeldfastMonitorOracle, "menace"},
		{"Wily Bandar", "Creature — Cat Monkey", kldB1WilyBandarOracle, "indestructible"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			if hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Fatalf("%s has %s before the ability", tc.name, tc.keyword)
			}
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)
			if !hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("%s has no %s after the ability", tc.name, tc.keyword)
			}
		})
	}
}

// --- artifact statics ---------------------------------------------------------------

func TestKld5539B1PumpedWhileYouControlAnArtifact(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, tough           int
		wantP, wantT           int
	}{
		{"Inventor's Apprentice", "Creature — Human Artificer", kldB1InventorsApprenticeOracle, 1, 2, 2, 3},
		{"Foundry Screecher", "Creature — Bat", kldB1FoundryScreecherOracle, 2, 1, 3, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
			b12Permanent(g, opp.ID, "Their Rock", "Artifact")
			if p, tough := ptOf(t, g, id); p != tc.power || tough != tc.tough {
				t.Errorf("with no artifact of yours it is %d/%d, want %d/%d", p, tough, tc.power, tc.tough)
			}
			b12Permanent(g, me.ID, "My Rock", "Artifact")
			if p, tough := ptOf(t, g, id); p != tc.wantP || tough != tc.wantT {
				t.Errorf("with an artifact it is %d/%d, want %d/%d", p, tough, tc.wantP, tc.wantT)
			}
		})
	}
}

// --- enters triggers -----------------------------------------------------------------

func TestKld5539B1ArborbackStomperGainsFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	castHoldCreature(t, g, "Arborback Stomper", "Creature — Beast", kldB1ArborbackStomperOracle, 5, 4)
	if me.Life != life+5 {
		t.Errorf("life %d -> %d, want +5", life, me.Life)
	}
}

func TestKld5539B1NimbleInnovatorDraws(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	hand := me.Hand.Size()
	castHoldCreature(t, g, "Nimble Innovator", "Creature — Vedalken Artificer", kldB1NimbleInnovatorOracle, 2, 2)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

// --- the spells --------------------------------------------------------------------

func TestKld5539B1AppetiteForTheUnnaturalDestroysAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	bear := kldB1Creature(g, opp.ID, "Their Bear", 2, 2, false)
	if err := castCatalogSpellErr(t, g, "Appetite for the Unnatural", "Instant", kldB1AppetiteForTheUnnaturalOracle, cardTarget(bear)); err == nil {
		t.Fatal("Appetite for the Unnatural targeted a creature")
	}
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	life := me.Life
	castCatalogSpell(t, g, "Appetite for the Unnatural", "Instant", kldB1AppetiteForTheUnnaturalOracle, cardTarget(rock))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

func TestKld5539B1EssenceExtractionDealsThreeAndGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	bear := kldB1Creature(g, opp.ID, "Their Bear", 3, 3, false)
	life := me.Life
	castCatalogSpell(t, g, "Essence Extraction", "Instant", kldB1EssenceExtractionOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the 3/3 survived 3 damage")
	}
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3", life, me.Life)
	}
}

func TestKld5539B1FragmentizeHitsManaValueFourOrLess(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	big := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big Rock", TypeLine: "Artifact",
		ManaCost: "{5}", Owner: opp.ID, Controller: opp.ID})
	if err := castCatalogSpellErr(t, g, "Fragmentize", "Sorcery", kldB1FragmentizeOracle, cardTarget(big)); err == nil {
		t.Fatal("Fragmentize targeted a mana value 5 artifact")
	}
	small := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Aura", TypeLine: "Enchantment",
		ManaCost: "{2}{W}{W}", Owner: opp.ID, Controller: opp.ID})
	castCatalogSpell(t, g, "Fragmentize", "Sorcery", kldB1FragmentizeOracle, cardTarget(small))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the mana value 4 enchantment survived")
	}
}

func TestKld5539B1LargerThanLifePumpsAndGrantsTrample(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	bear := kldB1Creature(g, me.ID, "Bear", 2, 2, false)
	castCatalogSpell(t, g, "Larger Than Life", "Sorcery", kldB1LargerThanLifeOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("the Bear is %d/%d, want 6/6", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear has no trample")
	}
}

func TestKld5539B1ServoExhibitionMakesTwoServos(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	castCatalogSpell(t, g, "Servo Exhibition", "Sorcery", kldB1ServoExhibitionOracle, nil)
	passPriorityAroundTable(t, g)
	if got := cardsNamed(g, "Servo", me.ID); got != 2 {
		t.Errorf("%d Servos, want 2", got)
	}
}

func TestKld5539B1SkywhalersShotDestroysPowerThreeOrGreater(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	small := kldB1Creature(g, opp.ID, "Small", 2, 2, false)
	if err := castCatalogSpellErr(t, g, "Skywhaler's Shot", "Instant", kldB1SkywhalersShotOracle, cardTarget(small)); err == nil {
		t.Fatal("Skywhaler's Shot targeted a 2-power creature")
	}
	big := kldB1Creature(g, opp.ID, "Big", 3, 3, false)
	castCatalogSpell(t, g, "Skywhaler's Shot", "Instant", kldB1SkywhalersShotOracle, cardTarget(big))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the 3-power creature survived")
	}
}

func TestKld5539B1SelectForInspectionBouncesATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	untapped := kldB1Creature(g, opp.ID, "Untapped", 2, 2, false)
	if err := castCatalogSpellErr(t, g, "Select for Inspection", "Instant", kldB1SelectForInspectionOracle, cardTarget(untapped)); err == nil {
		t.Fatal("Select for Inspection targeted an untapped creature")
	}
	tapped := kldB1Creature(g, opp.ID, "Tapped", 2, 2, true)
	castCatalogSpell(t, g, "Select for Inspection", "Instant", kldB1SelectForInspectionOracle, cardTarget(tapped))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) || !opp.Hand.Contains(tapped) {
		t.Error("the tapped creature did not return to its owner's hand")
	}
}

func TestKld5539B1RefurbishReturnsAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	bear := wthGraveyardCard(me, "Bear", kldB1BearType)
	if err := castCatalogSpellErr(t, g, "Refurbish", "Sorcery", kldB1RefurbishOracle, cardTarget(bear)); err == nil {
		t.Fatal("Refurbish targeted a creature card")
	}
	rock := wthGraveyardCard(me, "Old Rock", "Artifact")
	castCatalogSpell(t, g, "Refurbish", "Sorcery", kldB1RefurbishOracle, cardTarget(rock))
	passPriorityAroundTable(t, g)
	if _, ok := visBattlefieldNamed(g, me.ID, "Old Rock"); !ok {
		t.Error("the artifact card did not return to the battlefield under your control")
	}
}

func TestKld5539B1CeremoniousRejectionCountersOnlyColorless(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	red := kldB1OpponentCastsInstant(t, g, opp, "Their Bolt", "{R}", "R")
	id := b13HandCard(me, "Ceremonious Rejection", "Instant", kldB1CeremoniousRejectionOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(red)}); err == nil {
		t.Fatal("Ceremonious Rejection targeted a red spell")
	}

	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	trinket := kldB1OpponentCastsInstant(t, g, opp, "Their Trinket", "{2}")
	id = b13HandCard(me, "Ceremonious Rejection", "Instant", kldB1CeremoniousRejectionOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(trinket)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(trinket) {
		t.Error("the colorless spell was not countered")
	}
}

func TestKld5539B1MakeObsoleteShrinksOnlyOpponents(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	mine := kldB1Creature(g, me.ID, "Mine", 1, 1, false)
	theirs := kldB1Creature(g, opp.ID, "Theirs", 1, 1, false)
	bigger := kldB1Creature(g, opp.ID, "Bigger", 3, 3, false)
	castCatalogSpell(t, g, "Make Obsolete", "Instant", kldB1MakeObsoleteOracle, nil)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's 1/1 survived -1/-1")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your own 1/1 died")
	}
	if p, tough := ptOf(t, g, bigger); p != 2 || tough != 2 {
		t.Errorf("the opponent's 3/3 is %d/%d, want 2/2", p, tough)
	}
}

func TestKld5539B1AcrobaticManeuverFlickersAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	theirs := kldB1Creature(g, opp.ID, "Their Bear", 2, 2, false)
	if err := castCatalogSpellErr(t, g, "Acrobatic Maneuver", "Instant", kldB1AcrobaticManeuverOracle, cardTarget(theirs)); err == nil {
		t.Fatal("Acrobatic Maneuver targeted a creature an opponent controls")
	}
	mine := kldB1Creature(g, me.ID, "My Bear", 2, 2, false)
	castCatalogSpell(t, g, "Acrobatic Maneuver", "Instant", kldB1AcrobaticManeuverOracle, cardTarget(mine))
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if _, ok := visBattlefieldNamed(g, me.ID, "My Bear"); !ok {
		t.Error("the creature did not return to the battlefield")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

// --- equipment and vehicles ----------------------------------------------------------

func TestKld5539B1TorchGauntletGivesPlusTwo(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	gauntlet := pushCatalogPermanent(g, me.ID, "Torch Gauntlet", "Artifact — Equipment", kldB1TorchGauntletOracle, false)
	bear := kldB1Creature(g, me.ID, "Bear", 2, 2, false)
	equipTo(t, g, me.ID, gauntlet, bear)
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the equipped Bear is %d/%d, want 4/2", p, tough)
	}
}

func TestKld5539B1VehiclesCrewIntoKeywordedCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
		crew         int
		keywords     []string
	}{
		{"Sky Skiff", kldB1SkySkiffOracle, 2, 3, 1, []string{"flying"}},
		{"Ovalchase Dragster", kldB1OvalchaseDragsterOracle, 6, 1, 1, []string{"trample", "haste"}},
		{"Aradara Express", kldB1AradaraExpressOracle, 8, 6, 4, []string{"menace"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			ship := b12Push(g, me.ID, tc.name, kldB1VehicleType, tc.oracle, tc.power, tc.tough)
			if avB9Live(t, g, ship).IsCreature() {
				t.Fatal("the uncrewed Vehicle is a creature")
			}
			kldB1Crew(t, g, me, ship, tc.crew)
			if !avB9Live(t, g, ship).IsCreature() {
				t.Fatal("the crewed Vehicle is not a creature")
			}
			for _, kw := range tc.keywords {
				if !hasEffectiveKeyword(t, g, ship, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

func TestKld5539B1AradaraExpressNeedsCrewFour(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	ship := b12Push(g, me.ID, "Aradara Express", kldB1VehicleType, kldB1AradaraExpressOracle, 8, 6)
	small := b12Creature(g, me.ID, "Small", kldB1BearType, 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, ship, 0, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{small}}); err == nil {
		t.Fatal("Crew 4 was paid with power 3")
	}
}
