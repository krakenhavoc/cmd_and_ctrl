package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gpt_5599_b2_cards_test.go — Guildpact, batch 2 (#5599). Each test plays
// the real catalog entry and asserts what its text says.

const (
	gptB2CrystalSeerOracle   = "58838588-13d5-4948-babf-d09cc59d76a1"
	gptB2CullingSunOracle    = "e41d9b3a-0712-4174-8c5c-224065a63b16"
	gptB2DouseOracle         = "1fd754fd-c053-4638-b379-46d903357b01"
	gptB2DrownedOracle       = "62e86e38-2f8d-4cee-aa7e-4519add79d4d"
	gptB2DryadOracle         = "7cf59255-0ad7-4ca8-bc95-f3a8cc2643d5"
	gptB2FrazzleOracle       = "55c4df81-9513-424b-aa58-10190797bba7"
	gptB2GhorClanOracle      = "e3367ed8-16a3-4b04-bc6a-e0a3c794e6d3"
	gptB2GhostWardenOracle   = "66bd5999-742a-43e0-827c-ecb8ec7db242"
	gptB2SolifugeOracle      = "c590c1bd-430c-4515-9343-ba818c02bc5f"
	gptB2GigadrowseOracle    = "9cc64a91-82d2-4be1-812f-531863e99801"
	gptB2GristlebackOracle   = "36f4d359-f46c-4da4-a8c2-1fe1799ac3f7"
	gptB2HarrierOracle       = "65dc3191-0eb2-4c10-bf2c-ee0e24888870"
	gptB2ChronarchOracle     = "1da438f3-db1c-4713-a60c-e078f31d809c"
	gptB2LeapOracle          = "cf3b1070-2c08-49b9-aa78-fe676fab38aa"
	gptB2LionheartOracle     = "5058b9f1-9941-495d-9eed-554ae2f6d500"
	gptB2MartyredOracle      = "21bf5ae1-d7a0-406d-88fd-009dfa919096"
	gptB2MourningOracle      = "0406a534-c69c-40bb-801f-228a662ac1a2"
	gptB2OgreSavantOracle    = "2ce12817-a7b5-475a-87d0-718d77349a6d"
	gptB2OrzhovaOracle       = "8551a9cf-c54b-42d4-92d6-550f4890a3d7"
	gptB2OstiaryOracle       = "1c99c9d9-3ea8-43bc-96d7-4dea12770b5c"
	gptB2PrimevalOracle      = "b416d7ac-ec56-40e2-ba3d-bcecd1b73175"
	gptB2PyromaticsOracle    = "9c01e2a6-61e7-455c-b09f-8c0de2266782"
	gptB2PatriarchOracle     = "942f6260-055e-4cc7-bec6-3a46fd4b048e"
	gptB2SlumOracle          = "77275fc5-5f70-4d7f-938e-f32b5c19c532"
	gptB2RuneboggleOracle    = "22000392-99e9-4927-a488-1bcb576f6275"
	gptB2SchismotivateOracle = "d7081ec5-0070-4bc1-97c4-544ac08581ee"
	gptB2ScorchedOracle      = "dd05fd8e-f25e-4b9e-b162-afea1b248282"
	gptB2GrotesqueOracle     = "20c8b03c-303d-4e19-96ec-3fabae843352"
	gptB2SilhanaOracle       = "3b4b6af2-f929-4082-a5ae-61a4e074bf14"
	gptB2PitSkulkOracle      = "e2ce99b9-7098-4de1-95bf-609bd448848d"
	gptB2SkybreakerOracle    = "0a45d8d4-2a5e-4b5b-b430-fc2fd5fd82c0"
	gptB2StarvedOracle       = "0a57ac2e-276b-41ea-89f4-287f530114c6"
	gptB2SteamcoreOracle     = "e8f6caf4-4c6c-468a-9388-88c323ffe30b"
	gptB2ZeppelidOracle      = "f20640db-c5d4-4e00-9903-a33ff4f7820d"
	gptB2StreetbreakerOracle = "38bdb4f8-012e-460c-ad20-e51bec1e6745"
	gptB2HooliganOracle      = "a33073bf-66c1-425b-9f25-748b343f33fe"
	gptB2ToArmsOracle        = "90c97fc4-597c-45cf-96c6-3ba9558ede0d"
	gptB2TorchDrakeOracle    = "8408f99a-aa3c-40d0-9e26-dd527343e62b"
	gptB2TrainOracle         = "90ba0366-57ff-4dc4-a290-3925548129ce"
	gptB2DragonautsOracle    = "e9e56b1d-5ad2-4807-8c4c-1a1a6a25ecbf"
	gptB2CantorOracle        = "24472823-d33c-4d96-add1-233116f6f4f7"
	gptB2WithstandOracle     = "ebf2816b-e042-43e9-8734-d3a896334c69"
	gptB2WreakHavocOracle    = "b1d2519e-7340-4ea4-a62e-cf0740d4e57a"
)

func TestGptB2EveryCardIsRegisteredFull(t *testing.T) {
	names := map[string]string{
		gptB2CrystalSeerOracle:   "Crystal Seer",
		gptB2CullingSunOracle:    "Culling Sun",
		gptB2DouseOracle:         "Douse in Gloom",
		gptB2DrownedOracle:       "Drowned Rusalka",
		gptB2DryadOracle:         "Dryad Sophisticate",
		gptB2FrazzleOracle:       "Frazzle",
		gptB2GhorClanOracle:      "Ghor-Clan Bloodscale",
		gptB2GhostWardenOracle:   "Ghost Warden",
		gptB2SolifugeOracle:      "Giant Solifuge",
		gptB2GigadrowseOracle:    "Gigadrowse",
		gptB2GristlebackOracle:   "Gristleback",
		gptB2HarrierOracle:       "Harrier Griffin",
		gptB2ChronarchOracle:     "Izzet Chronarch",
		gptB2LeapOracle:          "Leap of Flame",
		gptB2LionheartOracle:     "Lionheart Maverick",
		gptB2MartyredOracle:      "Martyred Rusalka",
		gptB2MourningOracle:      "Mourning Thrull",
		gptB2OgreSavantOracle:    "Ogre Savant",
		gptB2OrzhovaOracle:       "Orzhova, the Church of Deals",
		gptB2OstiaryOracle:       "Ostiary Thrull",
		gptB2PrimevalOracle:      "Primeval Light",
		gptB2PyromaticsOracle:    "Pyromatics",
		gptB2PatriarchOracle:     "Revenant Patriarch",
		gptB2SlumOracle:          "Rumbling Slum",
		gptB2RuneboggleOracle:    "Runeboggle",
		gptB2SchismotivateOracle: "Schismotivate",
		gptB2ScorchedOracle:      "Scorched Rusalka",
		gptB2GrotesqueOracle:     "Shrieking Grotesque",
		gptB2SilhanaOracle:       "Silhana Starfletcher",
		gptB2PitSkulkOracle:      "Skarrgan Pit-Skulk",
		gptB2SkybreakerOracle:    "Skarrgan Skybreaker",
		gptB2StarvedOracle:       "Starved Rusalka",
		gptB2SteamcoreOracle:     "Steamcore Weird",
		gptB2ZeppelidOracle:      "Stratozeppelid",
		gptB2StreetbreakerOracle: "Streetbreaker Wurm",
		gptB2HooliganOracle:      "Tin Street Hooligan",
		gptB2ToArmsOracle:        "To Arms!",
		gptB2TorchDrakeOracle:    "Torch Drake",
		gptB2TrainOracle:         "Train of Thought",
		gptB2DragonautsOracle:    "Wee Dragonauts",
		gptB2CantorOracle:        "Wild Cantor",
		gptB2WithstandOracle:     "Withstand",
		gptB2WreakHavocOracle:    "Wreak Havoc",
	}
	if len(names) != 43 {
		t.Fatalf("%d cards in the batch, want 43", len(names))
	}
	for oracle, name := range names {
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

// gptB2CastPaid casts a creature from the active seat's hand in a main
// phase, strictly, off the given floating mana (one rune per mana), so
// the record of the mana spent is exact. The caller settles the stack.
func gptB2CastPaid(t *testing.T, g *game.Game, name, typeLine, cost, oracle, pool string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
			Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID,
		})
	})
	effectstest.FloatForTest(g, me, pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast %s off %s: %v", name, pool, err)
	}
	return id
}

