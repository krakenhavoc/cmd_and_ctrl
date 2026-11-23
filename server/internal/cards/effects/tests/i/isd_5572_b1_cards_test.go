package testsi

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// isd_5572_b1_cards_test.go — Innistrad, batch 1 (#5572): six vanilla
// bodies, creatures whose text is keywords, a block rule or a tribal
// anthem, creatures and permanents whose abilities an existing catalog
// card already prints, and spells built from existing bodies (eight of
// them with the shared Flashback). Each test plays the real catalog
// entry through the engine.

const (
	isdB1AbbeyGriffinOracle       = "6eb74f6e-1c9d-4caf-a0d8-fa472d435346"
	isdB1ChapelGeistOracle        = "f8219a99-1620-40aa-a257-fc3dfb239d51"
	isdB1MoonHeronOracle          = "57dfe4aa-1e9a-4891-851b-201edd1fc79c"
	isdB1MarkovPatricianOracle    = "d56602f3-64df-4c68-a4a3-95c37c021e72"
	isdB1VoicelessSpiritOracle    = "56cb7592-ad58-42dc-b568-4b7754bba53d"
	isdB1SpectralRiderOracle      = "713641c4-544e-4c25-ab11-10894be2cb53"
	isdB1GraveBrambleOracle       = "9e2f854b-17f7-4d52-927c-abdfef86c1f8"
	isdB1EliteInquisitorOracle    = "4f7e865b-6ab0-4361-b864-63c38cdb2fed"
	isdB1InvisibleStalkerOracle   = "1a137a67-4156-45c8-aae9-838f688e12bd"
	isdB1OrchardSpiritOracle      = "b6eb56a7-4dfe-4d51-a1df-55f892c1b014"
	isdB1BattlegroundGeistOracle  = "2450d35b-7d87-4885-b396-f8e1d3ac0b22"
	isdB1GallowsWardenOracle      = "06157e34-dbe1-4936-a718-3b1d70726794"
	isdB1ArmoredSkaabOracle       = "74d85903-6404-41d9-bc00-81c6b2f61bb0"
	isdB1CrosswayVampireOracle    = "543071da-fc89-4274-a2ae-d5d755d1f2a9"
	isdB1DarkthicketWolfOracle    = "a78ab15e-72e7-4d2b-bc42-b2960147ea7b"
	isdB1FeralRidgewolfOracle     = "3a799843-1304-4abf-b8d7-db616628ee3f"
	isdB1KessigWolfOracle         = "7b82045f-5a2b-4e6e-976c-986c80ecd8b2"
	isdB1LanternSpiritOracle      = "cf590bb7-2ea4-446a-9ca2-1e0706c2c41e"
	isdB1ManorSkeletonOracle      = "aa3f4938-ac4b-4f52-a8d3-481c1b387467"
	isdB1MausoleumGuardOracle     = "64f3af46-f8ab-4ba4-9dff-7412cefb4eea"
	isdB1PitchburnDevilsOracle    = "86e2bbdd-a788-4f76-9ebb-ef48d9bc9dcf"
	isdB1HollowhengeOracle        = "60116231-cc54-4c74-b3d9-74e517182f1f"
	isdB1StromkirkPatrolOracle    = "5c5fbb54-f0e3-4afe-9158-ff9eadaa338d"
	isdB1SilverchaseFoxOracle     = "f277b082-dfc8-45c6-bab5-b63cd218224c"
	isdB1GhoulcallersBellOracle   = "36e1485d-3b8c-4c23-9d12-6937fd216b25"
	isdB1CobbledWingsOracle       = "8d8682f3-9ef3-4aa7-9ea6-8a2ce09bff6f"
	isdB1SpectralFlightOracle     = "d127d1e6-f3a6-4f68-bea5-d7b4123476e5"
	isdB1SensoryDeprivationOracle = "0006faf6-7a61-426c-9034-579f2cfcfa83"
	isdB1GruesomeDeformityOracle  = "9b44cc9e-6ef0-4fee-b79a-3e34cb0a3407"
	isdB1BramblecrushOracle       = "367282d6-956b-470d-a131-d22beff610b0"
	isdB1SmiteTheMonstrousOracle  = "649512ed-4908-450a-9652-b2c80df5e24d"
	isdB1RebukeOracle             = "cf35ec1e-11e2-4028-89df-9755fd6772bf"
	isdB1UrgentExorcismOracle     = "2f87ebc6-68b0-4121-bfaf-b412b077d75b"
	isdB1VictimOfNightOracle      = "d92b9f63-17b6-4ed2-b6ae-083a8161863e"
	isdB1BrimstoneVolleyOracle    = "d12e374d-ed09-4d5e-bb7a-6c7d9ea80a1e"
	isdB1BumpInTheNightOracle     = "d9f2c571-b357-4d4d-94c5-fd47551bb842"
	isdB1DreamTwistOracle         = "d569d969-ce39-4314-b5f0-76b45f5c4c7f"
	isdB1GeistflameOracle         = "2f01c3d9-e0fc-4cdf-b4db-daeb8bb24bc3"
	isdB1FeelingOfDreadOracle     = "269208e1-353c-4334-941a-2177f6665fba"
	isdB1RallyThePeasantsOracle   = "3f0a0a15-613a-4864-9309-a340dc4cf94d"
	isdB1RollingTemblorOracle     = "e450014f-7152-476b-ae6e-0d49fcb91d18"
	isdB1PurifyTheGraveOracle     = "1dafc03b-6c45-4f28-8a1f-b4f1044adc97"
	isdB1TravelPreparationsOracle = "8432729e-5fd4-4661-8b79-e41b1d03eb7b"
	isdB1PreyUponOracle           = "b3c12848-528d-4005-878f-4f950a07240a"
	isdB1BearType                 = "Creature — Bear"
	isdB1SpiritType               = "Creature — Spirit"
	isdB1WolfType                 = "Creature — Wolf"
	isdB1AuraType                 = "Enchantment — Aura"
	isdB1ThrabenPurebloodsOracle  = "cf8aef54-3f43-413c-9f3f-d4a1b2090904"
	isdB1ThrabenPurebloodsName    = "Thraben Purebloods"
	isdB1ThrabenPurebloodsType    = "Creature — Dog"
)

