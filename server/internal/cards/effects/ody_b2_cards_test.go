package effects

import (
	"slices"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ody_b2_cards_test.go — Odyssey, batch 2 (#5613). Each test plays the
// real catalog entry through the engine and asserts what its text says.

const (
	odyB2DuskImpOracle          = "1389ae4a-c3a5-4678-9012-937a3cbaf7f7"
	odyB2HalberdierOracle       = "1c3c23b5-b771-4117-83b5-febf7e96281a"
	odyB2SmokeweaverOracle      = "e80be353-8b19-42d2-9141-5353a8409c8e"
	odyB2TreetopSentinelOracle  = "786afd52-0eae-4453-b1bd-cffc21699849"
	odyB2IridescentAngelOracle  = "28fa9211-96bb-4fef-81bf-8f58114a8df7"
	odyB2BelovedChaplainOracle  = "a07406b7-faa1-4ed3-ab84-314c40f3f7f1"
	odyB2DwarvenGruntOracle     = "6da00e2d-4287-48eb-8abe-d153318381fd"
	odyB2LeafDancerOracle       = "b2f92310-3725-44c2-8953-8b280437217e"
	odyB2WoodlandDruidOracle    = "96b3c4ff-b5f7-4315-9b39-2e2146ef10d1"
	odyB2DarkwaterEggOracle     = "c4893a24-3cbc-4011-bef4-50e0c4dce16e"
	odyB2MossfireEggOracle      = "364b3231-c0e7-45a8-90b9-a1cdb584dd7c"
	odyB2ShadowbloodEggOracle   = "9b58e7fa-4259-4d6b-8b5d-33fb37fe489f"
	odyB2SkycloudEggOracle      = "1c4b6543-777e-4c3b-a9fb-5b7210d458d5"
	odyB2SungrassEggOracle      = "80a49a1b-a202-4c14-b093-dc76eb0f42c7"
	odyB2AbandonedOutpostOracle = "5a70ccfa-d12d-4e62-a1a4-f05cda2fd442"
	odyB2BogWreckageOracle      = "bce32541-3bc4-4553-ace2-784a58f164aa"
	odyB2RavagedHighlandsOracle = "68919931-b995-4aa0-9d2a-e80eae2f335f"
	odyB2SeafloorDebrisOracle   = "eef64810-4187-40c1-8e59-1cb8773ebea3"
	odyB2TimberlandRuinsOracle  = "29c7f059-2eeb-40f9-8f50-03eba2d0d5e0"
	odyB2CabalPitOracle         = "92392467-a22f-4dd7-a0eb-393bef956dc0"
	odyB2CentaurGardenOracle    = "5cf92fd4-7c0b-4d8e-92f1-53dc2e0476fc"
	odyB2NomadStadiumOracle     = "4034bec6-e3c7-4d3f-81df-7c903977a606"
	odyB2RayOfDistortionOracle  = "1c838441-4713-49dd-97e6-4fb7155c36c8"
	odyB2BashToBitsOracle       = "894b8375-5d55-4138-8448-c38f83a17e4f"
	odyB2EarthRiftOracle        = "888a5071-24b1-48ce-a97b-39f108976c6d"
	odyB2DematerializeOracle    = "d2fbad7d-d05a-4321-9aff-a93bd102a3f4"
	odyB2FerventDenialOracle    = "2e56bc34-1e35-4b95-9525-cb58ac8ba060"
	odyB2ElephantAmbushOracle   = "09386931-59bd-481f-bccc-32e8a8abc0f9"
	odyB2CoffinPurgeOracle      = "9db5aa54-3deb-45ee-82a2-e88a4367c6d9"
	odyB2MorbidHungerOracle     = "6e7e9685-ebe9-4db2-b52c-f6ce95e37e56"
	odyB2ScorchingMissileOracle = "02e2dbde-2373-4f29-bd9d-787f1805b061"
	odyB2CephalidLooterOracle   = "412294bf-a25f-47c2-8043-12eb3e142b9e"
	odyB2CephalidBrokerOracle   = "18bb35a8-01ac-414f-9a45-e550060ff905"
)

func TestOdyB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		odyB2DuskImpOracle:          "Dusk Imp",
		odyB2HalberdierOracle:       "Halberdier",
		odyB2SmokeweaverOracle:      "Aven Smokeweaver",
		odyB2TreetopSentinelOracle:  "Treetop Sentinel",
		odyB2IridescentAngelOracle:  "Iridescent Angel",
		odyB2BelovedChaplainOracle:  "Beloved Chaplain",
		odyB2DwarvenGruntOracle:     "Dwarven Grunt",
		odyB2LeafDancerOracle:       "Leaf Dancer",
		odyB2WoodlandDruidOracle:    "Woodland Druid",
		odyB2DarkwaterEggOracle:     "Darkwater Egg",
		odyB2MossfireEggOracle:      "Mossfire Egg",
		odyB2ShadowbloodEggOracle:   "Shadowblood Egg",
		odyB2SkycloudEggOracle:      "Skycloud Egg",
		odyB2SungrassEggOracle:      "Sungrass Egg",
		odyB2AbandonedOutpostOracle: "Abandoned Outpost",
		odyB2BogWreckageOracle:      "Bog Wreckage",
		odyB2RavagedHighlandsOracle: "Ravaged Highlands",
		odyB2SeafloorDebrisOracle:   "Seafloor Debris",
		odyB2TimberlandRuinsOracle:  "Timberland Ruins",
		odyB2CabalPitOracle:         "Cabal Pit",
		odyB2CentaurGardenOracle:    "Centaur Garden",
		odyB2NomadStadiumOracle:     "Nomad Stadium",
		odyB2RayOfDistortionOracle:  "Ray of Distortion",
		odyB2BashToBitsOracle:       "Bash to Bits",
		odyB2EarthRiftOracle:        "Earth Rift",
		odyB2DematerializeOracle:    "Dematerialize",
		odyB2FerventDenialOracle:    "Fervent Denial",
		odyB2ElephantAmbushOracle:   "Elephant Ambush",
		odyB2CoffinPurgeOracle:      "Coffin Purge",
		odyB2MorbidHungerOracle:     "Morbid Hunger",
		odyB2ScorchingMissileOracle: "Scorching Missile",
		odyB2CephalidLooterOracle:   "Cephalid Looter",
		odyB2CephalidBrokerOracle:   "Cephalid Broker",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- keyword bodies ------------------------------------------------------------

func TestOdyB2KeywordCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Dusk Imp", "Creature — Imp", odyB2DuskImpOracle, []string{"flying"}},
		{"Halberdier", "Creature — Human Barbarian", odyB2HalberdierOracle, []string{"first strike"}},
		{"Aven Smokeweaver", "Creature — Bird Soldier", odyB2SmokeweaverOracle, []string{"flying", "protection from red"}},
		{"Treetop Sentinel", "Creature — Bird Soldier", odyB2TreetopSentinelOracle, []string{"flying", "protection from green"}},
		{"Iridescent Angel", "Creature — Angel", odyB2IridescentAngelOracle, []string{
			"flying", "protection from white", "protection from blue", "protection from black",
			"protection from red", "protection from green",
		}},
		{"Beloved Chaplain", "Creature — Human Cleric", odyB2BelovedChaplainOracle, []string{"protection from creatures"}},
		{"Dwarven Grunt", "Creature — Dwarf", odyB2DwarvenGruntOracle, []string{"mountainwalk"}},
		{"Leaf Dancer", "Creature — Centaur", odyB2LeafDancerOracle, []string{"forestwalk"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := blbB1Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			for _, kw := range tc.keywords {
				if !hasKeywordOnBattlefield(t, g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

func TestOdyB2WoodlandDruidIsVanilla(t *testing.T) {
	spec, ok := Lookup(odyB2WoodlandDruidOracle)
	if !ok {
		t.Fatal("Woodland Druid is not in the catalog")
	}
	if len(spec.PrintedKeywords) != 0 || len(spec.Activated) != 0 || len(spec.Static) != 0 ||
		len(spec.Triggered) != 0 || len(spec.ManaAbilities) != 0 {
		t.Errorf("Woodland Druid declares rules: %+v", spec)
	}
}

// odyB2BoltRefused reports whether a red Lightning Bolt cast by the active
// seat is refused when it targets `id`.
func odyB2BoltRefused(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: lightningBoltOracle,
		ManaCost: "{R}", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID,
	})
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: cardTarget(id)})
	me.ManaPool.EmptyPool()
	return err != nil
}

// Protection from red, and from each color, refuses a red spell's target
// (CR 702.16b); protection from green does not.
func TestOdyB2ProtectionRefusesARedBolt(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		refused      bool
	}{
		{"Aven Smokeweaver", odyB2SmokeweaverOracle, true},
		{"Iridescent Angel", odyB2IridescentAngelOracle, true},
		{"Treetop Sentinel", odyB2TreetopSentinelOracle, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := blbB1Push(g, g.Seats[g.Turn.ActiveSeat].ID, tc.name, "Creature — Test", tc.oracle, 4, 4)
			if got := odyB2BoltRefused(t, g, id); got != tc.refused {
				t.Errorf("Lightning Bolt refused = %v, want %v", got, tc.refused)
			}
		})
	}
}