// gptB2CardRef is a battlefield or stack card as a target.
func gptB2CardRef(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// --- vanilla and keywords ------------------------------------------------------------

func TestGptB2KeywordCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.GptB1Seats(g)
	sol := effectstest.B12Push(g, me.ID, "Giant Solifuge", "Creature — Insect", gptB2SolifugeOracle, 4, 1)
	for _, kw := range []string{"trample", "haste", "shroud"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, sol, kw) {
			t.Errorf("Giant Solifuge has no %s", kw)
		}
	}
	spec, _ := effects.Lookup(gptB2StreetbreakerOracle)
	if len(spec.PrintedKeywords) != 0 || len(spec.Activated) != 0 || len(spec.Triggered) != 0 || len(spec.Static) != 0 {
		t.Errorf("Streetbreaker Wurm declares rules: %+v", spec)
	}
}

func TestGptB2DryadSophisticateIsUnblockableAgainstANonbasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	dryad := effectstest.B12Push(g, me.ID, "Dryad Sophisticate", "Creature — Dryad", gptB2DryadOracle, 2, 1)
	bear := effectstest.B12Creature(g, them.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B12Permanent(g, them.ID, "Their Guildgate", "Land — Gate")
	if !effectstest.HasKeywordOnBattlefield(t, g, dryad, "nonbasic landwalk") {
		t.Fatal("the Dryad has no nonbasic landwalk")
	}
	effectstest.AttackIntoBlocks(t, g, them.ID, dryad)
	if r := effectstest.BlockRefusal(t, g, bear, dryad); r != game.BlockReasonLandwalk {
		t.Errorf("blocking the Dryad against a nonbasic land: reason %q, want landwalk", r)
	}
}

func TestGptB2StratozeppelidBlocksOnlyFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	zepp := effectstest.B12Push(g, me.ID, "Stratozeppelid", "Creature — Beast", gptB2ZeppelidOracle, 4, 4)
	ground := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	flyer := effectstest.TypedCreature(g, opp.ID, "Their Bird", "Creature — Bird", "", 2, 2, "flying")
	if !effectstest.HasKeywordOnBattlefield(t, g, zepp, "flying") {
		t.Error("Stratozeppelid has no flying")
	}
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		z, _ := g.LookupCardForEffect(zepp)
		gr, _ := g.LookupCardForEffect(ground)
		fl, _ := g.LookupCardForEffect(flyer)
		if r := g.BlockPairRefusalLocked(&gr, &z); r.Legal() {
			t.Error("Stratozeppelid blocked a creature without flying")
		}
		if r := g.BlockPairRefusalLocked(&fl, &z); !r.Legal() {
			t.Errorf("Stratozeppelid may not block a flyer: %q", r.Reason)
		}
	})
}

// --- self-pumps ---------------------------------------------------------------------

func TestGptB2GhorClanBloodscalePumpsOnlyOnceATurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.GptB1Seats(g)
	scale := effectstest.B12Push(g, me.ID, "Ghor-Clan Bloodscale", "Creature — Lizard Warrior", gptB2GhorClanOracle, 2, 1)
	if !effectstest.HasKeywordOnBattlefield(t, g, scale, "first strike") {
		t.Error("the Bloodscale has no first strike")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, scale, 0, game.ActivateAbilityParams{}, "C", "C", "C", "G")
	if p, tough := effectstest.PtOf(t, g, scale); p != 4 || tough != 3 {
		t.Fatalf("after one activation the Bloodscale is %d/%d, want 4/3", p, tough)
	}
	if err := effectstest.RavB2Activate(t, g, me.ID, scale, 0, game.ActivateAbilityParams{}, "C", "C", "C", "G"); err == nil {
		t.Error("the Bloodscale pumped twice in one turn")
	}
}

func TestGptB2LionheartMaverickAndTorchDrakePumpThemselves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.GptB1Seats(g)
	lion := effectstest.B12Push(g, me.ID, "Lionheart Maverick", "Creature — Human Knight", gptB2LionheartOracle, 1, 1)
	drake := effectstest.B12Push(g, me.ID, "Torch Drake", "Creature — Drake", gptB2TorchDrakeOracle, 2, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, lion, "vigilance") {
		t.Error("Lionheart Maverick has no vigilance")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, drake, "flying") {
		t.Error("Torch Drake has no flying")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, lion, 0, game.ActivateAbilityParams{}, "C", "C", "C", "C", "W")
	if p, tough := effectstest.PtOf(t, g, lion); p != 2 || tough != 3 {
		t.Errorf("the Maverick is %d/%d, want 2/3", p, tough)
	}
	effectstest.RavB2MustActivate(t, g, me.ID, drake, 0, game.ActivateAbilityParams{}, "C", "R")
	effectstest.RavB2MustActivate(t, g, me.ID, drake, 0, game.ActivateAbilityParams{}, "C", "R")
	if p, tough := effectstest.PtOf(t, g, drake); p != 4 || tough != 2 {
		t.Errorf("the Drake after two activations is %d/%d, want 4/2", p, tough)
	}
}

// --- tap abilities ------------------------------------------------------------------

func TestGptB2GhostWardenPumpsAndOstiaryThrullTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	warden := effectstest.B12Push(g, me.ID, "Ghost Warden", "Creature — Spirit", gptB2GhostWardenOracle, 1, 1)
	thrull := effectstest.B12Push(g, me.ID, "Ostiary Thrull", "Creature — Thrull", gptB2OstiaryOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.RavB2MustActivate(t, g, me.ID, warden, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the Warden's target is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.B16Tapped(t, g, warden) {
		t.Error("the Warden's ability did not tap it")
	}

	effectstest.RavB2MustActivate(t, g, me.ID, thrull, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)}, "W")
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Error("the Thrull's target is untapped")
	}
}

// --- the Rusalkas: sacrifice a creature ---------------------------------------------

