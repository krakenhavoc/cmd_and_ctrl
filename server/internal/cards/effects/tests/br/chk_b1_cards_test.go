package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// chk_b1_cards_test.go — Champions of Kamigawa, batch 1 (#5603). Each test
// plays the real catalog entry and asserts what its text says.

const (
	chkB1HondenRageOracle      = "0cb76b04-d9c3-46fc-a3f2-5f0cb0d91761"
	chkB1HondenWindsOracle     = "45beca46-480c-4771-ba46-78b6ac7d5c2c"
	chkB1HondenFireOracle      = "15ee7e1d-b13e-4fd0-bc41-345b6ed34f53"
	chkB1HondenWebOracle       = "a1fdfcc6-4b3c-40f8-a9e4-dff9ace74a59"
	chkB1HondenReachOracle     = "3776d11b-18fb-4451-a8f8-bda4c7c23776"
	chkB1TimeOfNeedOracle      = "4f2adfd0-c8ab-4bcc-ae55-cb0e798aec7f"
	chkB1EiganjoOracle         = "895a0e00-20a9-44f8-9215-66edcdf016b7"
	chkB1OkinaOracle           = "3ad69bfb-2e51-4fe9-8d2f-7d071a4f1c69"
	chkB1RyuseiOracle          = "5e99ba3b-dc2b-4c3c-84ce-228d4772cfee"
	chkB1KondaOracle           = "efea2342-bb24-472a-acfb-1b9e012af91b"
	chkB1SeshiroOracle         = "54fc373e-157f-4157-b7dd-74d689f4c0f9"
	chkB1LavaSpikeOracle       = "2837888c-bfa8-4955-9334-0605ad409f7e"
	chkB1ReachThroughOracle    = "c81ca8ff-92f3-481e-82f7-0673b6c74ea0"
	chkB1RendFleshOracle       = "079d4698-dc11-4d29-91b6-725d78dfb75d"
	chkB1EyeOfNowhereOracle    = "27ffdf11-bb7f-40bf-94c4-6bfbc41668e5"
	chkB1QuietPurityOracle     = "3f16dc14-3d3f-4ffa-90bc-9abc67db75cf"
	chkB1CleanfallOracle       = "826b1c93-4e5d-4436-8d59-2d7d970f609f"
	chkB1CounselOracle         = "62ddc5ae-ced9-4319-854c-1a114c6afc3f"
	chkB1WakingNightmareOracle = "f3f34ef2-1255-4da2-9f29-8c79fee7a2c5"
	chkB1PullUnderOracle       = "b99ddb27-a59a-4cd3-88b1-97c759ddbd93"
	chkB1SwallowingOracle      = "91cac048-90ca-41bc-9177-56c78453c4a2"
	chkB1UnnaturalSpeedOracle  = "fa46ced5-b701-4aa3-b4a1-36fb15b80696"
	chkB1DefianceOracle        = "2ac6f1c7-f4c5-4d45-9644-da49d8fe4758"
	chkB1RendSpiritOracle      = "cca698b3-b86f-45a2-aa3f-bab7d5a56e8a"
	chkB1ThoughtbindOracle     = "f31bb2b4-061c-4ba6-9f91-374f40fe6b05"
	chkB1SustainerOracle       = "d3039b42-25d5-4fe2-955a-0d910d7395bd"
	chkB1LeafcallerOracle      = "081db161-1ad0-4ee7-9d34-c415984080f0"
	chkB1SireOracle            = "68f1db12-82fb-4bf1-908d-db38fd67efe5"
	chkB1ThiefOfHopeOracle     = "d3fa1320-9e95-437a-aa36-102a2722fc42"
	chkB1KamiOfTheHuntOracle   = "96e68c7a-f187-4aa6-998c-238e7c833809"
	chkB1BlademasterOracle     = "2db73fff-1a89-4006-945d-a244ff2f0bef"
	chkB1EnforcersOracle       = "9d0d19e4-34f9-4bfa-9c8b-74356d223a18"
	chkB1GibberingOracle       = "8a5610bd-5f04-43f3-9e2e-d6c728ee899b"
	chkB1LunacyOracle          = "5c848e4a-f0f7-4659-b7b1-f468efba9563"
	chkB1PalaceFieldsOracle    = "e8d44867-9519-41f2-aa0f-4079444aeef9"
	chkB1HundredTalonOracle    = "100b4440-c588-423e-93d2-23c9b595614d"
	chkB1KumoOracle            = "3910a155-89dc-45fe-98ec-762e8e42f052"
	chkB1KodamaOracle          = "15d8f129-2518-45f4-9f34-a4fcd4d859af"
	chkB1DeathcurseOracle      = "0d40bc98-28e7-4fa5-b848-1cbac5345c12"
	chkB1KiteOracle            = "10aab5bc-5758-443b-ba4c-49f6c6d91262"
)

