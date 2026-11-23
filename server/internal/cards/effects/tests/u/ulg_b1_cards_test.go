package testsu

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ulg_b1_cards_test.go — Urza's Legacy, batch 1 (#5623). Each test plays
// the real catalog entry and asserts what its text says.

const (
	ulgB1DefenseGridOracle      = "22ba474e-f6e6-450d-a4cf-0739f0b51079"
	ulgB1LevitationOracle       = "94b703c4-5584-4913-8365-7e9f2f535c2d"
	ulgB1KnighthoodOracle       = "22f008b3-c263-4dcc-855f-403da9ca4cd9"
	ulgB1WatchtowerOracle       = "cabf7953-0fac-4dbb-b3ae-05e85e02b3fc"
	ulgB1EncampmentOracle       = "9cc02c16-8cf9-4ac7-9475-231c27121968"
	ulgB1MiscalculationOracle   = "9078db26-3a54-4a11-9aae-e45c38da7226"
	ulgB1TinkerOracle           = "254878f0-be90-4653-a395-0c41258fceaf"
	ulgB1EraseOracle            = "c2ceb15c-5d02-4cf4-a7c9-c1a40b7ca667"
	ulgB1OpportunityOracle      = "1f544a9f-c238-4858-be19-d6cd7d023dcc"
	ulgB1MightOfOaksOracle      = "8331f281-819b-4a0b-bad7-bd86dbedb877"
	ulgB1ArchivistOracle        = "d137586f-83b0-40af-8100-443460b07ac0"
	ulgB1ThornwindOracle        = "941a0e8d-c407-476b-9d0e-cffc1f8c6003"
	ulgB1WheelOracle            = "7a3e6e84-a8c2-472d-99fb-7cd14850b5ce"
	ulgB1RadiantsJudgmentOracle = "85fd04e9-b39c-4ce0-97a1-58f9670a298f"
	ulgB1SwatOracle             = "7bb5c5af-6943-4958-a666-0c322d5b9f33"
	ulgB1IronWillOracle         = "dc0decb9-32da-47aa-a51a-0d9c623c534a"
	ulgB1PurifyOracle           = "cfd1b276-ef5d-4f60-9f63-a548ad315e07"
	ulgB1RebuildOracle          = "496999a7-a82a-45ea-9c4e-1b6aabc781cd"
	ulgB1MultaniOracle          = "bce5b327-7da1-4136-b76f-069e2fef5d2a"
	ulgB1PlagueOracle           = "dc8210b4-8e17-4156-994b-e95e215e3e34"
	ulgB1GraniteGripOracle      = "19e09405-3294-4f98-a18f-60cf0a8a2e4e"
	ulgB1SleepersGuileOracle    = "eaaa3e41-dbf6-43ae-b65e-e1502ba4fa0b"
	ulgB1SluggishnessOracle     = "39eb4062-68a0-4f64-a9cc-79be87fd8173"
	ulgB1CessationOracle        = "a8b3a35e-0172-4f7b-b50f-571e59d7ca5f"
	ulgB1ShivanPhoenixOracle    = "49ca8516-624b-4aa8-8e13-d145590cdfda"
	ulgB1TreefolkOracle         = "4e9b878b-80a2-433b-941b-92d854e83d58"
	ulgB1FogOfGnatsOracle       = "41f58e06-5ce0-4b2d-a280-5bb3b6c56dfe"
	ulgB1SandscoutOracle        = "f8d5a65c-382b-4277-9c08-5dec9e395b02"
	ulgB1CutthroatOracle        = "6dc1305a-d6ca-4280-ba5d-5ed91f78ced0"
	ulgB1DebaserOracle          = "554de904-086c-428c-9f32-b0b2498a6fd5"
	ulgB1DefilerOracle          = "bed87e57-d410-478f-ae9d-2d7d00831e5e"
	ulgB1DenouncerOracle        = "b331c5fa-cd97-4fd8-b1df-d2c9e6a5a854"
	ulgB1DarkwatchElvesOracle   = "55ad9924-f217-4844-8a82-48746186bc53"
	ulgB1BloatedToadOracle      = "5ddcc7a5-c869-49fd-baa1-ab5f5f456f90"
	ulgB1AngelicCuratorOracle   = "cca29a9b-794f-4712-8e6c-36ce3da9cb8b"
	ulgB1DefenderOfChaosOracle  = "e5f04e8a-e9ae-4f00-8ada-f63b3454ed30"
	ulgB1DefenderOfLawOracle    = "656513cf-1bf0-431e-8d3a-2d1ec79d0bec"
	ulgB1WingSnareOracle        = "61ce6ec1-79d8-4dc9-ab2d-1623ed1459c1"
	ulgB1RackAndRuinOracle      = "e04552f4-41c1-40f3-996e-de796c9967e4"
	ulgB1PeaceAndQuietOracle    = "0517f736-e6eb-4fdb-8288-e0c8a9a44379"
	ulgB1ExpendableOracle       = "70ebd18a-c878-4a93-b7c4-3f7a2aa65f5e"
	ulgB1EvisceratorOracle      = "710e6d6a-caaa-4b76-82c8-e363e9438bec"
)