func TestGptB2StarvedRusalkaGainsALife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.GptB1Seats(g)
	rus := effectstest.B12Push(g, me.ID, "Starved Rusalka", "Creature — Spirit", gptB2StarvedOracle, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 2, 2)
	life := me.Life
	effectstest.RavB2MustActivate(t, g, me.ID, rus, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}, "G")
	if g.Battlefield.Contains(bear) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
}

func TestGptB2ScorchedRusalkaPingsAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	rus := effectstest.B12Push(g, me.ID, "Scorched Rusalka", "Creature — Spirit", gptB2ScorchedOracle, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 2, 2)
	life := opp.Life
	effectstest.RavB2MustActivate(t, g, me.ID, rus, 0,
		game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}, Targets: effectstest.PlayerRefs(opp.ID)}, "R")
	if opp.Life != life-1 {
		t.Errorf("target life %d, want %d", opp.Life, life-1)
	}
}

func TestGptB2DrownedRusalkaRummages(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.GptB1Seats(g)
	rus := effectstest.B12Push(g, me.ID, "Drowned Rusalka", "Creature — Spirit", gptB2DrownedOracle, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 2, 2)
	effectstest.PushHandCard(g, me)
	effectstest.RavB2MustActivate(t, g, me.ID, rus, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}, "U")
	if g.Battlefield.Contains(bear) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if n := effectstest.DiscardOwed(g, me.ID); n != 1 {
		t.Errorf("the controller owes %d discards, want 1", n)
	}
}

func TestGptB2MartyredRusalkaStopsAnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	rus := effectstest.B12Push(g, me.ID, "Martyred Rusalka", "Creature — Spirit", gptB2MartyredOracle, 1, 1)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 2, 2)
	held := effectstest.B12Creature(g, me.ID, "Held", "Creature — Bear", 2, 2)
	free := effectstest.B12Creature(g, me.ID, "Free", "Creature — Bear", 2, 2)
	effectstest.RavB2MustActivate(t, g, me.ID, rus, 0,
		game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Targets: effectstest.CardRefs(held)}, "W")
	effectstest.AdvanceToStepInTurn(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(held, them.ID); err == nil {
		t.Error("the target attacked")
	}
	if err := g.DeclareAttacker(free, them.ID); err != nil {
		t.Errorf("another creature was refused: %v", err)
	}
}

// --- sacrifice this creature ----------------------------------------------------------

func TestGptB2GristlebackLifeAmountIsItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.GptB1Seats(g)
	boar := effectstest.B12Push(g, me.ID, "Gristleback", "Creature — Boar Beast", gptB2GristlebackOracle, 2, 2)
	life := me.Life
	effectstest.RavB2MustActivate(t, g, me.ID, boar, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(boar) {
		t.Error("Gristleback was not sacrificed")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestGptB2SkarrganSkybreakerFlingsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	sky := effectstest.B12Push(g, me.ID, "Skarrgan Skybreaker", "Creature — Giant Shaman", gptB2SkybreakerOracle, 3, 3)
	life := opp.Life
	effectstest.RavB2MustActivate(t, g, me.ID, sky, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)}, "C")
	if g.Battlefield.Contains(sky) {
		t.Error("the Skybreaker was not sacrificed")
	}
	if opp.Life != life-3 {
		t.Errorf("target life %d, want %d", opp.Life, life-3)
	}
}

// --- bloodthirst ------------------------------------------------------------------------

func TestGptB2BloodthirstCards(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, t, n                int
	}{
		{"Skarrgan Pit-Skulk", "Creature — Human Warrior", gptB2PitSkulkOracle, 1, 1, 1},
		{"Skarrgan Skybreaker", "Creature — Giant Shaman", gptB2SkybreakerOracle, 3, 3, 3},
		{"Gristleback", "Creature — Boar Beast", gptB2GristlebackOracle, 2, 2, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			_, opp := effectstest.GptB1Seats(g)
			effectstest.GptB1DamagePlayer(g, opp.ID, 1)
			id := effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.p, tc.t)
			if n := effectstest.CounterCount(g, id, game.CounterPlusOne); n != tc.n {
				t.Errorf("%d +1/+1 counters, want %d", n, tc.n)
			}
		})
	}
}

func TestGptB2PitSkulkCantBeBlockedBySmallerCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.GptB1Seats(g)
	effectstest.GptB1DamagePlayer(g, opp.ID, 1)
	skulk := effectstest.CastHoldCreature(t, g, "Skarrgan Pit-Skulk", "Creature — Human Warrior", gptB2PitSkulkOracle, 1, 1)
	one := effectstest.B12Creature(g, opp.ID, "One", "Creature — Elf", 1, 1)
	two := effectstest.B12Creature(g, opp.ID, "Two", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		s, _ := g.LookupCardForEffect(skulk)
		o, _ := g.LookupCardForEffect(one)
		w, _ := g.LookupCardForEffect(two)
		if r := g.BlockPairRefusalLocked(&s, &o); r.Legal() {
			t.Error("a 1-power creature blocked the 2-power Pit-Skulk")
		}
		if r := g.BlockPairRefusalLocked(&s, &w); !r.Legal() {
			t.Errorf("a 2-power creature may not block the 2-power Pit-Skulk: %q", r.Reason)
		}
	})
}