func TestChkB1EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		chkB1HondenRageOracle:      "Honden of Infinite Rage",
		chkB1HondenWindsOracle:     "Honden of Seeing Winds",
		chkB1HondenFireOracle:      "Honden of Cleansing Fire",
		chkB1HondenWebOracle:       "Honden of Life's Web",
		chkB1HondenReachOracle:     "Honden of Night's Reach",
		chkB1TimeOfNeedOracle:      "Time of Need",
		chkB1EiganjoOracle:         "Eiganjo Castle",
		chkB1OkinaOracle:           "Okina, Temple to the Grandfathers",
		chkB1RyuseiOracle:          "Ryusei, the Falling Star",
		chkB1KondaOracle:           "Konda, Lord of Eiganjo",
		chkB1SeshiroOracle:         "Seshiro the Anointed",
		chkB1LavaSpikeOracle:       "Lava Spike",
		chkB1ReachThroughOracle:    "Reach Through Mists",
		chkB1RendFleshOracle:       "Rend Flesh",
		chkB1EyeOfNowhereOracle:    "Eye of Nowhere",
		chkB1QuietPurityOracle:     "Quiet Purity",
		chkB1CleanfallOracle:       "Cleanfall",
		chkB1CounselOracle:         "Counsel of the Soratami",
		chkB1WakingNightmareOracle: "Waking Nightmare",
		chkB1PullUnderOracle:       "Pull Under",
		chkB1SwallowingOracle:      "Swallowing Plague",
		chkB1UnnaturalSpeedOracle:  "Unnatural Speed",
		chkB1DefianceOracle:        "Hisoka's Defiance",
		chkB1RendSpiritOracle:      "Rend Spirit",
		chkB1ThoughtbindOracle:     "Thoughtbind",
		chkB1SustainerOracle:       "Orochi Sustainer",
		chkB1LeafcallerOracle:      "Orochi Leafcaller",
		chkB1SireOracle:            "Sire of the Storm",
		chkB1ThiefOfHopeOracle:     "Thief of Hope",
		chkB1KamiOfTheHuntOracle:   "Kami of the Hunt",
		chkB1BlademasterOracle:     "Kitsune Blademaster",
		chkB1EnforcersOracle:       "Samurai Enforcers",
		chkB1GibberingOracle:       "Gibbering Kami",
		chkB1LunacyOracle:          "Kami of Lunacy",
		chkB1PalaceFieldsOracle:    "Kami of the Palace Fields",
		chkB1HundredTalonOracle:    "Hundred-Talon Kami",
		chkB1KumoOracle:            "Venerable Kumo",
		chkB1KodamaOracle:          "Kodama of the North Tree",
		chkB1DeathcurseOracle:      "Deathcurse Ogre",
		chkB1KiteOracle:            "Journeyer's Kite",
	}
	if len(want) != 40 {
		t.Fatalf("the batch registers 40 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
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

func TestChkB1CardsHaveTheirPrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Ryusei, the Falling Star", "Legendary Creature — Dragon Spirit", chkB1RyuseiOracle, []string{"flying"}},
		{"Konda, Lord of Eiganjo", "Legendary Creature — Human Samurai", chkB1KondaOracle, []string{"vigilance", "indestructible"}},
		{"Kitsune Blademaster", "Creature — Fox Samurai", chkB1BlademasterOracle, []string{"first strike"}},
		{"Sire of the Storm", "Creature — Spirit", chkB1SireOracle, []string{"flying"}},
		{"Gibbering Kami", "Creature — Spirit", chkB1GibberingOracle, []string{"flying"}},
		{"Kami of Lunacy", "Creature — Spirit", chkB1LunacyOracle, []string{"flying"}},
		{"Kami of the Palace Fields", "Creature — Spirit", chkB1PalaceFieldsOracle, []string{"flying", "first strike"}},
		{"Hundred-Talon Kami", "Creature — Spirit", chkB1HundredTalonOracle, []string{"flying"}},
		{"Venerable Kumo", "Creature — Spirit", chkB1KumoOracle, []string{"reach"}},
		{"Kodama of the North Tree", "Legendary Creature — Spirit", chkB1KodamaOracle, []string{"trample", "shroud"}},
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

// --- the Hondens ------------------------------------------------------------------------

// chkB1HondenBoard puts a Honden, a second Shrine and a plain enchantment
// under seat 1, and a Shrine under seat 2 that must not count, then
// moves to seat 1's upkeep.
func chkB1HondenBoard(t *testing.T, name, oracle string) (g *game.Game, me, opp *game.Player) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[1], g.Seats[2]
	effectstest.B12Push(g, me.ID, name, "Legendary Enchantment — Shrine", oracle, 0, 0)
	effectstest.B12Permanent(g, me.ID, "Other Shrine", "Legendary Enchantment — Shrine")
	effectstest.B12Permanent(g, me.ID, "Plain Enchantment", "Enchantment")
	effectstest.B12Permanent(g, opp.ID, "Their Shrine", "Legendary Enchantment — Shrine")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	return g, me, opp
}

func TestChkB1HondensCountYourShrinesAtYourUpkeep(t *testing.T) {
	g, me, _ := chkB1HondenBoard(t, "Honden of Seeing Winds", chkB1HondenWindsOracle)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("Seeing Winds: hand %d → %d, want two draws for two Shrines", hand, me.Hand.Size())
	}

	g, me, _ = chkB1HondenBoard(t, "Honden of Cleansing Fire", chkB1HondenFireOracle)
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("Cleansing Fire: life %d → %d, want +4 for two Shrines", life, me.Life)
	}

	g, me, _ = chkB1HondenBoard(t, "Honden of Life's Web", chkB1HondenWebOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Spirit"); n != 2 {
		t.Errorf("Life's Web: %d Spirit tokens, want 2", n)
	}
}