func TestUlgB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		ulgB1DefenseGridOracle:      "Defense Grid",
		ulgB1LevitationOracle:       "Levitation",
		ulgB1KnighthoodOracle:       "Knighthood",
		ulgB1WatchtowerOracle:       "Forbidding Watchtower",
		ulgB1EncampmentOracle:       "Ghitu Encampment",
		ulgB1MiscalculationOracle:   "Miscalculation",
		ulgB1TinkerOracle:           "Tinker",
		ulgB1EraseOracle:            "Erase",
		ulgB1OpportunityOracle:      "Opportunity",
		ulgB1MightOfOaksOracle:      "Might of Oaks",
		ulgB1ArchivistOracle:        "Archivist",
		ulgB1ThornwindOracle:        "Thornwind Faeries",
		ulgB1WheelOracle:            "Wheel of Torture",
		ulgB1RadiantsJudgmentOracle: "Radiant's Judgment",
		ulgB1SwatOracle:             "Swat",
		ulgB1IronWillOracle:         "Iron Will",
		ulgB1PurifyOracle:           "Purify",
		ulgB1RebuildOracle:          "Rebuild",
		ulgB1MultaniOracle:          "Multani, Maro-Sorcerer",
		ulgB1PlagueOracle:           "Engineered Plague",
		ulgB1GraniteGripOracle:      "Granite Grip",
		ulgB1SleepersGuileOracle:    "Sleeper's Guile",
		ulgB1SluggishnessOracle:     "Sluggishness",
		ulgB1CessationOracle:        "Cessation",
		ulgB1ShivanPhoenixOracle:    "Shivan Phoenix",
		ulgB1TreefolkOracle:         "Weatherseed Treefolk",
		ulgB1FogOfGnatsOracle:       "Fog of Gnats",
		ulgB1SandscoutOracle:        "Viashino Sandscout",
		ulgB1CutthroatOracle:        "Viashino Cutthroat",
		ulgB1DebaserOracle:          "Phyrexian Debaser",
		ulgB1DefilerOracle:          "Phyrexian Defiler",
		ulgB1DenouncerOracle:        "Phyrexian Denouncer",
		ulgB1DarkwatchElvesOracle:   "Darkwatch Elves",
		ulgB1BloatedToadOracle:      "Bloated Toad",
		ulgB1AngelicCuratorOracle:   "Angelic Curator",
		ulgB1DefenderOfChaosOracle:  "Defender of Chaos",
		ulgB1DefenderOfLawOracle:    "Defender of Law",
		ulgB1WingSnareOracle:        "Wing Snare",
		ulgB1RackAndRuinOracle:      "Rack and Ruin",
		ulgB1PeaceAndQuietOracle:    "Peace and Quiet",
		ulgB1ExpendableOracle:       "Expendable Troops",
		ulgB1EvisceratorOracle:      "Eviscerator",
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

// --- printed keywords -----------------------------------------------------------------

func TestUlgB1CardsHaveTheirPrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Thornwind Faeries", "Creature — Faerie", ulgB1ThornwindOracle, []string{"flying"}},
		{"Multani, Maro-Sorcerer", "Legendary Creature — Elemental Sorcerer", ulgB1MultaniOracle, []string{"shroud"}},
		{"Shivan Phoenix", "Creature — Phoenix", ulgB1ShivanPhoenixOracle, []string{"flying"}},
		{"Weatherseed Treefolk", "Creature — Treefolk", ulgB1TreefolkOracle, []string{"trample"}},
		{"Fog of Gnats", "Creature — Insect", ulgB1FogOfGnatsOracle, []string{"flying"}},
		{"Viashino Sandscout", "Creature — Lizard Scout", ulgB1SandscoutOracle, []string{"haste"}},
		{"Viashino Cutthroat", "Creature — Lizard", ulgB1CutthroatOracle, []string{"haste"}},
		{"Phyrexian Debaser", "Creature — Phyrexian Carrier", ulgB1DebaserOracle, []string{"flying"}},
		{"Darkwatch Elves", "Creature — Elf", ulgB1DarkwatchElvesOracle, []string{"protection from black"}},
		{"Bloated Toad", "Creature — Frog", ulgB1BloatedToadOracle, []string{"protection from blue"}},
		{"Angelic Curator", "Creature — Angel Spirit", ulgB1AngelicCuratorOracle, []string{"flying", "protection from artifacts"}},
		{"Defender of Chaos", "Creature — Human Knight", ulgB1DefenderOfChaosOracle, []string{"flash", "protection from white"}},
		{"Defender of Law", "Creature — Human Knight", ulgB1DefenderOfLawOracle, []string{"flash", "protection from red"}},
		{"Eviscerator", "Creature — Phyrexian Horror", ulgB1EvisceratorOracle, []string{"protection from white"}},
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

