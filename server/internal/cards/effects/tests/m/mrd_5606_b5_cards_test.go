package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mrd_5606_b5_cards_test.go — Mirrodin, batch 5 (#5606, S82). Every test
// plays the real catalog entry through the engine.

const (
	mrdB5KrarkClanGruntOracle    = "e7be5ab3-a11f-4786-9b0b-1daa94f59440"
	mrdB5LeoninDenGuardOracle    = "651220f7-223a-49b3-995d-32f08eb3a16a"
	mrdB5LeoninSunStandardOracle = "e235f988-55d2-421b-9cfd-0d30d2727aed"
	mrdB5LoxodonMenderOracle     = "b260b5b4-048b-4a70-ab3a-b0dd1ceb7aa8"
	mrdB5LoxodonPunisherOracle   = "fc3869fd-1441-4d5e-9af5-f02857ed634a"
	mrdB5LumengridWardenOracle   = "ee16a8f3-2522-4c36-b4de-65484dfb7455"
	mrdB5LuminousAngelOracle     = "8d7cf56c-94fd-47d8-9e0f-cd3163688983"
	mrdB5MalachiteGolemOracle    = "54e71732-6575-48f4-838b-35d44eba51cf"
	mrdB5MegatogOracle           = "bde427fa-8a00-4eff-9371-81324bc75364"
	mrdB5MyrAdapterOracle        = "714f916a-10ac-4c54-b1c7-a97b7ad8b9e5"
	mrdB5NecrogenSpellbombOracle = "34993bd9-53bf-46b2-bc20-b6daf82382d9"
	mrdB5NeedlebugOracle         = "2d5fba80-c84f-4f84-84e4-9c797ac244fc"
	mrdB5NeurokHoversailOracle   = "3cc02a23-93b7-445a-9c3e-0e4942ef927b"
	mrdB5NimLasherOracle         = "0203e250-d60e-46c6-9c58-d620f0377be0"
	mrdB5NimReplicaOracle        = "68c3aa66-85f2-494c-8fb2-ad0348814506"
	mrdB5NimShriekerOracle       = "1dff7c82-e49f-45e8-8f66-d5801f420fa6"
	mrdB5OmegaMyrOracle          = "fd1b237a-7132-4875-b738-8ec65b85ea28"
	mrdB5PewterGolemOracle       = "c9d2c69e-6a4a-45b0-aff1-c19bcbe3b490"
	mrdB5PlatedSlagwurmOracle    = "4aec7624-e406-45a4-b2b6-2e8d29f6268a"
	mrdB5PyriteSpellbombOracle   = "2c10cae2-951a-4f4f-94e4-8713b58d07dd"
	mrdB5RegressOracle           = "2fa763ad-4d94-4c3a-a099-13ac22c09ce4"
	mrdB5RustsporeRamOracle      = "a45b3934-1c9b-4cff-98b1-c9ac2f7759ea"
	mrdB5ShrapnelBlastOracle     = "d3c623f3-c7c9-47c0-8096-6dee3267baff"
	mrdB5SkyhunterCubOracle      = "1e9ebdf3-306a-48e1-8096-6a1d289e108c"
	mrdB5SkyhunterPatrolOracle   = "aadcff6f-9207-4d90-a12d-4913c96867e2"
	mrdB5SlagwurmArmorOracle     = "20b60c93-124b-42c8-93fb-63bbd1888658"
	mrdB5SlithAscendantOracle    = "72f5095f-b4ba-45ba-82e5-9a22bddd544b"
	mrdB5SlithBloodletterOracle  = "fbad6038-93ed-4a28-b1f0-ada54608758b"
	mrdB5SlithFirewalkerOracle   = "ea1af948-686d-4b40-ad4e-cb48a0258269"
	mrdB5SlithPredatorOracle     = "fd9c984e-daa2-414c-bb90-ecc41aaa896d"
	mrdB5SoldierReplicaOracle    = "702693f0-09f6-4f19-8dc9-6b31641e3347"
	mrdB5SomberHoverguardOracle  = "3c88108e-d983-427a-9b85-b9cdd2d5083b"
	mrdB5SteelWallOracle         = "5ccb57e1-ca94-4b5a-8e5f-b8b5e692cfb9"
	mrdB5SunbeamSpellbombOracle  = "014c94d8-2c39-4d33-902b-fa2398406fd5"
	mrdB5TanglebloomOracle       = "87890f04-b831-4869-a7de-72789a466c71"
	mrdB5TelJiladArchersOracle   = "939203ec-3de4-44d8-932e-63a32e5f2d47"
	mrdB5TelJiladChosenOracle    = "de6854d1-bdb0-4b9a-b442-08d4d90a538d"
	mrdB5TelJiladExileOracle     = "58987ca8-ee9b-412e-b08f-440588234e00"
	mrdB5TempestOfLightOracle    = "9399ed27-31c7-4d6a-99eb-644b74e67344"
	mrdB5TitaniumGolemOracle     = "84fa0c22-4e11-4d57-9a45-633f0e1686de"
	mrdB5TowerOfChampionsOracle  = "f7880784-e1dc-4250-812e-eda41fe91e36"
	mrdB5TowerOfEonsOracle       = "74a71bbd-c307-481c-b0cf-70e24b0c8ad4"
	mrdB5TowerOfFortunesOracle   = "dfb23df9-f912-4daa-9842-01bd12b4a72a"
	mrdB5TrollAsceticOracle      = "0585ea6a-ef78-4386-a22c-6b8d8fc6c03c"
	mrdB5TrollsOfTelJiladOracle  = "4478e843-fd25-4ad1-93f8-2dee605386fc"
	mrdB5ViridianShamanOracle    = "5aa6d553-a144-4c5b-83fa-65e931903899"
	mrdB5VulshokBattlegearOracle = "708df587-3b13-42cb-8341-f476ab4cbe45"
	mrdB5VulshokBerserkerOracle  = "694f7e51-7b8b-4f00-bd77-a52eded4aaa1"
	mrdB5WallOfBloodOracle       = "b26c3563-d289-4ca3-8736-9d89e191a840"
	mrdB5WanderguardSentryOracle = "cd02e276-3c2d-4177-aa60-294a31bae0db"

	mrdB5ArtifactType = "Artifact"
)

// mrdB5Activate activates row `index` of a catalog permanent with Strict
// payment and lets the ability resolve.
func mrdB5Activate(t *testing.T, g *game.Game, p *game.Player, source uuid.UUID, index int, params game.ActivateAbilityParams) {
	t.Helper()
	params.Strict = true
	if err := g.ActivateCatalogAbility(p.ID, source, index, params); err != nil {
		t.Fatalf("activate row %d: %v", index, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// mrdB5Refused asserts that activating row `index` is refused.
func mrdB5Refused(t *testing.T, g *game.Game, p *game.Player, source uuid.UUID, index int, params game.ActivateAbilityParams, why string) {
	t.Helper()
	params.Strict = true
	if err := g.ActivateCatalogAbility(p.ID, source, index, params); err == nil {
		t.Fatal(why)
	}
}

// mrdB5Mana floats `n` colorless mana and the given colours.
func mrdB5Mana(p *game.Player, n int, colors ...string) {
	for i := 0; i < n; i++ {
		effectstest.B06AddMana(p, "C")
	}
	effectstest.B06AddMana(p, colors...)
}

// --- every card is in the catalog --------------------------------------------

func TestMrd5606B5EveryCardIsRegisteredFull(t *testing.T) {
	for name, oracle := range map[string]string{
		"Krark-Clan Grunt": mrdB5KrarkClanGruntOracle, "Leonin Den-Guard": mrdB5LeoninDenGuardOracle,
		"Leonin Sun Standard": mrdB5LeoninSunStandardOracle, "Loxodon Mender": mrdB5LoxodonMenderOracle,
		"Loxodon Punisher": mrdB5LoxodonPunisherOracle, "Lumengrid Warden": mrdB5LumengridWardenOracle,
		"Luminous Angel": mrdB5LuminousAngelOracle, "Malachite Golem": mrdB5MalachiteGolemOracle,
		"Megatog": mrdB5MegatogOracle, "Myr Adapter": mrdB5MyrAdapterOracle,
		"Necrogen Spellbomb": mrdB5NecrogenSpellbombOracle, "Needlebug": mrdB5NeedlebugOracle,
		"Neurok Hoversail": mrdB5NeurokHoversailOracle, "Nim Lasher": mrdB5NimLasherOracle,
		"Nim Replica": mrdB5NimReplicaOracle, "Nim Shrieker": mrdB5NimShriekerOracle,
		"Omega Myr": mrdB5OmegaMyrOracle, "Pewter Golem": mrdB5PewterGolemOracle,
		"Plated Slagwurm": mrdB5PlatedSlagwurmOracle, "Pyrite Spellbomb": mrdB5PyriteSpellbombOracle,
		"Regress": mrdB5RegressOracle, "Rustspore Ram": mrdB5RustsporeRamOracle,
		"Shrapnel Blast": mrdB5ShrapnelBlastOracle, "Skyhunter Cub": mrdB5SkyhunterCubOracle,
		"Skyhunter Patrol": mrdB5SkyhunterPatrolOracle, "Slagwurm Armor": mrdB5SlagwurmArmorOracle,
		"Slith Ascendant": mrdB5SlithAscendantOracle, "Slith Bloodletter": mrdB5SlithBloodletterOracle,
		"Slith Firewalker": mrdB5SlithFirewalkerOracle, "Slith Predator": mrdB5SlithPredatorOracle,
		"Soldier Replica": mrdB5SoldierReplicaOracle, "Somber Hoverguard": mrdB5SomberHoverguardOracle,
		"Steel Wall": mrdB5SteelWallOracle, "Sunbeam Spellbomb": mrdB5SunbeamSpellbombOracle,
		"Tanglebloom": mrdB5TanglebloomOracle, "Tel-Jilad Archers": mrdB5TelJiladArchersOracle,
		"Tel-Jilad Chosen": mrdB5TelJiladChosenOracle, "Tel-Jilad Exile": mrdB5TelJiladExileOracle,
		"Tempest of Light": mrdB5TempestOfLightOracle, "Titanium Golem": mrdB5TitaniumGolemOracle,
		"Tower of Champions": mrdB5TowerOfChampionsOracle, "Tower of Eons": mrdB5TowerOfEonsOracle,
		"Tower of Fortunes": mrdB5TowerOfFortunesOracle, "Troll Ascetic": mrdB5TrollAsceticOracle,
		"Trolls of Tel-Jilad": mrdB5TrollsOfTelJiladOracle, "Viridian Shaman": mrdB5ViridianShamanOracle,
		"Vulshok Battlegear": mrdB5VulshokBattlegearOracle, "Vulshok Berserker": mrdB5VulshokBerserkerOracle,
		"Wall of Blood": mrdB5WallOfBloodOracle, "Wanderguard Sentry": mrdB5WanderguardSentryOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered as %q, %v, caveats %v; want its name, Full and none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- printed keywords ----------------------------------------------------------

func TestMrd5606B5PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Steel Wall", "Artifact Creature — Wall", mrdB5SteelWallOracle, []string{"defender"}},
		{"Plated Slagwurm", "Creature — Wurm", mrdB5PlatedSlagwurmOracle, []string{"hexproof"}},
		{"Skyhunter Patrol", "Creature — Cat Knight", mrdB5SkyhunterPatrolOracle, []string{"flying", "first strike"}},
		{"Vulshok Berserker", "Creature — Human Berserker", mrdB5VulshokBerserkerOracle, []string{"haste"}},
		{"Needlebug", "Artifact Creature — Insect", mrdB5NeedlebugOracle, []string{"flash", "protection from artifacts"}},
		{"Tel-Jilad Chosen", "Creature — Elf Warrior", mrdB5TelJiladChosenOracle, []string{"protection from artifacts"}},
		{"Tel-Jilad Archers", "Creature — Elf Archer", mrdB5TelJiladArchersOracle, []string{"reach", "protection from artifacts"}},
		{"Somber Hoverguard", "Creature — Drone", mrdB5SomberHoverguardOracle, []string{"flying"}},
		{"Troll Ascetic", "Creature — Troll Shaman", mrdB5TrollAsceticOracle, []string{"hexproof"}},
		{"Wall of Blood", "Creature — Wall", mrdB5WallOfBloodOracle, []string{"defender"}},
		{"Nim Shrieker", "Creature — Zombie", mrdB5NimShriekerOracle, []string{"flying"}},
		{"Luminous Angel", "Creature — Angel", mrdB5LuminousAngelOracle, []string{"flying"}},
		{"Slith Ascendant", "Creature — Slith", mrdB5SlithAscendantOracle, []string{"flying"}},
		{"Slith Firewalker", "Creature — Slith", mrdB5SlithFirewalkerOracle, []string{"haste"}},
		{"Slith Predator", "Creature — Slith", mrdB5SlithPredatorOracle, []string{"trample"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			abilities := effectstest.EffectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !effectstest.HasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// Protection from artifacts: Pyrite Spellbomb, an artifact, can't target
// Needlebug or either Tel-Jilad card, and can target a plain creature.
func TestMrd5606B5ProtectionFromArtifactsStopsAnArtifactsAbility(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Needlebug", "Artifact Creature — Insect", mrdB5NeedlebugOracle},
		{"Tel-Jilad Chosen", "Creature — Elf Warrior", mrdB5TelJiladChosenOracle},
		{"Tel-Jilad Archers", "Creature — Elf Archer", mrdB5TelJiladArchersOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			bomb := effectstest.Mrd2Catalog(g, me.ID, "Pyrite Spellbomb", "Artifact", mrdB5PyriteSpellbombOracle, "{1}", 0, 0)
			warded := effectstest.B12Push(g, opp.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			effectstest.B06AddMana(me, "R")
			mrdB5Refused(t, g, me, bomb, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(warded)},
				"an artifact's ability targeted a creature with protection from artifacts")
		})
	}
}

// Needlebug's flash: it is cast at the beginning of combat.
func TestMrd5606B5NeedlebugHasFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B13HandCard(me, "Needlebug", "Artifact Creature — Insect", mrdB5NeedlebugOracle)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("Needlebug at the beginning of combat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("Needlebug did not resolve")
	}
}

func TestMrd5606B5SomberHoverguardHasAffinityForArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 2; i++ {
		effectstest.PushPermanent(g, me.ID, game.Card{Name: "Signet", TypeLine: "Artifact"})
	}
	effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Their Signet", TypeLine: "Artifact"})
	if got := effectstest.SelfPricedMV(t, g, me, mrdB5SomberHoverguardOracle, "Creature — Drone", "{5}{U}"); got != 4 {
		t.Errorf("with two artifacts it costs mana value %d, want 4", got)
	}
	for i := 0; i < 5; i++ {
		effectstest.PushPermanent(g, me.ID, game.Card{Name: "Signet", TypeLine: "Artifact"})
	}
	if got := effectstest.SelfPricedMV(t, g, me, mrdB5SomberHoverguardOracle, "Creature — Drone", "{5}{U}"); got != 1 {
		t.Errorf("with seven artifacts it costs mana value %d, want 1 (the {U} stays)", got)
	}
}

// --- abilities on the creature itself ------------------------------------------

func TestMrd5606B5GolemsGainAKeywordUntilEndOfTurn(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, color, keyword string
	}{
		{"Malachite Golem", mrdB5MalachiteGolemOracle, "G", "trample"},
		{"Titanium Golem", mrdB5TitaniumGolemOracle, "W", "first strike"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			golem := effectstest.Mrd2Catalog(g, me.ID, tc.name, "Artifact Creature — Golem", tc.oracle, "{6}", 3, 3)
			if effectstest.HasKeywordOnBattlefield(t, g, golem, tc.keyword) {
				t.Fatalf("%s has %s before the ability", tc.name, tc.keyword)
			}
			mrdB5Mana(me, 1, tc.color)
			mrdB5Activate(t, g, me, golem, 0, game.ActivateAbilityParams{})
			if !effectstest.HasKeywordOnBattlefield(t, g, golem, tc.keyword) {
				t.Errorf("%s has no %s after the ability", tc.name, tc.keyword)
			}
		})
	}
}

func TestMrd5606B5RegeneratorsShieldThemselves(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, color string
	}{
		{"Pewter Golem", "Artifact Creature — Golem", mrdB5PewterGolemOracle, "B"},
		{"Tel-Jilad Exile", "Creature — Troll Warrior", mrdB5TelJiladExileOracle, "G"},
		{"Troll Ascetic", "Creature — Troll Shaman", mrdB5TrollAsceticOracle, "G"},
		{"Slith Bloodletter", "Creature — Slith", mrdB5SlithBloodletterOracle, "B"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			id := effectstest.Mrd2Catalog(g, me.ID, tc.name, tc.typeLine, tc.oracle, "{3}", 2, 2)
			mrdB5Mana(me, 1, tc.color)
			mrdB5Activate(t, g, me, id, 0, game.ActivateAbilityParams{})
			effectstest.P1Destroy(t, g, id)
			if !g.Battlefield.Contains(id) || !effectstest.TappedForTest(t, g, id) {
				t.Errorf("the regenerated %s should stay on the battlefield, tapped", tc.name)
			}
		})
	}
}