// --- the Eggs --------------------------------------------------------------------

func TestOdyB2EggsAddTheirTwoColorsAndDraw(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		colors       []string
	}{
		{"Darkwater Egg", odyB2DarkwaterEggOracle, []string{"B", "U"}},
		{"Mossfire Egg", odyB2MossfireEggOracle, []string{"G", "R"}},
		{"Shadowblood Egg", odyB2ShadowbloodEggOracle, []string{"B", "R"}},
		{"Skycloud Egg", odyB2SkycloudEggOracle, []string{"U", "W"}},
		{"Sungrass Egg", odyB2SungrassEggOracle, []string{"G", "W"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			seedLibrary(me, "A", "B", "C")
			egg := pushCatalogPermanent(g, me.ID, tc.name, "Artifact", tc.oracle, false)
			apaMana(me, "C")
			if err := g.ActivateCatalogAbility(me.ID, egg, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Fatal("the Egg activated with only {1} to pay {2}")
			}
			me.ManaPool.EmptyPool()
			hand := me.Hand.Size()
			apaMana(me, "C", "C")
			if err := g.ActivateCatalogAbility(me.ID, egg, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			// Not a mana ability (CR 605.1a): nothing is added until it resolves.
			if len(g.StackMeta) != 1 || me.ManaPool.Total() != 0 || me.Hand.Size() != hand {
				t.Fatalf("stack %d, pool %d, hand %d (was %d): want the ability waiting on the stack",
					len(g.StackMeta), me.ManaPool.Total(), me.Hand.Size(), hand)
			}
			if !me.Graveyard.Contains(egg) {
				t.Error("the Egg was not sacrificed as a cost")
			}
			passPriorityAroundTable(t, g)
			got := poolColors(me)
			sort.Strings(got)
			if !slices.Equal(got, tc.colors) {
				t.Errorf("pool %v, want %v", got, tc.colors)
			}
			if me.Hand.Size() != hand+1 {
				t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
			}
		})
	}
}