// --- cycling --------------------------------------------------------------------------

func TestUlgB1CyclersCycle(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Miscalculation", "Instant", ulgB1MiscalculationOracle},
		{"Radiant's Judgment", "Instant", ulgB1RadiantsJudgmentOracle},
		{"Swat", "Instant", ulgB1SwatOracle},
		{"Iron Will", "Instant", ulgB1IronWillOracle},
		{"Rebuild", "Instant", ulgB1RebuildOracle},
		{"Darkwatch Elves", "Creature — Elf", ulgB1DarkwatchElvesOracle},
		{"Bloated Toad", "Creature — Frog", ulgB1BloatedToadOracle},
	} {
		effectstest.UsgB4Cycles(t, tc.name, tc.typeLine, tc.oracle)
	}
}

// --- Defense Grid ---------------------------------------------------------------------

func TestUlgB1DefenseGridTaxesSpellsCastOffTheirControllersTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, other.ID, "Defense Grid", "Artifact", ulgB1DefenseGridOracle, false)
	if got := effectstest.PriceInHand(t, g, active, "Lightning Bolt", "Instant", "{R}"); got != 1 {
		t.Errorf("a spell cast during its controller's turn costs %d, want 1", got)
	}
	if got := effectstest.PriceInHand(t, g, other, "Counterspell", "Instant", "{U}{U}"); got != 5 {
		t.Errorf("the Grid controller's spell on another turn costs %d, want 5", got)
	}
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	if got := effectstest.PriceInHand(t, g, third, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 5 {
		t.Errorf("a third player's spell on another turn costs %d, want 5", got)
	}
}

// --- Levitation, Knighthood -------------------------------------------------------------

func TestUlgB1AnthemsGrantOnlyYourCreaturesTheirKeyword(t *testing.T) {
	for _, tc := range []struct{ name, oracle, keyword string }{
		{"Levitation", ulgB1LevitationOracle, "flying"},
		{"Knighthood", ulgB1KnighthoodOracle, "first strike"},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		effectstest.B12Push(g, me.ID, tc.name, "Enchantment", tc.oracle, 0, 0)
		if !effectstest.HasKeywordOnBattlefield(t, g, mine, tc.keyword) {
			t.Errorf("%s: your creature lacks %s", tc.name, tc.keyword)
		}
		if effectstest.HasKeywordOnBattlefield(t, g, theirs, tc.keyword) {
			t.Errorf("%s: an opponent's creature has %s", tc.name, tc.keyword)
		}
		later := effectstest.B12Creature(g, me.ID, "Later Bear", "Creature — Bear", 2, 2)
		if !effectstest.HasKeywordOnBattlefield(t, g, later, tc.keyword) {
			t.Errorf("%s: a creature that arrived later lacks %s", tc.name, tc.keyword)
		}
	}
}

// --- Forbidding Watchtower, Ghitu Encampment ----------------------------------------------

func TestUlgB1ManlandsAnimateAndStayLands(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, sub, kw string
		power, tough          int
		cost, colours         []string
	}{
		{"Forbidding Watchtower", ulgB1WatchtowerOracle, "Soldier", "", 1, 5, []string{"C", "W"}, []string{"W"}},
		{"Ghitu Encampment", ulgB1EncampmentOracle, "Warrior", "first strike", 2, 1, []string{"C", "R"}, []string{"R"}},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		land := effectstest.S58p6Push(g, me.ID, tc.name, "Land", tc.oracle, "", 0, 0)
		effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
		if effectstest.S58p6View(t, g, land).IsCreature() {
			t.Fatalf("%s: an idle land is a creature", tc.name)
		}
		if err := effectstest.Cmdr2B13Activate(t, g, me, land, game.ActivateAbilityParams{}, tc.cost...); err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		c := effectstest.S58p6View(t, g, land)
		if !c.IsCreature() || !c.IsLand() || c.CurrentPower() != tc.power || c.CurrentToughness() != tc.tough {
			t.Errorf("%s: creature=%v land=%v %d/%d", tc.name, c.IsCreature(), c.IsLand(), c.CurrentPower(), c.CurrentToughness())
		}
		if !slices.Contains(c.Effective().Subtypes, tc.sub) {
			t.Errorf("%s: subtypes %v, want %s", tc.name, c.Effective().Subtypes, tc.sub)
		}
		if !slices.Equal(c.EffectiveColors(), tc.colours) {
			t.Errorf("%s: colours %v, want %v", tc.name, c.EffectiveColors(), tc.colours)
		}
		if tc.kw != "" {
			effectstest.AssertEffectiveKeyword(t, g, land, tc.kw)
		}
		if len(game.ManaAbilitiesForCard(c)) == 0 {
			t.Errorf("%s lost its mana ability", tc.name)
		}
		effectstest.DcCleanup(g)
		if effectstest.S58p6View(t, g, land).IsCreature() {
			t.Errorf("%s is still a creature after the cleanup step", tc.name)
		}

		g2 := effectstest.NewCatalogGame(t)
		id := effectstest.B12PlayFromHand(t, g2, tc.name, "Land", tc.oracle, game.CastSpellParams{})
		if !effectstest.IsTapped(g2, id) {
			t.Errorf("%s entered untapped", tc.name)
		}
	}
}