func TestChkB1HondensDoNothingOnAnotherPlayersUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.B12Push(g, me.ID, "Honden of Cleansing Fire", "Legendary Enchantment — Shrine", chkB1HondenFireOracle, 0, 0)
	effectstest.AdvanceToUpkeepOf(t, g, 2)
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("an opponent's upkeep changed life %d → %d", life, me.Life)
	}
}

func TestChkB1HondenOfInfiniteRageDealsDamagePerShrine(t *testing.T) {
	g, me, opp := chkB1HondenBoard(t, "Honden of Infinite Rage", chkB1HondenRageOracle)
	p := effectstest.PassUntilPickTarget(t, g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the upkeep trigger")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	life := opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d → %d, want 2 damage for two Shrines", life, opp.Life)
	}
}

func TestChkB1HondenOfNightsReachMakesAnOpponentDiscardPerShrine(t *testing.T) {
	g, me, opp := chkB1HondenBoard(t, "Honden of Night's Reach", chkB1HondenReachOracle)
	effectstest.D2dSetHand(opp, 3)
	p := effectstest.PassUntilPickTarget(t, g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the upkeep trigger")
	}
	if effectstest.HasID(p.PickTargetPlayers, me.ID) {
		t.Error("the Honden's controller is offered as the target opponent")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DiscardChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("no discard prompt for the target opponent")
	}
	if c.ChooseMin != 2 || c.ChooseMax != 2 {
		t.Errorf("discard bounds [%d,%d], want two for two Shrines", c.ChooseMin, c.ChooseMax)
	}
	effectstest.AnswerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID, opp.Hand.Cards[1].InstanceID)
	if opp.Hand.Size() != 1 {
		t.Errorf("the opponent holds %d cards, want 1", opp.Hand.Size())
	}
}

func TestChkB1TimeOfNeedFindsALegendaryCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	legend := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Some Legend", TypeLine: "Legendary Creature — Human"})
	bear := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Some Bear", TypeLine: "Creature — Bear"})
	land := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Some Legendary Land", TypeLine: "Legendary Land"})
	effectstest.CastCatalogSpell(t, g, "Time of Need", "Sorcery", chkB1TimeOfNeedOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	open := effectstest.ChkB1OpenSearch(g)
	if open == nil {
		t.Fatal("no search prompt")
	}
	effectstest.BeginSearchIfAsked(t, g, open)
	open = effectstest.ChkB1OpenSearch(g)
	if !effectstest.HasID(open.SearchCards, legend) || effectstest.HasID(open.SearchCards, bear) || effectstest.HasID(open.SearchCards, land) {
		t.Errorf("the search offers %v, want only the legendary creature card %v", open.SearchCards, legend)
	}
	effectstest.TakeSearchPrompts(t, g)
	if !me.Hand.Contains(legend) {
		t.Error("the legendary creature card did not reach the hand")
	}
}

func TestChkB1JourneyersKiteFindsABasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	dual := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Some Dual", TypeLine: "Land — Forest Island"})
	kite := effectstest.PushCatalogPermanent(g, me.ID, "Journeyer's Kite", "Artifact", chkB1KiteOracle, true)
	if err := effectstest.RavB2Activate(t, g, me.ID, kite, 0, game.ActivateAbilityParams{}, "C", "C"); err == nil {
		t.Fatal("the Kite activated for two mana")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, kite, 0, game.ActivateAbilityParams{}, "C")
	if !effectstest.IsTapped(g, kite) {
		t.Error("the Kite did not tap as the cost")
	}
	open := effectstest.ChkB1OpenSearch(g)
	if open == nil {
		t.Fatal("no search prompt")
	}
	effectstest.BeginSearchIfAsked(t, g, open)
	open = effectstest.ChkB1OpenSearch(g)
	if !effectstest.HasID(open.SearchCards, forest) || effectstest.HasID(open.SearchCards, dual) {
		t.Errorf("the search offers %v, want only the basic land %v", open.SearchCards, forest)
	}
	effectstest.TakeSearchPrompts(t, g)
	if !me.Hand.Contains(forest) {
		t.Error("the basic land did not reach the hand")
	}
}

// --- Eiganjo Castle, Okina ---------------------------------------------------------------

func TestChkB1LegendaryLandsTapForTheirColour(t *testing.T) {
	for _, tc := range []struct{ name, oracle, colour string }{
		{"Eiganjo Castle", chkB1EiganjoOracle, "W"},
		{"Okina, Temple to the Grandfathers", chkB1OkinaOracle, "G"},
	} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		land := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Legendary Land", tc.oracle, false)
		if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if n := me.ManaPool.CountColor(tc.colour); n != 1 || me.ManaPool.Total() != 1 {
			t.Errorf("%s: pool holds %d {%s} of %d, want 1 of 1", tc.name, n, tc.colour, me.ManaPool.Total())
		}
	}
}

func TestChkB1EiganjoCastleShieldsALegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	castle := effectstest.PushCatalogPermanent(g, me.ID, "Eiganjo Castle", "Legendary Land", chkB1EiganjoOracle, false)
	legend := effectstest.B12Creature(g, me.ID, "My Legend", "Legendary Creature — Human Samurai", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := effectstest.RavB2Activate(t, g, me.ID, castle, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}, "W"); err == nil {
		t.Fatal("the Castle targeted a nonlegendary creature")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, castle, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(legend)}, "W")
	effectstest.Damage(t, g, uuid.Nil, legend, 3)
	if c := effectstest.B12Card(t, g, legend); c.DamageMarked != 1 {
		t.Errorf("the legend has %d damage marked, want 1 (3 less the 2-point shield)", c.DamageMarked)
	}
}

func TestChkB1OkinaPumpsALegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	okina := effectstest.PushCatalogPermanent(g, me.ID, "Okina, Temple to the Grandfathers", "Legendary Land", chkB1OkinaOracle, false)
	legend := effectstest.B12Creature(g, me.ID, "My Legend", "Legendary Creature — Snake Monk", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := effectstest.RavB2Activate(t, g, me.ID, okina, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}, "G"); err == nil {
		t.Fatal("Okina targeted a nonlegendary creature")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, okina, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(legend)}, "G")
	if p, tough := effectstest.EffectivePTOf(t, g, legend); p != 3 || tough != 3 {
		t.Errorf("the legend is %d/%d, want 3/3", p, tough)
	}
	effectstest.DcCleanup(g)
	if p, tough := effectstest.EffectivePTOf(t, g, legend); p != 2 || tough != 2 {
		t.Errorf("the bonus outlived the cleanup step (%d/%d)", p, tough)
	}
}

// --- Ryusei, the Falling Star --------------------------------------------------------------