// --- the sacrifice lands -----------------------------------------------------------

var odyB2SacrificeLands = []struct{ name, oracle, color string }{
	{"Abandoned Outpost", odyB2AbandonedOutpostOracle, "W"},
	{"Bog Wreckage", odyB2BogWreckageOracle, "B"},
	{"Ravaged Highlands", odyB2RavagedHighlandsOracle, "R"},
	{"Seafloor Debris", odyB2SeafloorDebrisOracle, "U"},
	{"Timberland Ruins", odyB2TimberlandRuinsOracle, "G"},
}

func TestOdyB2SacrificeLandsEnterTapped(t *testing.T) {
	for _, tc := range odyB2SacrificeLands {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			land := b31PlayLand(t, g, tc.name, "Land", tc.oracle)
			if !b31Tapped(t, g, land) {
				t.Error("the land entered untapped")
			}
		})
	}
}

func TestOdyB2SacrificeLandsTapForTheirColorOrSacrificeForAny(t *testing.T) {
	for _, tc := range odyB2SacrificeLands {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			land := pushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
			if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatal(err)
			}
			if got := poolColors(me); !slices.Equal(got, []string{tc.color}) {
				t.Errorf("{T} row: pool %v, want [%s]", got, tc.color)
			}
			if !g.Battlefield.Contains(land) {
				t.Error("the {T} row sacrificed the land")
			}

			for _, color := range []string{"W", "U", "B", "R", "G"} {
				g := newCatalogGame(t)
				advanceToMain(t, g)
				me := g.Seats[g.Turn.ActiveSeat]
				land := pushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
				if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{Colors: []string{color}}); err != nil {
					t.Fatal(err)
				}
				if got := poolColors(me); !slices.Equal(got, []string{color}) {
					t.Errorf("sacrifice row picking %s: pool %v", color, got)
				}
				if g.Battlefield.Contains(land) || !me.Graveyard.Contains(land) {
					t.Error("the sacrifice row did not sacrifice the land")
				}
			}
		})
	}
}