// --- Miscalculation ---------------------------------------------------------------------

func ulgB1MiscalculationBoard(t *testing.T) (g *game.Game, victim uuid.UUID, caster *game.Player) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	caster, them := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim = effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	spell := effectstest.HandCardFull(them, "Miscalculation", "Instant", "", ulgB1MiscalculationOracle, []string{"U"})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	if err := g.CastSpell(them.ID, spell, game.CastSpellParams{Targets: effectstest.CardRefs(victim)}); err != nil {
		t.Fatalf("casting Miscalculation: %v", err)
	}
	effectstest.PassUntilTaxed(t, g, caster.ID)
	return g, victim, caster
}

func TestUlgB1MiscalculationCountersUnlessTwoIsPaid(t *testing.T) {
	g, victim, caster := ulgB1MiscalculationBoard(t)
	prompt := effectstest.AnswerPayUnless(t, g, caster.ID, false)
	if prompt.PayCost != "{2}" {
		t.Errorf("the tax is %q, want {2}", prompt.PayCost)
	}
	if g.Stack.Contains(victim) {
		t.Error("a declined tax must counter the spell")
	}

	g, victim, caster = ulgB1MiscalculationBoard(t)
	effectstest.FillPool(caster, 2)
	effectstest.AnswerPayUnless(t, g, caster.ID, true)
	if !g.Stack.Contains(victim) {
		t.Error("a paid tax countered the spell anyway")
	}
}

// --- Tinker -----------------------------------------------------------------------------

func TestUlgB1TinkerSacrificesAnArtifactAndFetchesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMainForCost(t, g)
	land := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	fodder := effectstest.B12Permanent(g, me.ID, "Old Trinket", "Artifact")
	want := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Big Robot", TypeLine: "Artifact Creature — Construct"})
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Tinker", TypeLine: "Sorcery", OracleID: ulgB1TinkerOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("Tinker was cast by sacrificing a land")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("CastSpell Tinker: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the artifact was not sacrificed as the cast was paid")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if !g.Battlefield.Contains(want) {
		t.Error("the artifact card did not reach the battlefield")
	}
}

// --- Erase ------------------------------------------------------------------------------

func TestUlgB1EraseExilesAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	ench := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Erase", "Instant", ulgB1EraseOracle, effectstest.CardTarget(bear)); err == nil {
		t.Fatal("Erase targeted a creature")
	}
	effectstest.CastCatalogSpell(t, g, "Erase", "Instant", ulgB1EraseOracle, effectstest.CardTarget(ench))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) || opp.Graveyard.Contains(ench) {
		t.Error("the enchantment was not exiled")
	}
	if effectstest.B12ZoneOf(g, ench) != game.ZoneExile {
		t.Errorf("the enchantment is in %q, want exile", effectstest.B12ZoneOf(g, ench))
	}
}

// --- Opportunity ------------------------------------------------------------------------

func TestUlgB1OpportunityMakesTheTargetDrawFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	hand := opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Opportunity", "Instant", ulgB1OpportunityOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - hand; got != 4 {
		t.Errorf("the target drew %d cards, want 4", got)
	}
}

// --- Might of Oaks, Iron Will ------------------------------------------------------------

func TestUlgB1PumpSpellsBoostUntilEndOfTurn(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
	}{
		{"Might of Oaks", ulgB1MightOfOaksOracle, 9, 9},
		{"Iron Will", ulgB1IronWillOracle, 2, 6},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.CastCatalogSpell(t, g, tc.name, "Instant", tc.oracle, effectstest.CardTarget(bear))
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.EffectivePTOf(t, g, bear); p != tc.power || tough != tc.tough {
			t.Errorf("%s: the target is %d/%d, want %d/%d", tc.name, p, tough, tc.power, tc.tough)
		}
		effectstest.DcCleanup(g)
		if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 2 {
			t.Errorf("%s: the bonus outlived the cleanup step (%d/%d)", tc.name, p, tough)
		}
	}
}