// isdB1Vanilla are the batch's bodies with no rules text.
var isdB1Vanilla = map[string]string{
	"075fd933-689c-4199-85bd-e7994d936145": "Fortress Crab",
	"c690f92f-f34f-4df0-9e01-d2ecf3b8f857": "Kindercatch",
	"228ee6b0-2012-4ef8-8aab-552feece801d": "Riot Devils",
	"be22faba-07e0-4036-b700-82769989250b": "Rotting Fensnake",
	isdB1ThrabenPurebloodsOracle:           isdB1ThrabenPurebloodsName,
	"fea95888-e16a-4209-9cd4-623f7f4d2f67": "Walking Corpse",
}

// isdB1Keyworded are the batch's creatures whose whole text is printed
// keywords.
var isdB1Keyworded = map[string]struct {
	name, typeLine string
	keywords       []string
}{
	isdB1AbbeyGriffinOracle:    {"Abbey Griffin", "Creature — Griffin", []string{"flying", "vigilance"}},
	isdB1ChapelGeistOracle:     {"Chapel Geist", isdB1SpiritType, []string{"flying"}},
	isdB1MoonHeronOracle:       {"Moon Heron", "Creature — Spirit Bird", []string{"flying"}},
	isdB1MarkovPatricianOracle: {"Markov Patrician", "Creature — Vampire", []string{"lifelink"}},
	isdB1VoicelessSpiritOracle: {"Voiceless Spirit", isdB1SpiritType, []string{"flying", "first strike"}},
	isdB1SpectralRiderOracle:   {"Spectral Rider", "Creature — Spirit Knight", []string{"intimidate"}},
	isdB1GraveBrambleOracle:    {"Grave Bramble", "Creature — Plant", []string{"defender", "protection from Zombies"}},
	isdB1EliteInquisitorOracle: {"Elite Inquisitor", "Creature — Human Soldier",
		[]string{"first strike", "vigilance", "protection from Vampires", "protection from Werewolves", "protection from Zombies"}},
}

func TestIsdB1EveryCardIsRegisteredFull(t *testing.T) {
	all := map[string]string{
		isdB1InvisibleStalkerOracle:   "Invisible Stalker",
		isdB1OrchardSpiritOracle:      "Orchard Spirit",
		isdB1BattlegroundGeistOracle:  "Battleground Geist",
		isdB1GallowsWardenOracle:      "Gallows Warden",
		isdB1ArmoredSkaabOracle:       "Armored Skaab",
		isdB1CrosswayVampireOracle:    "Crossway Vampire",
		isdB1DarkthicketWolfOracle:    "Darkthicket Wolf",
		isdB1FeralRidgewolfOracle:     "Feral Ridgewolf",
		isdB1KessigWolfOracle:         "Kessig Wolf",
		isdB1LanternSpiritOracle:      "Lantern Spirit",
		isdB1ManorSkeletonOracle:      "Manor Skeleton",
		isdB1MausoleumGuardOracle:     "Mausoleum Guard",
		isdB1PitchburnDevilsOracle:    "Pitchburn Devils",
		isdB1HollowhengeOracle:        "Hollowhenge Scavenger",
		isdB1StromkirkPatrolOracle:    "Stromkirk Patrol",
		isdB1SilverchaseFoxOracle:     "Silverchase Fox",
		isdB1GhoulcallersBellOracle:   "Ghoulcaller's Bell",
		isdB1CobbledWingsOracle:       "Cobbled Wings",
		isdB1SpectralFlightOracle:     "Spectral Flight",
		isdB1SensoryDeprivationOracle: "Sensory Deprivation",
		isdB1GruesomeDeformityOracle:  "Gruesome Deformity",
		isdB1BramblecrushOracle:       "Bramblecrush",
		isdB1SmiteTheMonstrousOracle:  "Smite the Monstrous",
		isdB1RebukeOracle:             "Rebuke",
		isdB1UrgentExorcismOracle:     "Urgent Exorcism",
		isdB1VictimOfNightOracle:      "Victim of Night",
		isdB1BrimstoneVolleyOracle:    "Brimstone Volley",
		isdB1BumpInTheNightOracle:     "Bump in the Night",
		isdB1DreamTwistOracle:         "Dream Twist",
		isdB1GeistflameOracle:         "Geistflame",
		isdB1FeelingOfDreadOracle:     "Feeling of Dread",
		isdB1RallyThePeasantsOracle:   "Rally the Peasants",
		isdB1RollingTemblorOracle:     "Rolling Temblor",
		isdB1PurifyTheGraveOracle:     "Purify the Grave",
		isdB1TravelPreparationsOracle: "Travel Preparations",
		isdB1PreyUponOracle:           "Prey Upon",
	}
	for oracle, name := range isdB1Vanilla {
		all[oracle] = name
	}
	for oracle, k := range isdB1Keyworded {
		all[oracle] = k.name
	}
	if len(all) != 50 {
		t.Fatalf("the batch lists %d cards, want 50", len(all))
	}
	for oracle, name := range all {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- fixtures -------------------------------------------------------------------

// isdB1Seats is the active player and the next seat.
func isdB1Seats(g *game.Game) (me, opp *game.Player) {
	a := g.Turn.ActiveSeat
	return g.Seats[a], g.Seats[(a+1)%len(g.Seats)]
}

// isdB1Reacher seeds a non-catalog creature with printed reach.
func isdB1Reacher(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Spider", TypeLine: "Creature — Spider",
		Power: 1, Toughness: 3, Owner: owner, Controller: owner, Keywords: []string{"reach"},
	})
}

