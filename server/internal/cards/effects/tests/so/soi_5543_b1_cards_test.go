package testsso

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soi_5543_b1_cards_test.go — Shadows over Innistrad, batch 1 (#5543):
// six vanilla bodies, five keyword creatures, the five tapped duals, seven
// delirium creatures, nine madness cards, five investigate cards and
// thirteen cards whose text a catalogued card already prints. Each test
// plays the real catalog entry through the engine and asserts what its
// text says. Delirium is seeded with b24SeedTypes (four is delirium,
// three is not).

const (
	soiB1FarbogRevenantOracle = "6d32139f-839f-441b-930f-151816143142"
	soiB1RancidRatsOracle     = "e3552730-fe09-443f-b432-9640b69c7bbb"

	soiB1MoldgrafOracle    = "558fe476-edbd-4f18-86ee-6d7f61331789"
	soiB1MoorlandOracle    = "f5fa9138-9640-4586-bc81-caa6b8eae8b6"
	soiB1OxOracle          = "224788aa-3876-4844-949e-918e6273151d"
	soiB1DireSwineOracle   = "61c10a4c-b1d5-4777-a813-d231980be2a7"
	soiB1ParishBladeOracle = "d1c2b124-5290-4e66-9f6b-4db4d662a74c"
	soiB1HoundOracle       = "7be2bc4e-3986-4f4f-bd5b-0a9f59b2f865"
	soiB1ScourgeWolfOracle = "fd74d0f6-bbe5-47f6-af15-a127489ca7f3"

	soiB1AlmsOracle         = "2013a774-e6ab-4f2d-a84f-fd277a637d7e"
	soiB1BitingRainOracle   = "485cfaca-9348-418b-a751-2ecb19f3f210"
	soiB1BloodmadOracle     = "c8a20fc7-025b-403a-b893-ff5efcbdc9d8"
	soiB1BrokenConcOracle   = "7b0fa80e-6b16-4a4a-9776-6a91f145fc63"
	soiB1GisasBiddingOracle = "0abc993b-18c7-4bb8-aa58-0d06a41cb48f"
	soiB1JustTheWindOracle  = "761ebd6a-8f53-46fa-b3cd-16616b773007"
	soiB1MurderousOracle    = "8fd8b4ef-69c8-43aa-9cbe-c46dba845ba8"
	soiB1TwinsOracle        = "378110e9-62d6-4dcf-a2e3-813081c82ec4"
	soiB1SenselessOracle    = "5036fe35-43c6-45af-99d5-7b79ebcfc406"

	soiB1DrownyardOracle = "c1ebbe89-8976-4ed7-a68a-8b08aac4616c"
	soiB1HumbleOracle    = "bea6606c-ea9a-4199-a8cd-b35d7c47fd7b"
	soiB1ScrutinyOracle  = "8578c462-7b2d-4b69-ad10-a2eab44ac98d"
	soiB1PressOracle     = "4960c3a0-77ca-48d8-b565-6c5f1e6522cf"
	soiB1SurviveOracle   = "e8fcc7ff-6274-4130-a3a5-99f545eeee30"

	soiB1EmberEyeOracle    = "0b21ab85-2405-4410-9624-fb3c7d11ba56"
	soiB1QuilledOracle     = "607f9177-fea0-464c-a15f-65842865546a"
	soiB1ChaplainsOracle   = "195837b4-e5e4-479d-8dab-6ac77f9fcd83"
	soiB1EtherealOracle    = "0e24c0d2-2333-425c-952e-9a13e6519cd3"
	soiB1SilverstrikeOracl = "8652fa06-33dc-4fe5-90cb-83cbfe5c681b"
	soiB1ApparatusOracle   = "ab8301da-f9f7-4d90-afae-bd7e60834c22"
	soiB1BloodseekerOracle = "bc548ecb-57c4-49ef-a7aa-312bf3665fab"
	soiB1EpitaphOracle     = "e3483dd4-0118-4143-9b55-51078a6f274a"
	soiB1LoamDryadOracle   = "1420b7db-859b-4ad5-9844-10413cc70478"
	soiB1SleepParalysisOrc = "5a8bd4a4-8b10-4eef-8341-28be2afa2c00"
	soiB1DenyOracle        = "b665a64b-4772-42dd-9fd2-fd8598e689ba"
	soiB1GrotesqueOracle   = "058a8475-649f-433a-b258-42bc03bf40ba"
	soiB1RushOracle        = "ec9975f4-dfff-4b72-9da8-8f2025f0524a"
)

// soiB1Vanilla are the batch's bodies with no rules text.
var soiB1Vanilla = map[string]string{
	"8f80a7be-3da3-4e84-82a8-dd5d4b2c3f32": "Devilthorn Fox",
	"b6578a36-c43c-4474-81bd-5711cc4f626e": "Hulking Devil",
	"1c7dd66f-a062-49ac-aa7c-1bf5ec069d73": "Seagraf Skaab",
	"c5c221e6-dafe-4661-b920-3faed9551802": "Thornhide Wolves",
	"e85d957b-35a8-48a9-8d63-90e3c680e0dd": "Vampire Noble",
	"cc2f945f-9048-4978-b61f-49eefb3d94d7": "Wicker Witch",
}

// soiB1Keyworded are the creatures whose whole text is printed keywords.
var soiB1Keyworded = map[string]struct {
	name, typeLine string
	keywords       []string
}{
	"c377c3e0-2b91-43c6-9f49-5dd6fbf94308": {"Silent Observer", "Creature — Spirit", []string{"flying"}},
	soiB1FarbogRevenantOracle:              {"Farbog Revenant", "Creature — Spirit", []string{"skulk", "lifelink"}},
	soiB1RancidRatsOracle:                  {"Rancid Rats", "Creature — Zombie Rat", []string{"skulk", "deathtouch"}},
	"d3fb0388-bfa8-4d22-831c-40a81a11d8c6": {"Sanguinary Mage", "Creature — Vampire Wizard", []string{"prowess"}},
	"d14d41a5-6f38-43c7-b95c-b4a723002c20": {"Niblis of Dusk", "Creature — Spirit", []string{"flying", "prowess"}},
}

// soiB1Duals are the five tapped duals.
var soiB1Duals = []struct{ oracle, name, a, b string }{
	{"941b0dd1-0df2-48ee-8829-615e9c3177a7", "Forsaken Sanctuary", "W", "B"},
	{"ad6a2776-801f-4743-8268-6d654122171e", "Foul Orchard", "B", "G"},
	{"c643365a-4255-4d23-adb9-0f8b456f0838", "Highland Lake", "U", "R"},
	{"90bccf66-58ec-445d-ac96-c6013054a1b4", "Stone Quarry", "R", "W"},
	{"e887fb3f-d4c9-4022-8f75-1de6ec94af96", "Woodland Stream", "G", "U"},
}

func TestSoiB1EveryCardIsRegisteredFull(t *testing.T) {
	all := map[string]string{
		soiB1MoldgrafOracle: "Moldgraf Scavenger", soiB1MoorlandOracle: "Moorland Drifter",
		soiB1OxOracle: "Inquisitor's Ox", soiB1DireSwineOracle: "Kessig Dire Swine",
		soiB1ParishBladeOracle: "Paranoid Parish-Blade", soiB1HoundOracle: "Hound of the Farbogs",
		soiB1ScourgeWolfOracle: "Scourge Wolf",
		soiB1AlmsOracle:        "Alms of the Vein", soiB1BitingRainOracle: "Biting Rain",
		soiB1BloodmadOracle: "Bloodmad Vampire", soiB1BrokenConcOracle: "Broken Concentration",
		soiB1GisasBiddingOracle: "Gisa's Bidding", soiB1JustTheWindOracle: "Just the Wind",
		soiB1MurderousOracle: "Murderous Compulsion", soiB1TwinsOracle: "Twins of Maurer Estate",
		soiB1SenselessOracle: "Senseless Rage",
		soiB1DrownyardOracle: "Drownyard Explorers", soiB1HumbleOracle: "Humble the Brute",
		soiB1ScrutinyOracle: "Jace's Scrutiny", soiB1PressOracle: "Press for Answers",
		soiB1SurviveOracle:  "Survive the Night",
		soiB1EmberEyeOracle: "Ember-Eye Wolf", soiB1QuilledOracle: "Quilled Wolf",
		soiB1ChaplainsOracle: "Chaplain's Blessing", soiB1EtherealOracle: "Ethereal Guidance",
		soiB1SilverstrikeOracl: "Silverstrike", soiB1ApparatusOracle: "Explosive Apparatus",
		soiB1BloodseekerOracle: "Ravenous Bloodseeker", soiB1EpitaphOracle: "Epitaph Golem",
		soiB1LoamDryadOracle: "Loam Dryad", soiB1SleepParalysisOrc: "Sleep Paralysis",
		soiB1DenyOracle: "Deny Existence", soiB1GrotesqueOracle: "Grotesque Mutation",
		soiB1RushOracle: "Rush of Adrenaline",
	}
	for oracle, name := range soiB1Vanilla {
		all[oracle] = name
	}
	for oracle, k := range soiB1Keyworded {
		all[oracle] = k.name
	}
	for _, d := range soiB1Duals {
		all[d.oracle] = d.name
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

// --- bodies, keywords and lands ----------------------------------------------

func TestSoiB1VanillaBodiesDeclareNoAbilities(t *testing.T) {
	for oracle, name := range soiB1Vanilla {
		spec, _ := effects.Lookup(oracle)
		if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 ||
			len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 || spec.Madness != "" {
			t.Errorf("%s is a vanilla body and declares no abilities", name)
		}
	}
}

func TestSoiB1KeywordedCreaturesHaveTheirKeywords(t *testing.T) {
	for oracle, k := range soiB1Keyworded {
		t.Run(k.name, func(t *testing.T) {
			spec, _ := effects.Lookup(oracle)
			if !slices.Equal(spec.PrintedKeywords, k.keywords) {
				t.Errorf("printed keywords %v, want %v", spec.PrintedKeywords, k.keywords)
			}
			if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.Activated) != 0 {
				t.Errorf("%s declares more than its keywords", k.name)
			}
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, k.name, k.typeLine, oracle, 1, 1)
			for _, kw := range k.keywords {
				if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s on the battlefield", k.name, kw)
				}
			}
		})
	}
}