// --- Archivist --------------------------------------------------------------------------

func TestUlgB1ArchivistTapsToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sick := effectstest.PushCatalogPermanent(g, me.ID, "Archivist", "Creature — Human Wizard", ulgB1ArchivistOracle, true)
	if err := effectstest.RavB2Activate(t, g, me.ID, sick, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a summoning-sick Archivist was tapped to draw")
	}
	archivist := effectstest.PushCatalogPermanent(g, me.ID, "Archivist", "Creature — Human Wizard", ulgB1ArchivistOracle, false)
	hand := me.Hand.Size()
	effectstest.RavB2MustActivate(t, g, me.ID, archivist, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one draw", hand, me.Hand.Size())
	}
	if !effectstest.IsTapped(g, archivist) {
		t.Error("Archivist did not tap as the cost")
	}
}

// --- Thornwind Faeries -------------------------------------------------------------------

func TestUlgB1ThornwindFaeriesPingsAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	faeries := effectstest.B12Push(g, me.ID, "Thornwind Faeries", "Creature — Faerie", ulgB1ThornwindOracle, 1, 1)
	life := opp.Life
	effectstest.RavB2MustActivate(t, g, me.ID, faeries, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if opp.Life != life-1 {
		t.Errorf("life %d → %d, want 1 damage", life, opp.Life)
	}
}

// --- Wheel of Torture --------------------------------------------------------------------

func TestUlgB1WheelOfTortureHurtsAnOpponentWithASmallHand(t *testing.T) {
	for _, hand := range []int{0, 1, 3, 5} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Wheel of Torture", "Artifact", ulgB1WheelOracle, 0, 0)
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		effectstest.D2dSetHand(opp, hand)
		start := opp.Life
		effectstest.PassPriorityAroundTable(t, g)
		want := max(3-hand, 0)
		if opp.Life != start-want {
			t.Errorf("hand of %d: life %d → %d, want -%d", hand, start, opp.Life, want)
		}
	}

	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.B12Push(g, me.ID, "Wheel of Torture", "Artifact", ulgB1WheelOracle, 0, 0)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.D2dSetHand(me, 0)
	start := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != start {
		t.Errorf("the controller's own upkeep dealt %d", start-me.Life)
	}
}

// --- Radiant's Judgment, Swat, Wing Snare -------------------------------------------------

func TestUlgB1ConditionalDestroySpellsCheckTheirTarget(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		legal, illegal         func(g *game.Game, owner uuid.UUID) uuid.UUID
	}{
		{"Radiant's Judgment", "Instant", ulgB1RadiantsJudgmentOracle,
			func(g *game.Game, o uuid.UUID) uuid.UUID {
				return effectstest.B12Creature(g, o, "Big", "Creature — Giant", 4, 4)
			},
			func(g *game.Game, o uuid.UUID) uuid.UUID {
				return effectstest.B12Creature(g, o, "Small", "Creature — Bear", 3, 3)
			}},
		{"Swat", "Instant", ulgB1SwatOracle,
			func(g *game.Game, o uuid.UUID) uuid.UUID {
				return effectstest.B12Creature(g, o, "Small", "Creature — Bear", 2, 2)
			},
			func(g *game.Game, o uuid.UUID) uuid.UUID {
				return effectstest.B12Creature(g, o, "Big", "Creature — Giant", 3, 3)
			}},
		{"Wing Snare", "Sorcery", ulgB1WingSnareOracle,
			func(g *game.Game, o uuid.UUID) uuid.UUID {
				return effectstest.IceB1Push(g, o, "Bird", "Creature — Bird", "", 1, 1, "flying")
			},
			func(g *game.Game, o uuid.UUID) uuid.UUID {
				return effectstest.B12Creature(g, o, "Bear", "Creature — Bear", 2, 2)
			}},
	} {
		g := effectstest.NewCatalogGame(t)
		_, opp := effectstest.SeatsForCombat(g)
		legal, illegal := tc.legal(g, opp.ID), tc.illegal(g, opp.ID)
		if err := effectstest.CastCatalogSpellErr(t, g, tc.name, tc.typeLine, tc.oracle, effectstest.CardTarget(illegal)); err == nil {
			t.Errorf("%s targeted a creature it can't", tc.name)
		}
		effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, effectstest.CardTarget(legal))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(legal) {
			t.Errorf("%s did not destroy its target", tc.name)
		}
		if !g.Battlefield.Contains(illegal) {
			t.Errorf("%s destroyed a creature it did not target", tc.name)
		}
	}
}