// --- enters triggers -----------------------------------------------------------------------

func TestGptB2CrystalSeerLooksAtFourAndBouncesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.GptB1Seats(g)
	effectstest.SeedLibrary(me, "First", "Second", "Third", "Fourth")
	seer := effectstest.CastHoldCreature(t, g, "Crystal Seer", "Creature — Vedalken Wizard", gptB2CrystalSeerOracle, 2, 2)
	c := effectstest.LookAtTopChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the Seer queued no look-at-the-top prompt")
	}
	if len(c.ScryCards) != 4 {
		t.Fatalf("looked at %d cards, want 4", len(c.ScryCards))
	}
	if err := g.ResolveLookAtTop(c.ID, me.ID, c.ScryCards); err != nil {
		t.Fatalf("ResolveLookAtTop: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RavB2MustActivate(t, g, me.ID, seer, 0, game.ActivateAbilityParams{}, "C", "C", "C", "C", "U")
	if !me.Hand.Contains(seer) {
		t.Error("the Seer did not return to its owner's hand")
	}
}

func TestGptB2IzzetChronarchReturnsAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.GptB1Seats(g)
	spell := effectstest.PushGraveyardCardTyped(me, "Spent Bolt", "Instant")
	effectstest.CastHoldCreature(t, g, "Izzet Chronarch", "Creature — Human Wizard", gptB2ChronarchOracle, 2, 2)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: spell}, true)
	if !me.Hand.Contains(spell) {
		t.Error("the instant did not return to hand")
	}
}

// The off-colour cycle: each triggers only when its colour paid for it.
func TestGptB2SpentCycleTriggersOnlyOffItsColour(t *testing.T) {
	type check func() bool
	for _, tc := range []struct {
		name, typeLine, cost, oracle string
		with, without                string
		p, t                         int
		setup                        func(g *game.Game, me, opp *game.Player) (game.TargetRef, check)
	}{
		{"Steamcore Weird", "Creature — Weird", "{3}{U}", gptB2SteamcoreOracle, "URCC", "UCCC", 1, 3,
			func(_ *game.Game, _, opp *game.Player) (game.TargetRef, check) {
				life := opp.Life
				return effectstest.PbPlayer(opp.ID), func() bool { return opp.Life == life-2 }
			}},
		{"Tin Street Hooligan", "Creature — Goblin Rogue", "{1}{R}", gptB2HooliganOracle, "RG", "RC", 2, 1,
			func(g *game.Game, _, opp *game.Player) (game.TargetRef, check) {
				rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
				return gptB2CardRef(rock), func() bool { return !g.Battlefield.Contains(rock) }
			}},
		{"Ogre Savant", "Creature — Ogre Wizard", "{4}{R}", gptB2OgreSavantOracle, "RUCCC", "RCCCC", 3, 2,
			func(g *game.Game, _, opp *game.Player) (game.TargetRef, check) {
				bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
				return gptB2CardRef(bear), func() bool { return opp.Hand.Contains(bear) }
			}},
		{"Shrieking Grotesque", "Creature — Gargoyle", "{2}{W}", gptB2GrotesqueOracle, "WBC", "WCC", 2, 1,
			func(g *game.Game, _, opp *game.Player) (game.TargetRef, check) {
				return effectstest.PbPlayer(opp.ID), func() bool { return effectstest.DiscardOwed(g, opp.ID) == 1 }
			}},
		{"Revenant Patriarch", "Creature — Spirit", "{4}{B}", gptB2PatriarchOracle, "BWCCC", "BCCCC", 4, 3,
			func(g *game.Game, _, opp *game.Player) (game.TargetRef, check) {
				return effectstest.PbPlayer(opp.ID), func() bool { return g.NextCombatPhaseSkipsForLocked(opp.ID) == 1 }
			}},
	} {
		t.Run(tc.name+" with its colour", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := effectstest.GptB1Seats(g)
			pick, happened := tc.setup(g, me, opp)
			gptB2CastPaid(t, g, tc.name, tc.typeLine, tc.cost, tc.oracle, tc.with, tc.p, tc.t)
			effectstest.AssistCardSettle(t, g, me.ID, &pick, true)
			if !happened() {
				t.Error("the enters trigger did nothing with its colour spent")
			}
		})
		t.Run(tc.name+" without its colour", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := effectstest.GptB1Seats(g)
			_, happened := tc.setup(g, me, opp)
			id := gptB2CastPaid(t, g, tc.name, tc.typeLine, tc.cost, tc.oracle, tc.without, tc.p, tc.t)
			// A nil pick fails the test if a target prompt opens.
			effectstest.AssistCardSettle(t, g, me.ID, nil, true)
			if !g.Battlefield.Contains(id) {
				t.Fatal("the creature did not resolve")
			}
			if happened() {
				t.Error("the enters trigger fired without its colour spent")
			}
		})
	}
}