func TestMrd5606B5WallOfBloodPaysLifeToGrow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	wall := effectstest.Mrd2Catalog(g, me.ID, "Wall of Blood", "Creature — Wall", mrdB5WallOfBloodOracle, "{2}{B}", 0, 2)
	life := me.Life
	for i := 0; i < 2; i++ {
		mrdB5Activate(t, g, me, wall, 0, game.ActivateAbilityParams{})
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
	if p, tt := effectstest.Mrd2Size(t, g, wall); p != 2 || tt != 4 {
		t.Errorf("the Wall is %d/%d, want 2/4", p, tt)
	}
}

func TestMrd5606B5SacrificeAnArtifactPumps(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, tough           int
		wantP, wantT           int
		keyword                string
	}{
		{"Megatog", "Creature — Atog", mrdB5MegatogOracle, 3, 4, 6, 7, "trample"},
		{"Krark-Clan Grunt", "Creature — Goblin Warrior", mrdB5KrarkClanGruntOracle, 2, 2, 3, 2, "first strike"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			id := effectstest.Mrd2Catalog(g, me.ID, tc.name, tc.typeLine, tc.oracle, "{2}{R}", tc.power, tc.tough)
			bear := effectstest.SeedBear(g, me.ID)
			rock := effectstest.MrdPermanent(g, me.ID, "My Rock", "Artifact", "{2}", 0, 0)
			mrdB5Refused(t, g, me, id, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}},
				"a creature that isn't an artifact paid the sacrifice")
			mrdB5Activate(t, g, me, id, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
			if g.Battlefield.Contains(rock) {
				t.Error("the artifact was not sacrificed")
			}
			if p, tt := effectstest.Mrd2Size(t, g, id); p != tc.wantP || tt != tc.wantT {
				t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tt, tc.wantP, tc.wantT)
			}
			if !effectstest.HasKeywordOnBattlefield(t, g, id, tc.keyword) {
				t.Errorf("%s has no %s", tc.name, tc.keyword)
			}
		})
	}
}