// --- Purify -----------------------------------------------------------------------------

func TestUlgB1PurifyDestroysArtifactsAndEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	art := effectstest.B12Permanent(g, opp.ID, "Their Artifact", "Artifact")
	ench := effectstest.B12Permanent(g, me.ID, "My Enchantment", "Enchantment")
	golem := effectstest.B12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 3, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Purify", "Sorcery", ulgB1PurifyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{art, ench, golem} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived Purify", effectstest.B12Card(t, g, id).Name)
		}
	}
	for _, id := range []uuid.UUID{bear, land} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was destroyed", effectstest.B12Card(t, g, id).Name)
		}
	}
}

// --- Rebuild ----------------------------------------------------------------------------

func TestUlgB1RebuildReturnsEveryArtifactToItsOwnersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	mine := effectstest.B12Permanent(g, me.ID, "My Artifact", "Artifact")
	stolen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stolen Artifact", TypeLine: "Artifact", Owner: opp.ID, Controller: me.ID,
	})
	ench := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	effectstest.CastCatalogSpell(t, g, "Rebuild", "Instant", ulgB1RebuildOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine) {
		t.Error("your artifact did not return to your hand")
	}
	if !opp.Hand.Contains(stolen) || me.Hand.Contains(stolen) {
		t.Error("a stolen artifact did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(ench) {
		t.Error("Rebuild returned an enchantment")
	}
}

// --- Multani, Maro-Sorcerer ---------------------------------------------------------------

func TestUlgB1MultaniCountsEveryHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	total := 0
	for i, p := range g.Seats {
		effectstest.D2dSetHand(p, i+1)
		total += i + 1
	}
	multani := effectstest.B12Push(g, g.Seats[0].ID, "Multani, Maro-Sorcerer", "Legendary Creature — Elemental Sorcerer", ulgB1MultaniOracle, 0, 0)
	if p, tough := effectstest.EffectivePTOf(t, g, multani); p != total || tough != total {
		t.Errorf("Multani is %d/%d, want %d/%d", p, tough, total, total)
	}
	last := g.Seats[3]
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: last.ID},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: last.ID}, last.Hand.Cards[0].InstanceID); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, multani); p != total-1 {
		t.Errorf("Multani's power is %d after a discard, want %d", p, total-1)
	}
}

// --- Engineered Plague --------------------------------------------------------------------

func TestUlgB1EngineeredPlagueShrinksTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Elf", "Creature — Elf", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf Druid", 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastAndPause(t, g, "Engineered Plague", "Enchantment", ulgB1PlagueOracle)
	effectstest.AnswerCreatureType(t, g, me.ID, "Elf")
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 1 || tough != 1 {
		t.Errorf("your Elf is %d/%d, want 1/1", p, tough)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("a 1/1 Elf survived at 0 toughness")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("a Bear is %d/%d, want 2/2", p, tough)
	}
}

// --- Granite Grip -----------------------------------------------------------------------

func TestUlgB1GraniteGripCountsTheAurasControllersMountains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	for i := 0; i < 3; i++ {
		effectstest.SeedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	effectstest.SeedLandOnBattlefield(g, opp.ID, "Mountain", "Basic Land — Mountain")
	effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AuraCast(t, g, "Granite Grip", ulgB1GraniteGripOracle, theirs)
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 5 || tough != 2 {
		t.Errorf("the enchanted creature is %d/%d, want 5/2 (three of your Mountains)", p, tough)
	}
}

// --- Sleeper's Guile, Sluggishness, Cessation -----------------------------------------------

func TestUlgB1SleepersGuileGivesFear(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AuraCast(t, g, "Sleeper's Guile", ulgB1SleepersGuileOracle, mine)
	green := effectstest.B12Creature(g, them.ID, "Green Bear", "Creature — Bear", 2, 2)
	black := effectstest.ExoB1Push(g, them.ID, "Black Bear", "Creature — Bear", "", 2, 2, "B")
	effectstest.AttackIntoBlocks(t, g, them.ID, mine)
	if effectstest.BlockRefusal(t, g, green, mine) == "" {
		t.Error("a green creature blocked a creature with fear")
	}
	if r := effectstest.BlockRefusal(t, g, black, mine); r != "" {
		t.Errorf("a black creature could not block: %s", r)
	}
}

func TestUlgB1SluggishnessStopsBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, them.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AuraCast(t, g, "Sluggishness", ulgB1SluggishnessOracle, blocker)
	effectstest.AttackIntoBlocks(t, g, them.ID, attacker)
	if effectstest.BlockRefusal(t, g, blocker, attacker) == "" {
		t.Error("the enchanted creature blocked")
	}
}