func TestChkB1RyuseiDiesAndBurnsEveryCreatureWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ryusei := effectstest.B12Push(g, me.ID, "Ryusei, the Falling Star", "Legendary Creature — Dragon Spirit", chkB1RyuseiOracle, 5, 5)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	big := effectstest.B12Creature(g, opp.ID, "Their Wurm", "Creature — Wurm", 6, 6)
	bird := effectstest.IceB1Push(g, opp.ID, "Their Bird", "Creature — Bird", "", 1, 1, "flying")
	life := opp.Life
	effectstest.KillCreature(t, g, me.ID, ryusei)
	effectstest.DrcSettle(t, g, me.ID)
	g.RunStateChecksForTest()
	for _, id := range []uuid.UUID{mine, theirs} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived 5 damage", effectstest.B12Card(t, g, id).Name)
		}
	}
	if !g.Battlefield.Contains(bird) {
		t.Error("a creature with flying was hit")
	}
	if c := effectstest.B12Card(t, g, big); !g.Battlefield.Contains(big) || c.DamageMarked != 5 {
		t.Errorf("the 6/6 has %d damage marked, want 5 and alive", c.DamageMarked)
	}
	if opp.Life != life {
		t.Errorf("a player was dealt damage (%d → %d)", life, opp.Life)
	}
}

// --- bushido: Konda, Kitsune Blademaster, Samurai Enforcers ----------------------------------

func TestChkB1BushidoCreaturesGrowWhenBlocked(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, tough, n        int
	}{
		{"Konda, Lord of Eiganjo", "Legendary Creature — Human Samurai", chkB1KondaOracle, 3, 3, 5},
		{"Kitsune Blademaster", "Creature — Fox Samurai", chkB1BlademasterOracle, 2, 2, 1},
		{"Samurai Enforcers", "Creature — Human Samurai", chkB1EnforcersOracle, 4, 4, 2},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventBecomesBlocked, Actor: opp.ID, Source: id, CardID: id, Target: id})
		})
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.EffectivePTOf(t, g, id); p != tc.power+tc.n || tough != tc.tough+tc.n {
			t.Errorf("%s is %d/%d after becoming blocked, want %d/%d", tc.name, p, tough, tc.power+tc.n, tc.tough+tc.n)
		}
	}
}

func TestChkB1KondaIsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	konda := effectstest.B12Push(g, g.Seats[0].ID, "Konda, Lord of Eiganjo", "Legendary Creature — Human Samurai", chkB1KondaOracle, 3, 3)
	effectstest.DestroyIt(t, g, konda)
	if !g.Battlefield.Contains(konda) {
		t.Error("Konda was destroyed")
	}
}

// --- Seshiro the Anointed ------------------------------------------------------------------

func TestChkB1SeshiroPumpsOtherSnakesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seshiro := effectstest.B12Push(g, me.ID, "Seshiro the Anointed", "Legendary Creature — Snake Monk", chkB1SeshiroOracle, 3, 4)
	snake := effectstest.B12Creature(g, me.ID, "My Snake", "Creature — Snake Warrior", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Snake", "Creature — Snake", 1, 1)
	for _, tc := range []struct {
		id           uuid.UUID
		power, tough int
		what         string
	}{
		{seshiro, 3, 4, "Seshiro itself"}, {snake, 3, 3, "your other Snake"},
		{bear, 2, 2, "your non-Snake"}, {theirs, 1, 1, "an opponent's Snake"},
	} {
		if p, tough := effectstest.EffectivePTOf(t, g, tc.id); p != tc.power || tough != tc.tough {
			t.Errorf("%s is %d/%d, want %d/%d", tc.what, p, tough, tc.power, tc.tough)
		}
	}
}

func TestChkB1SeshiroDrawsWhenASnakeConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Seshiro the Anointed", "Legendary Creature — Snake Monk", chkB1SeshiroOracle, 3, 4)
	snake := effectstest.B12Creature(g, me.ID, "My Snake", "Creature — Snake Warrior", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, snake, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one draw for the one Snake that connected", hand, me.Hand.Size())
	}
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Error("the non-Snake asked for a draw too")
	}
}

// --- burn, draw, removal -----------------------------------------------------------------

func TestChkB1LavaSpikeHitsAPlayerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Lava Spike", "Sorcery — Arcane", chkB1LavaSpikeOracle, effectstest.CardTarget(bear)); err == nil {
		t.Error("Lava Spike targeted a creature")
	}
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Lava Spike", "Sorcery — Arcane", chkB1LavaSpikeOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want 3 damage", life, opp.Life)
	}
}

