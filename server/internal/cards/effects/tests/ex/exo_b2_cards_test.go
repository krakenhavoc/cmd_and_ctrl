package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exo_b2_cards_test.go — Exodus, batch 2 (#5626). Each test plays the
// real catalog entry and asserts what its text says.

const (
	exoB2WyvernOracle         = "acaae632-3203-4867-ad03-32619c3fcfe6"
	exoB2PeaceOfMindOracle    = "4f8c5fd7-f280-4b0c-bb84-6ff9b258c50f"
	exoB2FugueOracle          = "724c8a20-bfb3-4c63-8cfa-5ad4c72ebe2f"
	exoB2ScareTacticsOracle   = "d7548770-5c60-4c3f-b1d6-554ab5f931f9"
	exoB2ShieldMateOracle     = "98e7bfd4-fcb6-47b1-ad0b-c54b85dc2a79"
	exoB2AlligatorOracle      = "2bf84938-6049-4d03-b09e-85bd0763a187"
	exoB2TreasureHunterOracle = "71ae3bc0-28da-4e91-bffd-a49fe545b7a9"
	exoB2BaboonsOracle        = "1d9bd0a5-3ba0-44e5-9cb3-966f6b9dc14a"
	exoB2WaywardSoulOracle    = "319b4715-c3eb-49bf-b8d2-4d41e601ebe4"
	exoB2PiranhaOracle        = "fad69c20-d6f4-4cf5-8e4c-bbda8d268edd"
	exoB2WolverinesOracle     = "ed076c39-9bfc-424a-8ce8-8572f7c0de73"
	exoB2PygmyTrollOracle     = "6f18f41e-df8d-4b59-87cc-ea167823fbda"
	exoB2RootwallaOracle      = "e20b80f2-9913-4886-a61b-04f3fd27c2ad"
	exoB2DriftersOracle       = "1afa006b-1190-4047-a9fe-4215f20df8f2"
	exoB2HoundsOracle         = "0184d74b-e9aa-4b5a-8db0-b15cecbc17fb"
	exoB2WelkinHawkOracle     = "4547625a-0caf-4ecc-a65f-0fd2cad70294"
)

func TestExoB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		exoB2WyvernOracle:         "Sabertooth Wyvern",
		exoB2PeaceOfMindOracle:    "Peace of Mind",
		exoB2FugueOracle:          "Fugue",
		exoB2ScareTacticsOracle:   "Scare Tactics",
		exoB2ShieldMateOracle:     "Shield Mate",
		exoB2AlligatorOracle:      "Rootwater Alligator",
		exoB2TreasureHunterOracle: "Treasure Hunter",
		exoB2BaboonsOracle:        "Ravenous Baboons",
		exoB2WaywardSoulOracle:    "Wayward Soul",
		exoB2PiranhaOracle:        "School of Piranha",
		exoB2WolverinesOracle:     "Rabid Wolverines",
		exoB2PygmyTrollOracle:     "Pygmy Troll",
		exoB2RootwallaOracle:      "Plated Rootwalla",
		exoB2DriftersOracle:       "Thalakos Drifters",
		exoB2HoundsOracle:         "Vampire Hounds",
		exoB2WelkinHawkOracle:     "Welkin Hawk",
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

// exoB2Activate floats mana for the active player's ability 0 on id and
// activates it.
func exoB2Activate(t *testing.T, g *game.Game, who *game.Player, id uuid.UUID, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	if mana != "" {
		effectstest.DcMana(t, g, who.ID, mana)
	}
	params.Strict = true
	return g.ActivateCatalogAbility(who.ID, id, 0, params)
}

// --- Sabertooth Wyvern ----------------------------------------------------------------

func TestExoB2SabertoothWyvernHasFlyingAndFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	wyvern := effectstest.ExoB1Push(g, me.ID, "Sabertooth Wyvern", "Creature — Drake", exoB2WyvernOracle, 3, 2, "R")
	for _, kw := range []string{"flying", "first strike"} {
		if !effectstest.HasEffectiveKeyword(t, g, wyvern, kw) {
			t.Errorf("Sabertooth Wyvern does not have %s", kw)
		}
	}
	bear := effectstest.ExoB1Push(g, them.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2, "G")
	effectstest.AttackIntoBlocks(t, g, them.ID, wyvern)
	if effectstest.BlockRefusal(t, g, bear, wyvern) == "" {
		t.Error("a creature without flying or reach was allowed to block the Wyvern")
	}
}

