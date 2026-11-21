package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// iko_5512_b1_cards_test.go — Ikoria: Lair of Behemoths, batch 1
// (#5512). Every test plays the real catalog entry through the engine.

const (
	ikoGloomPangolinOracle      = "bcbdeccc-2baa-4158-bf87-8b88e7a19248"
	ikoSavaiSabertoothOracle    = "24f2cf61-22db-4e34-898f-6baa2957d3ae"
	ikoManedServalOracle        = "ac230f29-6d3b-4fe1-bcb0-6ffab95a1bd7"
	ikoMosscoatGoriakOracle     = "28cc015f-1f3e-417a-9762-de25109fee5c"
	ikoImposingVantasaurOracle  = "a0d5fce0-a439-496a-9fcf-9271503f3ce8"
	ikoLavaSerpentOracle        = "4d94d807-97d9-4cdf-a63c-d8b82c2acc6e"
	ikoIndathaCrystalOracle     = "3617c4ec-dc70-4fa8-94f3-fbc48abcaeea"
	ikoKetriaCrystalOracle      = "3ee5d7b6-0263-4239-a5fb-de1e1a220c55"
	ikoRaugrinCrystalOracle     = "ae4a8470-4517-42c7-a7f6-ef7635e10fe5"
	ikoSavaiCrystalOracle       = "cd0c1165-673d-4936-8254-deedd0ebbe97"
	ikoZagothCrystalOracle      = "87edb566-8929-4961-9c03-d3657a8c6feb"
	ikoBoonOfTheWishGiverOracle = "1720d38a-c3c1-4294-85c1-a91fc7179d6f"
	ikoNeutralizeOracle         = "c2c60973-b4d5-43fa-a0c4-b5688b28df05"
	ikoWiltOracle               = "77b6c3bb-bd05-4547-8f3e-9c50879973f1"
	ikoEasyPreyOracle           = "b783246a-e00b-4cdf-9111-fe0afb896a8e"
	ikoRakingClawsOracle        = "c2248d3d-3157-4610-868d-2b3fae2b1ec8"
	ikoCoordinatedChargeOracle  = "688a0b7d-bf00-45ab-ab3f-9d19f52f019a"
	ikoHamperingSnareOracle     = "4fd3dafe-e6a4-49d5-950a-e2fcc77eda98"
	ikoSuffocatingFumesOracle   = "80c7b1b4-4e37-4aef-9432-d629d71b7045"
	ikoFrostveilAmbushOracle    = "5cf069cf-3a10-4427-a02b-5405429b5652"
	ikoGoForBloodOracle         = "11cf103c-f8a3-4892-942d-2c4472dd5716"
	ikoBarrierBreachOracle      = "581425e6-3128-4f83-9efc-17028fc65078"
	ikoBladeBanishOracle        = "d8838e16-96ca-4f73-ad61-7e98a76c4baf"
	ikoClashOfTitansOracle      = "198c05ca-fb2d-4f15-a28f-1dad36466a4f"
	ikoUnexpectedFangsOracle    = "dfa8415d-2474-46ca-b92f-6e12e3268cf5"
	ikoFullyGrownOracle         = "c98163f4-4a78-43f8-9855-4a23ebb9ceeb"
	ikoHeightenedReflexesOracle = "513fd896-ab78-44fb-b136-f7c93288dc63"
	ikoSpontaneousFlightOracle  = "04da8c46-9703-47f2-974b-c144464ec380"
	ikoSuddenSpinneretsOracle   = "5b83fede-851c-4e9d-98ba-d88f6caffc1b"
	ikoForbiddenFriendshipOracl = "d6192840-5e5d-48fa-a256-3b1e8c1c307a"
	ikoDaysquadMarshalOracle    = "42de2d46-c98d-4ae5-a53d-97531592529b"
	ikoGarrisonCatOracle        = "e7bc7c44-946c-4131-b764-add0c5f21020"
	ikoHoneyMammothOracle       = "41da0e8c-ba86-446d-9719-667695c7f1cc"
	ikoExcavationMoleOracle     = "68e0b691-f332-4b0a-8173-bb8a48b463a1"
	ikoFacetReaderOracle        = "7aefe7a9-d696-4e2e-b1a5-d041d9e4e213"
	ikoCheckpointOfficerOracle  = "03e25120-b4e4-409e-8dbc-7574ea4c78d6"
	ikoAlmightyBrushwaggOracle  = "d8a3a89d-9d27-40ae-a1aa-d55cc6a79b0b"
	ikoAdaptiveShimmererOracle  = "8003ca2d-0c1d-47c1-bcd1-4609615c7a82"
	ikoFarfinderOracle          = "7d306de8-a93b-429b-9436-0501be97dc8d"
	ikoAvianOddityOracle        = "d6a03a52-7e85-46de-920c-93a3f6479c4c"
	ikoVoidBeckonerOracle       = "7185e345-d8a1-4dd3-a071-f90f265634b4"
	ikoTitanothRexOracle        = "8656a32b-eb95-402b-9daf-0b6d876b4b13"
	ikoSplendorMareOracle       = "fd2a00d1-9f21-4d54-9458-024fa03354b8"
	ikoSanctuarySmasherOracle   = "29925297-b6d4-4fc8-a585-1425c1482ec6"
	ikoDreamtailHeronOracle     = "1cadb1e1-19b6-4ce5-8eb5-7851cec7dfa3"
	ikoVulpikeetOracle          = "11d0c8a4-1dd0-4707-9c1d-19ed27d1934e"
	ikoGlowstoneRecluseOracle   = "b2b76070-f103-4f83-ad21-119e6352f2cd"
	ikoMajesticAuricornOracle   = "9c13f77a-800c-4fb6-b54a-89a96e12d279"
	ikoMysteriousEggOracle      = "d3b62851-6013-49e5-8808-fa61b1bdfe98"

	ikoBearType = "Creature — Bear"
)