// Skulk: a creature with greater power can't block Rancid Rats; one with
// equal power can (CR 702.118b).
func TestSoiB1RancidRatsSkulksPastABiggerBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rats := effectstest.B12Push(g, me.ID, "Rancid Rats", "Creature — Zombie Rat", soiB1RancidRatsOracle, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.PorBlock(t, g, rats, opp.ID, bear); err == nil {
		t.Error("a 2-power creature blocked Rancid Rats")
	}
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	rats = effectstest.B12Push(g, me.ID, "Rancid Rats", "Creature — Zombie Rat", soiB1RancidRatsOracle, 1, 1)
	squirrel := effectstest.B12Creature(g, opp.ID, "Squirrel", "Creature — Squirrel", 1, 1)
	if err := effectstest.PorBlock(t, g, rats, opp.ID, squirrel); err != nil {
		t.Errorf("a 1-power creature couldn't block Rancid Rats: %v", err)
	}
}

func TestSoiB1DualsEnterTappedAndTapForTheirTwoColours(t *testing.T) {
	for _, l := range soiB1Duals {
		t.Run(l.name, func(t *testing.T) {
			spec, _ := effects.Lookup(l.oracle)
			if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{"+l.a+"|"+l.b+"}" {
				t.Errorf("mana abilities %+v, want one producing {%s} or {%s}", spec.ManaAbilities, l.a, l.b)
			}
			g := effectstest.NewCatalogGame(t)
			id := effectstest.PlayLandFromHand(t, g, l.name, l.oracle)
			effectstest.Top100AssertEnteredTapped(t, g, id, l.name)
		})
	}
}

