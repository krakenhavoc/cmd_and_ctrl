package testsr

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rna_5521_b1_cards_test.go — Ravnica Allegiance, batch 1 (#5521): five
// vanilla bodies, keyword and riot creatures, afterlife and adapt, and
// cards whose text an existing catalog card already prints. Each test
// plays the real catalog entry through the engine and asserts what its
// text says.

const (
	rnaB1AxebaneBeast       = "8f1d5c8d-98e1-47b7-8347-38887253e34a"
	rnaB1CatacombCrocodile  = "edc52483-19c7-4d3a-8915-6b1f41b2d6e5"
	rnaB1CoralCommando      = "03ce2235-50d9-4f67-b588-166b9b37c664"
	rnaB1FeralMaaka         = "d4556b82-f96e-4f88-a885-3e84afdb98a1"
	rnaB1ProwlingCaracal    = "0eabe5e9-cb03-4481-9360-5e79eccd0631"
	rnaB1Humongulus         = "395922cc-d524-4241-a7d8-9c7959f0b9db"
	rnaB1NoxiousGroodion    = "aed587bf-c6e5-4850-baeb-2715cb27bf4c"
	rnaB1RubbleSlinger      = "4de340bf-eab6-4066-896d-7084a9c13c2d"
	rnaB1GhorClanWrecker    = "b113285c-80b2-46dd-874e-201d26979518"
	rnaB1RampagingRendhorn  = "b867deb4-a088-492e-95fe-487071ebfe8d"
	rnaB1WreckingBeast      = "181a7b5a-3794-46ae-b6ff-48849ee1470f"
	rnaB1ZhurTaaGoblin      = "1003304f-af01-4235-af82-6ef901bd00bf"
	rnaB1KnightArbiter      = "9adc1435-3094-427d-88f8-91d55c5a924d"
	rnaB1RubblebeltRecluse  = "3ef0fff2-6010-42de-a7ce-e4cac295a481"
	rnaB1KnightOfSorrows    = "50f644d6-2597-4ad6-8ea8-6b09056f90c0"
	rnaB1DebtorsTransport   = "72f89173-2998-45e6-94da-5b5bb9738a6b"
	rnaB1Ministrant         = "0cee5539-d10b-4de2-aee8-e02f0e24ac26"
	rnaB1ImperiousOligarch  = "b0381c2a-40e2-4002-82e3-ae2d687daed3"
	rnaB1OrzhovEnforcer     = "f60b879e-82b4-442d-a544-2349429fe4c9"
	rnaB1Aeromunculus       = "a0cd19c1-9125-4627-83ba-d42c489b4e7d"
	rnaB1SkatewingSpy       = "49701c15-c6f7-4b72-9e89-d34148e5c675"
	rnaB1TrollbredGuardian  = "e9c756a5-f4d4-4b84-be60-ad4bc0d1acf9"
	rnaB1CivicStalwart      = "80d7e395-f774-44b5-9442-fa2fc03aaa29"
	rnaB1SenateGriffin      = "7ddf9c12-8f75-481c-a218-d42f925c6503"
	rnaB1SylvanBrushstrider = "67a494dc-deac-43ee-9e47-20fae9788a9f"
	rnaB1WatchfulGiant      = "8239c63d-097a-449a-953a-cd480e594aa8"
	rnaB1HaazdaOfficer      = "35023c04-ae62-47bf-80a0-faa2d7a67788"
	rnaB1FaerieDuelist      = "ccbf7141-69c9-407d-b68e-c1c63ddef213"
	rnaB1ImpassionedOrator  = "c7107a2d-2dcf-42e9-9ea9-d0fc0d6d2ec6"
	rnaB1SenateGuildmage    = "76f1cc7a-084e-410d-92cb-b8f6ab8da4e9"
	rnaB1ThirstingShade     = "cba60174-a4de-41e1-a4d9-8f7bb1387e0e"
	rnaB1GyreEngineer       = "27eef5c1-259a-4d37-93af-7451f50905af"
	rnaB1EliteArrester      = "f25af8de-3edb-400b-9b6c-ec446a8314fa"
	rnaB1SenateCourier      = "4429e14f-a5f7-42e9-a8c3-b59428f92047"
	rnaB1GruulLocket        = "7018ccd0-96e6-458f-b412-197edcb2bafd"
	rnaB1RakdosLocket       = "ab784412-0722-4501-9aa0-fd68d307f6fd"
	rnaB1SimicLocket        = "c567fed8-df00-4579-901e-8754079105bf"
	rnaB1Quench             = "fa1b722a-109e-4a43-bd7b-818292728cb3"
	rnaB1Ragefire           = "91b8c0bc-2599-4423-a23f-5dc95b23cf25"
	rnaB1StormStrike        = "f0748ff2-94b1-45a9-8d56-8459bd86704a"
	rnaB1BurnBright         = "ed14c7fa-632c-47ee-9db0-8c492771121a"
	rnaB1GetThePoint        = "57c704bc-8dff-4b90-b4e1-38d12e41b3d1"
	rnaB1ExposeToDaylight   = "c8195fcd-1858-4f23-aa73-b654d25035aa"
	rnaB1GrotesqueDemise    = "eed4651b-4c45-4136-9e3d-44471031368f"
	rnaB1BringToTrial       = "968b6277-584c-4947-98ac-4d6e5f8d6754"
	rnaB1RubbleReading      = "894d96b1-c82a-4141-a2f3-76da505bd5e5"
	rnaB1AppliedBiomancy    = "93ee2ea1-4f9b-47dd-b7d4-b0ed5a26d719"
	rnaB1Deface             = "390290b3-77cc-4180-8e6c-9cb808f6ac7a"
	rnaB1SpiritOfTheSpires  = "bf10ce45-c6d6-461e-ab46-06b4205d31fb"
	rnaB1AzoriusSkyguard    = "bbfdb5eb-2140-4d50-a667-1b4ac75d03a5"

	rnaB1BearType = "Creature — Bear"
)

// rnaB1All is every card of the batch, oracle ID to name.
var rnaB1All = map[string]string{
	rnaB1AxebaneBeast: "Axebane Beast", rnaB1CatacombCrocodile: "Catacomb Crocodile",
	rnaB1CoralCommando: "Coral Commando", rnaB1FeralMaaka: "Feral Maaka",
	rnaB1ProwlingCaracal: "Prowling Caracal", rnaB1Humongulus: "Humongulus",
	rnaB1NoxiousGroodion: "Noxious Groodion", rnaB1RubbleSlinger: "Rubble Slinger",
	rnaB1GhorClanWrecker: "Ghor-Clan Wrecker", rnaB1RampagingRendhorn: "Rampaging Rendhorn",
	rnaB1WreckingBeast: "Wrecking Beast", rnaB1ZhurTaaGoblin: "Zhur-Taa Goblin",
	rnaB1KnightArbiter: "Azorius Knight-Arbiter", rnaB1RubblebeltRecluse: "Rubblebelt Recluse",
	rnaB1KnightOfSorrows: "Knight of Sorrows", rnaB1DebtorsTransport: "Debtors' Transport",
	rnaB1Ministrant: "Ministrant of Obligation", rnaB1ImperiousOligarch: "Imperious Oligarch",
	rnaB1OrzhovEnforcer: "Orzhov Enforcer", rnaB1Aeromunculus: "Aeromunculus",
	rnaB1SkatewingSpy: "Skatewing Spy", rnaB1TrollbredGuardian: "Trollbred Guardian",
	rnaB1CivicStalwart: "Civic Stalwart", rnaB1SenateGriffin: "Senate Griffin",
	rnaB1SylvanBrushstrider: "Sylvan Brushstrider", rnaB1WatchfulGiant: "Watchful Giant",
	rnaB1HaazdaOfficer: "Haazda Officer", rnaB1FaerieDuelist: "Faerie Duelist",
	rnaB1ImpassionedOrator: "Impassioned Orator", rnaB1SenateGuildmage: "Senate Guildmage",
	rnaB1ThirstingShade: "Thirsting Shade", rnaB1GyreEngineer: "Gyre Engineer",
	rnaB1EliteArrester: "Elite Arrester", rnaB1SenateCourier: "Senate Courier",
	rnaB1GruulLocket: "Gruul Locket", rnaB1RakdosLocket: "Rakdos Locket",
	rnaB1SimicLocket: "Simic Locket", rnaB1Quench: "Quench",
	rnaB1Ragefire: "Ragefire", rnaB1StormStrike: "Storm Strike",
	rnaB1BurnBright: "Burn Bright", rnaB1GetThePoint: "Get the Point",
	rnaB1ExposeToDaylight: "Expose to Daylight", rnaB1GrotesqueDemise: "Grotesque Demise",
	rnaB1BringToTrial: "Bring to Trial", rnaB1RubbleReading: "Rubble Reading",
	rnaB1AppliedBiomancy: "Applied Biomancy", rnaB1Deface: "Deface",
	rnaB1SpiritOfTheSpires: "Spirit of the Spires", rnaB1AzoriusSkyguard: "Azorius Skyguard",
}

// rnaB1Activate has p activate row index of source without the strict
// mana check, and resolves it.
func rnaB1Activate(t *testing.T, g *game.Game, p *game.Player, source uuid.UUID, index int, params game.ActivateAbilityParams) {
	t.Helper()
	if err := g.ActivateCatalogAbility(p.ID, source, index, params); err != nil {
		t.Fatalf("activate row %d: %v", index, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestRnaB1EveryCardIsRegisteredFull(t *testing.T) {
	if len(rnaB1All) != 50 {
		t.Fatalf("the batch lists %d cards, want 50", len(rnaB1All))
	}
	for oracle, name := range rnaB1All {
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

// --- bodies and keywords -----------------------------------------------------

func TestRnaB1VanillaBodiesDeclareNoAbilities(t *testing.T) {
	for _, oracle := range []string{rnaB1AxebaneBeast, rnaB1CatacombCrocodile, rnaB1CoralCommando, rnaB1FeralMaaka, rnaB1ProwlingCaracal} {
		spec, _ := effects.Lookup(oracle)
		if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 ||
			len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 {
			t.Errorf("%s is a vanilla body and declares no abilities", spec.Name)
		}
	}
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Catacomb Crocodile", "Creature — Crocodile", rnaB1CatacombCrocodile, 3, 7)
	if p, tough := effectstest.PtOf(t, g, id); p != 3 || tough != 7 {
		t.Errorf("Catacomb Crocodile is %d/%d, want 3/7", p, tough)
	}
}

func TestRnaB1PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		oracle, typeLine string
		keywords         []string
		live             []string
	}{
		{rnaB1Humongulus, "Creature — Homunculus", []string{"hexproof"}, []string{"hexproof"}},
		{rnaB1NoxiousGroodion, "Creature — Beast", []string{"deathtouch"}, []string{"deathtouch"}},
		{rnaB1RubbleSlinger, "Creature — Human Warrior", []string{"reach"}, []string{"reach"}},
		{rnaB1GhorClanWrecker, "Creature — Human Warrior", []string{game.KeywordRiot, "menace"}, []string{"menace"}},
		{rnaB1RampagingRendhorn, "Creature — Beast", []string{game.KeywordRiot}, nil},
		{rnaB1WreckingBeast, "Creature — Beast", []string{game.KeywordRiot, "trample"}, []string{"trample"}},
		{rnaB1ZhurTaaGoblin, "Creature — Goblin Berserker", []string{game.KeywordRiot}, nil},
		{rnaB1KnightArbiter, "Creature — Human Knight", []string{"vigilance"}, []string{"vigilance"}},
		{rnaB1ImperiousOligarch, "Creature — Human Cleric", []string{"vigilance"}, []string{"vigilance"}},
		{rnaB1OrzhovEnforcer, "Creature — Human Rogue", []string{"deathtouch"}, []string{"deathtouch"}},
		{rnaB1Aeromunculus, "Creature — Homunculus Mutant", []string{"flying"}, []string{"flying"}},
		{rnaB1SenateGriffin, "Creature — Griffin", []string{"flying"}, []string{"flying"}},
		{rnaB1FaerieDuelist, "Creature — Faerie Rogue", []string{"flash", "flying"}, []string{"flying"}},
		{rnaB1ThirstingShade, "Creature — Shade", []string{"lifelink"}, []string{"lifelink"}},
		{rnaB1SenateCourier, "Creature — Bird", []string{"flying"}, []string{"flying"}},
		{rnaB1SpiritOfTheSpires, "Creature — Spirit", []string{"flying"}, []string{"flying"}},
		{rnaB1AzoriusSkyguard, "Creature — Human Knight", []string{"flying", "first strike"}, []string{"flying", "first strike"}},
	} {
		name := rnaB1All[tc.oracle]
		t.Run(name, func(t *testing.T) {
			spec, _ := effects.Lookup(tc.oracle)
			if !slices.Equal(spec.PrintedKeywords, tc.keywords) {
				t.Errorf("printed keywords %v, want %v", spec.PrintedKeywords, tc.keywords)
			}
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, name, tc.typeLine, tc.oracle, 2, 2)
			for _, kw := range tc.live {
				if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s on the battlefield", name, kw)
				}
			}
		})
	}
}

// Riot (CR 702.136a): each riot creature asks as it enters, and the
// counter answer gives it a +1/+1 counter.
func TestRnaB1RiotCreaturesAskAsTheyEnter(t *testing.T) {
	for _, tc := range []struct {
		oracle, typeLine string
		power, tough     int
	}{
		{rnaB1GhorClanWrecker, "Creature — Human Warrior", 2, 2},
		{rnaB1RampagingRendhorn, "Creature — Beast", 4, 4},
		{rnaB1WreckingBeast, "Creature — Beast", 6, 6},
		{rnaB1ZhurTaaGoblin, "Creature — Goblin Berserker", 2, 2},
	} {
		name := rnaB1All[tc.oracle]
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.CastHoldCreature(t, g, name, tc.typeLine, tc.oracle, tc.power, tc.tough)
			effectstest.AnswerRiotFor(t, g, true)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 1 {
				t.Errorf("%s has %d +1/+1 counters after the counter answer, want 1", name, got)
			}
		})
	}
}