// isdB1Flashback casts a card from the active player's graveyard for its
// flashback cost.
func isdB1Flashback(t *testing.T, g *game.Game, name, typeLine, oracle string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.SeedGraveyardCard(t, g, name, typeLine, oracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", Targets: targets,
	}); err != nil {
		t.Fatalf("flashback %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

// --- bodies and keywords ----------------------------------------------------------

func TestIsdB1VanillaBodiesDeclareNoAbilities(t *testing.T) {
	for oracle, name := range isdB1Vanilla {
		spec, _ := effects.Lookup(oracle)
		if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 ||
			len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 {
			t.Errorf("%s is a vanilla body and declares no abilities", name)
		}
	}
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, isdB1ThrabenPurebloodsName, isdB1ThrabenPurebloodsType, isdB1ThrabenPurebloodsOracle, 3, 5)
	if p, tough := effectstest.PtOf(t, g, id); p != 3 || tough != 5 {
		t.Errorf("Thraben Purebloods is %d/%d, want 3/5", p, tough)
	}
}

func TestIsdB1KeywordedCreaturesHaveTheirKeywords(t *testing.T) {
	for oracle, k := range isdB1Keyworded {
		t.Run(k.name, func(t *testing.T) {
			spec, _ := effects.Lookup(oracle)
			if !slices.Equal(spec.PrintedKeywords, k.keywords) {
				t.Errorf("printed keywords %v, want %v", spec.PrintedKeywords, k.keywords)
			}
			if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.Activated) != 0 {
				t.Errorf("%s declares more than its keywords", k.name)
			}
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, k.name, k.typeLine, oracle, 2, 2)
			for _, kw := range k.keywords {
				if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s on the battlefield", k.name, kw)
				}
			}
		})
	}
}

// Elite Inquisitor can't be blocked by a Werewolf (CR 702.16f); a Human
// blocks it.
func TestIsdB1EliteInquisitorIsProtectedFromWerewolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	inq := effectstest.B12Push(g, me.ID, "Elite Inquisitor", "Creature — Human Soldier", isdB1EliteInquisitorOracle, 2, 2)
	wolf := effectstest.B12Creature(g, opp.ID, "Werewolf", "Creature — Werewolf", 2, 2)
	if err := effectstest.PorBlock(t, g, inq, opp.ID, wolf); err == nil {
		t.Error("a Werewolf blocked Elite Inquisitor")
	}
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	inq = effectstest.B12Push(g, me.ID, "Elite Inquisitor", "Creature — Human Soldier", isdB1EliteInquisitorOracle, 2, 2)
	human := effectstest.B12Creature(g, opp.ID, "Peasant", "Creature — Human", 1, 1)
	if err := effectstest.PorBlock(t, g, inq, opp.ID, human); err != nil {
		t.Errorf("a Human couldn't block Elite Inquisitor: %v", err)
	}
}

// Defender: Grave Bramble can't attack (CR 702.3b).
func TestIsdB1GraveBrambleCantAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bramble := effectstest.B12Push(g, me.ID, "Grave Bramble", "Creature — Plant", isdB1GraveBrambleOracle, 3, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bramble, opp.ID); err == nil {
		t.Error("Grave Bramble attacked")
	}
}

// Intimidate: a green creature can't block the white Spectral Rider
// (CR 702.13b).
func TestIsdB1SpectralRiderHasIntimidate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rider := effectstest.B12Push(g, me.ID, "Spectral Rider", "Creature — Spirit Knight", isdB1SpectralRiderOracle, 2, 2)
	effectstest.HmlSetColors(g, rider, "W")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", isdB1BearType, 2, 2)
	effectstest.HmlSetColors(g, bear, "G")
	if err := effectstest.PorBlock(t, g, rider, opp.ID, bear); err == nil {
		t.Error("a green creature blocked a white creature with intimidate")
	}
}