func TestGptB2RevenantPatriarchCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	pat := effectstest.B12Push(g, opp.ID, "Revenant Patriarch", "Creature — Spirit", gptB2PatriarchOracle, 4, 3)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		a, _ := g.LookupCardForEffect(bear)
		b, _ := g.LookupCardForEffect(pat)
		if r := g.BlockPairRefusalLocked(&a, &b); r.Legal() {
			t.Error("Revenant Patriarch may block")
		}
	})
}

// --- combat and cast triggers ----------------------------------------------------------------

func TestGptB2MourningThrullGainsTheDamageItDeals(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	thrull := effectstest.B12Push(g, me.ID, "Mourning Thrull", "Creature — Thrull", gptB2MourningOracle, 1, 1)
	if !effectstest.HasKeywordOnBattlefield(t, g, thrull, "flying") {
		t.Error("Mourning Thrull has no flying")
	}
	life := me.Life
	effectstest.Pr7bAttack(t, g, opp.ID, thrull)
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("controller life %d, want %d", me.Life, life+1)
	}
}

func TestGptB2WeeDragonautsGrowsOnInstants(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.GptB1Seats(g)
	naut := effectstest.B12Push(g, me.ID, "Wee Dragonauts", "Creature — Faerie Wizard", gptB2DragonautsOracle, 1, 3)
	effectstest.CastCatalogSpell(t, g, "Quick Study", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, naut); p != 3 || tough != 3 {
		t.Errorf("after one instant the Dragonauts are %d/%d, want 3/3", p, tough)
	}
}

// --- upkeep triggers ---------------------------------------------------------------------------

func TestGptB2HarrierGriffinTapsAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, victimOwner := g.Seats[1], g.Seats[2]
	effectstest.B12Push(g, owner.ID, "Harrier Griffin", "Creature — Griffin", gptB2HarrierOracle, 3, 3)
	victim := effectstest.B12Creature(g, victimOwner.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AssistCardSettle(t, g, owner.ID, &game.TargetRef{Kind: game.TargetCard, ID: victim}, true)
	if !effectstest.B16Tapped(t, g, victim) {
		t.Error("the Griffin's target is untapped")
	}
}

func TestGptB2RumblingSlumDamagesEachPlayerAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	effectstest.B12Push(g, owner.ID, "Rumbling Slum", "Creature — Elemental", gptB2SlumOracle, 5, 5)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Life != before[p.ID]-1 {
			t.Errorf("player %s life %d, want %d", p.ID, p.Life, before[p.ID]-1)
		}
	}
}

// --- mana ---------------------------------------------------------------------------------------

func TestGptB2SilhanaStarfletcherTapsForTheChosenColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.GptB1Seats(g)
	cast := effectstest.CastHoldCreature(t, g, "Silhana Starfletcher", "Creature — Elf Druid Archer", gptB2SilhanaOracle, 1, 3)
	effectstest.AnswerColor(t, g, me.ID, "W")
	if c := effectstest.B12Card(t, g, cast); c.ChosenColor != "W" {
		t.Errorf("chosen colour %q, want W", c.ChosenColor)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, cast, "reach") {
		t.Error("Silhana has no reach")
	}

	elf := effectstest.B12Push(g, me.ID, "Silhana Starfletcher", "Creature — Elf Druid Archer", gptB2SilhanaOracle, 1, 3)
	g.WithWriteLock(func() { g.QueueColorChoiceForEffect(me.ID, elf, "Silhana Starfletcher", nil, game.ColorForMana) })
	effectstest.AnswerColor(t, g, me.ID, "G")
	if err := g.ActivateManaAbility(me.ID, elf, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool = %v, want [G]", got)
	}
}

func TestGptB2WildCantorSacrificesForAnyColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.GptB1Seats(g)
	effectstest.AdvanceToMain(t, g)
	cantor := effectstest.B12Push(g, me.ID, "Wild Cantor", "Creature — Human Druid", gptB2CantorOracle, 1, 1)
	if err := g.ActivateManaAbility(me.ID, cantor, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("sacrifice for mana: %v", err)
	}
	if pick := effectstest.PendingOfKind(g, game.PendingChoiceMana); pick != nil {
		if err := g.ResolveManaChoice(pick.ID, me.ID, "B"); err != nil {
			t.Fatalf("choose a colour: %v", err)
		}
	}
	if g.Battlefield.Contains(cantor) {
		t.Error("the Cantor was not sacrificed")
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("the pool holds %d mana, want 1", me.ManaPool.Total())
	}
}