// --- the threshold lands -------------------------------------------------------------

func TestOdyB2ThresholdLandsTapForPainfulMana(t *testing.T) {
	for _, tc := range []struct{ name, oracle, color string }{
		{"Cabal Pit", odyB2CabalPitOracle, "B"},
		{"Centaur Garden", odyB2CentaurGardenOracle, "G"},
		{"Nomad Stadium", odyB2NomadStadiumOracle, "W"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			land := pushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
			life := me.Life
			if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatal(err)
			}
			if got := poolColors(me); !slices.Equal(got, []string{tc.color}) {
				t.Errorf("pool %v, want [%s]", got, tc.color)
			}
			if me.Life != life-1 {
				t.Errorf("life %d -> %d, want 1 damage", life, me.Life)
			}
		})
	}
}

// odyB2ThresholdLand puts the land under the active seat with six cards in
// its graveyard and the mana for the threshold row in its pool.
func odyB2ThresholdLand(t *testing.T, g *game.Game, name, oracle, color string) (*game.Player, uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	land := pushCatalogPermanent(g, me.ID, name, "Land", oracle, false)
	acGraveyardCards(me, 6)
	advanceToMain(t, g)
	b06AddMana(me, color)
	return me, land
}

func TestOdyB2CabalPitShrinksATargetCreatureAtThreshold(t *testing.T) {
	g := newCatalogGame(t)
	me, pit := odyB2ThresholdLand(t, g, "Cabal Pit", odyB2CabalPitOracle, "B")
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	acRefused(t, g, me.ID, pit, 0, game.ActivateAbilityParams{Targets: cardTarget(ogre)})
	acGraveyardCards(me, 1)
	b16Activate(t, g, me.ID, pit, 0, game.ActivateAbilityParams{Targets: cardTarget(ogre)})
	if g.Battlefield.Contains(pit) {
		t.Error("Cabal Pit was not sacrificed")
	}
	if p, tough := ptOf(t, g, ogre); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tough)
	}
}

func TestOdyB2CentaurGardenGrowsATargetCreatureAtThreshold(t *testing.T) {
	g := newCatalogGame(t)
	me, garden := odyB2ThresholdLand(t, g, "Centaur Garden", odyB2CentaurGardenOracle, "G")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	acRefused(t, g, me.ID, garden, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	acGraveyardCards(me, 1)
	b16Activate(t, g, me.ID, garden, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if p, tough := ptOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the target is %d/%d, want 5/5", p, tough)
	}
	advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if p, tough := ptOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("next turn the target is %d/%d, want 2/2", p, tough)
	}
}