// --- Peace of Mind ----------------------------------------------------------------------

func TestExoB2PeaceOfMindGainsThreeLifeForADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	peace := effectstest.PushCatalogPermanent(g, me.ID, "Peace of Mind", "Enchantment", exoB2PeaceOfMindOracle, false)
	if err := exoB2Activate(t, g, me, peace, "{W}", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("Peace of Mind was activated with an empty hand")
	}
	fodder := exoB1HandCard(me, "Fodder")
	life := me.Life
	if err := exoB2Activate(t, g, me, peace, "", game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(fodder) {
		t.Error("the discarded card is not in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 3 {
		t.Errorf("gained %d life, want 3", got)
	}
	other := exoB1HandCard(me, "Other")
	if err := exoB2Activate(t, g, me, peace, "", game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{other}}); err == nil {
		t.Error("Peace of Mind was activated without {W}")
	}
}

// --- Fugue ------------------------------------------------------------------------------

func TestExoB2FugueMakesTheTargetDiscardThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	opp.Hand.Cards = nil
	var hand []uuid.UUID
	for _, n := range []string{"A", "B", "C", "D"} {
		hand = append(hand, exoB1HandCard(opp, n))
	}
	effectstest.CastCatalogSpell(t, g, "Fugue", "Sorcery", exoB2FugueOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DiscardChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("Fugue did not ask its target to discard")
	}
	if c.ChooseMin != 3 || c.ChooseMax != 3 {
		t.Errorf("prompt bounds = [%d,%d], want [3,3]", c.ChooseMin, c.ChooseMax)
	}
	effectstest.AnswerDiscard(t, g, opp.ID, hand[0], hand[1], hand[2])
	if opp.Hand.Size() != 1 || !opp.Hand.Contains(hand[3]) {
		t.Errorf("hand has %d cards, want only the one not discarded", opp.Hand.Size())
	}
	for _, id := range hand[:3] {
		if !opp.Graveyard.Contains(id) {
			t.Errorf("discarded card %s is not in the graveyard", id)
		}
	}
}

// --- Scare Tactics ----------------------------------------------------------------------

func TestExoB2ScareTacticsGivesYourCreaturesPlusOneZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Scare Tactics", "Instant", exoB2ScareTacticsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 3 || tough != 2 {
		t.Errorf("your creature is %d/%d, want 3/2", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
	effectstest.DcCleanup(g)
	if p, _ := effectstest.EffectivePTOf(t, g, mine); p != 2 {
		t.Errorf("the bonus outlived the cleanup step: power %d", p)
	}
}

// --- Shield Mate ------------------------------------------------------------------------

func TestExoB2ShieldMateSacrificesForPlusZeroFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.SeatsForCombat(g)
	mate := effectstest.PushCatalogPermanent(g, me.ID, "Shield Mate", "Creature — Human Soldier", exoB2ShieldMateOracle, true)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := exoB2Activate(t, g, me, mate, "", game.ActivateAbilityParams{Targets: effectstest.CardTarget(theirs)}); err != nil {
		t.Fatalf("activate (summoning sick, no tap needed): %v", err)
	}
	if g.Battlefield.Contains(mate) || !me.Graveyard.Contains(mate) {
		t.Error("Shield Mate was not sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 2 || tough != 6 {
		t.Errorf("the target is %d/%d, want 2/6", p, tough)
	}
}

// --- Rootwater Alligator -----------------------------------------------------------------