// ikoCreature puts a non-catalog creature with a mana cost on the
// battlefield for owner.
func ikoCreature(g *game.Game, owner uuid.UUID, name, manaCost string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: ikoBearType, ManaCost: manaCost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// ikoTokensNamed counts owner's creature tokens with the given name.
func ikoTokensNamed(g *game.Game, owner uuid.UUID, name string) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == owner && c.IsToken() && c.IsCreature() {
				n++
			}
		}
	})
	return n
}

// ikoMutatorInHand puts a catalog mutate card in owner's hand at a main
// phase.
func ikoMutatorInHand(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	return effectstest.RprMutatorInHand(t, g, owner, game.Card{
		Name: name, TypeLine: typeLine, OracleID: oracle,
		Keywords: []string{"mutate"}, Power: power, Toughness: toughness,
	})
}

// --- every card is in the catalog ------------------------------------------

func TestIko5512EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		ikoGloomPangolinOracle, ikoSavaiSabertoothOracle, ikoManedServalOracle, ikoMosscoatGoriakOracle,
		ikoImposingVantasaurOracle, ikoLavaSerpentOracle, ikoIndathaCrystalOracle, ikoKetriaCrystalOracle,
		ikoRaugrinCrystalOracle, ikoSavaiCrystalOracle, ikoZagothCrystalOracle, ikoBoonOfTheWishGiverOracle,
		ikoNeutralizeOracle, ikoWiltOracle, ikoEasyPreyOracle, ikoRakingClawsOracle,
		ikoCoordinatedChargeOracle, ikoHamperingSnareOracle, ikoSuffocatingFumesOracle, ikoFrostveilAmbushOracle,
		ikoGoForBloodOracle, ikoBarrierBreachOracle, ikoBladeBanishOracle, ikoClashOfTitansOracle,
		ikoUnexpectedFangsOracle, ikoFullyGrownOracle, ikoHeightenedReflexesOracle, ikoSpontaneousFlightOracle,
		ikoSuddenSpinneretsOracle, ikoForbiddenFriendshipOracl, ikoDaysquadMarshalOracle, ikoGarrisonCatOracle,
		ikoHoneyMammothOracle, ikoExcavationMoleOracle, ikoFacetReaderOracle, ikoCheckpointOfficerOracle,
		ikoAlmightyBrushwaggOracle, ikoAdaptiveShimmererOracle, ikoFarfinderOracle, ikoAvianOddityOracle,
		ikoVoidBeckonerOracle, ikoTitanothRexOracle, ikoSplendorMareOracle, ikoSanctuarySmasherOracle,
		ikoDreamtailHeronOracle, ikoVulpikeetOracle, ikoGlowstoneRecluseOracle, ikoMajesticAuricornOracle,
		ikoMysteriousEggOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// --- vanilla and printed keywords ----------------------------------------------

