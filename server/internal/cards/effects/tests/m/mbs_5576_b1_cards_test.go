package testsm

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mbs_5576_b1_cards_test.go — Mirrodin Besieged, batch 1 (#5576). Each
// test plays the real catalog entry through the engine and asserts what
// its text says.

const (
	mbsB1BlightwidowOracle    = "282003f4-862a-4f80-8f45-e470fc0fe2fe"
	mbsB1FlensermiteOracle    = "4b63bf36-cd0e-41d8-b0ca-aac459d91f3f"
	mbsB1GlissasCourierOracle = "25fb2a67-fd7f-4a06-bd59-118d82c86d6e"
	mbsB1KothsCourierOracle   = "59f57c25-ebb7-49df-8cdd-c6f069d1358d"
	mbsB1GargoyleOracle       = "1070ede7-c81d-4c77-a1a9-6f243432a640"
	mbsB1MirranCrusaderOracle = "fe69d9bd-2a60-4b33-a0d8-1ca18c2b6705"
	mbsB1PhyCrusaderOracle    = "91760bdf-a82a-4f3d-b925-3bc81ffbd41f"
	mbsB1DigesterOracle       = "9e7b03ae-dcf5-4892-9000-aee7e2cd2fed"
	mbsB1PriestsOracle        = "793fe31a-2b11-49f8-9851-719469ad1726"
	mbsB1ScourgeServantOracle = "561d514d-060e-491f-a703-8a7f9da5204d"
	mbsB1TangleMantisOracle   = "d52397bf-e3fb-4894-996d-7c1cc8d92603"
	mbsB1TineShrikeOracle     = "e20ad2f7-cb41-4e6e-8f3b-bd260f4a2ecf"
	mbsB1HexplateGolemOracle  = "bacb9d63-5950-4bcd-a056-40c02ecf4ce8"
	mbsB1OgreResisterOracle   = "d626bd7f-c993-49ff-b05b-262330d8540c"
	mbsB1SlagwurmOracle       = "5114f3e0-339d-46d6-9d44-97fa0afbd7c5"
	mbsB1AccorderOracle       = "6891c8b7-7112-40dc-8b94-b84ad1885e5e"
	mbsB1WardriverOracle      = "10849bbb-50ca-4194-a7cb-fa872272b29b"
	mbsB1PartisanOracle       = "9d2bfbb2-84df-4d7c-9989-7f5577f90d97"
	mbsB1RingleaderOracle     = "dff72513-b67a-4732-a8fc-269cbf21ae99"
	mbsB1JuggernautOracle     = "009fcd1e-5c5f-435b-afc1-099f1622e45c"
	mbsB1ArdentRecruitOracle  = "9275108c-7eaf-4edd-aa6d-e5263d5ac7fd"
	mbsB1RhinoOracle          = "834efbf7-7b93-42ac-ab20-0b7b8ca58a16"
	mbsB1SignalPestOracle     = "724f3d2f-0513-4efd-961f-56a7ea7b5731"
	mbsB1BladedSentinelOracle = "00f579f4-2d18-4c7c-91b2-34726b384535"
	mbsB1GustSkimmerOracle    = "3f49dfa6-d0d4-4b58-8417-ed11c062e220"
	mbsB1DrossRipperOracle    = "bf3b7489-e9d6-4b3c-9fb8-3bb727584db9"
	mbsB1TangleHulkOracle     = "f6d0360c-ebc6-48b5-80ad-d52dadcdd5f5"
	mbsB1FleshEaterImpOracle  = "b1aae085-2dca-425c-911a-19a89a0a4993"
	mbsB1BlisterstickOracle   = "6f76fec2-6fd0-4d33-b39a-f1684922d478"
	mbsB1PeaceStriderOracle   = "b0e96742-77c0-4de5-84b0-0d59a8bee3e0"
	mbsB1PierceStriderOracle  = "e2bbf963-794d-4600-bb92-cb9d1d00205a"
	mbsB1CausticHoundOracle   = "032cc574-f7e7-4753-9a30-76cbd45d9563"
	mbsB1CoreProwlerOracle    = "0ad44727-41c3-471c-b038-1bf43c5f296a"
	mbsB1OculusOracle         = "e793ede9-044a-44e9-bc9c-99b0034503eb"
	mbsB1NeurokCommandoOracle = "48936751-429b-4c40-88f5-72acde6278cc"
	mbsB1CopperCarapaceOracle = "8ffbff60-a59d-4f71-864e-3c16cb61e43b"
	mbsB1ViridianClawOracle   = "11ba27f4-decf-4447-9808-0d58844655cf"
	mbsB1FlayerHuskOracle     = "5cfdfcac-57fb-4507-a97a-cccbac2d1327"
	mbsB1SkinwingOracle       = "100c877d-4cce-4b96-8f21-a4f8a43e7d84"
	mbsB1StrandwalkerOracle   = "55a4b186-c6d4-4b08-bf45-481429693948"
	mbsB1CorrosionOracle      = "33eefb22-51a6-4630-9bd2-0895997f6f33"
	mbsB1CrushOracle          = "ad4dda1e-2236-45bf-b2f3-f63e790d837d"
	mbsB1IntoTheCoreOracle    = "ad08ad3a-bc58-43da-b7f3-4a478132da8a"
	mbsB1GeyserOracle         = "6ee7b00b-b875-48e2-92f2-889c154b0f82"
	mbsB1MorbidPlunderOracle  = "1bb7009b-aeb0-4545-be4c-c8c5d723bcf5"
	mbsB1TurnTheTideOracle    = "2546db90-1f6e-43f7-a116-2fefe67d489d"
	mbsB1PredationOracle      = "120446cc-f55b-40db-915d-2669c1083e56"
	mbsB1MastersCallOracle    = "d6c9c1f0-050f-40e2-ab6c-6566d0af418c"
	mbsB1BanishmentOracle     = "567b676b-54f3-44c7-8654-4d2b0262ea12"
	mbsB1SicknessOracle       = "142af8f1-6d6b-4043-8cc4-6632481972c1"

	mbsB1BearType = "Creature — Bear"
)

func TestMbsB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mbsB1BlightwidowOracle:    "Blightwidow",
		mbsB1FlensermiteOracle:    "Flensermite",
		mbsB1GlissasCourierOracle: "Glissa's Courier",
		mbsB1KothsCourierOracle:   "Koth's Courier",
		mbsB1GargoyleOracle:       "Lumengrid Gargoyle",
		mbsB1MirranCrusaderOracle: "Mirran Crusader",
		mbsB1PhyCrusaderOracle:    "Phyrexian Crusader",
		mbsB1DigesterOracle:       "Phyrexian Digester",
		mbsB1PriestsOracle:        "Priests of Norn",
		mbsB1ScourgeServantOracle: "Scourge Servant",
		mbsB1TangleMantisOracle:   "Tangle Mantis",
		mbsB1TineShrikeOracle:     "Tine Shrike",
		mbsB1HexplateGolemOracle:  "Hexplate Golem",
		mbsB1OgreResisterOracle:   "Ogre Resister",
		mbsB1SlagwurmOracle:       "Quilled Slagwurm",
		mbsB1AccorderOracle:       "Accorder Paladin",
		mbsB1WardriverOracle:      "Goblin Wardriver",
		mbsB1PartisanOracle:       "Loxodon Partisan",
		mbsB1RingleaderOracle:     "Kuldotha Ringleader",
		mbsB1JuggernautOracle:     "Phyrexian Juggernaut",
		mbsB1ArdentRecruitOracle:  "Ardent Recruit",
		mbsB1RhinoOracle:          "Razorfield Rhino",
		mbsB1SignalPestOracle:     "Signal Pest",
		mbsB1BladedSentinelOracle: "Bladed Sentinel",
		mbsB1GustSkimmerOracle:    "Gust-Skimmer",
		mbsB1DrossRipperOracle:    "Dross Ripper",
		mbsB1TangleHulkOracle:     "Tangle Hulk",
		mbsB1FleshEaterImpOracle:  "Flesh-Eater Imp",
		mbsB1BlisterstickOracle:   "Blisterstick Shaman",
		mbsB1PeaceStriderOracle:   "Peace Strider",
		mbsB1PierceStriderOracle:  "Pierce Strider",
		mbsB1CausticHoundOracle:   "Caustic Hound",
		mbsB1CoreProwlerOracle:    "Core Prowler",
		mbsB1OculusOracle:         "Oculus",
		mbsB1NeurokCommandoOracle: "Neurok Commando",
		mbsB1CopperCarapaceOracle: "Copper Carapace",
		mbsB1ViridianClawOracle:   "Viridian Claw",
		mbsB1FlayerHuskOracle:     "Flayer Husk",
		mbsB1SkinwingOracle:       "Skinwing",
		mbsB1StrandwalkerOracle:   "Strandwalker",
		mbsB1CorrosionOracle:      "Creeping Corrosion",
		mbsB1CrushOracle:          "Crush",
		mbsB1IntoTheCoreOracle:    "Into the Core",
		mbsB1GeyserOracle:         "Quicksilver Geyser",
		mbsB1MorbidPlunderOracle:  "Morbid Plunder",
		mbsB1TurnTheTideOracle:    "Turn the Tide",
		mbsB1PredationOracle:      "Unnatural Predation",
		mbsB1MastersCallOracle:    "Master's Call",
		mbsB1BanishmentOracle:     "Banishment Decree",
		mbsB1SicknessOracle:       "Spread the Sickness",
	} {
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

// --- printed keywords ------------------------------------------------------------------

func TestMbsB1CardsHaveTheirPrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Blightwidow", "Creature — Phyrexian Spider", mbsB1BlightwidowOracle, []string{"reach", "infect"}},
		{"Flensermite", "Creature — Phyrexian Gremlin", mbsB1FlensermiteOracle, []string{"infect", "lifelink"}},
		{"Glissa's Courier", "Creature — Phyrexian Horror", mbsB1GlissasCourierOracle, []string{"mountainwalk"}},
		{"Koth's Courier", "Creature — Human Rogue", mbsB1KothsCourierOracle, []string{"forestwalk"}},
		{"Lumengrid Gargoyle", "Artifact Creature — Gargoyle", mbsB1GargoyleOracle, []string{"flying"}},
		{"Mirran Crusader", "Creature — Human Knight", mbsB1MirranCrusaderOracle, []string{"double strike", "protection from black", "protection from green"}},
		{"Phyrexian Crusader", "Creature — Phyrexian Zombie Knight", mbsB1PhyCrusaderOracle, []string{"first strike", "protection from red", "protection from white", "infect"}},
		{"Phyrexian Digester", "Artifact Creature — Phyrexian Construct", mbsB1DigesterOracle, []string{"infect"}},
		{"Priests of Norn", "Creature — Phyrexian Cleric", mbsB1PriestsOracle, []string{"vigilance", "infect"}},
		{"Scourge Servant", "Creature — Phyrexian Zombie", mbsB1ScourgeServantOracle, []string{"infect"}},
		{"Tangle Mantis", "Creature — Insect", mbsB1TangleMantisOracle, []string{"trample"}},
		{"Tine Shrike", "Creature — Phyrexian Bird", mbsB1TineShrikeOracle, []string{"flying", "infect"}},
		{"Phyrexian Juggernaut", "Artifact Creature — Phyrexian Juggernaut", mbsB1JuggernautOracle, []string{"infect"}},
		{"Core Prowler", "Artifact Creature — Phyrexian Horror", mbsB1CoreProwlerOracle, []string{"infect"}},
		{"Flesh-Eater Imp", "Creature — Phyrexian Imp", mbsB1FleshEaterImpOracle, []string{"flying", "infect"}},
		{"Neurok Commando", "Creature — Human Rogue", mbsB1NeurokCommandoOracle, []string{"shroud"}},
	} {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
		for _, kw := range tc.keywords {
			if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s lacks %s", tc.name, kw)
			}
		}
	}
}