// --- delirium ------------------------------------------------------------------

func TestSoiB1DeliriumCreaturesChangeOnlyWithFourCardTypes(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, tough           int
		dPower, dTough         int
		keyword                string
	}{
		{"Moldgraf Scavenger", "Creature — Fungus", soiB1MoldgrafOracle, 0, 4, 3, 4, ""},
		{"Moorland Drifter", "Creature — Spirit", soiB1MoorlandOracle, 2, 2, 2, 2, "flying"},
		{"Inquisitor's Ox", "Creature — Ox", soiB1OxOracle, 2, 5, 3, 5, "vigilance"},
		{"Kessig Dire Swine", "Creature — Boar Horror", soiB1DireSwineOracle, 6, 6, 6, 6, "trample"},
		{"Paranoid Parish-Blade", "Creature — Human Soldier", soiB1ParishBladeOracle, 3, 2, 4, 2, "first strike"},
		{"Hound of the Farbogs", "Creature — Zombie Dog", soiB1HoundOracle, 5, 3, 5, 3, "menace"},
		{"Scourge Wolf", "Creature — Wolf Horror", soiB1ScourgeWolfOracle, 2, 2, 2, 2, "double strike"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
			effectstest.B24SeedTypes(me, 3)
			effectstest.S29Bump(g)
			if p, tou := effectstest.PtOf(t, g, id); p != tc.power || tou != tc.tough {
				t.Errorf("three card types: %d/%d, want %d/%d", p, tou, tc.power, tc.tough)
			}
			if tc.keyword != "" && effectstest.HasKeywordOf(g, id, tc.keyword) {
				t.Errorf("three card types: has %s", tc.keyword)
			}
			effectstest.GyCard(me, "Old Forest", "Basic Land — Forest", "", "")
			effectstest.S29Bump(g)
			if p, tou := effectstest.PtOf(t, g, id); p != tc.dPower || tou != tc.dTough {
				t.Errorf("four card types: %d/%d, want %d/%d", p, tou, tc.dPower, tc.dTough)
			}
			if tc.keyword != "" && !effectstest.HasKeywordOf(g, id, tc.keyword) {
				t.Errorf("four card types: no %s", tc.keyword)
			}
		})
	}
}

func TestSoiB1ScourgeWolfAlwaysHasFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wolf := effectstest.B12Push(g, me.ID, "Scourge Wolf", "Creature — Wolf Horror", soiB1ScourgeWolfOracle, 2, 2)
	if !effectstest.HasKeywordOf(g, wolf, "first strike") {
		t.Error("Scourge Wolf has no first strike without delirium")
	}
}

// --- madness -------------------------------------------------------------------

func TestSoiB1MadnessCardsGrowBothHalves(t *testing.T) {
	for _, tc := range []struct{ name, oracle, cost string }{
		{"Alms of the Vein", soiB1AlmsOracle, "{B}"},
		{"Biting Rain", soiB1BitingRainOracle, "{2}{B}"},
		{"Bloodmad Vampire", soiB1BloodmadOracle, "{1}{R}"},
		{"Broken Concentration", soiB1BrokenConcOracle, "{3}{U}"},
		{"Gisa's Bidding", soiB1GisasBiddingOracle, "{2}{B}"},
		{"Just the Wind", soiB1JustTheWindOracle, "{U}"},
		{"Murderous Compulsion", soiB1MurderousOracle, "{1}{B}"},
		{"Twins of Maurer Estate", soiB1TwinsOracle, "{2}{B}"},
		{"Senseless Rage", soiB1SenselessOracle, "{1}{R}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spec, _ := effects.Lookup(tc.oracle)
			if spec.Madness != tc.cost {
				t.Errorf("madness cost %q, want %q", spec.Madness, tc.cost)
			}
			reps, trigs := effectstest.MadnessAbilitiesFor(tc.oracle)
			if reps != 1 || trigs != 1 {
				t.Errorf("madness halves: %d replacements, %d triggers; want 1 and 1 (CR 702.35a)", reps, trigs)
			}
		})
	}
}