func TestIko5512VanillaCreaturesHaveNoAbilities(t *testing.T) {
	for _, oracle := range []string{ikoGloomPangolinOracle, ikoSavaiSabertoothOracle} {
		spec, _ := effects.Lookup(oracle)
		if len(spec.PrintedKeywords)+len(spec.Triggered)+len(spec.Activated)+len(spec.Static) != 0 {
			t.Errorf("%s is vanilla but declares abilities", spec.Name)
		}
	}
}

func TestIko5512KeywordCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
	}{
		{"Maned Serval", "Creature — Cat", ikoManedServalOracle, "vigilance"},
		{"Mosscoat Goriak", "Creature — Beast", ikoMosscoatGoriakOracle, "vigilance"},
		{"Imposing Vantasaur", "Creature — Dinosaur", ikoImposingVantasaurOracle, "vigilance"},
		{"Lava Serpent", "Creature — Elemental Serpent", ikoLavaSerpentOracle, "haste"},
		{"Excavation Mole", "Creature — Mole", ikoExcavationMoleOracle, "trample"},
		{"Almighty Brushwagg", "Creature — Brushwagg", ikoAlmightyBrushwaggOracle, "trample"},
		{"Farfinder", "Creature — Fox", ikoFarfinderOracle, "vigilance"},
		{"Adaptive Shimmerer", "Creature — Insect", ikoAdaptiveShimmererOracle, "flash"},
		{"Avian Oddity", "Creature — Bird", ikoAvianOddityOracle, "flying"},
		{"Void Beckoner", "Creature — Nightmare Horror", ikoVoidBeckonerOracle, "deathtouch"},
		{"Titanoth Rex", "Creature — Dinosaur Beast", ikoTitanothRexOracle, "trample"},
		{"Splendor Mare", "Creature — Elk Unicorn", ikoSplendorMareOracle, "lifelink"},
		{"Sanctuary Smasher", "Creature — Rhino Beast", ikoSanctuarySmasherOracle, "first strike"},
		{"Dreamtail Heron", "Creature — Elemental Bird", ikoDreamtailHeronOracle, "flying"},
		{"Vulpikeet", "Creature — Fox Bird", ikoVulpikeetOracle, "flying"},
		{"Glowstone Recluse", "Creature — Spider", ikoGlowstoneRecluseOracle, "reach"},
		{"Majestic Auricorn", "Creature — Unicorn", ikoMajesticAuricornOracle, "vigilance"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, id), tc.keyword) {
				t.Errorf("%s has no %s", tc.name, tc.keyword)
			}
		})
	}
}

// --- cycling ------------------------------------------------------------------------

// Every card here cycles from hand for its printed cost and draws.
func TestIko5512CyclingDrawsACard(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, mana string
	}{
		{"Imposing Vantasaur", "Creature — Dinosaur", ikoImposingVantasaurOracle, "{C}"},
		{"Lava Serpent", "Creature — Elemental Serpent", ikoLavaSerpentOracle, "{C}{C}"},
		{"Indatha Crystal", "Artifact", ikoIndathaCrystalOracle, "{C}{C}"},
		{"Ketria Crystal", "Artifact", ikoKetriaCrystalOracle, "{C}{C}"},
		{"Raugrin Crystal", "Artifact", ikoRaugrinCrystalOracle, "{C}{C}"},
		{"Savai Crystal", "Artifact", ikoSavaiCrystalOracle, "{C}{C}"},
		{"Zagoth Crystal", "Artifact", ikoZagothCrystalOracle, "{C}{C}"},
		{"Boon of the Wish-Giver", "Sorcery", ikoBoonOfTheWishGiverOracle, "{C}"},
		{"Neutralize", "Instant", ikoNeutralizeOracle, "{C}{C}"},
		{"Wilt", "Instant", ikoWiltOracle, "{C}{C}"},
		{"Easy Prey", "Instant", ikoEasyPreyOracle, "{C}{C}"},
		{"Raking Claws", "Instant", ikoRakingClawsOracle, "{C}{C}"},
		{"Coordinated Charge", "Instant", ikoCoordinatedChargeOracle, "{C}{C}"},
		{"Hampering Snare", "Instant", ikoHamperingSnareOracle, "{C}{C}"},
		{"Suffocating Fumes", "Instant", ikoSuffocatingFumesOracle, "{C}{C}"},
		{"Frostveil Ambush", "Instant", ikoFrostveilAmbushOracle, "{C}"},
		{"Go for Blood", "Sorcery", ikoGoForBloodOracle, "{C}"},
		{"Barrier Breach", "Instant", ikoBarrierBreachOracle, "{C}{C}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id, me := effectstest.CycleFromHand(t, g, tc.name, tc.typeLine, tc.oracle, tc.mana)
			if !me.Graveyard.Contains(id) {
				t.Fatal("the cycled card is not in the graveyard")
			}
			hand := me.Hand.Size()
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size(); got != hand+1 {
				t.Errorf("hand %d -> %d, want the cycling draw", hand, got)
			}
		})
	}
}