// Invisible Stalker can't be blocked, and an opponent's spell can't
// target it while it can target an ordinary creature.
func TestIsdB1InvisibleStalkerIsUnblockableAndHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stalker := effectstest.B12Push(g, me.ID, "Invisible Stalker", "Creature — Human Rogue", isdB1InvisibleStalkerOracle, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, stalker, "hexproof") {
		t.Error("Invisible Stalker has no hexproof")
	}
	effectstest.AssertRestrictions(t, g, stalker, game.CantBeBlocked)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", isdB1BearType, 2, 2)
	flame := effectstest.PushCatalogHandCard(opp, "Geistflame", "Instant", isdB1GeistflameOracle)
	if err := g.CastSpell(opp.ID, flame, game.CastSpellParams{Targets: effectstest.CardTarget(stalker)}); err == nil {
		t.Fatal("an opponent's spell targeted Invisible Stalker")
	}
	if err := g.CastSpell(opp.ID, flame, game.CastSpellParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Errorf("the same spell couldn't target an ordinary creature: %v", err)
	}

	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	stalker = effectstest.B12Push(g, me.ID, "Invisible Stalker", "Creature — Human Rogue", isdB1InvisibleStalkerOracle, 1, 1)
	blocker := effectstest.B12Creature(g, opp.ID, "Bear", isdB1BearType, 2, 2)
	if err := effectstest.PorBlock(t, g, stalker, opp.ID, blocker); err == nil {
		t.Error("Invisible Stalker was blocked")
	}
}

// Orchard Spirit: a ground creature can't block it; reach and flying can.
func TestIsdB1OrchardSpiritIsBlockedOnlyByFlyingOrReach(t *testing.T) {
	for _, tc := range []struct {
		name    string
		blocker func(g *game.Game, owner uuid.UUID) uuid.UUID
		ok      bool
	}{
		{"ground", func(g *game.Game, owner uuid.UUID) uuid.UUID {
			return effectstest.B12Creature(g, owner, "Bear", isdB1BearType, 2, 2)
		}, false},
		{"reach", isdB1Reacher, true},
		{"flying", func(g *game.Game, owner uuid.UUID) uuid.UUID { return effectstest.PorFlyer(g, owner, "Bird") }, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			spirit := effectstest.B12Push(g, me.ID, "Orchard Spirit", isdB1SpiritType, isdB1OrchardSpiritOracle, 2, 2)
			blocker := tc.blocker(g, opp.ID)
			err := effectstest.PorBlock(t, g, spirit, opp.ID, blocker)
			if tc.ok && err != nil {
				t.Errorf("a %s creature couldn't block Orchard Spirit: %v", tc.name, err)
			}
			if !tc.ok && err == nil {
				t.Errorf("a %s creature blocked Orchard Spirit", tc.name)
			}
		})
	}
}

// Battleground Geist and Gallows Warden pump only the OTHER Spirits their
// controller controls.
func TestIsdB1SpiritLordsPumpYourOtherSpirits(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
	}{
		{"Battleground Geist", isdB1BattlegroundGeistOracle, 2, 1},
		{"Gallows Warden", isdB1GallowsWardenOracle, 1, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			lord := effectstest.B12Push(g, me.ID, tc.name, isdB1SpiritType, tc.oracle, 3, 3)
			mine := effectstest.B12Creature(g, me.ID, "My Spirit", isdB1SpiritType, 1, 1)
			bear := effectstest.B12Creature(g, me.ID, "My Bear", isdB1BearType, 1, 1)
			theirs := effectstest.B12Creature(g, opp.ID, "Their Spirit", isdB1SpiritType, 1, 1)
			if p, tt := effectstest.PtOf(t, g, mine); p != tc.power || tt != tc.tough {
				t.Errorf("my Spirit is %d/%d, want %d/%d", p, tt, tc.power, tc.tough)
			}
			for _, id := range []uuid.UUID{bear, theirs} {
				if p, tt := effectstest.PtOf(t, g, id); p != 1 || tt != 1 {
					t.Errorf("an unpumped creature is %d/%d, want 1/1", p, tt)
				}
			}
			if p, tt := effectstest.PtOf(t, g, lord); p != 3 || tt != 3 {
				t.Errorf("%s pumped itself to %d/%d", tc.name, p, tt)
			}
			if !effectstest.HasEffectiveKeyword(t, g, lord, "flying") {
				t.Errorf("%s has no flying", tc.name)
			}
		})
	}
}

// --- creatures with abilities -------------------------------------------------------

func TestIsdB1ArmoredSkaabMillsFourOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := isdB1Seats(g)
	effectstest.AdvanceToMain(t, g)
	lib, yard := me.Library.Size(), me.Graveyard.Size()
	effectstest.CastHoldCreature(t, g, "Armored Skaab", "Creature — Zombie Warrior", isdB1ArmoredSkaabOracle, 1, 4)
	if got := lib - me.Library.Size(); got != 4 {
		t.Errorf("library lost %d cards, want 4", got)
	}
	if got := me.Graveyard.Size() - yard; got != 4 {
		t.Errorf("graveyard gained %d cards, want 4", got)
	}
}

func TestIsdB1CrosswayVampireStopsATargetFromBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := isdB1Seats(g)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	other := effectstest.B12Creature(g, opp.ID, "Other Wall", "Creature — Wall", 0, 4)
	effectstest.CastHoldCreature(t, g, "Crossway Vampire", "Creature — Vampire", isdB1CrosswayVampireOracle, 3, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, wall)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, wall, game.CantBlock)
	effectstest.AssertRestrictions(t, g, other, 0)
}

func TestIsdB1DarkthicketWolfPumpsOnlyOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := isdB1Seats(g)
	wolf := effectstest.B12Push(g, me.ID, "Darkthicket Wolf", isdB1WolfType, isdB1DarkthicketWolfOracle, 2, 2)
	effectstest.ApaMana(me, "G", "C", "C")
	effectstest.ApaActivate(t, g, me, wolf, 0, game.ActivateAbilityParams{})
	effectstest.ApaMana(me, "G", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, wolf, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Errorf("second activation this turn: err = %v, want ErrConditionNotMet", err)
	}
	if p, tough := effectstest.PtOf(t, g, wolf); p != 4 || tough != 4 {
		t.Errorf("Darkthicket Wolf is %d/%d, want 4/4", p, tough)
	}
}

func TestIsdB1FeralRidgewolfPumpsRepeatedly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := isdB1Seats(g)
	wolf := effectstest.B12Push(g, me.ID, "Feral Ridgewolf", isdB1WolfType, isdB1FeralRidgewolfOracle, 1, 2)
	for i := 0; i < 2; i++ {
		effectstest.ApaMana(me, "R", "C")
		effectstest.ApaActivate(t, g, me, wolf, 0, game.ActivateAbilityParams{})
	}
	if p, tough := effectstest.PtOf(t, g, wolf); p != 5 || tough != 2 {
		t.Errorf("Feral Ridgewolf is %d/%d, want 5/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, wolf, "trample") {
		t.Error("Feral Ridgewolf has no trample")
	}
}

func TestIsdB1KessigWolfGainsFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := isdB1Seats(g)
	wolf := effectstest.B12Push(g, me.ID, "Kessig Wolf", isdB1WolfType, isdB1KessigWolfOracle, 3, 1)
	if effectstest.HasEffectiveKeyword(t, g, wolf, "first strike") {
		t.Fatal("Kessig Wolf has first strike before activating")
	}
	effectstest.ApaMana(me, "R", "C")
	effectstest.ApaActivate(t, g, me, wolf, 0, game.ActivateAbilityParams{})
	if !effectstest.HasEffectiveKeyword(t, g, wolf, "first strike") {
		t.Error("Kessig Wolf has no first strike after activating")
	}
}

func TestIsdB1LanternSpiritReturnsToItsOwnersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := isdB1Seats(g)
	spirit := effectstest.B12Push(g, me.ID, "Lantern Spirit", isdB1SpiritType, isdB1LanternSpiritOracle, 2, 1)
	if !effectstest.HasEffectiveKeyword(t, g, spirit, "flying") {
		t.Error("Lantern Spirit has no flying")
	}
	effectstest.ApaMana(me, "U")
	effectstest.ApaActivate(t, g, me, spirit, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(spirit) || !me.Hand.Contains(spirit) {
		t.Error("Lantern Spirit is not back in its owner's hand")
	}
}

func TestIsdB1ManorSkeletonRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := isdB1Seats(g)
	skeleton := effectstest.B12Push(g, me.ID, "Manor Skeleton", "Creature — Skeleton", isdB1ManorSkeletonOracle, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, skeleton, "haste") {
		t.Error("Manor Skeleton has no haste")
	}
	effectstest.ApaMana(me, "C", "B")
	effectstest.ApaActivate(t, g, me, skeleton, 0, game.ActivateAbilityParams{})
	effectstest.B25Destroy(g, skeleton)
	if !g.Battlefield.Contains(skeleton) {
		t.Fatal("the shielded Manor Skeleton was destroyed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B25Destroy(g, skeleton)
	if g.Battlefield.Contains(skeleton) {
		t.Error("the shield was not used up: Manor Skeleton survived a second destroy")
	}
}

func TestIsdB1MausoleumGuardLeavesTwoFlyingSpirits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := isdB1Seats(g)
	guard := effectstest.B12Push(g, me.ID, "Mausoleum Guard", "Creature — Human Scout", isdB1MausoleumGuardOracle, 2, 2)
	effectstest.B25Destroy(g, guard)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Spirit"); got != 2 {
		t.Fatalf("%d Spirit tokens, want 2", got)
	}
	var spirits []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && c.Controller == me.ID {
			spirits = append(spirits, c.InstanceID)
		}
	}
	for _, id := range spirits {
		if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
			t.Error("a Spirit token has no flying")
		}
	}
}

func TestIsdB1PitchburnDevilsDealsThreeWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := isdB1Seats(g)
	devils := effectstest.B12Push(g, me.ID, "Pitchburn Devils", "Creature — Devil", isdB1PitchburnDevilsOracle, 3, 3)
	life := opp.Life
	effectstest.B25Destroy(g, devils)
	effectstest.Vow1Pick(t, g, me, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want -3", life, opp.Life)
	}
}

func TestIsdB1HollowhengeScavengerGainsFiveOnlyAfterADeath(t *testing.T) {
	for _, died := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := isdB1Seats(g)
		bear := effectstest.B12Creature(g, opp.ID, "Their Bear", isdB1BearType, 2, 2)
		if died {
			effectstest.B25Destroy(g, bear)
			g.RunStateChecksForTest()
		}
		life := me.Life
		effectstest.CastHoldCreature(t, g, "Hollowhenge Scavenger", "Creature — Elemental", isdB1HollowhengeOracle, 4, 5)
		effectstest.PassPriorityAroundTable(t, g)
		want := life
		if died {
			want = life + 5
		}
		if me.Life != want {
			t.Errorf("died=%v: life %d -> %d, want %d", died, life, me.Life, want)
		}
	}
}

func TestIsdB1StromkirkPatrolGrowsOnCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := isdB1Seats(g)
	patrol := effectstest.B12Push(g, me.ID, "Stromkirk Patrol", "Creature — Vampire Soldier", isdB1StromkirkPatrolOracle, 4, 3)
	effectstest.DealCombatDamageToPlayer(g, patrol, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, patrol); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

func TestIsdB1SilverchaseFoxExilesAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := isdB1Seats(g)
	fox := effectstest.B12Push(g, me.ID, "Silverchase Fox", "Creature — Fox", isdB1SilverchaseFoxOracle, 2, 2)
	ench := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	effectstest.ApaMana(me, "W", "C")
	effectstest.ApaActivate(t, g, me, fox, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(ench)})
	if !effectstest.InExile(g, ench) {
		t.Error("the enchantment was not exiled")
	}
	if g.Battlefield.Contains(fox) || !me.Graveyard.Contains(fox) {
		t.Error("Silverchase Fox was not sacrificed")
	}
}

// --- artifacts and Auras ----------------------------------------------------------

func TestIsdB1GhoulcallersBellMillsEachPlayerOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := isdB1Seats(g)
	bell := effectstest.B12Push(g, me.ID, "Ghoulcaller's Bell", "Artifact", isdB1GhoulcallersBellOracle, 0, 0)
	libs := make([]int, len(g.Seats))
	yards := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		libs[i], yards[i] = p.Library.Size(), p.Graveyard.Size()
	}
	effectstest.ApaActivate(t, g, me, bell, 0, game.ActivateAbilityParams{})
	for i, p := range g.Seats {
		if libs[i]-p.Library.Size() != 1 || p.Graveyard.Size()-yards[i] != 1 {
			t.Errorf("seat %d: library %d -> %d, graveyard %d -> %d, want one card milled",
				i, libs[i], p.Library.Size(), yards[i], p.Graveyard.Size())
		}
	}
	if !effectstest.AlliTapped(t, g, bell) {
		t.Error("the Bell did not tap")
	}
}

func TestIsdB1CobbledWingsGrantsFlying(t *testing.T) {
	spec, _ := effects.Lookup(isdB1CobbledWingsOracle)
	if len(spec.Activated) != 1 {
		t.Fatalf("Cobbled Wings has %d activated rows, want the equip row", len(spec.Activated))
	}
	g := effectstest.NewCatalogGame(t)
	me, _ := isdB1Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", isdB1BearType, 2, 2)
	wings := effectstest.B12Push(g, me.ID, "Cobbled Wings", "Artifact — Equipment", isdB1CobbledWingsOracle, 0, 0)
	effectstest.BrAttach(t, g, wings, bear)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the equipped creature has no flying")
	}
}

func TestIsdB1SpectralFlightIsCastOntoACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := isdB1Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", isdB1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Spectral Flight", isdB1AuraType, isdB1SpectralFlightOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the enchanted creature is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the enchanted creature has no flying")
	}
}

func TestIsdB1SensoryDeprivationAndGruesomeDeformity(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := isdB1Seats(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	deprivation := effectstest.B12Push(g, me.ID, "Sensory Deprivation", isdB1AuraType, isdB1SensoryDeprivationOracle, 0, 0)
	effectstest.BrAttach(t, g, deprivation, theirs)
	if p, tough := effectstest.PtOf(t, g, theirs); p != 1 || tough != 4 {
		t.Errorf("the deprived creature is %d/%d, want 1/4", p, tough)
	}
	mine := effectstest.B12Creature(g, me.ID, "My Bear", isdB1BearType, 2, 2)
	deformity := effectstest.B12Push(g, me.ID, "Gruesome Deformity", isdB1AuraType, isdB1GruesomeDeformityOracle, 0, 0)
	effectstest.BrAttach(t, g, deformity, mine)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "intimidate") {
		t.Error("the deformed creature has no intimidate")
	}
}

// --- destroy spells -----------------------------------------------------------------

func TestIsdB1BramblecrushDestroysOnlyANoncreaturePermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := isdB1Seats(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", isdB1BearType, 2, 2)
	ench := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	if err := effectstest.CastCatalogSpellErr(t, g, "Bramblecrush", "Sorcery", isdB1BramblecrushOracle, effectstest.CardRefs(bear)); err == nil {
		t.Fatal("Bramblecrush targeted a creature")
	}
	effectstest.CastCatalogSpell(t, g, "Bramblecrush", "Sorcery", isdB1BramblecrushOracle, effectstest.CardRefs(ench))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) || !opp.Graveyard.Contains(ench) {
		t.Error("the enchantment was not destroyed")
	}
}

func TestIsdB1SmiteTheMonstrousNeedsPowerFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := isdB1Seats(g)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Ogre", 3, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 4, 4)
	if err := effectstest.CastCatalogSpellErr(t, g, "Smite the Monstrous", "Instant", isdB1SmiteTheMonstrousOracle, effectstest.CardRefs(small)); err == nil {
		t.Fatal("Smite the Monstrous targeted a 3-power creature")
	}
	effectstest.CastCatalogSpell(t, g, "Smite the Monstrous", "Instant", isdB1SmiteTheMonstrousOracle, effectstest.CardRefs(big))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the 4-power creature was not destroyed")
	}
}