func TestGptB2OrzhovaTapsForColourlessAndDrains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	church := effectstest.B12Push(g, me.ID, "Orzhova, the Church of Deals", "Land", gptB2OrzhovaOracle, 0, 0)
	mine, theirs := me.Life, opp.Life
	effectstest.RavB2MustActivate(t, g, me.ID, church, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)}, "C", "C", "C", "W", "B")
	if opp.Life != theirs-1 || me.Life != mine+1 {
		t.Errorf("life: target %d (want %d), controller %d (want %d)", opp.Life, theirs-1, me.Life, mine+1)
	}
	if !effectstest.B16Tapped(t, g, church) {
		t.Error("the drain did not tap the land")
	}
	effectstest.Untap(g, church)
	if err := g.ActivateManaAbility(me.ID, church, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool = %v, want [C]", got)
	}
}

// --- instants and sorceries ----------------------------------------------------------------------

func TestGptB2DouseInGloomDamagesAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 3)
	life := me.Life
	effectstest.CastAt(t, g, "Douse in Gloom", "Instant", gptB2DouseOracle, gptB2CardRef(bear))
	c, ok := effectstest.BattlefieldCard(g, bear)
	if !ok || c.DamageMarked != 2 {
		t.Errorf("the 2/3 target: on battlefield %v, damage %d, want 2", ok, c.DamageMarked)
	}
	if me.Life != life+2 {
		t.Errorf("caster life %d, want %d", me.Life, life+2)
	}
}

func TestGptB2SchismotivateSplitsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	effectstest.CastAt(t, g, "Schismotivate", "Instant", gptB2SchismotivateOracle, gptB2CardRef(mine), gptB2CardRef(theirs))
	if p, tough := effectstest.PtOf(t, g, mine); p != 6 || tough != 2 {
		t.Errorf("the first target is %d/%d, want 6/2", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 1 || tough != 5 {
		t.Errorf("the second target is %d/%d, want 1/5", p, tough)
	}
}

func TestGptB2LeapOfFlamePumpsAndGrants(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.GptB1Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastAt(t, g, "Leap of Flame", "Instant", gptB2LeapOracle, gptB2CardRef(bear))
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the target is %d/%d, want 3/2", p, tough)
	}
	for _, kw := range []string{"flying", "first strike"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, bear, kw) {
			t.Errorf("the target has no %s", kw)
		}
	}
}

func TestGptB2GigadrowseTapsTargetPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.GptB1Seats(g)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.CastAt(t, g, "Gigadrowse", "Instant", gptB2GigadrowseOracle, gptB2CardRef(rock))
	if !effectstest.B16Tapped(t, g, rock) {
		t.Error("the target is untapped")
	}
}

func TestGptB2PyromaticsReplicatesOntoASecondTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a, b := g.Seats[(g.Turn.ActiveSeat+1)%4], g.Seats[(g.Turn.ActiveSeat+2)%4]
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.AvB24Hand(g, me, game.Card{Name: "Pyromatics", TypeLine: "Instant", ManaCost: "{1}{R}", OracleID: gptB2PyromaticsOracle})
	effectstest.FloatForTest(g, me, "RRCC")
	lifeA, lifeB := a.Life, b.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: effectstest.PlayerRefs(a.ID), OptionalCosts: []int{0}, Strict: true}); err != nil {
		t.Fatalf("cast with one replicate: %v", err)
	}
	if it := effectstest.TriggerOnStack(g, spell); it == nil || it.XValue != 1 {
		t.Fatalf("replicate trigger = %+v, want one counting one payment", it)
	}
	for i := 0; i < 64; i++ {
		if effectstest.StackFullyEmpty(g) && effectstest.LatestPickTarget(g, me.ID) == nil {
			break
		}
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, effectstest.PbPlayer(b.ID)); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if a.Life != lifeA-1 || b.Life != lifeB-1 {
		t.Errorf("life: first target %d (want %d), copy's target %d (want %d)", a.Life, lifeA-1, b.Life, lifeB-1)
	}
}

func TestGptB2TrainOfThoughtDrawsOncePerCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.AvB24Hand(g, me, game.Card{Name: "Train of Thought", TypeLine: "Sorcery", ManaCost: "{1}{U}", OracleID: gptB2TrainOracle})
	effectstest.FloatForTest(g, me, "UUUCCC")
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{OptionalCosts: []int{0, 0}, Strict: true}); err != nil {
		t.Fatalf("cast with two replicates: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand-1+3 {
		t.Errorf("hand %d, want %d (the spell and two copies each draw one)", got, hand-1+3)
	}
}