func TestRnaB1KnightArbiterCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.B12Push(g, me.ID, "Azorius Knight-Arbiter", "Creature — Human Knight", rnaB1KnightArbiter, 2, 5)
	effectstest.AssertRestrictions(t, g, id, game.CantBeBlocked)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", rnaB1BearType, 2, 2)
	if err := effectstest.PorBlock(t, g, id, opp.ID, bear); err == nil {
		t.Error("Azorius Knight-Arbiter was blocked")
	}
}

func TestRnaB1RubblebeltRecluseMustAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	recluse := effectstest.B12Push(g, me.ID, "Rubblebelt Recluse", "Creature — Ogre Berserker", rnaB1RubblebeltRecluse, 6, 5)
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != recluse {
		t.Fatalf("pass with the Recluse home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(recluse, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
}

func TestRnaB1KnightOfSorrowsBlocksTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[1].ID, "Knight of Sorrows", "Creature — Human Knight", rnaB1KnightOfSorrows, 3, 3)
	if got := effectstest.MbCapacity(t, g, id); got != 2 {
		t.Errorf("block capacity %d, want 2", got)
	}
}

// --- afterlife ---------------------------------------------------------------

func TestRnaB1AfterlifeMakesSpirits(t *testing.T) {
	for _, tc := range []struct {
		oracle, typeLine string
		spirits          int
	}{
		{rnaB1KnightOfSorrows, "Creature — Human Knight", 1},
		{rnaB1DebtorsTransport, "Creature — Thrull", 2},
		{rnaB1Ministrant, "Creature — Human Cleric", 2},
		{rnaB1ImperiousOligarch, "Creature — Human Cleric", 1},
		{rnaB1OrzhovEnforcer, "Creature — Human Rogue", 1},
	} {
		name := rnaB1All[tc.oracle]
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.B12Push(g, me.ID, name, tc.typeLine, tc.oracle, 2, 2)
			effectstest.B27Kill(g, id)
			g.RunStateChecksForTest()
			effectstest.PassPriorityAroundTable(t, g)
			if n, _ := effectstest.CountTokensNamed(g, me.ID, "Spirit"); n != tc.spirits {
				t.Errorf("%s left %d Spirit tokens, want %d", name, n, tc.spirits)
			}
		})
	}
}