func TestExoB2RootwaterAlligatorRegeneratesForAForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gator := effectstest.B12Push(g, me.ID, "Rootwater Alligator", "Creature — Crocodile", exoB2AlligatorOracle, 3, 2)
	plains := effectstest.SeedLand(g, me.ID, "Plains", "Basic Land — Plains", "")
	if err := exoB2Activate(t, g, me, gator, "", game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{plains}}); err == nil {
		t.Fatal("the Alligator sacrificed a Plains")
	}
	forest := effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	if err := exoB2Activate(t, g, me, gator, "", game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{forest}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(forest) {
		t.Error("the Forest was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.UsgB3Destroy(t, g, gator)
	if !g.Battlefield.Contains(gator) {
		t.Fatal("the regenerated Alligator was destroyed")
	}
	if !effectstest.B13Tapped(t, g, gator) {
		t.Error("a regenerated creature is tapped")
	}
}

// --- Treasure Hunter ---------------------------------------------------------------------

func TestExoB2TreasureHunterMayReturnAnArtifactCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.PushGraveyardPermanent(me, "Mind Stone", "Artifact", "{2}")
	effectstest.PushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}{G}")
	effectstest.CastCatalogSpell(t, g, "Treasure Hunter", "Creature — Human", exoB2TreasureHunterOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil || !effectstest.HasID(p.PickTargetCards, rock) || len(p.PickTargetCards) != 1 {
		t.Fatalf("the artifact card, and only it, is offered: %+v", p)
	}
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) {
		t.Error("the artifact card did not return to hand")
	}
}

// --- Ravenous Baboons --------------------------------------------------------------------

func TestExoB2RavenousBaboonsDestroysANonbasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	coffers := effectstest.SeedLandOnBattlefield(g, opp.ID, "Cabal Coffers", "Land")
	basic := effectstest.SeedLandOnBattlefield(g, opp.ID, "Swamp", "Basic Land — Swamp")
	effectstest.CastCatalogSpell(t, g, "Ravenous Baboons", "Creature — Monkey", exoB2BaboonsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, basic) {
		t.Error("a basic land is offered")
	}
	effectstest.PickCard(t, g, me.ID, coffers)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(coffers) || !opp.Graveyard.Contains(coffers) {
		t.Error("the nonbasic land was not destroyed")
	}
	if !g.Battlefield.Contains(basic) {
		t.Error("the basic land left the battlefield")
	}
}

// --- Wayward Soul ------------------------------------------------------------------------

func TestExoB2WaywardSoulPutsItselfOnTopOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	soul := effectstest.B12Push(g, me.ID, "Wayward Soul", "Creature — Spirit", exoB2WaywardSoulOracle, 3, 2)
	if !effectstest.HasEffectiveKeyword(t, g, soul, "flying") {
		t.Error("Wayward Soul does not have flying")
	}
	if err := exoB2Activate(t, g, me, soul, "{U}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(soul) {
		t.Fatal("Wayward Soul is still on the battlefield")
	}
	top, err := me.Library.Top()
	if err != nil || top.InstanceID != soul {
		t.Errorf("the top of its owner's library is %q (%v), want Wayward Soul", top.Name, err)
	}
}

// --- School of Piranha -------------------------------------------------------------------

func TestExoB2SchoolOfPiranhaUpkeep(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		fish := effectstest.B12Push(g, me.ID, "School of Piranha", "Creature — Fish", exoB2PiranhaOracle, 3, 3)
		effectstest.WalkToUpkeepOfSeat0(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if pay {
			effectstest.DcMana(t, g, me.ID, "{U}{U}")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		if g.Battlefield.Contains(fish) != pay {
			t.Errorf("pay=%v: on the battlefield = %v", pay, g.Battlefield.Contains(fish))
		}
	}
}

// --- becomes blocked by a creature ---------------------------------------------------------

func TestExoB2RabidWolverinesAndPygmyTrollGrowOncePerBlocker(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		power, toughness       int
	}{
		{"Rabid Wolverines", exoB2WolverinesOracle, "Creature — Wolverine", 4, 4},
		{"Pygmy Troll", exoB2PygmyTrollOracle, "Creature — Troll", 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			a := effectstest.B12Creature(g, opp.ID, "Blocker A", "Creature — Bear", 1, 1)
			b := effectstest.B12Creature(g, opp.ID, "Blocker B", "Creature — Bear", 1, 1)
			effectstest.WoeS2Block(g, opp.ID, a, id)
			effectstest.WoeS2Block(g, opp.ID, b, id)
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.EffectivePTOf(t, g, id); p != tc.power+2 || tough != tc.toughness+2 {
				t.Errorf("blocked by two creatures it is %d/%d, want %d/%d", p, tough, tc.power+2, tc.toughness+2)
			}
		})
		t.Run(tc.name+" blocking is not becoming blocked", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := effectstest.B12Push(g, opp.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
			effectstest.WoeS2Block(g, opp.ID, id, attacker)
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.EffectivePTOf(t, g, id); p != tc.power || tough != tc.toughness {
				t.Errorf("blocking it is %d/%d, want %d/%d", p, tough, tc.power, tc.toughness)
			}
		})
	}
}