// "When you cycle this card, put a <keyword> counter on target creature
// you control": the counter lands, and the cycling still draws.
func TestIko5512CycledCountersOnYourCreature(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, mana, kind string
	}{
		{"Avian Oddity", "Creature — Bird", ikoAvianOddityOracle, "{C}{C}{U}", game.CounterFlying},
		{"Void Beckoner", "Creature — Nightmare Horror", ikoVoidBeckonerOracle, "{C}{C}{B}", game.CounterDeathtouch},
		{"Titanoth Rex", "Creature — Dinosaur Beast", ikoTitanothRexOracle, "{C}{G}", game.CounterTrample},
		{"Splendor Mare", "Creature — Elk Unicorn", ikoSplendorMareOracle, "{C}{W}", game.CounterLifelink},
		{"Sanctuary Smasher", "Creature — Rhino Beast", ikoSanctuarySmasherOracle, "{C}{C}{R}", game.CounterFirstStrike},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			bear := ikoCreature(g, me.ID, "Bear", "{1}{G}", 2, 2)
			id, _ := effectstest.CycleFromHand(t, g, tc.name, tc.typeLine, tc.oracle, tc.mana)
			if !me.Graveyard.Contains(id) {
				t.Fatal("the cycled card is not in the graveyard")
			}
			hand := me.Hand.Size()
			effectstest.AnswerAnyPendingTargetPrompts(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CounterOn(g, bear, tc.kind); got != 1 {
				t.Errorf("%d %s counters on the Bear, want 1", got, tc.kind)
			}
			if !effectstest.HasEffectiveKeyword(t, g, bear, tc.kind) {
				t.Errorf("the Bear does not have %s", tc.kind)
			}
			if got := me.Hand.Size(); got != hand+1 {
				t.Errorf("hand %d -> %d, want the cycling draw", hand, got)
			}
		})
	}
}

// --- the Crystals ------------------------------------------------------------------

func TestIko5512CrystalsTapForTheirWedge(t *testing.T) {
	for _, tc := range []struct{ name, oracle, produced string }{
		{"Indatha Crystal", ikoIndathaCrystalOracle, "{W|B|G}"},
		{"Ketria Crystal", ikoKetriaCrystalOracle, "{G|U|R}"},
		{"Raugrin Crystal", ikoRaugrinCrystalOracle, "{U|R|W}"},
		{"Savai Crystal", ikoSavaiCrystalOracle, "{R|W|B}"},
		{"Zagoth Crystal", ikoZagothCrystalOracle, "{B|G|U}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spec, _ := effects.Lookup(tc.oracle)
			if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != tc.produced || !spec.ManaAbilities[0].Cost.Tap {
				t.Fatalf("mana abilities %+v, want one {T} producing %s", spec.ManaAbilities, tc.produced)
			}
			g := effectstest.NewCatalogGame(t)
			id := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, tc.name, "Artifact", tc.oracle, false)
			if n := len(game.ManaAbilitiesForCard(effectstest.FindCardAnywhere(t, g, id))); n != 1 {
				t.Errorf("the Crystal on the battlefield has %d mana abilities, want 1", n)
			}
		})
	}
}

// --- spells ----------------------------------------------------------------------------

func TestIko5512BoonDrawsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Boon of the Wish-Giver", "Sorcery", ikoBoonOfTheWishGiverOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+4 {
		t.Errorf("hand %d -> %d, want four more", hand, got)
	}
}