// --- activated abilities on other objects --------------------------------------

func TestMrd5606B5LeoninSunStandardPumpsOnlyYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	standard := effectstest.Mrd2Catalog(g, me.ID, "Leonin Sun Standard", "Artifact", mrdB5LeoninSunStandardOracle, "{2}", 0, 0)
	mine := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.SeedBear(g, opp.ID)
	mrdB5Mana(me, 1, "W")
	mrdB5Activate(t, g, me, standard, 0, game.ActivateAbilityParams{})
	if p, tt := effectstest.Mrd2Size(t, g, mine); p != 3 || tt != 3 {
		t.Errorf("your creature is %d/%d, want 3/3", p, tt)
	}
	if p, tt := effectstest.Mrd2Size(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tt)
	}
}

func TestMrd5606B5LoxodonMenderRegeneratesAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	mender := effectstest.Mrd2Catalog(g, me.ID, "Loxodon Mender", "Creature — Elephant Cleric", mrdB5LoxodonMenderOracle, "{5}{W}", 3, 3)
	bear := effectstest.SeedBear(g, me.ID)
	rock := effectstest.MrdPermanent(g, me.ID, "My Rock", "Artifact", "{2}", 0, 0)
	effectstest.B06AddMana(me, "W")
	mrdB5Refused(t, g, me, mender, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)},
		"a creature that isn't an artifact was accepted as the target")
	mrdB5Activate(t, g, me, mender, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(rock)})
	effectstest.P1Destroy(t, g, rock)
	if !g.Battlefield.Contains(rock) {
		t.Error("the shielded artifact was destroyed")
	}
	if !effectstest.TappedForTest(t, g, mender) {
		t.Error("the Mender taps as part of the cost")
	}
}