// Landwalk is enforced at the block declaration (CR 702.14c).
func TestMbsB1CouriersHaveLandwalk(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, land string
	}{
		{"Glissa's Courier", "Creature — Phyrexian Horror", mbsB1GlissasCourierOracle, "Mountain"},
		{"Koth's Courier", "Creature — Human Rogue", mbsB1KothsCourierOracle, "Forest"},
	} {
		g := effectstest.NewCatalogGame(t)
		me, them := effectstest.SeatsForCombat(g)
		courier := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 3)
		bear := effectstest.B12Creature(g, them.ID, "Their Bear", mbsB1BearType, 2, 2)
		effectstest.SeedLand(g, them.ID, tc.land, "Basic Land — "+tc.land, "")
		effectstest.AttackIntoBlocks(t, g, them.ID, courier)
		if r := effectstest.BlockRefusal(t, g, bear, courier); r != game.BlockReasonLandwalk {
			t.Errorf("%s: block refusal %q, want landwalk", tc.name, r)
		}
	}
}

// Protection from black and from green: neither colour can block it, red can.
func TestMbsB1MirranCrusaderCantBeBlockedByBlackOrGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	crusader := effectstest.B12Push(g, me.ID, "Mirran Crusader", "Creature — Human Knight", mbsB1MirranCrusaderOracle, 2, 2)
	black := effectstest.ExoB1Push(g, them.ID, "Black Bear", mbsB1BearType, "", 2, 2, "B")
	green := effectstest.ExoB1Push(g, them.ID, "Green Bear", mbsB1BearType, "", 2, 2, "G")
	red := effectstest.ExoB1Push(g, them.ID, "Red Bear", mbsB1BearType, "", 2, 2, "R")
	effectstest.AttackIntoBlocks(t, g, them.ID, crusader)
	if effectstest.BlockRefusal(t, g, black, crusader) == "" {
		t.Error("a black creature blocked a creature with protection from black")
	}
	if effectstest.BlockRefusal(t, g, green, crusader) == "" {
		t.Error("a green creature blocked a creature with protection from green")
	}
	if r := effectstest.BlockRefusal(t, g, red, crusader); r != "" {
		t.Errorf("a red creature could not block: %s", r)
	}
}