func TestIsdB1RebukeDestroysAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := isdB1Seats(g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacking Ogre", "Creature — Ogre", 3, 3)
	idle := effectstest.B12Creature(g, me.ID, "Idle Bear", isdB1BearType, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	rebuke := effectstest.PushCatalogHandCard(opp, "Rebuke", "Instant", isdB1RebukeOracle)
	if err := g.CastSpell(opp.ID, rebuke, game.CastSpellParams{Targets: effectstest.CardTarget(idle)}); err == nil {
		t.Fatal("a creature that is not attacking was accepted as the target")
	}
	if err := g.CastSpell(opp.ID, rebuke, game.CastSpellParams{Targets: effectstest.CardTarget(attacker)}); err != nil {
		t.Fatalf("cast Rebuke at the attacker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacking creature was not destroyed")
	}
	if !g.Battlefield.Contains(idle) {
		t.Error("the idle creature left the battlefield")
	}
}

func TestIsdB1UrgentExorcismDestroysASpiritOrAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := isdB1Seats(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", isdB1BearType, 2, 2)
	spirit := effectstest.B12Creature(g, opp.ID, "Their Spirit", isdB1SpiritType, 1, 1)
	ench := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	if err := effectstest.CastCatalogSpellErr(t, g, "Urgent Exorcism", "Instant", isdB1UrgentExorcismOracle, effectstest.CardRefs(bear)); err == nil {
		t.Fatal("Urgent Exorcism targeted a creature that is not a Spirit")
	}
	effectstest.CastCatalogSpell(t, g, "Urgent Exorcism", "Instant", isdB1UrgentExorcismOracle, effectstest.CardRefs(spirit))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Urgent Exorcism", "Instant", isdB1UrgentExorcismOracle, effectstest.CardRefs(ench))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(spirit) {
		t.Error("the Spirit was not destroyed")
	}
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment was not destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear left the battlefield")
	}
}

func TestIsdB1VictimOfNightSparesTheMonsters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := isdB1Seats(g)
	for _, typeLine := range []string{"Creature — Vampire", "Creature — Human Werewolf", "Creature — Zombie"} {
		monster := effectstest.B12Creature(g, opp.ID, "Monster", typeLine, 2, 2)
		if err := effectstest.CastCatalogSpellErr(t, g, "Victim of Night", "Instant", isdB1VictimOfNightOracle, effectstest.CardRefs(monster)); err == nil {
			t.Errorf("Victim of Night targeted a %s", typeLine)
		}
	}
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", isdB1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Victim of Night", "Instant", isdB1VictimOfNightOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not destroyed")
	}
}

// --- damage, life and mill ----------------------------------------------------------

func TestIsdB1BrimstoneVolleyDealsFiveAfterADeath(t *testing.T) {
	for _, died := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		_, opp := isdB1Seats(g)
		bear := effectstest.B12Creature(g, opp.ID, "Their Bear", isdB1BearType, 2, 2)
		if died {
			effectstest.B25Destroy(g, bear)
			g.RunStateChecksForTest()
		}
		life := opp.Life
		effectstest.CastCatalogSpell(t, g, "Brimstone Volley", "Instant", isdB1BrimstoneVolleyOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
		effectstest.PassPriorityAroundTable(t, g)
		want := 3
		if died {
			want = 5
		}
		if got := life - opp.Life; got != want {
			t.Errorf("died=%v: dealt %d, want %d", died, got, want)
		}
	}
}

func TestIsdB1BumpInTheNightTargetsAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := isdB1Seats(g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Bump in the Night", "Sorcery", isdB1BumpInTheNightOracle, []game.TargetRef{effectstest.PlayerTarget(me.ID)}); err == nil {
		t.Fatal("Bump in the Night targeted its caster")
	}
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Bump in the Night", "Sorcery", isdB1BumpInTheNightOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want -3", life, opp.Life)
	}
}

func TestIsdB1DreamTwistMillsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := isdB1Seats(g)
	lib, yard := opp.Library.Size(), opp.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Dream Twist", "Instant", isdB1DreamTwistOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if lib-opp.Library.Size() != 3 || opp.Graveyard.Size()-yard != 3 {
		t.Errorf("library %d -> %d, graveyard %d -> %d, want three milled", lib, opp.Library.Size(), yard, opp.Graveyard.Size())
	}
}

func TestIsdB1GeistflameDealsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := isdB1Seats(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", isdB1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Geistflame", "Instant", isdB1GeistflameOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, bear); got != 1 {
		t.Errorf("the Bear has %d damage, want 1", got)
	}
}

func TestIsdB1FeelingOfDreadTapsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := isdB1Seats(g)
	a := effectstest.B12Creature(g, opp.ID, "Bear A", isdB1BearType, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", isdB1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Feeling of Dread", "Instant", isdB1FeelingOfDreadOracle, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AlliTapped(t, g, a) || !effectstest.AlliTapped(t, g, b) {
		t.Error("both targets should be tapped")
	}
}