func TestUlgB1CessationStopsAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AuraCast(t, g, "Cessation", ulgB1CessationOracle, bear)
	effectstest.AdvanceToStepInTurn(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, them.ID); err == nil {
		t.Error("the enchanted creature attacked")
	}
}

func TestUlgB1ReturningAurasGoBackToHand(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Sleeper's Guile", ulgB1SleepersGuileOracle},
		{"Sluggishness", ulgB1SluggishnessOracle},
		{"Cessation", ulgB1CessationOracle},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		aura := effectstest.AuraCast(t, g, tc.name, tc.oracle, bear)
		effectstest.KillCreature(t, g, opp.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(aura) {
			t.Errorf("%s did not return to its owner's hand", tc.name)
		}
	}
}

// --- Shivan Phoenix, Weatherseed Treefolk ------------------------------------------------------

func TestUlgB1DiesReturnToHand(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Shivan Phoenix", "Creature — Phoenix", ulgB1ShivanPhoenixOracle},
		{"Weatherseed Treefolk", "Creature — Treefolk", ulgB1TreefolkOracle},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 3, 4)
		effectstest.KillCreature(t, g, me.ID, id)
		effectstest.DrcSettle(t, g, me.ID)
		if !me.Hand.Contains(id) || me.Graveyard.Contains(id) {
			t.Errorf("%s did not return to hand", tc.name)
		}
		stolen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: tc.name, TypeLine: tc.typeLine, OracleID: tc.oracle,
			Power: 3, Toughness: 4, Owner: me.ID, Controller: opp.ID,
		})
		effectstest.KillCreature(t, g, me.ID, stolen)
		effectstest.DrcSettle(t, g, opp.ID)
		if !me.Hand.Contains(stolen) {
			t.Errorf("a stolen %s did not go to its owner's hand", tc.name)
		}
		exiled := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 3, 4)
		effectstest.ExileCreature(t, g, exiled)
		if me.Hand.Contains(exiled) {
			t.Errorf("an exiled %s returned to hand", tc.name)
		}
	}
}

// --- Fog of Gnats -----------------------------------------------------------------------

func TestUlgB1FogOfGnatsRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gnats := effectstest.B12Push(g, me.ID, "Fog of Gnats", "Creature — Insect", ulgB1FogOfGnatsOracle, 1, 1)
	if err := effectstest.IceB1Activate(g, me.ID, gnats, 0); err == nil {
		t.Fatal("the regeneration was activated with an empty pool")
	}
	effectstest.DcMana(t, g, me.ID, "{B}")
	if err := effectstest.IceB1Activate(g, me.ID, gnats, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DestroyIt(t, g, gnats)
	if !g.Battlefield.Contains(gnats) {
		t.Fatal("Fog of Gnats was destroyed through its shield")
	}
	if !effectstest.FindCardAnywhere(t, g, gnats).Tapped {
		t.Error("the regenerated creature is untapped")
	}
}

// --- Viashino Sandscout, Viashino Cutthroat ----------------------------------------------------

func TestUlgB1ViashinosReturnAtEachEndStep(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Viashino Sandscout", "Creature — Lizard Scout", ulgB1SandscoutOracle},
		{"Viashino Cutthroat", "Creature — Lizard", ulgB1CutthroatOracle},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		id := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, true)
		effectstest.AdvanceToEndStepOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(id) || !me.Hand.Contains(id) {
			t.Errorf("%s did not return to hand at your end step", tc.name)
		}

		g2 := effectstest.NewCatalogGame(t)
		me2 := g2.Seats[0]
		effectstest.AdvanceToUpkeepOf(t, g2, 1)
		id2 := effectstest.PushCatalogPermanent(g2, me2.ID, tc.name, tc.typeLine, tc.oracle, true)
		effectstest.AdvanceToEndStepOf(t, g2, 1)
		effectstest.PassPriorityAroundTable(t, g2)
		if g2.Battlefield.Contains(id2) || !me2.Hand.Contains(id2) {
			t.Errorf("%s did not return to hand at an opponent's end step", tc.name)
		}
	}
}

// --- Phyrexian Debaser, Defiler, Denouncer -------------------------------------------------