// --- adapt -------------------------------------------------------------------

// Adapt (CR 701.46a): the counters go on once; a second activation with
// a counter already there adds none.
func TestRnaB1AdaptPutsCountersOnlyOnce(t *testing.T) {
	for _, tc := range []struct {
		oracle, typeLine string
		n                int
		grants           string
	}{
		{rnaB1Aeromunculus, "Creature — Homunculus Mutant", 1, ""},
		{rnaB1SkatewingSpy, "Creature — Vedalken Rogue Mutant", 2, "flying"},
		{rnaB1TrollbredGuardian, "Creature — Troll Frog Warrior", 2, "trample"},
	} {
		name := rnaB1All[tc.oracle]
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.B12Push(g, me.ID, name, tc.typeLine, tc.oracle, 2, 3)
			if tc.grants != "" && effectstest.HasEffectiveKeyword(t, g, id, tc.grants) {
				t.Errorf("%s has %s with no counter", name, tc.grants)
			}
			rnaB1Activate(t, g, me, id, 0, game.ActivateAbilityParams{})
			if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != tc.n {
				t.Fatalf("%s has %d counters after adapting, want %d", name, got, tc.n)
			}
			rnaB1Activate(t, g, me, id, 0, game.ActivateAbilityParams{})
			if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != tc.n {
				t.Errorf("%s has %d counters after adapting again, want %d", name, got, tc.n)
			}
			if tc.grants != "" && !effectstest.HasEffectiveKeyword(t, g, id, tc.grants) {
				t.Errorf("%s has no %s with a +1/+1 counter", name, tc.grants)
			}
		})
	}
}