func TestMrd5606B5TrollsOfTelJiladRegenerateAGreenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	trolls := effectstest.Mrd2Catalog(g, me.ID, "Trolls of Tel-Jilad", "Creature — Troll Shaman", mrdB5TrollsOfTelJiladOracle, "{5}{G}{G}", 5, 6)
	elf := effectstest.B16Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1, "G")
	goblin := effectstest.B16Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1, "R")
	mrdB5Mana(me, 1, "G")
	mrdB5Refused(t, g, me, trolls, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(goblin)},
		"a red creature was accepted as the target green creature")
	mrdB5Activate(t, g, me, trolls, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(elf)})
	effectstest.P1Destroy(t, g, elf)
	if !g.Battlefield.Contains(elf) {
		t.Error("the shielded green creature was destroyed")
	}
}

func TestMrd5606B5LifeGainArtifacts(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		generic      int
		colors       []string
		gain         int
	}{
		{"Tanglebloom", mrdB5TanglebloomOracle, 1, nil, 1},
		{"Tower of Eons", mrdB5TowerOfEonsOracle, 8, nil, 10},
		{"Sunbeam Spellbomb", mrdB5SunbeamSpellbombOracle, 0, []string{"W"}, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			id := effectstest.Mrd2Catalog(g, me.ID, tc.name, mrdB5ArtifactType, tc.oracle, "{1}", 0, 0)
			life := me.Life
			mrdB5Mana(me, tc.generic, tc.colors...)
			mrdB5Activate(t, g, me, id, 0, game.ActivateAbilityParams{})
			if me.Life != life+tc.gain {
				t.Errorf("life %d, want %d", me.Life, life+tc.gain)
			}
		})
	}
}