func TestChkB1DrawSpellsDraw(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		n                      int
	}{
		{"Reach Through Mists", "Instant — Arcane", chkB1ReachThroughOracle, 1},
		{"Counsel of the Soratami", "Sorcery", chkB1CounselOracle, 2},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		hand := me.Hand.Size()
		effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Size() != hand+tc.n {
			t.Errorf("%s: hand %d → %d, want +%d", tc.name, hand, me.Hand.Size(), tc.n)
		}
	}
}

func TestChkB1DestroySpellsCheckTheirTarget(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle         string
		legalType, illegalType         string
		legalCreature, illegalCreature bool
	}{
		{"Rend Flesh", "Instant — Arcane", chkB1RendFleshOracle, "Creature — Bear", "Creature — Spirit", true, true},
		{"Rend Spirit", "Instant", chkB1RendSpiritOracle, "Creature — Spirit", "Creature — Bear", true, true},
		{"Quiet Purity", "Instant — Arcane", chkB1QuietPurityOracle, "Enchantment", "Creature — Bear", false, true},
	} {
		g := effectstest.NewCatalogGame(t)
		_, opp := effectstest.SeatsForCombat(g)
		mk := func(name, typeLine string, creature bool) uuid.UUID {
			if creature {
				return effectstest.B12Creature(g, opp.ID, name, typeLine, 2, 2)
			}
			return effectstest.B12Permanent(g, opp.ID, name, typeLine)
		}
		legal := mk("Legal", tc.legalType, tc.legalCreature)
		illegal := mk("Illegal", tc.illegalType, tc.illegalCreature)
		if err := effectstest.CastCatalogSpellErr(t, g, tc.name, tc.typeLine, tc.oracle, effectstest.CardTarget(illegal)); err == nil {
			t.Errorf("%s targeted a %s", tc.name, tc.illegalType)
		}
		effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, effectstest.CardTarget(legal))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(legal) || !opp.Graveyard.Contains(legal) {
			t.Errorf("%s did not destroy its target", tc.name)
		}
		if !g.Battlefield.Contains(illegal) {
			t.Errorf("%s destroyed a permanent it did not target", tc.name)
		}
	}
}

func TestChkB1EyeOfNowhereReturnsAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	ench := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	effectstest.CastCatalogSpell(t, g, "Eye of Nowhere", "Sorcery — Arcane", chkB1EyeOfNowhereOracle, effectstest.CardTarget(ench))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) || !opp.Hand.Contains(ench) {
		t.Error("the enchantment did not return to its owner's hand")
	}
}

func TestChkB1CleanfallDestroysEveryEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	mine := effectstest.B12Permanent(g, me.ID, "My Enchantment", "Enchantment")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment — Aura")
	art := effectstest.B12Permanent(g, opp.ID, "Their Artifact", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Cleanfall", "Sorcery — Arcane", chkB1CleanfallOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived Cleanfall", effectstest.B12Card(t, g, id).Name)
		}
	}
	for _, id := range []uuid.UUID{art, bear} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was destroyed", effectstest.B12Card(t, g, id).Name)
		}
	}
}

func TestChkB1WakingNightmareDiscardsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	effectstest.D2dSetHand(opp, 3)
	effectstest.CastCatalogSpell(t, g, "Waking Nightmare", "Sorcery — Arcane", chkB1WakingNightmareOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DiscardChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("no discard prompt for the target")
	}
	if c.ChooseMin != 2 || c.ChooseMax != 2 {
		t.Errorf("discard bounds [%d,%d], want [2,2]", c.ChooseMin, c.ChooseMax)
	}
	effectstest.AnswerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID, opp.Hand.Cards[1].InstanceID)
	if opp.Hand.Size() != 1 {
		t.Errorf("the target holds %d cards, want 1", opp.Hand.Size())
	}
}

func TestChkB1PullUnderShrinksUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	wurm := effectstest.B12Creature(g, opp.ID, "Their Wurm", "Creature — Wurm", 6, 6)
	effectstest.CastCatalogSpell(t, g, "Pull Under", "Instant — Arcane", chkB1PullUnderOracle, effectstest.CardTarget(wurm))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, wurm); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tough)
	}
	effectstest.DcCleanup(g)
	if p, tough := effectstest.EffectivePTOf(t, g, wurm); p != 6 || tough != 6 {
		t.Errorf("the shrink outlived the cleanup step (%d/%d)", p, tough)
	}
}

func TestChkB1SwallowingPlagueDealsXAndGainsX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	giant := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	life := me.Life
	effectstest.CastXSpell(t, g, "Swallowing Plague", "Sorcery — Arcane", chkB1SwallowingOracle, "{X}{B}{B}", 3, effectstest.CardTarget(giant))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.B12Card(t, g, giant); c.DamageMarked != 3 {
		t.Errorf("the target has %d damage marked, want 3", c.DamageMarked)
	}
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
}

func TestChkB1UnnaturalSpeedGrantsHasteUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Unnatural Speed", "Instant — Arcane", chkB1UnnaturalSpeedOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("the target lacks haste")
	}
	effectstest.DcCleanup(g)
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("haste outlived the cleanup step")
	}
}

// --- counterspells --------------------------------------------------------------------------

func TestChkB1HisokasDefianceCountersSpiritOrArcaneSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if err := effectstest.CastCatalogSpellErr(t, g, "Hisoka's Defiance", "Instant", chkB1DefianceOracle, effectstest.B16TargetCard(bear)); err == nil {
		t.Error("Hisoka's Defiance targeted a Bear spell")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, tc := range []struct{ Name, TypeLine string }{
		{"A Spirit", "Creature — Spirit"},
		{"An Arcane Spell", "Sorcery — Arcane"},
	} {
		spell := effectstest.CastCatalogSpell(t, g, tc.Name, tc.TypeLine, "", nil)
		effectstest.CastCatalogSpell(t, g, "Hisoka's Defiance", "Instant", chkB1DefianceOracle, effectstest.B16TargetCard(spell))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(spell) || !me.Graveyard.Contains(spell) {
			t.Errorf("%s was not countered", tc.Name)
		}
	}
}

func TestChkB1ThoughtbindCountersManaValueFourOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	big := effectstest.CastWithCost(t, g, "Wurm", "Creature — Wurm", "{3}{G}{G}", "")
	if err := effectstest.CastCatalogSpellErr(t, g, "Thoughtbind", "Instant", chkB1ThoughtbindOracle, effectstest.B16TargetCard(big)); err == nil {
		t.Error("Thoughtbind targeted a mana value 5 spell")
	}
	effectstest.PassPriorityAroundTable(t, g)
	small := effectstest.CastWithCost(t, g, "Elk", "Creature — Elk", "{2}{G}{G}", "")
	effectstest.CastCatalogSpell(t, g, "Thoughtbind", "Instant", chkB1ThoughtbindOracle, effectstest.B16TargetCard(small))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || !me.Graveyard.Contains(small) {
		t.Error("the mana value 4 spell was not countered")
	}
}

// --- mana creatures -----------------------------------------------------------------------

func TestChkB1OrochiSustainerTapsForGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sick := effectstest.PushCatalogPermanent(g, me.ID, "Orochi Sustainer", "Creature — Snake Shaman", chkB1SustainerOracle, true)
	if err := g.ActivateManaAbility(me.ID, sick, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a summoning-sick Sustainer tapped for mana")
	}
	ready := effectstest.PushCatalogPermanent(g, me.ID, "Orochi Sustainer", "Creature — Snake Shaman", chkB1SustainerOracle, false)
	if err := g.ActivateManaAbility(me.ID, ready, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if n := me.ManaPool.CountColor("G"); n != 1 || me.ManaPool.Total() != 1 {
		t.Errorf("pool holds %d {G} of %d, want 1 of 1", n, me.ManaPool.Total())
	}
}

func TestChkB1OrochiLeafcallerFiltersGreenIntoAnyColour(t *testing.T) {
	for _, colour := range []string{"W", "U", "B", "R", "G"} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		caller := effectstest.PushCatalogPermanent(g, me.ID, "Orochi Leafcaller", "Creature — Snake Shaman", chkB1LeafcallerOracle, true)
		if err := g.ActivateManaAbility(me.ID, caller, 0, game.ManaAbilityParams{Colors: []string{colour}}); err == nil {
			t.Fatal("the ability activated with no {G} to pay")
		}
		effectstest.ApaMana(me, "G")
		if err := g.ActivateManaAbility(me.ID, caller, 0, game.ManaAbilityParams{Colors: []string{colour}}); err != nil {
			t.Fatalf("a summoning-sick Leafcaller can still filter: %v", err)
		}
		if n := me.ManaPool.CountColor(colour); n != 1 || me.ManaPool.Total() != 1 {
			t.Errorf("picking %s: pool holds %d of %d, want 1 of 1", colour, n, me.ManaPool.Total())
		}
		if effectstest.AlliTapped(t, g, caller) {
			t.Error("the ability tapped the Leafcaller")
		}
	}
}