func TestExoB2PygmyTrollRegeneratesForG(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	troll := effectstest.B12Push(g, me.ID, "Pygmy Troll", "Creature — Troll", exoB2PygmyTrollOracle, 1, 1)
	if err := exoB2Activate(t, g, me, troll, "{G}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.UsgB3Destroy(t, g, troll)
	if !g.Battlefield.Contains(troll) {
		t.Fatal("the regenerated Pygmy Troll was destroyed")
	}
}

// --- Plated Rootwalla ---------------------------------------------------------------------

func TestExoB2PlatedRootwallaPumpsOnlyOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walla := effectstest.B12Push(g, me.ID, "Plated Rootwalla", "Creature — Lizard", exoB2RootwallaOracle, 3, 3)
	if err := exoB2Activate(t, g, me, walla, "{G}{G}{G}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, walla); p != 6 || tough != 6 {
		t.Errorf("the Rootwalla is %d/%d, want 6/6", p, tough)
	}
	if err := exoB2Activate(t, g, me, walla, "{G}{G}{G}", game.ActivateAbilityParams{}); err == nil {
		t.Error("the Rootwalla was activated a second time in one turn")
	}
	effectstest.DcCleanup(g)
	if p, tough := effectstest.EffectivePTOf(t, g, walla); p != 3 || tough != 3 {
		t.Errorf("after cleanup the Rootwalla is %d/%d, want 3/3", p, tough)
	}
}

// --- Thalakos Drifters --------------------------------------------------------------------

func TestExoB2ThalakosDriftersGainsShadowForADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	drifters := effectstest.B12Push(g, me.ID, "Thalakos Drifters", "Creature — Thalakos", exoB2DriftersOracle, 3, 3)
	if effectstest.HasEffectiveKeyword(t, g, drifters, "shadow") {
		t.Fatal("the Drifters have shadow before the activation")
	}
	if err := exoB2Activate(t, g, me, drifters, "", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the ability was activated with an empty hand")
	}
	fodder := exoB1HandCard(me, "Fodder")
	if err := exoB2Activate(t, g, me, drifters, "", game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, drifters, "shadow") {
		t.Fatal("the Drifters did not gain shadow")
	}
	effectstest.DcCleanup(g)
	if effectstest.HasEffectiveKeyword(t, g, drifters, "shadow") {
		t.Error("the Drifters kept shadow past the cleanup step")
	}
}

// --- Vampire Hounds -----------------------------------------------------------------------

func TestExoB2VampireHoundsGrowForACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	hounds := effectstest.B12Push(g, me.ID, "Vampire Hounds", "Creature — Vampire Dog", exoB2HoundsOracle, 2, 2)
	sorcery := exoB1HandCard(me, "Not a Creature")
	if err := exoB2Activate(t, g, me, hounds, "", game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{sorcery}}); err == nil {
		t.Fatal("the Hounds discarded a noncreature card")
	}
	creature := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: creature, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	if err := exoB2Activate(t, g, me, hounds, "", game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{creature}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, hounds); p != 4 || tough != 4 {
		t.Errorf("the Hounds are %d/%d, want 4/4", p, tough)
	}
}

// --- Welkin Hawk --------------------------------------------------------------------------

func TestExoB2WelkinHawkFetchesAnotherHawkWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hawk := effectstest.B12Push(g, me.ID, "Welkin Hawk", "Creature — Bird", exoB2WelkinHawkOracle, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, hawk, "flying") {
		t.Error("Welkin Hawk has no flying")
	}
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Welkin Owl", TypeLine: "Creature — Bird"},
		game.Card{Name: "Welkin Hawk", TypeLine: "Creature — Bird", OracleID: exoB2WelkinHawkOracle},
	)
	effectstest.AdvanceToMain(t, g)
	effectstest.UsgB3Destroy(t, g, hawk)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search after agreeing")
	}
	if effectstest.SearchOptionNamed(g, c, "Welkin Owl") != uuid.Nil {
		t.Error("a card not named Welkin Hawk was offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Welkin Hawk")
	found := false
	for _, card := range me.Hand.Cards {
		if card.Name == "Welkin Hawk" {
			found = true
		}
	}
	if !found {
		t.Error("the second Welkin Hawk is not in hand")
	}
}