// --- battle cry ------------------------------------------------------------------------

func TestMbsB1BattleCryPumpsEachOtherAttacker(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, tough           int
	}{
		{"Accorder Paladin", "Creature — Human Knight", mbsB1AccorderOracle, 3, 1},
		{"Goblin Wardriver", "Creature — Goblin Warrior", mbsB1WardriverOracle, 2, 2},
		{"Loxodon Partisan", "Creature — Elephant Soldier", mbsB1PartisanOracle, 3, 4},
		{"Kuldotha Ringleader", "Creature — Giant Berserker", mbsB1RingleaderOracle, 4, 4},
		{"Signal Pest", "Artifact Creature — Pest", mbsB1SignalPestOracle, 0, 1},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		crier := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
		other := effectstest.B12Creature(g, me.ID, "My Bear", mbsB1BearType, 2, 2)
		idle := effectstest.B12Creature(g, me.ID, "My Idle Bear", mbsB1BearType, 2, 2)
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		for _, id := range []uuid.UUID{crier, other} {
			if err := g.DeclareAttacker(id, opp.ID); err != nil {
				t.Fatalf("%s: DeclareAttacker: %v", tc.name, err)
			}
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.EffectivePower(t, g, other); got != 3 {
			t.Errorf("%s: the other attacker's power %d, want 3", tc.name, got)
		}
		if got := effectstest.EffectivePower(t, g, crier); got != tc.power {
			t.Errorf("%s: its own power %d, want %d", tc.name, got, tc.power)
		}
		if got := effectstest.EffectivePower(t, g, idle); got != 2 {
			t.Errorf("%s: a creature that did not attack has power %d, want 2", tc.name, got)
		}
	}
}