// --- "Whenever you cast a Spirit or Arcane spell" ---------------------------------------------

func TestChkB1SireOfTheStormMayDrawOnASpiritOrArcaneSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Sire of the Storm", "Creature — Spirit", chkB1SireOracle, 3, 3)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil || me.Hand.Size() != hand {
		t.Error("a Bear spell triggered Sire of the Storm")
	}
	effectstest.CastCatalogSpell(t, g, "An Arcane Spell", "Instant — Arcane", "", nil)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one draw", hand, me.Hand.Size())
	}
}

func TestChkB1ThiefOfHopeDrainsOnASpiritOrArcaneSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Thief of Hope", "Creature — Spirit", chkB1ThiefOfHopeOracle, 2, 2)
	mine, theirs := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "A Spirit", "Creature — Spirit", "", nil)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-1 || me.Life != mine+1 {
		t.Errorf("life: yours %d → %d, theirs %d → %d; want +1 and -1", mine, me.Life, theirs, opp.Life)
	}
}

func TestChkB1KamiOfTheHuntGrowsPerSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kami := effectstest.B12Push(g, me.ID, "Kami of the Hunt", "Creature — Spirit", chkB1KamiOfTheHuntOracle, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "A Spirit", "Creature — Spirit", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "An Arcane Spell", "Sorcery — Arcane", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, kami); p != 4 || tough != 4 {
		t.Errorf("the Kami is %d/%d, want 4/4 (two Spirit or Arcane spells)", p, tough)
	}
	effectstest.DcCleanup(g)
	if p, tough := effectstest.EffectivePTOf(t, g, kami); p != 2 || tough != 2 {
		t.Errorf("the bonus outlived the cleanup step (%d/%d)", p, tough)
	}
}

// --- soulshift ------------------------------------------------------------------------------

func TestChkB1SoulshiftersReturnASmallerSpirit(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, cost string
		n                  int
	}{
		{"Thief of Hope", chkB1ThiefOfHopeOracle, "{2}{B}", 2},
		{"Gibbering Kami", chkB1GibberingOracle, "{3}{B}", 3},
		{"Kami of Lunacy", chkB1LunacyOracle, "{4}{B}{B}", 5},
		{"Kami of the Palace Fields", chkB1PalaceFieldsOracle, "{5}{W}", 5},
		{"Hundred-Talon Kami", chkB1HundredTalonOracle, "{4}{W}", 4},
		{"Venerable Kumo", chkB1KumoOracle, "{4}{G}", 4},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		kami := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: tc.name, TypeLine: "Creature — Spirit", OracleID: tc.oracle,
			ManaCost: tc.cost, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		})
		fits, tooBig, bear := uuid.New(), uuid.New(), uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: fits, Name: "Fits", TypeLine: "Creature — Spirit",
			ManaCost: effectstest.ChkB1GenericCost(tc.n), Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: tooBig, Name: "Too Big", TypeLine: "Creature — Spirit",
			ManaCost: effectstest.ChkB1GenericCost(tc.n + 1), Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Bear Card", TypeLine: "Creature — Bear",
			ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
		effectstest.KillCreature(t, g, me.ID, kami)
		p := effectstest.PassUntilPickTarget(t, g, me.ID)
		if p == nil {
			t.Fatalf("%s: no soulshift target prompt", tc.name)
		}
		if !effectstest.HasID(p.PickTargetCards, fits) || effectstest.HasID(p.PickTargetCards, tooBig) || effectstest.HasID(p.PickTargetCards, bear) {
			t.Errorf("%s: soulshift offers %v, want the mana value %d Spirit %v", tc.name, p.PickTargetCards, tc.n, fits)
		}
		effectstest.PickCard(t, g, me.ID, fits)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(fits) {
			t.Errorf("%s: soulshift did not return the Spirit card", tc.name)
		}
	}
}

// --- Deathcurse Ogre ----------------------------------------------------------------------

func TestChkB1DeathcurseOgreDrainsEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ogre := effectstest.B12Push(g, me.ID, "Deathcurse Ogre", "Creature — Ogre Warrior", chkB1DeathcurseOracle, 3, 3)
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Life
	}
	effectstest.KillCreature(t, g, me.ID, ogre)
	effectstest.DrcSettle(t, g, me.ID)
	for i, p := range g.Seats {
		if p.Life != before[i]-3 {
			t.Errorf("seat %d: life %d → %d, want -3", i, before[i], p.Life)
		}
	}
}