func TestIko5512NeutralizeCountersASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.B13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", nil)
	id := effectstest.B13HandCard(me, "Neutralize", "Instant", ikoNeutralizeOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(bolt)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the spell was not countered")
	}
}

func TestIko5512WiltDestroysAnEnchantmentNotACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	bear := ikoCreature(g, opp.ID, "Bear", "{1}{G}", 2, 2)
	rune := effectstest.B12Permanent(g, opp.ID, "Rune", "Enchantment")
	if err := effectstest.CastCatalogSpellErr(t, g, "Wilt", "Instant", ikoWiltOracle, effectstest.CardTarget(bear)); err == nil {
		t.Fatal("Wilt targeted a creature")
	}
	effectstest.CastCatalogSpell(t, g, "Wilt", "Instant", ikoWiltOracle, effectstest.CardTarget(rune))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rune) {
		t.Error("the enchantment survived")
	}
}

func TestIko5512EasyPreyDestroysOnlyManaValueTwoOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	small := ikoCreature(g, opp.ID, "Small", "{1}{G}", 2, 2)
	big := ikoCreature(g, opp.ID, "Big", "{2}{G}", 3, 3)
	if err := effectstest.CastCatalogSpellErr(t, g, "Easy Prey", "Instant", ikoEasyPreyOracle, effectstest.CardTarget(big)); err == nil {
		t.Fatal("Easy Prey targeted a creature with mana value 3")
	}
	effectstest.CastCatalogSpell(t, g, "Easy Prey", "Instant", ikoEasyPreyOracle, effectstest.CardTarget(small))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the mana value 2 creature survived")
	}
}

func TestIko5512BladeBanishExilesOnlyPowerFourOrGreater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	three := ikoCreature(g, opp.ID, "Three", "{2}{G}", 3, 3)
	four := ikoCreature(g, opp.ID, "Four", "{3}{G}", 4, 4)
	if err := effectstest.CastCatalogSpellErr(t, g, "Blade Banish", "Instant", ikoBladeBanishOracle, effectstest.CardTarget(three)); err == nil {
		t.Fatal("Blade Banish targeted a creature with power 3")
	}
	effectstest.CastCatalogSpell(t, g, "Blade Banish", "Instant", ikoBladeBanishOracle, effectstest.CardTarget(four))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(four) || !g.Exile.Contains(four) {
		t.Error("the power 4 creature was not exiled")
	}
}

func TestIko5512BarrierBreachExilesUpToThreeEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	var runes []uuid.UUID
	for i := 0; i < 3; i++ {
		runes = append(runes, effectstest.B12Permanent(g, opp.ID, "Rune", "Enchantment"))
	}
	effectstest.CastCatalogSpell(t, g, "Barrier Breach", "Instant", ikoBarrierBreachOracle, effectstest.CardTargets(runes...))
	effectstest.PassPriorityAroundTable(t, g)
	for i, id := range runes {
		if !g.Exile.Contains(id) {
			t.Errorf("enchantment %d was not exiled", i)
		}
	}
}

func TestIko5512RakingClawsGrantsDoubleStrikeForTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	bear := ikoCreature(g, me.ID, "Bear", "{1}{G}", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Raking Claws", "Instant", ikoRakingClawsOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "double strike") {
		t.Fatal("the Bear has no double strike")
	}
	effectstest.AdvanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if effectstest.HasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("double strike outlasted the turn")
	}
}

func TestIko5512TeamPumpsAndShrinks(t *testing.T) {
	for _, tc := range []struct {
		name, oracle                 string
		mineP, mineT, theirP, theirT int
	}{
		{"Coordinated Charge", ikoCoordinatedChargeOracle, 4, 4, 2, 3},
		{"Hampering Snare", ikoHamperingSnareOracle, 2, 3, 0, 3},
		{"Suffocating Fumes", ikoSuffocatingFumesOracle, 2, 3, 1, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
			mine := ikoCreature(g, me.ID, "Mine", "{1}{G}", 2, 3)
			theirs := ikoCreature(g, opp.ID, "Theirs", "{1}{G}", 2, 3)
			effectstest.CastCatalogSpell(t, g, tc.name, "Instant", tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.PtOf(t, g, mine); p != tc.mineP || tough != tc.mineT {
				t.Errorf("your creature is %d/%d, want %d/%d", p, tough, tc.mineP, tc.mineT)
			}
			if p, tough := effectstest.PtOf(t, g, theirs); p != tc.theirP || tough != tc.theirT {
				t.Errorf("their creature is %d/%d, want %d/%d", p, tough, tc.theirP, tc.theirT)
			}
		})
	}
}