// --- attacks each combat if able ----------------------------------------------------------

func TestMbsB1MustAttackCreaturesAttackEachCombat(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
	}{
		{"Kuldotha Ringleader", "Creature — Giant Berserker", mbsB1RingleaderOracle},
		{"Phyrexian Juggernaut", "Artifact Creature — Phyrexian Juggernaut", mbsB1JuggernautOracle},
	} {
		g := effectstest.NewCatalogGame(t)
		seat := g.Turn.ActiveSeat
		me, them := effectstest.SeatsForCombat(g)
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 4, 4)
		effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
		var re *game.AttackRequirementError
		if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != id {
			t.Fatalf("%s: pass with it at home = %v, want a refusal naming it", tc.name, err)
		}
		if err := g.DeclareAttacker(id, them.ID); err != nil {
			t.Fatalf("%s: DeclareAttacker: %v", tc.name, err)
		}
		if err := g.PassPriority(); err != nil {
			t.Errorf("%s: pass with it attacking: %v", tc.name, err)
		}
	}
}

// --- Signal Pest's evasion ---------------------------------------------------------------

func TestMbsB1SignalPestBlockedOnlyByFlyingOrReach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	pest := effectstest.B12Push(g, me.ID, "Signal Pest", "Artifact Creature — Pest", mbsB1SignalPestOracle, 0, 1)
	ground := effectstest.B12Creature(g, them.ID, "Their Bear", mbsB1BearType, 2, 2)
	flyer := effectstest.IceB1Push(g, them.ID, "Their Bird", "Creature — Bird", "", 1, 1, "flying")
	spider := effectstest.IceB1Push(g, them.ID, "Their Spider", "Creature — Spider", "", 1, 3, "reach")
	effectstest.AttackIntoBlocks(t, g, them.ID, pest)
	if effectstest.BlockRefusal(t, g, ground, pest) == "" {
		t.Error("a creature without flying or reach blocked Signal Pest")
	}
	for _, id := range []uuid.UUID{flyer, spider} {
		if r := effectstest.BlockRefusal(t, g, id, pest); r != "" {
			t.Errorf("%s could not block: %s", effectstest.B12Card(t, g, id).Name, r)
		}
	}
}