// --- enters triggers ---------------------------------------------------------

func TestRnaB1CivicStalwartPumpsYourTeam(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", rnaB1BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", rnaB1BearType, 2, 2)
	effectstest.CastHoldCreature(t, g, "Civic Stalwart", "Creature — Elephant Soldier", rnaB1CivicStalwart, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, mine); p != 3 || tt != 3 {
		t.Errorf("my bear is %d/%d, want 3/3", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("their bear is %d/%d, want 2/2", p, tt)
	}
}

func TestRnaB1SenateGriffinScriesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.CastHoldCreature(t, g, "Senate Griffin", "Creature — Griffin", rnaB1SenateGriffin, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ScryChoiceFor(g, me.ID); c == nil || len(c.ScryCards) != 1 {
		t.Errorf("scry prompt %v, want one looking at one card", c)
	}
}

func TestRnaB1SylvanBrushstriderGainsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	life := me.Life
	effectstest.CastHoldCreature(t, g, "Sylvan Brushstrider", "Creature — Beast", rnaB1SylvanBrushstrider, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestRnaB1WatchfulGiantMakesAHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	effectstest.CastHoldCreature(t, g, "Watchful Giant", "Creature — Giant Soldier", rnaB1WatchfulGiant, 3, 6)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Human"); n != 1 {
		t.Errorf("%d Human tokens, want 1", n)
	}
}

func TestRnaB1HaazdaOfficerPumpsYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", rnaB1BearType, 2, 2)
	effectstest.RfdCast(t, g, "Haazda Officer", "Creature — Human Soldier", rnaB1HaazdaOfficer, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickIfAsked(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, mine); p != 3 || tt != 3 {
		t.Errorf("the target is %d/%d, want 3/3", p, tt)
	}
}

func TestRnaB1FaerieDuelistShrinksAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	victim := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", rnaB1BearType, 2, 2)
	effectstest.RfdCast(t, g, "Faerie Duelist", "Creature — Faerie Rogue", rnaB1FaerieDuelist, 1, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, mine) {
		t.Error("the controller's own creature is offered")
	}
	effectstest.PickIfAsked(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, victim); p != 1 || tt != 3 {
		t.Errorf("the target is %d/%d, want 1/3", p, tt)
	}
}

func TestRnaB1ImpassionedOratorGainsForAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	effectstest.B12Push(g, me.ID, "Impassioned Orator", "Creature — Human Cleric", rnaB1ImpassionedOrator, 2, 2)
	life := me.Life
	effectstest.CastHoldCreature(t, g, "Bear", rnaB1BearType, "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
}

// --- activated abilities -----------------------------------------------------

func TestRnaB1SenateGuildmageGainsAndLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mage := effectstest.B12Push(g, me.ID, "Senate Guildmage", "Creature — Human Wizard", rnaB1SenateGuildmage, 2, 2)
	life := me.Life
	rnaB1Activate(t, g, me, mage, 0, game.ActivateAbilityParams{})
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	mage = effectstest.B12Push(g, me.ID, "Senate Guildmage", "Creature — Human Wizard", rnaB1SenateGuildmage, 2, 2)
	hand := me.Hand.Size()
	rnaB1Activate(t, g, me, mage, 1, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d before the discard, want %d", got, hand+1)
	}
	if n := effectstest.DiscardOwed(g, me.ID); n != 1 {
		t.Errorf("owes %d discards, want 1", n)
	}
}