func TestMrd5606B5TowerOfFortunesDrawsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	tower := effectstest.Mrd2Catalog(g, me.ID, "Tower of Fortunes", mrdB5ArtifactType, mrdB5TowerOfFortunesOracle, "{4}", 0, 0)
	hand := me.Hand.Size()
	mrdB5Mana(me, 8)
	mrdB5Activate(t, g, me, tower, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("drew %d, want 4", got)
	}
	if !effectstest.TappedForTest(t, g, tower) {
		t.Error("the Tower taps as part of the cost")
	}
}

func TestMrd5606B5TowerOfChampionsGivesPlusSixPlusSix(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	tower := effectstest.Mrd2Catalog(g, me.ID, "Tower of Champions", mrdB5ArtifactType, mrdB5TowerOfChampionsOracle, "{4}", 0, 0)
	bear := effectstest.SeedBear(g, me.ID)
	mrdB5Mana(me, 8)
	mrdB5Activate(t, g, me, tower, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if p, tt := effectstest.Mrd2Size(t, g, bear); p != 8 || tt != 8 {
		t.Errorf("the creature is %d/%d, want 8/8", p, tt)
	}
}

func TestMrd5606B5SpellbombsDrawWithTheSecondRow(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Pyrite Spellbomb", mrdB5PyriteSpellbombOracle},
		{"Necrogen Spellbomb", mrdB5NecrogenSpellbombOracle},
		{"Sunbeam Spellbomb", mrdB5SunbeamSpellbombOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			bomb := effectstest.Mrd2Catalog(g, me.ID, tc.name, "Artifact", tc.oracle, "{1}", 0, 0)
			hand := me.Hand.Size()
			mrdB5Mana(me, 1)
			mrdB5Activate(t, g, me, bomb, 1, game.ActivateAbilityParams{})
			if g.Battlefield.Contains(bomb) {
				t.Error("the Spellbomb was not sacrificed")
			}
			if got := me.Hand.Size() - hand; got != 1 {
				t.Errorf("drew %d, want 1", got)
			}
		})
	}
}