func TestIko5512SuffocatingFumesKillsAnOpponentsOneToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	elf := ikoCreature(g, opp.ID, "Elf", "{G}", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Suffocating Fumes", "Instant", ikoSuffocatingFumesOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(elf) {
		t.Error("a 1/1 survived -1/-1")
	}
}

func TestIko5512FrostveilAmbushTapsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	a := ikoCreature(g, opp.ID, "A", "{1}{G}", 2, 2)
	b := ikoCreature(g, opp.ID, "B", "{1}{G}", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Frostveil Ambush", "Instant", ikoFrostveilAmbushOracle, effectstest.CardTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HmlTapped(t, g, a) || !effectstest.HmlTapped(t, g, b) {
		t.Error("both targets should be tapped")
	}
}

func TestIko5512GoForBloodFightsACreatureYouDontControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	mine := ikoCreature(g, me.ID, "Mine", "{2}{G}", 3, 3)
	other := ikoCreature(g, me.ID, "Other", "{1}{G}", 2, 2)
	theirs := ikoCreature(g, opp.ID, "Theirs", "{1}{G}", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Go for Blood", "Sorcery", ikoGoForBloodOracle, effectstest.CardTargets(mine, other)); err == nil {
		t.Fatal("the second target was a creature you control")
	}
	effectstest.CastCatalogSpell(t, g, "Go for Blood", "Sorcery", ikoGoForBloodOracle, effectstest.CardTargets(mine, theirs))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(theirs) {
		t.Error("the 2/2 survived the fight")
	}
	if got := effectstest.FindCardAnywhere(t, g, mine).DamageMarked; got != 2 {
		t.Errorf("your creature has %d damage, want 2", got)
	}
}

func TestIko5512ClashOfTitansFightsTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	big := ikoCreature(g, opp.ID, "Big", "{3}{G}", 4, 4)
	small := ikoCreature(g, opp.ID, "Small", "{1}{G}", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Clash of Titans", "Instant", ikoClashOfTitansOracle, effectstest.CardTargets(big, small))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(small) {
		t.Error("the 2/2 survived the fight")
	}
	if got := effectstest.FindCardAnywhere(t, g, big).DamageMarked; got != 2 {
		t.Errorf("the 4/4 has %d damage, want 2", got)
	}
}

func TestIko5512UnexpectedFangsPutsBothCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	bear := ikoCreature(g, me.ID, "Bear", "{1}{G}", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Unexpected Fangs", "Instant", ikoUnexpectedFangsOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, bear, game.CounterPlusOne) != 1 || effectstest.CounterOn(g, bear, game.CounterLifelink) != 1 {
		t.Errorf("counters %d +1/+1 and %d lifelink, want one of each",
			effectstest.CounterOn(g, bear, game.CounterPlusOne), effectstest.CounterOn(g, bear, game.CounterLifelink))
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("the Bear has no lifelink")
	}
}

// The pump ends at cleanup; the keyword counter stays.
func TestIko5512PumpsThatLeaveAKeywordCounter(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, kind string
		p, tough           int
		untaps             bool
	}{
		{"Fully Grown", ikoFullyGrownOracle, game.CounterTrample, 5, 5, false},
		{"Heightened Reflexes", ikoHeightenedReflexesOracle, game.CounterFirstStrike, 3, 2, false},
		{"Spontaneous Flight", ikoSpontaneousFlightOracle, game.CounterFlying, 4, 4, false},
		{"Sudden Spinnerets", ikoSuddenSpinneretsOracle, game.CounterReach, 3, 5, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			bear := ikoCreature(g, me.ID, "Bear", "{1}{G}", 2, 2)
			g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
			effectstest.CastCatalogSpell(t, g, tc.name, "Instant", tc.oracle, effectstest.CardTarget(bear))
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.PtOf(t, g, bear); p != tc.p || tough != tc.tough {
				t.Errorf("the Bear is %d/%d, want %d/%d", p, tough, tc.p, tc.tough)
			}
			if effectstest.CounterOn(g, bear, tc.kind) != 1 || !effectstest.HasEffectiveKeyword(t, g, bear, tc.kind) {
				t.Errorf("the Bear has no %s counter", tc.kind)
			}
			if got := effectstest.HmlTapped(t, g, bear); got == tc.untaps {
				t.Errorf("tapped = %v after the spell, want %v", got, !tc.untaps)
			}
			effectstest.AdvanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
			if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 2 {
				t.Errorf("the pump outlasted the turn: %d/%d", p, tough)
			}
			if !effectstest.HasEffectiveKeyword(t, g, bear, tc.kind) {
				t.Errorf("the %s counter's keyword did not last", tc.kind)
			}
		})
	}
}

func TestIko5512ForbiddenFriendshipMakesTwoTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.CastCatalogSpell(t, g, "Forbidden Friendship", "Sorcery", ikoForbiddenFriendshipOracl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if ikoTokensNamed(g, me.ID, "Dinosaur") != 1 || ikoTokensNamed(g, me.ID, "Human Soldier") != 1 {
		t.Fatalf("want one Dinosaur and one Human Soldier")
	}
	var dino uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Dinosaur" && c.Controller == me.ID {
				dino = c.InstanceID
			}
		}
	})
	if !effectstest.HasEffectiveKeyword(t, g, dino, "haste") {
		t.Error("the Dinosaur has no haste")
	}
}

// --- creatures ----------------------------------------------------------------------------

func TestIko5512DaysquadMarshalMakesAHumanSoldier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Daysquad Marshal", "Creature — Human Soldier", ikoDaysquadMarshalOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if n := ikoTokensNamed(g, me.ID, "Human Soldier"); n != 1 {
		t.Errorf("%d Human Soldier tokens, want 1", n)
	}
}

func TestIko5512GarrisonCatLeavesAHumanSoldier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cat := effectstest.B12Push(g, me.ID, "Garrison Cat", "Creature — Cat", ikoGarrisonCatOracle, 1, 1)
	effectstest.B27Kill(g, cat)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := ikoTokensNamed(g, me.ID, "Human Soldier"); n != 1 {
		t.Errorf("%d Human Soldier tokens, want 1", n)
	}
}

func TestIko5512HoneyMammothGainsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	effectstest.CastHoldCreature(t, g, "Honey Mammoth", "Creature — Elephant", ikoHoneyMammothOracle, 6, 6)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
}

func TestIko5512ExcavationMoleMillsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	library := me.Library.Size()
	effectstest.CastHoldCreature(t, g, "Excavation Mole", "Creature — Mole", ikoExcavationMoleOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := library - me.Library.Size(); got != 3 {
		t.Errorf("milled %d, want 3", got)
	}
}

func TestIko5512FacetReaderLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	reader := effectstest.B12Push(g, me.ID, "Facet Reader", "Creature — Human Wizard", ikoFacetReaderOracle, 1, 2)
	effectstest.ApaMana(me, "C")
	hand := me.Hand.Size()
	effectstest.ApaActivate(t, g, me, reader, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d before the discard, want one more", hand, got)
	}
	if n := effectstest.DiscardOwed(g, me.ID); n != 1 {
		t.Errorf("owes %d discards, want 1", n)
	}
	if !effectstest.HmlTapped(t, g, reader) {
		t.Error("Facet Reader did not tap")
	}
}

func TestIko5512CheckpointOfficerTapsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	officer := effectstest.B12Push(g, me.ID, "Checkpoint Officer", "Creature — Human Soldier", ikoCheckpointOfficerOracle, 1, 2)
	bear := ikoCreature(g, opp.ID, "Bear", "{1}{G}", 2, 2)
	effectstest.ApaMana(me, "C", "W")
	effectstest.ApaActivate(t, g, me, officer, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if !effectstest.HmlTapped(t, g, bear) || !effectstest.HmlTapped(t, g, officer) {
		t.Error("the Officer and its target should both be tapped")
	}
}

func TestIko5512AlmightyBrushwaggPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	wagg := effectstest.B12Push(g, me.ID, "Almighty Brushwagg", "Creature — Brushwagg", ikoAlmightyBrushwaggOracle, 1, 1)
	effectstest.ApaMana(me, "C", "C", "C", "G")
	effectstest.ApaActivate(t, g, me, wagg, 0, game.ActivateAbilityParams{})
	if p, tough := effectstest.PtOf(t, g, wagg); p != 4 || tough != 4 {
		t.Errorf("the Brushwagg is %d/%d, want 4/4", p, tough)
	}
}