func TestOdyB2NomadStadiumGainsFourAtThreshold(t *testing.T) {
	g := newCatalogGame(t)
	me, stadium := odyB2ThresholdLand(t, g, "Nomad Stadium", odyB2NomadStadiumOracle, "W")
	acRefused(t, g, me.ID, stadium, 0, game.ActivateAbilityParams{})
	acGraveyardCards(me, 1)
	life := me.Life
	b16Activate(t, g, me.ID, stadium, 0, game.ActivateAbilityParams{})
	if me.Life != life+4 {
		t.Errorf("life %d -> %d, want +4", life, me.Life)
	}
	if g.Battlefield.Contains(stadium) {
		t.Error("Nomad Stadium was not sacrificed")
	}
}

// --- flashback --------------------------------------------------------------------

var odyB2FlashbackCards = []struct{ name, oracle, cost string }{
	{"Ray of Distortion", odyB2RayOfDistortionOracle, "{4}{W}{W}"},
	{"Bash to Bits", odyB2BashToBitsOracle, "{4}{R}{R}"},
	{"Earth Rift", odyB2EarthRiftOracle, "{5}{R}{R}"},
	{"Dematerialize", odyB2DematerializeOracle, "{5}{U}{U}"},
	{"Fervent Denial", odyB2FerventDenialOracle, "{5}{U}{U}"},
	{"Elephant Ambush", odyB2ElephantAmbushOracle, "{6}{G}{G}"},
	{"Coffin Purge", odyB2CoffinPurgeOracle, "{B}"},
	{"Morbid Hunger", odyB2MorbidHungerOracle, "{7}{B}{B}"},
	{"Scorching Missile", odyB2ScorchingMissileOracle, "{9}{R}"},
}

func TestOdyB2FlashbackCardsOfferTheirPrintedPriceFromTheGraveyard(t *testing.T) {
	for _, c := range odyB2FlashbackCards {
		t.Run(c.name, func(t *testing.T) {
			if !game.CardCastableFromZone(c.oracle, game.ZoneGraveyard) {
				t.Fatalf("%s does not open the graveyard", c.name)
			}
			offers := game.AlternativeCostsOfferedFromZone(c.oracle, game.ZoneGraveyard)
			if len(offers) != 1 || offers[0].Key != "flashback" {
				t.Fatalf("%s graveyard offers = %+v, want one flashback", c.name, offers)
			}
			if offers[0].ManaCost != c.cost || !offers[0].ExileOnLeavingStack {
				t.Errorf("%s flashback pays %q (exile %v), want %q and exile", c.name, offers[0].ManaCost, offers[0].ExileOnLeavingStack, c.cost)
			}
		})
	}
}

// The three destroy spells take only their printed kind of permanent, from
// hand and again with flashback.
func TestOdyB2DestroySpellsHitTheirKindFromHandAndGraveyard(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, cost string
		kinds                        []string
		wrong                        string
	}{
		{"Ray of Distortion", "Instant", odyB2RayOfDistortionOracle, "{4}{W}{W}", []string{"Artifact", "Enchantment"}, "Land"},
		{"Bash to Bits", "Instant", odyB2BashToBitsOracle, "{4}{R}{R}", []string{"Artifact"}, "Enchantment"},
		{"Earth Rift", "Sorcery", odyB2EarthRiftOracle, "{5}{R}{R}", []string{"Land"}, "Artifact"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			_, opp, _ := blbB1Seats(g)
			wrong := b12Permanent(g, opp.ID, "Wrong", tc.wrong)
			if err := castCatalogSpellErr(t, g, tc.name, tc.typeLine, tc.oracle, cardTarget(wrong)); err == nil {
				t.Fatalf("%s targeted a %s", tc.name, tc.wrong)
			}
			for _, kind := range tc.kinds {
				g := newCatalogGame(t)
				_, opp, _ := blbB1Seats(g)
				target := b12Permanent(g, opp.ID, "Target", kind)
				castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, cardTarget(target))
				passPriorityAroundTable(t, g)
				if g.Battlefield.Contains(target) {
					t.Errorf("the %s survived %s", kind, tc.name)
				}
			}
			again := b12Permanent(g, opp.ID, "Again", tc.kinds[0])
			id := seedGraveyardCard(t, g, tc.name, tc.typeLine, tc.oracle)
			odyB1Flashback(t, g, id, tc.cost, cardTarget(again)...)
			if g.Battlefield.Contains(again) {
				t.Errorf("the flashed-back %s did not destroy the %s", tc.name, tc.kinds[0])
			}
		})
	}
}