func TestMrd5606B5PyriteSpellbombDealsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bomb := effectstest.Mrd2Catalog(g, me.ID, "Pyrite Spellbomb", "Artifact", mrdB5PyriteSpellbombOracle, "{1}", 0, 0)
	life := opp.Life
	effectstest.B06AddMana(me, "R")
	mrdB5Activate(t, g, me, bomb, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if opp.Life != life-2 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-2)
	}
}

func TestMrd5606B5NecrogenSpellbombMakesTheTargetDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bomb := effectstest.Mrd2Catalog(g, me.ID, "Necrogen Spellbomb", "Artifact", mrdB5NecrogenSpellbombOracle, "{1}", 0, 0)
	effectstest.D2dSetHand(opp, 3)
	effectstest.B06AddMana(me, "B")
	mrdB5Activate(t, g, me, bomb, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if n := effectstest.DiscardOwed(g, opp.ID); n != 1 {
		t.Errorf("the target player owes %d discards, want 1", n)
	}
}

func TestMrd5606B5NimReplicaShrinksTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	replica := effectstest.Mrd2Catalog(g, me.ID, "Nim Replica", "Artifact Creature — Zombie", mrdB5NimReplicaOracle, "{3}", 3, 1)
	bear := effectstest.SeedBear(g, opp.ID)
	mrdB5Mana(me, 2, "B")
	mrdB5Activate(t, g, me, replica, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if g.Battlefield.Contains(replica) {
		t.Error("the Replica is sacrificed as the cost")
	}
	if p, tt := effectstest.Mrd2Size(t, g, bear); p != 1 || tt != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tt)
	}
}

func TestMrd5606B5SoldierReplicaShootsOnlyACombatant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	replica := effectstest.Mrd2Catalog(g, me.ID, "Soldier Replica", "Artifact Creature — Soldier", mrdB5SoldierReplicaOracle, "{3}", 1, 3)
	raider := effectstest.B12Creature(g, me.ID, "Raider", "Creature — Human", 2, 4)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Human", 2, 4)
	effectstest.BrAttack(t, g, raider)
	mrdB5Mana(me, 1, "W")
	mrdB5Refused(t, g, me, replica, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(idle)},
		"the Replica targeted a creature that isn't attacking or blocking")
	mrdB5Activate(t, g, me, replica, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(raider)})
	if got := effectstest.FindCardAnywhere(t, g, raider).DamageMarked; got != 3 {
		t.Errorf("the attacker has %d damage, want 3", got)
	}
}