// --- metalcraft --------------------------------------------------------------------------

func TestMbsB1MetalcraftPumpsWithThreeArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	recruit := effectstest.B12Push(g, me.ID, "Ardent Recruit", "Creature — Human Soldier", mbsB1ArdentRecruitOracle, 1, 1)
	rhino := effectstest.B12Push(g, me.ID, "Razorfield Rhino", "Artifact Creature — Rhino", mbsB1RhinoOracle, 4, 4)
	effectstest.B12Permanent(g, me.ID, "Trinket A", "Artifact")
	if p, tough := effectstest.EffectivePTOf(t, g, recruit); p != 1 || tough != 1 {
		t.Errorf("Ardent Recruit with two artifacts is %d/%d, want 1/1", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, rhino); p != 4 || tough != 4 {
		t.Errorf("Razorfield Rhino with two artifacts is %d/%d, want 4/4", p, tough)
	}
	effectstest.B12Permanent(g, me.ID, "Trinket B", "Artifact")
	if p, tough := effectstest.EffectivePTOf(t, g, recruit); p != 3 || tough != 3 {
		t.Errorf("Ardent Recruit with three artifacts is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, rhino); p != 6 || tough != 6 {
		t.Errorf("Razorfield Rhino with three artifacts is %d/%d, want 6/6", p, tough)
	}
}

// --- activated abilities -------------------------------------------------------------------

func TestMbsB1SelfGrantsLastUntilEndOfTurn(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, colour, keyword string
	}{
		{"Bladed Sentinel", "Artifact Creature — Construct", mbsB1BladedSentinelOracle, "W", "vigilance"},
		{"Gust-Skimmer", "Artifact Creature — Insect", mbsB1GustSkimmerOracle, "U", "flying"},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
		if effectstest.HasKeywordOnBattlefield(t, g, id, tc.keyword) {
			t.Fatalf("%s has %s before activating", tc.name, tc.keyword)
		}
		effectstest.RavB2MustActivate(t, g, me.ID, id, 0, game.ActivateAbilityParams{}, tc.colour)
		if !effectstest.HasKeywordOnBattlefield(t, g, id, tc.keyword) {
			t.Errorf("%s did not gain %s", tc.name, tc.keyword)
		}
		effectstest.DcCleanup(g)
		if effectstest.HasKeywordOnBattlefield(t, g, id, tc.keyword) {
			t.Errorf("%s kept %s past the cleanup step", tc.name, tc.keyword)
		}
	}
}