// Gisa's Bidding looted away and cast from exile for {2}{B}.
func TestSoiB1GisasBiddingIsCastForItsMadnessCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bidding := effectstest.TorB1LootIntoExile(t, g, "Gisa's Bidding", "Sorcery", soiB1GisasBiddingOracle)
	effectstest.DcMana(t, g, me.ID, "{C}{C}{B}")
	if err := g.CastSpell(me.ID, bidding, game.CastSpellParams{
		Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyMadness,
	}); err != nil {
		t.Fatalf("the madness cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Zombie"); n != 2 {
		t.Errorf("Zombie tokens %d, want 2", n)
	}
}

// Twins of Maurer Estate looted away and cast from exile for {2}{B}.
func TestSoiB1TwinsOfMaurerEstateIsCastForItsMadnessCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	twins := effectstest.TorB1LootIntoExile(t, g, "Twins of Maurer Estate", "Creature — Vampire", soiB1TwinsOracle)
	effectstest.DcMana(t, g, me.ID, "{C}{C}{B}")
	if err := g.CastSpell(me.ID, twins, game.CastSpellParams{
		Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyMadness,
	}); err != nil {
		t.Fatalf("the madness cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(twins) {
		t.Error("the madness-cast Twins are not on the battlefield")
	}
}

func TestSoiB1AlmsOfTheVeinDrainsAnOpponentForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Alms of the Vein", "Sorcery", soiB1AlmsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err == nil {
		t.Fatal("Alms of the Vein targeted its caster")
	}
	mine, theirs := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Alms of the Vein", "Sorcery", soiB1AlmsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine+3 || opp.Life != theirs-3 {
		t.Errorf("life: me %d → %d, opponent %d → %d; want +3 and -3", mine, me.Life, theirs, opp.Life)
	}
}

func TestSoiB1BitingRainShrinksEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "My Giant", "Creature — Giant", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Biting Rain", "Sorcery", soiB1BitingRainOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tou := effectstest.PtOf(t, g, mine); p != 1 || tou != 1 {
		t.Errorf("my giant is %d/%d, want 1/1", p, tou)
	}
	if p, tou := effectstest.PtOf(t, g, theirs); p != 2 || tou != 2 {
		t.Errorf("their giant is %d/%d, want 2/2", p, tou)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the 2/2 survived -2/-2")
	}
}

func TestSoiB1BloodmadVampireGrowsOnCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vamp := effectstest.B12Push(g, me.ID, "Bloodmad Vampire", "Creature — Vampire Berserker", soiB1BloodmadOracle, 4, 1)
	effectstest.DealCombatDamageToPlayer(g, vamp, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tou := effectstest.PtOf(t, g, vamp); p != 5 || tou != 2 {
		t.Errorf("Bloodmad Vampire is %d/%d, want 5/2", p, tou)
	}
}

func TestSoiB1BrokenConcentrationAndDenyExistenceCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	fox := effectstest.CastCatalogSpell(t, g, "Devilthorn Fox", "Creature — Fox", "8f80a7be-3da3-4e84-82a8-dd5d4b2c3f32", nil)
	deny := effectstest.PushCatalogHandCard(opp, "Deny Existence", "Instant", soiB1DenyOracle)
	if err := g.CastSpell(opp.ID, deny, game.CastSpellParams{Targets: effectstest.CardTarget(fox)}); err != nil {
		t.Fatalf("cast Deny Existence at a creature spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fox) || me.Graveyard.Contains(fox) || !g.Exile.Contains(fox) {
		t.Error("the creature spell Deny Existence countered is not in exile")
	}

	guidance := effectstest.CastCatalogSpell(t, g, "Ethereal Guidance", "Sorcery", soiB1EtherealOracle, nil)
	deny2 := effectstest.PushCatalogHandCard(opp, "Deny Existence", "Instant", soiB1DenyOracle)
	if err := g.CastSpell(opp.ID, deny2, game.CastSpellParams{Targets: effectstest.CardTarget(guidance)}); err == nil {
		t.Fatal("Deny Existence accepted a sorcery as its target creature spell")
	}
	broken := effectstest.PushCatalogHandCard(opp, "Broken Concentration", "Instant", soiB1BrokenConcOracle)
	if err := g.CastSpell(opp.ID, broken, game.CastSpellParams{Targets: effectstest.CardTarget(guidance)}); err != nil {
		t.Fatalf("cast Broken Concentration: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(guidance) {
		t.Error("the countered sorcery is not in its owner's graveyard")
	}
}

func TestSoiB1JustTheWindBouncesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Just the Wind", "Instant", soiB1JustTheWindOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the creature is not in its owner's hand")
	}
}