// --- statics and Equipment -------------------------------------------------------

func TestMrd5606B5NimCreaturesCountYourArtifacts(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
	}{
		{"Nim Lasher", mrdB5NimLasherOracle, 1, 1},
		{"Nim Shrieker", mrdB5NimShriekerOracle, 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			nim := effectstest.Mrd2Catalog(g, me.ID, tc.name, "Creature — Zombie", tc.oracle, "{2}{B}", tc.power, tc.tough)
			effectstest.MrdPermanent(g, me.ID, "Rock", "Artifact", "{1}", 0, 0)
			effectstest.MrdPermanent(g, me.ID, "Golem", "Artifact Creature — Golem", "{2}", 1, 1)
			effectstest.MrdPermanent(g, opp.ID, "Their Rock", "Artifact", "{1}", 0, 0)
			if p, tt := effectstest.Mrd2Size(t, g, nim); p != tc.power+2 || tt != tc.tough {
				t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tt, tc.power+2, tc.tough)
			}
		})
	}
}

func TestMrd5606B5PunisherAndAdapterCountTheirEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	punisher := effectstest.Mrd2Catalog(g, me.ID, "Loxodon Punisher", "Creature — Elephant Soldier", mrdB5LoxodonPunisherOracle, "{3}{W}", 2, 2)
	adapter := effectstest.Mrd2Catalog(g, me.ID, "Myr Adapter", "Artifact Creature — Myr", mrdB5MyrAdapterOracle, "{3}", 1, 1)
	battlegear := effectstest.SeedEquipment(g, me.ID, "Vulshok Battlegear", mrdB5VulshokBattlegearOracle)
	armor := effectstest.SeedEquipment(g, me.ID, "Slagwurm Armor", mrdB5SlagwurmArmorOracle)
	if p, tt := effectstest.Mrd2Size(t, g, punisher); p != 2 || tt != 2 {
		t.Fatalf("the unequipped Punisher is %d/%d, want 2/2", p, tt)
	}
	effectstest.EquipTo(t, g, me.ID, battlegear, punisher)
	if p, tt := effectstest.Mrd2Size(t, g, punisher); p != 7 || tt != 7 {
		t.Errorf("the Punisher with the Battlegear is %d/%d, want 7/7 (2 + 2 counted + 3)", p, tt)
	}
	effectstest.EquipTo(t, g, me.ID, armor, adapter)
	if p, tt := effectstest.Mrd2Size(t, g, adapter); p != 2 || tt != 8 {
		t.Errorf("the Adapter with the Armor is %d/%d, want 2/8 (1 + 1 counted + 0/6)", p, tt)
	}
}

func TestMrd5606B5EquippedCatsGrowAndGainAKeyword(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, tough           int
		keyword                string
	}{
		{"Leonin Den-Guard", "Creature — Cat Soldier", mrdB5LeoninDenGuardOracle, 1, 3, "vigilance"},
		{"Skyhunter Cub", "Creature — Cat Knight", mrdB5SkyhunterCubOracle, 2, 2, "flying"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			cat := effectstest.Mrd2Catalog(g, me.ID, tc.name, tc.typeLine, tc.oracle, "{1}{W}", tc.power, tc.tough)
			sail := effectstest.SeedEquipment(g, me.ID, "Neurok Hoversail", mrdB5NeurokHoversailOracle)
			if effectstest.HasKeywordOnBattlefield(t, g, cat, tc.keyword) {
				t.Fatalf("the unequipped %s has %s", tc.name, tc.keyword)
			}
			effectstest.EquipTo(t, g, me.ID, sail, cat)
			if p, tt := effectstest.Mrd2Size(t, g, cat); p != tc.power+1 || tt != tc.tough+1 {
				t.Errorf("the equipped %s is %d/%d, want %d/%d", tc.name, p, tt, tc.power+1, tc.tough+1)
			}
			for _, kw := range []string{tc.keyword, "flying"} {
				if !effectstest.HasKeywordOnBattlefield(t, g, cat, kw) {
					t.Errorf("the equipped %s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// --- triggered abilities ---------------------------------------------------------

func TestMrd5606B5SlithsGrowOnCombatDamageToAPlayer(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Slith Ascendant", mrdB5SlithAscendantOracle},
		{"Slith Bloodletter", mrdB5SlithBloodletterOracle},
		{"Slith Firewalker", mrdB5SlithFirewalkerOracle},
		{"Slith Predator", mrdB5SlithPredatorOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			slith := effectstest.B12Push(g, me.ID, tc.name, "Creature — Slith", tc.oracle, 1, 1)
			bear := effectstest.SeedBear(g, opp.ID)
			effectstest.DealCombatDamage(g, me.ID, slith, bear)
			effectstest.PassPriorityAroundTable(t, g)
			if got := mrdPlusOneCounters(t, g, slith); got != 0 {
				t.Fatalf("combat damage to a creature put %d counters on it", got)
			}
			effectstest.DealCombatDamage(g, me.ID, slith, opp.ID)
			effectstest.PassPriorityAroundTable(t, g)
			if got := mrdPlusOneCounters(t, g, slith); got != 1 {
				t.Errorf("%s has %d +1/+1 counters, want 1", tc.name, got)
			}
		})
	}
}

func TestMrd5606B5EntersAndDestroys(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		victim                 string
	}{
		{"Viridian Shaman", "Creature — Elf Shaman", mrdB5ViridianShamanOracle, "Artifact"},
		{"Rustspore Ram", "Artifact Creature — Sheep", mrdB5RustsporeRamOracle, "Artifact — Equipment"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			victim := effectstest.B12Permanent(g, opp.ID, "Victim", tc.victim)
			bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
			effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 2, 2)
			effectstest.B04WaitForPick(t, g, me.ID)
			p := effectstest.LatestPickTarget(g, me.ID)
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
				t.Fatal("a creature was accepted as the target")
			}
			effectstest.AnswerPickTarget(t, g, victim)
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(victim) {
				t.Error("the target survived")
			}
		})
	}
}