func TestMbsB1DrossRipperPumpsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ripper := effectstest.B12Push(g, me.ID, "Dross Ripper", "Artifact Creature — Phyrexian Dog", mbsB1DrossRipperOracle, 3, 3)
	effectstest.RavB2MustActivate(t, g, me.ID, ripper, 0, game.ActivateAbilityParams{}, "C", "C", "B")
	if p, tough := effectstest.EffectivePTOf(t, g, ripper); p != 4 || tough != 4 {
		t.Errorf("Dross Ripper is %d/%d, want 4/4", p, tough)
	}
}

func TestMbsB1TangleHulkRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hulk := effectstest.B12Push(g, me.ID, "Tangle Hulk", "Artifact Creature — Phyrexian Beast", mbsB1TangleHulkOracle, 5, 3)
	effectstest.RavB2MustActivate(t, g, me.ID, hulk, 0, game.ActivateAbilityParams{}, "C", "C", "G")
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(hulk); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Battlefield.Contains(hulk) {
		t.Error("the regenerated Tangle Hulk was destroyed")
	}
}

func TestMbsB1FleshEaterImpSacrificesACreatureToGrow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	imp := effectstest.B12Push(g, me.ID, "Flesh-Eater Imp", "Creature — Phyrexian Imp", mbsB1FleshEaterImpOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", mbsB1BearType, 2, 2)
	effectstest.RavB2MustActivate(t, g, me.ID, imp, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not sacrificed")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, imp); p != 3 || tough != 3 {
		t.Errorf("Flesh-Eater Imp is %d/%d, want 3/3", p, tough)
	}
}

// --- enters triggers ---------------------------------------------------------------------

func TestMbsB1BlisterstickShamanPingsAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	life := opp.Life
	effectstest.PlsB1Cast(t, g, "Blisterstick Shaman", "Creature — Goblin Shaman", mbsB1BlisterstickOracle, "", 2, 1)
	pick := effectstest.PassUntilPickTarget(t, g, me.ID)
	if pick == nil {
		t.Fatal("no target was asked for")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d → %d, want 1 damage", life, opp.Life)
	}
}

func TestMbsB1PeaceStriderGainsThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	effectstest.PlsB1Cast(t, g, "Peace Strider", "Artifact Creature — Construct", mbsB1PeaceStriderOracle, "", 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
}

func TestMbsB1PierceStriderDrainsAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	life, mine := opp.Life, me.Life
	effectstest.PlsB1Cast(t, g, "Pierce Strider", "Artifact Creature — Phyrexian Construct", mbsB1PierceStriderOracle, "", 3, 3)
	pick := effectstest.PassUntilPickTarget(t, g, me.ID)
	if pick == nil {
		t.Fatal("no opponent was asked for")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("the opponent's life %d → %d, want -3", life, opp.Life)
	}
	if me.Life != mine {
		t.Errorf("the controller's life changed %d → %d", mine, me.Life)
	}
}

// --- dies and combat-damage triggers ----------------------------------------------------------

func TestMbsB1CausticHoundMakesEachPlayerLoseFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hound := effectstest.B12Push(g, me.ID, "Caustic Hound", "Creature — Phyrexian Dog", mbsB1CausticHoundOracle, 4, 4)
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Life
	}
	effectstest.KillCreature(t, g, me.ID, hound)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Life != before[i]-4 {
			t.Errorf("seat %d life %d → %d, want -4", i, before[i], p.Life)
		}
	}
}

func TestMbsB1CoreProwlerProliferatesAsItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	prowler := effectstest.B12Push(g, me.ID, "Core Prowler", "Artifact Creature — Phyrexian Horror", mbsB1CoreProwlerOracle, 2, 2)
	bear := effectstest.CounterBear(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	effectstest.KillCreature(t, g, me.ID, prowler)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerProliferate(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
}

func TestMbsB1OculusMayDrawAsItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oculus := effectstest.B12Push(g, me.ID, "Oculus", "Creature — Phyrexian Homunculus", mbsB1OculusOracle, 1, 1)
	hand := me.Hand.Size()
	effectstest.KillCreature(t, g, me.ID, oculus)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one draw", hand, me.Hand.Size())
	}
}

func TestMbsB1NeurokCommandoMayDrawOnCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	commando := effectstest.B12Push(g, me.ID, "Neurok Commando", "Creature — Human Rogue", mbsB1NeurokCommandoOracle, 2, 1)
	hand := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, commando, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one draw", hand, me.Hand.Size())
	}
}

// --- Equipment -----------------------------------------------------------------------------

func TestMbsB1EquipmentBoostsAndGrants(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
		keywords     []string
	}{
		{"Copper Carapace", mbsB1CopperCarapaceOracle, 4, 4, nil},
		{"Viridian Claw", mbsB1ViridianClawOracle, 3, 2, []string{"first strike"}},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		bear := effectstest.B12Creature(g, me.ID, "Bear", mbsB1BearType, 2, 2)
		eq := effectstest.B12Push(g, me.ID, tc.name, "Artifact — Equipment", tc.oracle, 0, 0)
		effectstest.FillPool(me, 3)
		effectstest.EquipTo(t, g, me.ID, eq, bear)
		if p, tough := effectstest.EffectivePTOf(t, g, bear); p != tc.power || tough != tc.tough {
			t.Errorf("%s: the creature is %d/%d, want %d/%d", tc.name, p, tough, tc.power, tc.tough)
		}
		for _, kw := range tc.keywords {
			if !effectstest.HasKeywordOnBattlefield(t, g, bear, kw) {
				t.Errorf("%s: the creature lacks %s", tc.name, kw)
			}
		}
	}
}

func TestMbsB1CopperCarapaceStopsBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", mbsB1BearType, 2, 2)
	blocker := effectstest.B12Creature(g, them.ID, "Their Bear", mbsB1BearType, 2, 2)
	effectstest.SomB1Equip(t, g, them.ID, "Copper Carapace", mbsB1CopperCarapaceOracle, blocker)
	effectstest.AttackIntoBlocks(t, g, them.ID, attacker)
	if effectstest.BlockRefusal(t, g, blocker, attacker) == "" {
		t.Error("the equipped creature blocked")
	}
}

func TestMbsB1LivingWeaponsMakeAnEquippedGerm(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
		keyword      string
	}{
		{"Flayer Husk", mbsB1FlayerHuskOracle, 1, 1, ""},
		{"Skinwing", mbsB1SkinwingOracle, 2, 2, "flying"},
		{"Strandwalker", mbsB1StrandwalkerOracle, 2, 4, "reach"},
	} {
		g := effectstest.NewCatalogGame(t)
		_, germ := effectstest.CastLivingWeapon(t, g, tc.name, tc.oracle)
		if p, tough := effectstest.EffectivePTOf(t, g, germ); p != tc.power || tough != tc.tough {
			t.Errorf("%s: the Germ is %d/%d, want %d/%d", tc.name, p, tough, tc.power, tc.tough)
		}
		if tc.keyword != "" && !effectstest.HasKeywordOnBattlefield(t, g, germ, tc.keyword) {
			t.Errorf("%s: the Germ lacks %s", tc.name, tc.keyword)
		}
	}
}

// --- spells --------------------------------------------------------------------------------

func TestMbsB1CreepingCorrosionDestroysEveryArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	mine := effectstest.B12Permanent(g, me.ID, "My Trinket", "Artifact")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Myr", "Artifact Creature — Myr", 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", mbsB1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Creeping Corrosion", "Sorcery", mbsB1CorrosionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("an artifact survived")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a nonartifact creature was destroyed")
	}
}

func TestMbsB1CrushDestroysANoncreatureArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	trinket := effectstest.B12Permanent(g, opp.ID, "Their Trinket", "Artifact")
	myr := effectstest.B12Creature(g, opp.ID, "Their Myr", "Artifact Creature — Myr", 1, 1)
	if err := effectstest.CastCatalogSpellErr(t, g, "Crush", "Instant", mbsB1CrushOracle, effectstest.CardTarget(myr)); err == nil {
		t.Fatal("Crush targeted an artifact creature")
	}
	effectstest.CastCatalogSpell(t, g, "Crush", "Instant", mbsB1CrushOracle, effectstest.CardTarget(trinket))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, trinket) != game.ZoneGraveyard {
		t.Errorf("the artifact is in %q, want the graveyard", effectstest.B12ZoneOf(g, trinket))
	}
}

func TestMbsB1IntoTheCoreExilesTwoArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	a := effectstest.B12Permanent(g, opp.ID, "Trinket A", "Artifact")
	b := effectstest.B12Permanent(g, opp.ID, "Trinket B", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Into the Core", "Instant", mbsB1IntoTheCoreOracle, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if effectstest.B12ZoneOf(g, id) != game.ZoneExile {
			t.Errorf("an artifact is in %q, want exile", effectstest.B12ZoneOf(g, id))
		}
	}
}

func TestMbsB1QuicksilverGeyserBouncesTwoNonlandPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	art := effectstest.B12Permanent(g, opp.ID, "Their Trinket", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", mbsB1BearType, 2, 2)
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	if err := effectstest.CastCatalogSpellErr(t, g, "Quicksilver Geyser", "Instant", mbsB1GeyserOracle, effectstest.CardTarget(land)); err == nil {
		t.Fatal("Quicksilver Geyser targeted a land")
	}
	effectstest.CastCatalogSpell(t, g, "Quicksilver Geyser", "Instant", mbsB1GeyserOracle, effectstest.CardRefs(art, bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(art) || !opp.Hand.Contains(bear) {
		t.Error("the permanents did not return to their owner's hand")
	}
}

func TestMbsB1MorbidPlunderReturnsTwoCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.WthGraveyardCard(me, "Dead Bear", mbsB1BearType)
	b := effectstest.WthGraveyardCard(me, "Dead Ogre", "Creature — Ogre")
	effectstest.CastCatalogSpell(t, g, "Morbid Plunder", "Sorcery", mbsB1MorbidPlunderOracle, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the creature cards did not return to hand")
	}
}

func TestMbsB1TurnTheTideShrinksOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", mbsB1BearType, 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", mbsB1BearType, 3, 3)
	effectstest.CastCatalogSpell(t, g, "Turn the Tide", "Instant", mbsB1TurnTheTideOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 1 || tough != 3 {
		t.Errorf("the opponent's creature is %d/%d, want 1/3", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("your creature is %d/%d, want 3/3", p, tough)
	}
	effectstest.DcCleanup(g)
	if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 3 {
		t.Errorf("the shrink outlived the cleanup step (power %d)", p)
	}
}

func TestMbsB1UnnaturalPredationPumpsAndGrantsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", mbsB1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Unnatural Predation", "Instant", mbsB1PredationOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the target is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "trample") {
		t.Error("the target lacks trample")
	}
}

func TestMbsB1MastersCallMakesTwoMyr(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Master's Call", "Instant", mbsB1MastersCallOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CardsNamed(g, "Myr", me.ID); n != 2 {
		t.Errorf("%d Myr tokens, want 2", n)
	}
}

func TestMbsB1BanishmentDecreePutsAPermanentOnTopOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	effectstest.SeedLibrary(opp, "Under")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", mbsB1BearType, 2, 2)
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	if err := effectstest.CastCatalogSpellErr(t, g, "Banishment Decree", "Instant", mbsB1BanishmentOracle, effectstest.CardTarget(land)); err == nil {
		t.Fatal("Banishment Decree targeted a land")
	}
	effectstest.CastCatalogSpell(t, g, "Banishment Decree", "Instant", mbsB1BanishmentOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, bear) {
		t.Fatal("the target is still on the battlefield")
	}
	if top := effectstest.LibraryTopIDs(opp, 1); top[0] != bear {
		t.Error("the target is not on top of its owner's library")
	}
}

func TestMbsB1SpreadTheSicknessDestroysThenProliferates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", mbsB1BearType, 2, 2)
	bear := effectstest.CounterBear(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	effectstest.CastCatalogSpell(t, g, "Spread the Sickness", "Sorcery", mbsB1SicknessOracle, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerProliferate(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, victim) != game.ZoneGraveyard {
		t.Errorf("the target is in %q, want the graveyard", effectstest.B12ZoneOf(g, victim))
	}
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
}