func TestGptB2ToArmsUntapsAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.GptB1Seats(g)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.B16Tap(g, a)
	effectstest.B16Tap(g, b)
	hand := me.Hand.Size()
	effectstest.CastAt(t, g, "To Arms!", "Instant", gptB2ToArmsOracle)
	if effectstest.B16Tapped(t, g, a) || effectstest.B16Tapped(t, g, b) {
		t.Error("a creature is still tapped")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestGptB2WithstandPreventsThreeAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.GptB1Seats(g)
	hand := me.Hand.Size()
	effectstest.CastAt(t, g, "Withstand", "Instant", gptB2WithstandOracle, effectstest.PbPlayer(me.ID))
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	life := me.Life
	effectstest.GptB1DamagePlayer(g, me.ID, 4)
	if me.Life != life-1 {
		t.Errorf("after 4 damage with 3 prevented, life %d, want %d", me.Life, life-1)
	}
}

func TestGptB2WreakHavocDestroysALandAndCantBeCountered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.GptB1Seats(g)
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	effectstest.CastAt(t, g, "Wreak Havoc", "Sorcery", gptB2WreakHavocOracle, gptB2CardRef(land))
	if g.Battlefield.Contains(land) {
		t.Error("the land survived")
	}
	if spec, _ := effects.Lookup(gptB2WreakHavocOracle); !spec.CantBeCountered {
		t.Error("Wreak Havoc can be countered")
	}
}

func TestGptB2CullingSunDestroysTheCheapCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	three := effectstest.B27Push(g, opp.ID, "Three", "Creature — Bear", "", "{2}{G}", 2, 2, "G")
	one := effectstest.B27Push(g, me.ID, "One", "Creature — Elf", "", "{G}", 1, 1, "G")
	four := effectstest.B27Push(g, opp.ID, "Four", "Creature — Giant", "", "{3}{R}", 4, 4, "R")
	effectstest.CastAt(t, g, "Culling Sun", "Sorcery", gptB2CullingSunOracle)
	if g.Battlefield.Contains(three) || g.Battlefield.Contains(one) {
		t.Error("a creature with mana value 3 or less survived")
	}
	if !g.Battlefield.Contains(four) {
		t.Error("the mana value 4 creature was destroyed")
	}
}

func TestGptB2PrimevalLightDestroysOnlyTheTargetsEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	mine := effectstest.B12Permanent(g, me.ID, "My Enchantment", "Enchantment")
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.CastAt(t, g, "Primeval Light", "Sorcery", gptB2PrimevalOracle, effectstest.PbPlayer(opp.ID))
	if g.Battlefield.Contains(theirs) {
		t.Error("the target's enchantment survived")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(rock) {
		t.Error("a permanent outside the sweep was destroyed")
	}
}

// gptB2SpellOnStack puts a sorcery with the given colours on the stack
// for the active seat and hands priority to the next seat.
func gptB2SpellOnStack(t *testing.T, g *game.Game, colors ...string) (me, opp *game.Player, victim uuid.UUID) {
	t.Helper()
	me, opp = effectstest.GptB1Seats(g)
	effectstest.ToMainForCost(t, g)
	victim = effectstest.HandCardFull(me, "Their Spell", "Sorcery", "", "", colors)
	if err := g.CastSpell(me.ID, victim, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the victim: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	return me, opp, victim
}

func TestGptB2FrazzleCountersOnlyNonblue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, blue := gptB2SpellOnStack(t, g, "U")
	bad := effectstest.HandCardFull(opp, "Frazzle", "Instant", "", gptB2FrazzleOracle, []string{"U"})
	if err := g.CastSpell(opp.ID, bad, game.CastSpellParams{Targets: effectstest.CardRefs(blue)}); err == nil {
		t.Fatal("Frazzle targeted a blue spell")
	}

	g = effectstest.NewCatalogGame(t)
	me, opp, red := gptB2SpellOnStack(t, g, "R")
	effectstest.CastFromSeat(t, g, opp, "Frazzle", "Instant", gptB2FrazzleOracle, effectstest.CardRefs(red))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(red) || !me.Graveyard.Contains(red) {
		t.Error("the red spell was not countered")
	}
}

func TestGptB2RuneboggleCountersUnlessPaidThenDraws(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp, victim := gptB2SpellOnStack(t, g)
		hand := opp.Hand.Size()
		effectstest.CastFromSeat(t, g, opp, "Runeboggle", "Instant", gptB2RuneboggleOracle, effectstest.CardRefs(victim))
		effectstest.PassUntilTaxed(t, g, me.ID)
		if pay {
			me.ManaPool.AddMana(game.ManaToken{Color: "C"})
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		countered := me.Graveyard.Contains(victim) && !g.Stack.Contains(victim)
		if countered == pay {
			t.Errorf("paid %v: countered %v", pay, countered)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Hand.Size() != hand+1 {
			t.Errorf("paid %v: caster's hand %d, want %d", pay, opp.Hand.Size(), hand+1)
		}
	}
}