func TestRnaB1ThirstingShadePumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shade := effectstest.B12Push(g, me.ID, "Thirsting Shade", "Creature — Shade", rnaB1ThirstingShade, 1, 1)
	rnaB1Activate(t, g, me, shade, 0, game.ActivateAbilityParams{})
	if p, tt := effectstest.PtOf(t, g, shade); p != 2 || tt != 2 {
		t.Errorf("Thirsting Shade is %d/%d, want 2/2", p, tt)
	}
}

func TestRnaB1EliteArresterTapsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arrester := effectstest.B12Push(g, me.ID, "Elite Arrester", "Creature — Human Soldier", rnaB1EliteArrester, 0, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", rnaB1BearType, 2, 2)
	rnaB1Activate(t, g, me, arrester, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if !effectstest.AlliTapped(t, g, bear) {
		t.Error("the target is untapped")
	}
	if !effectstest.AlliTapped(t, g, arrester) {
		t.Error("Elite Arrester did not tap as a cost")
	}
}

func TestRnaB1SenateCourierGainsVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	courier := effectstest.B12Push(g, me.ID, "Senate Courier", "Creature — Bird", rnaB1SenateCourier, 1, 4)
	if effectstest.HasEffectiveKeyword(t, g, courier, "vigilance") {
		t.Fatal("Senate Courier has vigilance before the activation")
	}
	rnaB1Activate(t, g, me, courier, 0, game.ActivateAbilityParams{})
	if !effectstest.HasEffectiveKeyword(t, g, courier, "vigilance") {
		t.Error("Senate Courier has no vigilance after the activation")
	}
}

func TestRnaB1GyreEngineerAddsTwoMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	engineer := effectstest.B12Push(g, me.ID, "Gyre Engineer", "Creature — Vedalken Wizard", rnaB1GyreEngineer, 1, 1)
	if err := g.ActivateManaAbility(me.ID, engineer, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := me.ManaPool.Total(); got != 2 {
		t.Errorf("pool = %d, want 2", got)
	}
}

func TestRnaB1LocketsMakeManaAndSacrificeToDraw(t *testing.T) {
	for _, tc := range []struct {
		oracle string
		color  string
	}{
		{rnaB1GruulLocket, "R"},
		{rnaB1RakdosLocket, "B"},
		{rnaB1SimicLocket, "U"},
	} {
		name := rnaB1All[tc.oracle]
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			locket := effectstest.PushCatalogPermanent(g, me.ID, name, "Artifact", tc.oracle, false)
			if err := g.ActivateManaAbility(me.ID, locket, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("tap for mana: %v", err)
			}
			if pick := effectstest.RiderLatestManaPick(g, me.ID); pick != nil {
				if err := g.ResolveManaChoice(pick.ID, me.ID, tc.color); err != nil {
					t.Fatalf("ResolveManaChoice: %v", err)
				}
			}
			if got := me.ManaPool.Total(); got != 1 {
				t.Errorf("pool = %d, want 1", got)
			}

			g = effectstest.NewCatalogGame(t)
			me = g.Seats[0]
			locket = effectstest.PushCatalogPermanent(g, me.ID, name, "Artifact", tc.oracle, false)
			for i := 0; i < 4; i++ {
				me.ManaPool.AddMana(game.ManaToken{Color: tc.color})
			}
			hand := me.Hand.Size()
			if err := g.ActivateCatalogAbility(me.ID, locket, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if g.Battlefield.Contains(locket) {
				t.Error("the Locket was not sacrificed as a cost")
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size() - hand; got != 2 {
				t.Errorf("drew %d, want 2", got)
			}
		})
	}
}