func TestOdyB2DematerializeReturnsAPermanentToItsOwner(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	relic := b12Permanent(g, opp.ID, "Relic", "Enchantment")
	castCatalogSpell(t, g, "Dematerialize", "Sorcery", odyB2DematerializeOracle, cardTarget(relic))
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(relic) {
		t.Fatal("the permanent did not return to its owner's hand")
	}

	land := b12Permanent(g, opp.ID, "Land", "Land")
	id := seedGraveyardCard(t, g, "Dematerialize", "Sorcery", odyB2DematerializeOracle)
	odyB1Flashback(t, g, id, "{5}{U}{U}", cardTarget(land)...)
	if !opp.Hand.Contains(land) {
		t.Error("the flashed-back Dematerialize did not return the land")
	}
}

func TestOdyB2FerventDenialCountersFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	advanceToMain(t, g)
	life := me.Life
	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: lightningBoltOracle,
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{playerTarget(me.ID)}}); err != nil {
		t.Fatalf("the opponent's Bolt: %v", err)
	}
	denial := seedGraveyardCard(t, g, "Fervent Denial", "Instant", odyB2FerventDenialOracle)
	odyB1Flashback(t, g, denial, "{5}{U}{U}", game.TargetRef{Kind: game.TargetCard, ID: bolt})
	if me.Life != life {
		t.Errorf("life %d -> %d: the countered Bolt dealt damage", life, me.Life)
	}
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the countered Bolt is not in its owner's graveyard")
	}
}

func TestOdyB2ElephantAmbushMakesAnElephantFromHandAndGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Elephant Ambush", "Instant", odyB2ElephantAmbushOracle, nil)
	passPriorityAroundTable(t, g)
	if got := b43TokensNamed(g, me.ID, "Elephant"); got != 1 {
		t.Fatalf("%d Elephants after the hand cast, want 1", got)
	}
	id := seedGraveyardCard(t, g, "Elephant Ambush", "Instant", odyB2ElephantAmbushOracle)
	odyB1Flashback(t, g, id, "{6}{G}{G}")
	if got := b43TokensNamed(g, me.ID, "Elephant"); got != 2 {
		t.Fatalf("%d Elephants after the flashback, want 2", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Elephant" {
			if p, tough := ptOf(t, g, c.InstanceID); p != 3 || tough != 3 || !c.HasColor("G") {
				t.Errorf("the Elephant is %d/%d green=%v, want a green 3/3", p, tough, c.HasColor("G"))
			}
		}
	}
}

func TestOdyB2CoffinPurgeExilesACardFromAnyGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	theirs := pushGraveyardCardTyped(opp, "Their Card", "Creature — Bear")
	castCatalogSpell(t, g, "Coffin Purge", "Instant", odyB2CoffinPurgeOracle, cardTarget(theirs))
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) {
		t.Fatal("the opponent's graveyard card was not exiled")
	}

	another := pushGraveyardCardTyped(opp, "Another", "Sorcery")
	purge := seedGraveyardCard(t, g, "Coffin Purge", "Instant", odyB2CoffinPurgeOracle)
	odyB1Flashback(t, g, purge, "{B}", cardTarget(another)...)
	if !g.Exile.Contains(another) {
		t.Error("the flashed-back Coffin Purge did not exile its target")
	}
}

func TestOdyB2MorbidHungerDealsThreeAndGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	myLife, oppLife := me.Life, opp.Life
	castCatalogSpell(t, g, "Morbid Hunger", "Sorcery", odyB2MorbidHungerOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-3 || me.Life != myLife+3 {
		t.Fatalf("opponent %d -> %d, me %d -> %d: want 3 damage and 3 life", oppLife, opp.Life, myLife, me.Life)
	}

	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	hunger := seedGraveyardCard(t, g, "Morbid Hunger", "Sorcery", odyB2MorbidHungerOracle)
	myLife = me.Life
	odyB1Flashback(t, g, hunger, "{7}{B}{B}", cardTarget(ogre)...)
	if g.Battlefield.Contains(ogre) {
		t.Error("the flashed-back Morbid Hunger did not kill the 3/3")
	}
	if me.Life != myLife+3 {
		t.Errorf("life %d -> %d after the flashback, want +3", myLife, me.Life)
	}
}

func TestOdyB2ScorchingMissileHitsOnlyAPlayerOrPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Scorching Missile", "Sorcery", odyB2ScorchingMissileOracle, cardTarget(bear)); err == nil {
		t.Fatal("Scorching Missile targeted a creature")
	}
	g = newCatalogGame(t)
	_, opp, _ = blbB1Seats(g)
	life := opp.Life
	castCatalogSpell(t, g, "Scorching Missile", "Sorcery", odyB2ScorchingMissileOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Fatalf("opponent life %d -> %d, want 4 damage", life, opp.Life)
	}
	missile := seedGraveyardCard(t, g, "Scorching Missile", "Sorcery", odyB2ScorchingMissileOracle)
	odyB1Flashback(t, g, missile, "{9}{R}", playerTarget(opp.ID))
	if opp.Life != life-8 {
		t.Errorf("opponent life %d after the flashback, want %d", opp.Life, life-8)
	}
}

// --- the Cephalids ------------------------------------------------------------------

func TestOdyB2CephalidsMakeATargetPlayerDrawThenDiscard(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		n                      int
	}{
		{"Cephalid Looter", "Creature — Octopus Rogue", odyB2CephalidLooterOracle, 1},
		{"Cephalid Broker", "Creature — Octopus", odyB2CephalidBrokerOracle, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me, opp, _ := blbB1Seats(g)
			seedLibrary(opp, "A", "B", "C")
			src := pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			hand := opp.Hand.Size()
			b16Activate(t, g, me.ID, src, 0, game.ActivateAbilityParams{Targets: acPlayerRef(opp.ID)})
			if !b16Tapped(t, g, src) {
				t.Error("the ability did not tap its source")
			}
			if opp.Hand.Size() != hand+tc.n {
				t.Fatalf("target's hand %d -> %d, want +%d before the discard", hand, opp.Hand.Size(), tc.n)
			}
			if got := discardOwed(g, opp.ID); got != tc.n {
				t.Fatalf("the target owes %d discards, want %d", got, tc.n)
			}
			discard := make([]uuid.UUID, 0, tc.n)
			for _, c := range opp.Hand.Cards[:tc.n] {
				discard = append(discard, c.InstanceID)
			}
			answerDiscard(t, g, opp.ID, discard...)
			if opp.Hand.Size() != hand {
				t.Errorf("target's hand %d -> %d, want back to %d", hand, opp.Hand.Size(), hand)
			}
		})
	}
}

// A summoning-sick Looter can't use its {T} ability (CR 302.6).
func TestOdyB2CephalidLooterNeedsToBeUnsick(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, _ := blbB1Seats(g)
	looter := pushCatalogPermanent(g, me.ID, "Cephalid Looter", "Creature — Octopus Rogue", odyB2CephalidLooterOracle, true)
	if err := g.ActivateCatalogAbility(me.ID, looter, 0, game.ActivateAbilityParams{Targets: acPlayerRef(opp.ID)}); err == nil {
		t.Error("a summoning-sick Looter activated its {T} ability")
	}
}