func TestSoiB1MurderousCompulsionDestroysOnlyATappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	untapped := effectstest.B12Creature(g, opp.ID, "Awake Bear", "Creature — Bear", 2, 2)
	tapped := effectstest.B12Creature(g, opp.ID, "Tired Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == tapped {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if err := effectstest.CastCatalogSpellErr(t, g, "Murderous Compulsion", "Sorcery", soiB1MurderousOracle, effectstest.CardTarget(untapped)); err == nil {
		t.Fatal("Murderous Compulsion targeted an untapped creature")
	}
	effectstest.CastCatalogSpell(t, g, "Murderous Compulsion", "Sorcery", soiB1MurderousOracle, effectstest.CardTarget(tapped))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped creature was not destroyed")
	}
}

func TestSoiB1SenselessRageGivesPlusTwoPlusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Senseless Rage", effectstest.AuraTypeLine, soiB1SenselessOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tou := effectstest.PtOf(t, g, bear); p != 4 || tou != 4 {
		t.Errorf("the enchanted Bear is %d/%d, want 4/4", p, tou)
	}
}

// --- investigate -----------------------------------------------------------------

func TestSoiB1DrownyardExplorersInvestigatesOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Drownyard Explorers", "Creature — Human Wizard", soiB1DrownyardOracle, 2, 4)
	if n := effectstest.TmtB1Count(g, me.ID, "Clue"); n != 1 {
		t.Errorf("Clues %d, want 1", n)
	}
}

func TestSoiB1HumbleTheBruteDestroysOnlyABigCreatureAndInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	small := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	big := effectstest.B12Creature(g, opp.ID, "Giant", "Creature — Giant", 4, 4)
	if err := effectstest.CastCatalogSpellErr(t, g, "Humble the Brute", "Instant", soiB1HumbleOracle, effectstest.CardTarget(small)); err == nil {
		t.Fatal("Humble the Brute targeted a 3-power creature")
	}
	effectstest.CastCatalogSpell(t, g, "Humble the Brute", "Instant", soiB1HumbleOracle, effectstest.CardTarget(big))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the 4-power creature was not destroyed")
	}
	if n := effectstest.TmtB1Count(g, me.ID, "Clue"); n != 1 {
		t.Errorf("Clues %d, want 1", n)
	}
}

func TestSoiB1JacesScrutinyShrinksPowerAndInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	giant := effectstest.B12Creature(g, opp.ID, "Giant", "Creature — Giant", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Jace's Scrutiny", "Instant", soiB1ScrutinyOracle, effectstest.CardTarget(giant))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tou := effectstest.PtOf(t, g, giant); p != 0 || tou != 4 {
		t.Errorf("the Giant is %d/%d, want 0/4", p, tou)
	}
	if n := effectstest.TmtB1Count(g, me.ID, "Clue"); n != 1 {
		t.Errorf("Clues %d, want 1", n)
	}
}

func TestSoiB1PressForAnswersTapsHoldsAndInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	oppSeat := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Press for Answers", "Sorcery", soiB1PressOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, bear); !c.Tapped {
		t.Fatal("the creature was not tapped")
	}
	if n := effectstest.TmtB1Count(g, me.ID, "Clue"); n != 1 {
		t.Errorf("Clues %d, want 1", n)
	}
	effectstest.AdvanceToUpkeepOf(t, g, oppSeat)
	if c, _ := effectstest.BattlefieldCard(g, bear); !c.Tapped {
		t.Error("the creature untapped during its controller's next untap step")
	}
}

func TestSoiB1SurviveTheNightPumpsGrantsIndestructibleAndInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Survive the Night", "Instant", soiB1SurviveOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tou := effectstest.PtOf(t, g, bear); p != 3 || tou != 2 {
		t.Errorf("the Bear is %d/%d, want 3/2", p, tou)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the Bear has no indestructible")
	}
	if n := effectstest.TmtB1Count(g, me.ID, "Clue"); n != 1 {
		t.Errorf("Clues %d, want 1", n)
	}
}