// Rustspore Ram's target is an Equipment: a plain artifact is refused.
func TestMrd5606B5RustsporeRamTargetsOnlyEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	sword := effectstest.B12Permanent(g, opp.ID, "Sword", "Artifact — Equipment")
	effectstest.CastHoldCreature(t, g, "Rustspore Ram", "Artifact Creature — Sheep", mrdB5RustsporeRamOracle, 1, 3)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: rock}); err == nil {
		t.Fatal("an artifact that isn't an Equipment was accepted as the target")
	}
	effectstest.AnswerPickTarget(t, g, sword)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(sword) || !g.Battlefield.Contains(rock) {
		t.Error("the Equipment should be destroyed and the other artifact left alone")
	}
}

func TestMrd5606B5WanderguardSentryLooksAtAnOpponentsHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	effectstest.HandCardFull(opp, "Secret", "Instant", "{U}", "", nil)
	effectstest.CastCatalogSpell(t, g, "Wanderguard Sentry", "Creature — Drone", mrdB5WanderguardSentryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	for i := range opp.Hand.Cards {
		if !opp.Hand.Cards[i].IsKnownTo(me.ID) {
			t.Error("the Sentry's controller did not see the opponent's hand")
		}
		if opp.Hand.Cards[i].IsKnownTo(other.ID) {
			t.Error("\"look at\" showed the hand to another player")
		}
	}
}

func TestMrd5606B5LuminousAngelMakesASpiritEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	effectstest.B12Push(g, owner.ID, "Luminous Angel", "Creature — Angel", mrdB5LuminousAngelOracle, 4, 4)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AnswerLatestTriggerPrompt(t, g, owner.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	spirit := effectstest.FindBattlefieldByName(g, "Spirit")
	if spirit == uuid.Nil {
		t.Fatal("no Spirit token")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, spirit, "flying") {
		t.Error("the Spirit has no flying")
	}
}

// --- the spells ------------------------------------------------------------------

func TestMrd5606B5RegressReturnsAPermanentToItsOwnersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	rock := uuid.New()
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: rock, Name: "Their Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Regress", "Instant", mrdB5RegressOracle, effectstest.CardTarget(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Hand.Contains(rock) {
		t.Error("the permanent did not return to its owner's hand")
	}
}

func TestMrd5606B5TempestOfLightDestroysEveryEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	mine := effectstest.B12Permanent(g, me.ID, "My Aura", "Enchantment")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Tempest of Light", "Instant", mrdB5TempestOfLightOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("an enchantment survived")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("an artifact was destroyed")
	}
}

func TestMrd5606B5ShrapnelBlastSacrificesAnArtifactToDealFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	life := opp.Life
	id := effectstest.B13HandCard(me, "Shrapnel Blast", "Instant", mrdB5ShrapnelBlastOracle)
	targets := []game.TargetRef{effectstest.PlayerTarget(opp.ID)}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets}); err == nil {
		t.Fatal("cast with no sacrifice")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a creature that isn't an artifact paid the sacrifice")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets, SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if opp.Life != life-5 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-5)
	}
}