func TestIko5512AdaptiveShimmererFlashesInWithThreeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B13HandCard(me, "Adaptive Shimmerer", "Creature — Insect", ikoAdaptiveShimmererOracle)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("Adaptive Shimmerer at the beginning of combat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("it did not survive its entry")
	}
	if got := effectstest.CounterOn(g, id, game.CounterPlusOne); got != 3 {
		t.Errorf("%d +1/+1 counters, want 3", got)
	}
}

func TestIko5512FarfinderFetchesABasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.CastHoldCreature(t, g, "Farfinder", "Creature — Fox", ikoFarfinderOracle, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Fatal("no search prompt")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(forest) {
		t.Error("the Forest is not in hand")
	}
}

// --- mutate ---------------------------------------------------------------------------------

func TestIko5512DreamtailHeronDrawsWhenItMutates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elves := effectstest.LlanowarOn(g, me.ID)
	heron := ikoMutatorInHand(t, g, me.ID, "Dreamtail Heron", "Creature — Elemental Bird", ikoDreamtailHeronOracle, 3, 4)
	if err := effectstest.CastMutating(g, me.ID, heron, elves, false); err != nil {
		t.Fatalf("cast mutating: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.ResolveMutate(t, g, me.ID, true)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want the mutate draw", hand, got)
	}
	if f := effectstest.ReadMerged(t, g, elves); !f.Flying || len(f.Components) != 2 {
		t.Errorf("under: the Elves have the Heron's flying: %+v", f)
	}
}

func TestIko5512DreamtailHeronCastNormallyDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.CastHoldCreature(t, g, "Dreamtail Heron", "Creature — Elemental Bird", ikoDreamtailHeronOracle, 3, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d, want no draw without a mutate", hand, got)
	}
}

func TestIko5512MutateCountersGoOnTheMergedCreature(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness, want int
		under                  bool
	}{
		{"Vulpikeet", "Creature — Fox Bird", ikoVulpikeetOracle, 2, 3, 1, false},
		{"Glowstone Recluse", "Creature — Spider", ikoGlowstoneRecluseOracle, 2, 3, 2, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			elves := effectstest.LlanowarOn(g, me.ID)
			id := ikoMutatorInHand(t, g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			if err := effectstest.CastMutating(g, me.ID, id, elves, false); err != nil {
				t.Fatalf("cast mutating: %v", err)
			}
			effectstest.ResolveMutate(t, g, me.ID, tc.under)
			if got := effectstest.CounterOn(g, elves, game.CounterPlusOne); got != tc.want {
				t.Errorf("%d +1/+1 counters on the merged creature, want %d", got, tc.want)
			}
		})
	}
}

func TestIko5512MajesticAuricornGainsFourWhenItMutates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elves := effectstest.LlanowarOn(g, me.ID)
	auricorn := ikoMutatorInHand(t, g, me.ID, "Majestic Auricorn", "Creature — Unicorn", ikoMajesticAuricornOracle, 4, 4)
	if err := effectstest.CastMutating(g, me.ID, auricorn, elves, false); err != nil {
		t.Fatalf("cast mutating: %v", err)
	}
	life := me.Life
	effectstest.ResolveMutate(t, g, me.ID, false)
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
	if f := effectstest.ReadMerged(t, g, elves); f.Name != "Majestic Auricorn" || f.Power != 4 {
		t.Errorf("over: the permanent is the Auricorn: %+v", f)
	}
}

// A Vulpikeet mutating onto the Egg fires both "whenever this creature
// mutates" abilities: two +1/+1 counters on the merged permanent.
func TestIko5512MysteriousEggGrowsWhenSomethingMutatesOntoIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	egg := effectstest.B12Push(g, me.ID, "Mysterious Egg", "Creature — Egg", ikoMysteriousEggOracle, 0, 2)
	keet := ikoMutatorInHand(t, g, me.ID, "Vulpikeet", "Creature — Fox Bird", ikoVulpikeetOracle, 2, 3)
	if err := effectstest.CastMutating(g, me.ID, keet, egg, false); err != nil {
		t.Fatalf("cast mutating: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.B35SettleTriggerOrders(t, g, me.ID)
	if got := effectstest.CounterOn(g, egg, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters on the Egg, want 2", got)
	}
}