// --- the rest ------------------------------------------------------------------

func TestSoiB1WolvesPumpThemselves(t *testing.T) {
	for _, tc := range []struct {
		name, oracle   string
		power, tough   int
		mana           []string
		wantP, wantT   int
		printedKeyword string
	}{
		{"Ember-Eye Wolf", soiB1EmberEyeOracle, 1, 2, []string{"C", "R"}, 3, 2, "haste"},
		{"Quilled Wolf", soiB1QuilledOracle, 2, 2, []string{"C", "C", "C", "C", "C", "G"}, 6, 6, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			wolf := effectstest.B12Push(g, me.ID, tc.name, "Creature — Wolf", tc.oracle, tc.power, tc.tough)
			if tc.printedKeyword != "" && !effectstest.HasEffectiveKeyword(t, g, wolf, tc.printedKeyword) {
				t.Errorf("%s has no %s", tc.name, tc.printedKeyword)
			}
			effectstest.AdvanceToMain(t, g)
			effectstest.B06AddMana(me, tc.mana[:len(tc.mana)-1]...)
			if err := g.ActivateCatalogAbility(me.ID, wolf, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Fatal("activated one mana short")
			}
			effectstest.B06AddMana(me, tc.mana[len(tc.mana)-1])
			if err := g.ActivateCatalogAbility(me.ID, wolf, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if p, tou := effectstest.PtOf(t, g, wolf); p != tc.wantP || tou != tc.wantT {
				t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tou, tc.wantP, tc.wantT)
			}
		})
	}
}

func TestSoiB1ChaplainsBlessingGainsFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Chaplain's Blessing", "Sorcery", soiB1ChaplainsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life %d → %d, want +5", life, me.Life)
	}
}

func TestSoiB1EtherealGuidancePumpsOnlyYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Ethereal Guidance", "Sorcery", soiB1EtherealOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tou := effectstest.PtOf(t, g, mine); p != 4 || tou != 3 {
		t.Errorf("my Bear is %d/%d, want 4/3", p, tou)
	}
	if p, tou := effectstest.PtOf(t, g, theirs); p != 2 || tou != 2 {
		t.Errorf("their Bear is %d/%d, want 2/2", p, tou)
	}
}

func TestSoiB1SilverstrikeDestroysAnAttackerAndGainsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	effectstest.SocB34AttackWith(t, g, opp.ID, attacker)
	strike := effectstest.PushCatalogHandCard(me, "Silverstrike", "Instant", soiB1SilverstrikeOracl)
	if err := g.CastSpell(me.ID, strike, game.CastSpellParams{Targets: effectstest.CardTarget(idle)}); err == nil {
		t.Fatal("Silverstrike targeted a creature that isn't attacking")
	}
	life := me.Life
	if err := g.CastSpell(me.ID, strike, game.CastSpellParams{Targets: effectstest.CardTarget(attacker)}); err != nil {
		t.Fatalf("cast Silverstrike at the attacker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacking creature was not destroyed")
	}
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
}