// --- statics -----------------------------------------------------------------

func TestRnaB1SpiritOfTheSpiresToughensOtherFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spirit := effectstest.B12Push(g, me.ID, "Spirit of the Spires", "Creature — Spirit", rnaB1SpiritOfTheSpires, 2, 4)
	flyer := effectstest.PorFlyer(g, me.ID, "Bird")
	walker := effectstest.B12Creature(g, me.ID, "Bear", rnaB1BearType, 2, 2)
	if p, tt := effectstest.PtOf(t, g, flyer); p != 1 || tt != 2 {
		t.Errorf("the flyer is %d/%d, want 1/2", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, walker); p != 2 || tt != 2 {
		t.Errorf("the ground creature is %d/%d, want 2/2", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, spirit); p != 2 || tt != 4 {
		t.Errorf("Spirit of the Spires is %d/%d, want 2/4", p, tt)
	}
}

func TestRnaB1AzoriusSkyguardShrinksOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Azorius Skyguard", "Creature — Human Knight", rnaB1AzoriusSkyguard, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", rnaB1BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", rnaB1BearType, 2, 2)
	if p, tt := effectstest.PtOf(t, g, theirs); p != 1 || tt != 2 {
		t.Errorf("their bear is %d/%d, want 1/2", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, mine); p != 2 || tt != 2 {
		t.Errorf("my bear is %d/%d, want 2/2", p, tt)
	}
}

// --- spells ------------------------------------------------------------------

func TestRnaB1QuenchTaxesTheSpell(t *testing.T) {
	for _, pay := range []bool{false, true} {
		name := "declined"
		if pay {
			name = "paid"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, them := g.Seats[0], g.Seats[1]
			effectstest.ToMainForCost(t, g)
			victim := effectstest.CastCatalogSpell(t, g, "Their Bear", rnaB1BearType, "", nil)
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
			effectstest.CastFromSeat(t, g, them, "Quench", "Instant", rnaB1Quench, effectstest.CardRefs(victim))
			effectstest.PassUntilTaxed(t, g, me.ID)
			if pay {
				me.ManaPool.AddMana(game.ManaToken{Color: "C"})
				me.ManaPool.AddMana(game.ManaToken{Color: "C"})
			}
			effectstest.AnswerPayUnless(t, g, me.ID, pay)
			if g.Stack.Contains(victim) == !pay {
				t.Errorf("paid=%v: the spell is still on the stack = %v", pay, g.Stack.Contains(victim))
			}
		})
	}
}

func TestRnaB1RagefireDealsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.WoeB1Seats(g)
	victim := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Ragefire", "Sorcery", rnaB1Ragefire, effectstest.CardRefs(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, victim).DamageMarked; got != 3 {
		t.Errorf("the target has %d damage, want 3", got)
	}
}

func TestRnaB1StormStrikePumpsGrantsFirstStrikeAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", rnaB1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Storm Strike", "Instant", rnaB1StormStrike, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the target is %d/%d, want 3/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the target has no first strike")
	}
	if c := effectstest.ScryChoiceFor(g, me.ID); c == nil || len(c.ScryCards) != 1 {
		t.Errorf("scry prompt %v, want one looking at one card", c)
	}
}

func TestRnaB1BurnBrightPumpsOnlyYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", rnaB1BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", rnaB1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Burn Bright", "Instant", rnaB1BurnBright, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, mine); p != 4 || tt != 2 {
		t.Errorf("my bear is %d/%d, want 4/2", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("their bear is %d/%d, want 2/2", p, tt)
	}
}