func TestUlgB1CarriersSacrificeToShrink(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		n            int
	}{
		{"Phyrexian Debaser", ulgB1DebaserOracle, 2},
		{"Phyrexian Defiler", ulgB1DefilerOracle, 3},
		{"Phyrexian Denouncer", ulgB1DenouncerOracle, 1},
	} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := effectstest.SeatsForCombat(g)
		sick := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Creature — Phyrexian Carrier", tc.oracle, true)
		target := effectstest.B12Creature(g, opp.ID, "Big Beast", "Creature — Beast", 5, 5)
		if err := effectstest.RavB2Activate(t, g, me.ID, sick, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(target)}); err == nil {
			t.Errorf("%s: a summoning-sick Carrier was tapped", tc.name)
		}
		carrier := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Creature — Phyrexian Carrier", tc.oracle, false)
		effectstest.RavB2MustActivate(t, g, me.ID, carrier, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(target)})
		if !me.Graveyard.Contains(carrier) {
			t.Errorf("%s was not sacrificed", tc.name)
		}
		if p, tough := effectstest.EffectivePTOf(t, g, target); p != 5-tc.n || tough != 5-tc.n {
			t.Errorf("%s: the target is %d/%d, want %d/%d", tc.name, p, tough, 5-tc.n, 5-tc.n)
		}
	}
}

// --- Angelic Curator ------------------------------------------------------------------------

func TestUlgB1AngelicCuratorCantBeTargetedByArtifactSources(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	curator := effectstest.B12Push(g, opp.ID, "Angelic Curator", "Creature — Angel Spirit", ulgB1AngelicCuratorOracle, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	pinger := effectstest.PushCatalogPermanent(g, me.ID, "Thornwind Faeries", "Artifact Creature — Faerie", ulgB1ThornwindOracle, false)
	if err := effectstest.RavB2Activate(t, g, me.ID, pinger, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(curator)}); err == nil {
		t.Error("an artifact source targeted Angelic Curator")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, pinger, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if !g.Battlefield.Contains(curator) {
		t.Error("Angelic Curator left the battlefield")
	}
}

// --- Defender of Chaos, Defender of Law ----------------------------------------------------

func TestUlgB1DefendersHaveFlash(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Defender of Chaos", ulgB1DefenderOfChaosOracle},
		{"Defender of Law", ulgB1DefenderOfLawOracle},
	} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		me := g.Seats[0]
		for i := 0; g.Turn.PriorityHolder != 0 && i < 4; i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
		if err := effectstest.S58p6CastNow(g, me, tc.name, "Creature — Human Knight", tc.oracle, nil); err != nil {
			t.Errorf("%s could not be cast during an opponent's upkeep: %v", tc.name, err)
		}
	}
}

// --- Rack and Ruin, Peace and Quiet ---------------------------------------------------------

func TestUlgB1DestroyTwoTargets(t *testing.T) {
	for _, tc := range []struct{ name, oracle, typeLine string }{
		{"Rack and Ruin", ulgB1RackAndRuinOracle, "Artifact"},
		{"Peace and Quiet", ulgB1PeaceAndQuietOracle, "Enchantment"},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		a := effectstest.B12Permanent(g, opp.ID, "First", tc.typeLine)
		b := effectstest.B12Permanent(g, me.ID, "Second", tc.typeLine)
		spare := effectstest.B12Permanent(g, opp.ID, "Spare", tc.typeLine)
		bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		if err := effectstest.CastCatalogSpellErr(t, g, tc.name, "Instant", tc.oracle, effectstest.CardRefs(a, bear)); err == nil {
			t.Errorf("%s targeted a creature", tc.name)
		}
		effectstest.CastCatalogSpell(t, g, tc.name, "Instant", tc.oracle, effectstest.CardRefs(a, b))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
			t.Errorf("%s left a target on the battlefield", tc.name)
		}
		if !g.Battlefield.Contains(spare) {
			t.Errorf("%s destroyed a permanent it did not target", tc.name)
		}
	}
}

// --- Expendable Troops ---------------------------------------------------------------------

func TestUlgB1ExpendableTroopsHitAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	troops := effectstest.B12Push(g, me.ID, "Expendable Troops", "Creature — Human Soldier", ulgB1ExpendableOracle, 2, 1)
	idle := effectstest.B12Creature(g, opp.ID, "Idle Bear", "Creature — Bear", 2, 2)
	if err := effectstest.RavB2Activate(t, g, me.ID, troops, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(idle)}); err == nil {
		t.Fatal("the Troops targeted a creature that is not attacking or blocking")
	}
	attacker := effectstest.B12Creature(g, opp.ID, "Attacking Bear", "Creature — Bear", 2, 2)
	effectstest.CrtBlock(g, uuid.Nil, attacker, me.ID)
	effectstest.RavB2MustActivate(t, g, me.ID, troops, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(attacker)})
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacking 2/2 survived 2 damage")
	}
	if !me.Graveyard.Contains(troops) {
		t.Error("the Troops were not sacrificed")
	}
}

// --- Eviscerator ---------------------------------------------------------------------------

func TestUlgB1EvisceratorCostsFiveLifeToEnter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Eviscerator", "Creature — Phyrexian Horror", ulgB1EvisceratorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-5 {
		t.Errorf("life %d → %d, want -5", life, me.Life)
	}
}