func TestSoiB1ExplosiveApparatusSacrificesForTwoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	app := effectstest.PushCatalogPermanent(g, me.ID, "Explosive Apparatus", "Artifact", soiB1ApparatusOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C")
	life := opp.Life
	effectstest.B16Activate(t, g, me.ID, app, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if opp.Life != life-2 {
		t.Errorf("opponent life %d → %d, want -2", life, opp.Life)
	}
	if g.Battlefield.Contains(app) || !me.Graveyard.Contains(app) {
		t.Error("Explosive Apparatus was not sacrificed")
	}
}

func TestSoiB1RavenousBloodseekerDiscardsForPlusTwoMinusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seeker := effectstest.B12Push(g, me.ID, "Ravenous Bloodseeker", "Creature — Vampire Berserker", soiB1BloodseekerOracle, 1, 3)
	effectstest.AdvanceToMain(t, g)
	card := effectstest.PushCatalogHandCard(me, "Fodder", "Sorcery", "")
	if err := g.ActivateCatalogAbility(me.ID, seeker, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{card}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tou := effectstest.PtOf(t, g, seeker); p != 3 || tou != 1 {
		t.Errorf("Ravenous Bloodseeker is %d/%d, want 3/1", p, tou)
	}
	if me.Hand.Contains(card) {
		t.Error("the card was not discarded")
	}
}

func TestSoiB1EpitaphGolemBottomsACardFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	golem := effectstest.B12Push(g, me.ID, "Epitaph Golem", "Artifact Creature — Golem", soiB1EpitaphOracle, 3, 5)
	effectstest.SeedLibrary(me, "Top")
	dead := effectstest.PushGraveyardPermanent(me, "Dead Thing", "Creature — Bear", "{1}{G}")
	theirs := effectstest.PushGraveyardPermanent(opp, "Their Dead Thing", "Creature — Bear", "{1}{G}")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, golem, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)}); err == nil {
		t.Error("a card in an opponent's graveyard was accepted")
	}
	effectstest.B16Activate(t, g, me.ID, golem, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(dead)})
	if me.Graveyard.Contains(dead) {
		t.Error("the card is still in the graveyard")
	}
	if got := effectstest.LibraryBottom(me, 1); len(got) != 1 || got[0] != dead {
		t.Errorf("library bottom %v, want the card", got)
	}
}

func TestSoiB1LoamDryadTapsItselfAndACreatureForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dryad := effectstest.B12Push(g, me.ID, "Loam Dryad", "Creature — Dryad Horror", soiB1LoamDryadOracle, 1, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := effectstest.PushCatalogPermanent(g, me.ID, "My Land", "Land", "", false)
	for name, ids := range map[string][]uuid.UUID{
		"nothing else": nil,
		"itself":       {dryad},
		"a land":       {land},
	} {
		if err := g.ActivateManaAbility(me.ID, dryad, 0, game.ManaAbilityParams{TapIDs: ids}); err == nil {
			t.Fatalf("%s paid the cost", name)
		}
	}
	effectstest.ActivateManaFor(t, g, me.ID, dryad, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{bear}})
	effectstest.ResolveManaPickAnyColor(t, g, me.ID)
	effectstest.AssertTapped(t, g, true, dryad, bear)
	if me.ManaPool.Total() != 1 {
		t.Errorf("mana pool = %v, want one mana", me.ManaPool)
	}
}

func TestSoiB1SleepParalysisTapsAndHoldsItsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.PushPermanentForTest(g, me.ID, "Host", "", "Creature — Bear")
	aura := effectstest.CastCatalogSpell(t, g, "Sleep Paralysis", effectstest.AuraTypeLine, soiB1SleepParalysisOrc,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.AttachmentHostOf(t, g, aura); got.Kind != game.TargetCard || got.ID != host {
		t.Fatalf("Sleep Paralysis attached to %+v, want host %s", got, host)
	}
	if c, _ := effectstest.BattlefieldCard(g, host); !c.Tapped {
		t.Fatal("Sleep Paralysis's enters trigger did not tap its enchanted creature")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if c, _ := effectstest.BattlefieldCard(g, host); !c.Tapped {
		t.Error("enchanted creature untapped during its controller's untap step")
	}
}

func TestSoiB1PumpsGrantTheirKeyword(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, keyword string
		power, tough          int
	}{
		{"Grotesque Mutation", soiB1GrotesqueOracle, "lifelink", 5, 3},
		{"Rush of Adrenaline", soiB1RushOracle, "trample", 4, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
			effectstest.CastCatalogSpell(t, g, tc.name, "Instant", tc.oracle, effectstest.CardTarget(bear))
			effectstest.PassPriorityAroundTable(t, g)
			if p, tou := effectstest.PtOf(t, g, bear); p != tc.power || tou != tc.tough {
				t.Errorf("the Bear is %d/%d, want %d/%d", p, tou, tc.power, tc.tough)
			}
			if !effectstest.HasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("the Bear has no %s", tc.keyword)
			}
		})
	}
}