// Get the Point, Expose to Daylight and Rubble Reading destroy their
// target and then scry.
func TestRnaB1DestroyThenScry(t *testing.T) {
	for _, tc := range []struct {
		oracle, typeLine string
		targetType       string
		scry             int
	}{
		{rnaB1GetThePoint, "Instant", rnaB1BearType, 1},
		{rnaB1ExposeToDaylight, "Instant", "Artifact", 1},
		{rnaB1RubbleReading, "Sorcery", "Basic Land — Forest", 2},
	} {
		name := rnaB1All[tc.oracle]
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := effectstest.WoeB1Seats(g)
			effectstest.SeedLibrary(me, "A", "B", "C")
			var victim uuid.UUID
			if tc.targetType == rnaB1BearType {
				victim = effectstest.B12Creature(g, opp.ID, "Victim", tc.targetType, 2, 2)
			} else {
				victim = effectstest.B12Permanent(g, opp.ID, "Victim", tc.targetType)
			}
			effectstest.CastCatalogSpell(t, g, name, tc.typeLine, tc.oracle, effectstest.CardRefs(victim))
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(victim) || !opp.Graveyard.Contains(victim) {
				t.Errorf("%s did not destroy its target", name)
			}
			if c := effectstest.ScryChoiceFor(g, me.ID); c == nil || len(c.ScryCards) != tc.scry {
				t.Errorf("scry prompt %v, want one looking at %d cards", c, tc.scry)
			}
		})
	}
}

func TestRnaB1PowerGatedExiles(t *testing.T) {
	for _, tc := range []struct {
		oracle, typeLine string
		legal, illegal   int
	}{
		{rnaB1GrotesqueDemise, "Instant", 3, 4},
		{rnaB1BringToTrial, "Sorcery", 4, 3},
	} {
		name := rnaB1All[tc.oracle]
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			_, opp := effectstest.WoeB1Seats(g)
			wrong := effectstest.B12Creature(g, opp.ID, "Wrong", "Creature — Giant", tc.illegal, 5)
			if err := effectstest.CastCatalogSpellErr(t, g, name, tc.typeLine, tc.oracle, effectstest.CardRefs(wrong)); err == nil {
				t.Errorf("%s accepted a creature with power %d", name, tc.illegal)
			}
			g = effectstest.NewCatalogGame(t)
			_, opp = effectstest.WoeB1Seats(g)
			right := effectstest.B12Creature(g, opp.ID, "Right", "Creature — Giant", tc.legal, 5)
			effectstest.CastCatalogSpell(t, g, name, tc.typeLine, tc.oracle, effectstest.CardRefs(right))
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(right) || !g.Exile.Contains(right) {
				t.Errorf("%s did not exile a creature with power %d", name, tc.legal)
			}
		})
	}
}

func TestRnaB1AppliedBiomancyBothModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", rnaB1BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", rnaB1BearType, 2, 2)
	err := effectstest.WthCastWithModes(t, g, "Applied Biomancy", "Instant", rnaB1AppliedBiomancy, []int{0, 1}, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Mode: 0},
		{Kind: game.TargetCard, ID: theirs, Mode: 1},
	})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, mine); p != 3 || tt != 3 {
		t.Errorf("my bear is %d/%d, want 3/3", p, tt)
	}
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("their bear was not returned to its owner's hand")
	}
}

func TestRnaB1DefaceEachMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.WoeB1Seats(g)
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	if err := effectstest.WthCastWithModes(t, g, "Deface", "Sorcery", rnaB1Deface, []int{0}, effectstest.CardRefs(relic)); err != nil {
		t.Fatalf("cast at an artifact: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact survived")
	}

	g = effectstest.NewCatalogGame(t)
	_, opp = effectstest.WoeB1Seats(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", rnaB1BearType, 2, 2)
	if err := effectstest.WthCastWithModes(t, g, "Deface", "Sorcery", rnaB1Deface, []int{1}, effectstest.CardRefs(bear)); err == nil {
		t.Error("Deface accepted a creature without defender")
	}

	g = effectstest.NewCatalogGame(t)
	_, opp = effectstest.WoeB1Seats(g)
	wall := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4, Keywords: []string{"defender"}})
	if err := effectstest.WthCastWithModes(t, g, "Deface", "Sorcery", rnaB1Deface, []int{1}, effectstest.CardRefs(wall)); err != nil {
		t.Fatalf("cast at a defender: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) {
		t.Error("the defender survived")
	}
}