func TestIsdB1RallyThePeasantsPumpsOnlyYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := isdB1Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", isdB1BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", isdB1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Rally the Peasants", "Instant", isdB1RallyThePeasantsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, mine); p != 4 || tough != 2 {
		t.Errorf("my Bear is %d/%d, want 4/2", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their Bear is %d/%d, want 2/2", p, tough)
	}
}

func TestIsdB1RollingTemblorSparesFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := isdB1Seats(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", isdB1BearType, 2, 2)
	bird := effectstest.PorFlyer(g, opp.ID, "Their Bird")
	effectstest.CastCatalogSpell(t, g, "Rolling Temblor", "Sorcery", isdB1RollingTemblorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(bear) {
		t.Error("the ground creature survived 2 damage")
	}
	if !g.Battlefield.Contains(bird) || effectstest.DamageOn(t, g, bird) != 0 {
		t.Error("the flyer was damaged")
	}
}

func TestIsdB1PurifyTheGraveExilesAGraveyardCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := isdB1Seats(g)
	card := effectstest.PushGraveyardCardTyped(opp, "Their Card", "Sorcery")
	effectstest.CastCatalogSpell(t, g, "Purify the Grave", "Instant", isdB1PurifyTheGraveOracle, effectstest.CardRefs(card))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, card) || opp.Graveyard.Contains(card) {
		t.Error("the graveyard card was not exiled")
	}
}

func TestIsdB1TravelPreparationsCountersTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := isdB1Seats(g)
	a := effectstest.B12Creature(g, me.ID, "Bear A", isdB1BearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", isdB1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Travel Preparations", "Sorcery", isdB1TravelPreparationsOracle, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOneCounters(g, a) != 1 || effectstest.PlusOneCounters(g, b) != 1 {
		t.Errorf("counters %d and %d, want one each", effectstest.PlusOneCounters(g, a), effectstest.PlusOneCounters(g, b))
	}
}

func TestIsdB1PreyUponFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := isdB1Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", isdB1BearType, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Prey Upon", "Sorcery", isdB1PreyUponOracle, effectstest.CardRefs(theirs, mine)); err == nil {
		t.Fatal("Prey Upon accepted the creatures in the wrong slots")
	}
	effectstest.CastCatalogSpell(t, g, "Prey Upon", "Sorcery", isdB1PreyUponOracle, effectstest.CardRefs(mine, theirs))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(theirs) {
		t.Error("their Bear survived the fight")
	}
	if got := effectstest.DamageOn(t, g, mine); got != 2 {
		t.Errorf("my Ogre has %d damage, want 2", got)
	}
}

// --- flashback ------------------------------------------------------------------------

// Every flashback card in the batch casts from its owner's graveyard and
// is exiled as it resolves (CR 702.34a).
func TestIsdB1FlashbackCardsCastFromTheGraveyardAndAreExiled(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		targets                func(g *game.Game, me, opp *game.Player) []game.TargetRef
	}{
		{"Bump in the Night", "Sorcery", isdB1BumpInTheNightOracle, func(_ *game.Game, _, opp *game.Player) []game.TargetRef {
			return []game.TargetRef{effectstest.PlayerTarget(opp.ID)}
		}},
		{"Dream Twist", "Instant", isdB1DreamTwistOracle, func(_ *game.Game, _, opp *game.Player) []game.TargetRef {
			return []game.TargetRef{effectstest.PlayerTarget(opp.ID)}
		}},
		{"Geistflame", "Instant", isdB1GeistflameOracle, func(_ *game.Game, _, opp *game.Player) []game.TargetRef {
			return []game.TargetRef{effectstest.PlayerTarget(opp.ID)}
		}},
		{"Feeling of Dread", "Instant", isdB1FeelingOfDreadOracle, func(g *game.Game, _, opp *game.Player) []game.TargetRef {
			return effectstest.CardRefs(effectstest.B12Creature(g, opp.ID, "Their Bear", isdB1BearType, 2, 2))
		}},
		{"Rally the Peasants", "Instant", isdB1RallyThePeasantsOracle, nil},
		{"Rolling Temblor", "Sorcery", isdB1RollingTemblorOracle, nil},
		{"Purify the Grave", "Instant", isdB1PurifyTheGraveOracle, func(_ *game.Game, _, opp *game.Player) []game.TargetRef {
			return effectstest.CardRefs(effectstest.PushGraveyardCardTyped(opp, "Their Card", "Sorcery"))
		}},
		{"Travel Preparations", "Sorcery", isdB1TravelPreparationsOracle, func(g *game.Game, me, _ *game.Player) []game.TargetRef {
			return effectstest.CardRefs(effectstest.B12Creature(g, me.ID, "My Bear", isdB1BearType, 2, 2))
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spec, _ := effects.Lookup(tc.oracle)
			if !slices.Contains(spec.CastableZones, game.ZoneGraveyard) || len(spec.AlternativeCosts) != 1 {
				t.Fatalf("%s declares no flashback", tc.name)
			}
			g := effectstest.NewCatalogGame(t)
			me, opp := isdB1Seats(g)
			var targets []game.TargetRef
			if tc.targets != nil {
				targets = tc.targets(g, me, opp)
			}
			id := isdB1Flashback(t, g, tc.name, tc.typeLine, tc.oracle, targets)
			if !effectstest.InExile(g, id) || me.Graveyard.Contains(id) {
				t.Errorf("%s was not exiled after its flashback", tc.name)
			}
		})
	}
}
