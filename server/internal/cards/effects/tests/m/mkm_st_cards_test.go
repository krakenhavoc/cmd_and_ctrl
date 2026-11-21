package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mkm_st_cards_test.go — Murders at Karlov Manor, S–T (#4469). Each
// test plays the real catalog entry and asserts what its text says.

const (
	mkmSTSaviorOracle      = "cc3a6c83-a671-4ce8-b98e-434989fb6778"
	mkmSTAutomatonOracle   = "be2ab552-df82-49a0-8f0b-ee66de34e1a7"
	mkmSTInformantOracle   = "5cf3727b-eb7b-414e-982c-3acf4c05ddf1"
	mkmSTSliceOracle       = "4d31d3f4-4748-4713-808f-24699fcd7b10"
	mkmSTDualleechOracle   = "ccb1b532-92b8-4ebf-b3c2-fb5924b116f5"
	mkmSTGorehoundOracle   = "d2b1c5eb-06f0-4ec4-9446-791faac099a7"
	mkmSTEnervationOracle  = "5b4508a3-1d21-46ec-8a50-cd3171a7348a"
	mkmSTDetonationOracle  = "91f7672c-3090-48f4-8586-0426341c3159"
	mkmSTChaseOracle       = "789899eb-3a33-4a41-b6b5-5e2f55a9af51"
	mkmSTWentThisWayOracle = "47951720-9584-4ca3-a73d-3d259102b2f8"
	mkmSTThinkingCapOracle = "a54fb3de-0581-461e-be33-dc4f4d16a33e"
	mkmSTPantherOracle     = "efddf972-94c4-49d1-a844-b4aba70c93aa"
	mkmSTToxinOracle       = "1f3b0b15-2c38-4aaa-b527-afbb540a6715"
	mkmSTTrostaniOracle    = "5c90b951-826e-4595-aacc-958904f3f794"
	mkmSTTorchOracle       = "51473069-68dc-4e87-b345-d7cbcfdeea74"
)

func TestMkmSTEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mkmSTSaviorOracle:      "Sanguine Savior",
		mkmSTAutomatonOracle:   "Sanitation Automaton",
		mkmSTInformantOracle:   "Shady Informant",
		mkmSTSliceOracle:       "Slice from the Shadows",
		mkmSTDualleechOracle:   "Slimy Dualleech",
		mkmSTGorehoundOracle:   "Snarling Gorehound",
		mkmSTEnervationOracle:  "Soul Enervation",
		mkmSTDetonationOracle:  "Suspicious Detonation",
		mkmSTChaseOracle:       "The Chase Is On",
		mkmSTWentThisWayOracle: "They Went This Way",
		mkmSTThinkingCapOracle: "Thinking Cap",
		mkmSTPantherOracle:     "Topiary Panther",
		mkmSTToxinOracle:       "Toxin Analysis",
		mkmSTTrostaniOracle:    "Trostani, Three Whispers",
		mkmSTTorchOracle:       "Torch the Witness",
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

// TestMkmSTDisguiseAndUncounterableDeclarations pins the two disguise
// costs and the two "can't be countered" spells.
func TestMkmSTDisguiseAndUncounterableDeclarations(t *testing.T) {
	for _, tc := range []struct{ name, oracle, faceUp string }{
		{"Sanguine Savior", mkmSTSaviorOracle, "{W/B}{W/B}"},
		{"Shady Informant", mkmSTInformantOracle, "{2}{B/R}{B/R}"},
	} {
		ac := game.AlternativeCostByKey(tc.oracle, "disguise")
		if ac == nil || ac.FaceDown == nil {
			t.Errorf("%s offers no disguise cast", tc.name)
			continue
		}
		if ac.ManaCost != effects.MorphCastCost || ac.FaceDown.Kind != game.FaceDownDisguised || ac.FaceDown.FaceUpCost != tc.faceUp {
			t.Errorf("%s: cast %q, kind %q, face-up %q; want %q, disguised, %q",
				tc.name, ac.ManaCost, ac.FaceDown.Kind, ac.FaceDown.FaceUpCost, effects.MorphCastCost, tc.faceUp)
		}
	}
	for _, oracle := range []string{mkmSTSliceOracle, mkmSTDetonationOracle} {
		if spec, _ := effects.Lookup(oracle); !spec.CantBeCountered {
			t.Errorf("%s can be countered", spec.Name)
		}
	}
}

// mkmSTOpponent is the seat after the active one.
func mkmSTOpponent(g *game.Game) *game.Player {
	return g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// --- Sanguine Savior -----------------------------------------------------------

func TestMkmSTSanguineSaviorGivesAnotherCreatureLifelinkWhenTurnedUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	savior := effectstest.MkmB1CastFaceDown(t, g, "Sanguine Savior", "Creature — Vampire Cleric", mkmSTSaviorOracle, 2, 1)

	effectstest.MkmB1TurnFaceUp(t, g, me, savior, "W", "B")
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: savior}); err == nil {
		t.Fatal("the Savior itself was accepted; the target is another creature")
	}
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.HasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("the bear did not gain lifelink")
	}
	effectstest.AssertKeywords(t, g, savior, "flying", "lifelink")
}

// --- Sanitation Automaton ------------------------------------------------------

func TestMkmSTSanitationAutomatonSurveilsOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Top", "Second")
	effectstest.CastHoldCreature(t, g, "Sanitation Automaton", "Artifact Creature — Construct", mkmSTAutomatonOracle, 2, 1)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the entry trigger queued no surveil")
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("surveil looked at %d cards, want 1", len(c.ScryCards))
	}
}

// --- Shady Informant -----------------------------------------------------------

func TestMkmSTShadyInformantDeals2WhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := mkmSTOpponent(g)
	informant := effectstest.B12Push(g, me.ID, "Shady Informant", "Creature — Ogre Rogue", mkmSTInformantOracle, 4, 2)
	before := opp.Life

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(informant) })
	effectstest.B04WaitForPick(t, g, me.ID)
	mkmB1PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Life != before-2 {
		t.Errorf("opponent life %d -> %d, want -2", before, opp.Life)
	}
}

func TestMkmSTShadyInformantFaceDownDiesQuietly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := mkmSTOpponent(g)
	informant := effectstest.MkmB1CastFaceDown(t, g, "Shady Informant", "Creature — Ogre Rogue", mkmSTInformantOracle, 4, 2)
	before := opp.Life

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(informant) })
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a face-down Informant has no abilities (CR 708.2a), but its death asked for a target")
	}
	if opp.Life != before {
		t.Errorf("opponent life %d -> %d, want unchanged", before, opp.Life)
	}
}

// --- Slice from the Shadows ----------------------------------------------------

func TestMkmSTSliceFromTheShadowsGivesMinusXMinusX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := mkmSTOpponent(g)
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	effectstest.CastXSpell(t, g, "Slice from the Shadows", "Instant", mkmSTSliceOracle, "{X}{B}", 2,
		[]game.TargetRef{{Kind: game.TargetCard, ID: ogre}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.MkmB1PT(t, g, ogre); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tough)
	}
}

// --- Slimy Dualleech -----------------------------------------------------------

func TestMkmSTSlimyDualleechPumpsASmallCreatureAtYourCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Slimy Dualleech", "Creature — Leech", mkmSTDualleechOracle, 2, 4)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ogre := effectstest.B12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)

	ch := effectstest.PickTargetPrompt(t, g)
	for _, bad := range []uuid.UUID{ogre, theirs} {
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatal("a creature with power 3, or an opponent's, was accepted")
		}
	}
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if p, tough := effectstest.MkmB1PT(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the bear is %d/%d, want 3/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Error("the bear did not gain deathtouch")
	}
}

// --- Snarling Gorehound --------------------------------------------------------

func TestMkmSTSnarlingGorehoundSurveilsForASmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Top", "Second")
	hound := effectstest.B12Push(g, me.ID, "Snarling Gorehound", "Creature — Dog", mkmSTGorehoundOracle, 1, 1)
	effectstest.AssertKeywords(t, g, hound, "menace")

	effectstest.CastHoldCreature(t, g, "Pyrotechnic Performer", effectstest.MkmMPPerformerTypeLine, effectstest.MkmMPPerformerOracle, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) != nil {
		t.Fatal("a 3/2 entering made the Gorehound surveil")
	}

	effectstest.CastHoldCreature(t, g, "Neighborhood Guardian", "Creature — Unicorn", effectstest.MkmMPGuardianOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SurveilChoiceFor(g, me.ID); c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("a 2/2 entered and the Gorehound did not surveil 1: %+v", c)
	}
}

// --- Soul Enervation -----------------------------------------------------------

func TestMkmSTSoulEnervationShrinksATargetOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := mkmSTOpponent(g)
	wurm := effectstest.B12Creature(g, opp.ID, "Their Wurm", "Creature — Wurm", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Soul Enervation", "Enchantment", mkmSTEnervationOracle, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, wurm)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.MkmB1PT(t, g, wurm); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tough)
	}
}

func TestMkmSTSoulEnervationDrainsOnceForABatch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Soul Enervation", "Enchantment", mkmSTEnervationOracle, 0, 0)
	a := effectstest.MkmAC2PushGraveyard(g, me, "Dead A", "Creature — Zombie", "", "{1}")
	b := effectstest.MkmAC2PushGraveyard(g, me, "Dead B", "Creature — Zombie", "", "{1}")
	spell := effectstest.MkmAC2PushGraveyard(g, me, "Old Spell", "Sorcery", "", "{1}")
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}

	g.WithWriteLock(func() { _ = g.ExileCardsForEffect([]uuid.UUID{spell}) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before[me.ID] {
		t.Fatalf("a sorcery card leaving drained: life %d -> %d", before[me.ID], me.Life)
	}

	g.WithWriteLock(func() { _ = g.ExileCardsForEffect([]uuid.UUID{a, b}) })
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := before[p.ID] - 1
		if p.ID == me.ID {
			want = before[p.ID] + 1
		}
		if p.Life != want {
			t.Errorf("seat %s life %d -> %d, want %d", p.ID, before[p.ID], p.Life, want)
		}
	}
}

// --- Suspicious Detonation -----------------------------------------------------

func TestMkmSTSuspiciousDetonationCostsThreeLessAfterAnArtifactSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	card := game.Card{InstanceID: uuid.New(), Name: "Suspicious Detonation", TypeLine: "Sorcery",
		ManaCost: "{4}{R}", OracleID: mkmSTDetonationOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	price := func() int {
		t.Helper()
		p, err := g.PriceCast(me.ID, card, game.CastSpellParams{})
		if err != nil {
			t.Fatalf("PriceCast: %v", err)
		}
		return p.Total.ManaValue()
	}
	if got := price(); got != 5 {
		t.Fatalf("nothing sacrificed: %d, want 5", got)
	}
	mkmSacrifice(t, g, me.ID, effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	if got := price(); got != 5 {
		t.Fatalf("a creature sacrificed: %d, want 5", got)
	}
	mkmSacrifice(t, g, me.ID, effectstest.B12Creature(g, me.ID, "Rock", "Artifact", 0, 0))
	if got := price(); got != 2 {
		t.Fatalf("an artifact sacrificed: %d, want 2", got)
	}
}

func TestMkmSTSuspiciousDetonationDeals4ToTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := mkmSTOpponent(g)
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Suspicious Detonation", "Sorcery", mkmSTDetonationOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: ogre}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the 4/4 survived 4 damage")
	}
}

// --- The Chase Is On -----------------------------------------------------------

func TestMkmSTTheChaseIsOnPumpsGrantsFirstStrikeAndInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "The Chase Is On", "Instant", mkmSTChaseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	if p, tough := effectstest.MkmB1PT(t, g, bear); p != 5 || tough != 2 {
		t.Errorf("the target is %d/%d, want 5/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the target did not gain first strike")
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

// --- They Went This Way --------------------------------------------------------

func TestMkmSTTheyWentThisWayFetchesATappedBasicAndInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	basic := effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Forest", "Basic Land — Forest"))[0]
	effectstest.CastCatalogSpell(t, g, "They Went This Way", "Sorcery", mkmSTWentThisWayOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{basic}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(basic) {
		t.Fatal("the basic did not reach the battlefield")
	}
	if !effectstest.TmtB1Card(t, g, basic).Tapped {
		t.Error("the basic entered untapped")
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

// --- Thinking Cap --------------------------------------------------------------

func TestMkmSTThinkingCapEquipsADetectiveForOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	thinkingCap := effectstest.PushCatalogPermanent(g, me.ID, "Thinking Cap", "Artifact — Equipment", mkmSTThinkingCapOracle, false)
	detective := effectstest.B12Creature(g, me.ID, "My Detective", "Creature — Human Detective", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	effectstest.MkmAC2AddMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, thinkingCap, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("Equip Detective {1} attached to a Bear")
	}
	effectstest.B16Activate(t, g, me.ID, thinkingCap, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: detective}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.MkmB1PT(t, g, detective); p != 2 || tough != 3 {
		t.Errorf("the equipped Detective is %d/%d, want 2/3", p, tough)
	}

	effectstest.MkmAC2AddMana(t, g, me, "{C}{C}{C}")
	effectstest.B16Activate(t, g, me.ID, thinkingCap, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.MkmB1PT(t, g, bear); p != 3 || tough != 4 {
		t.Errorf("the equipped Bear is %d/%d, want 3/4", p, tough)
	}
	if p, tough := effectstest.MkmB1PT(t, g, detective); p != 1 || tough != 1 {
		t.Errorf("the Detective kept the bonus after the Cap moved: %d/%d", p, tough)
	}
}

// --- Topiary Panther -----------------------------------------------------------

func TestMkmSTTopiaryPantherBasicLandcycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if spec, _ := effects.Lookup(mkmSTPantherOracle); len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "trample" {
		t.Errorf("printed keywords %v, want trample", spec.PrintedKeywords)
	}
	me := g.Seats[g.Turn.ActiveSeat]
	basic := effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Forest", "Basic Land — Forest"))[0]
	id, _ := effectstest.CycleFromHand(t, g, "Topiary Panther", "Creature — Plant Cat", mkmSTPantherOracle, "{C}{G}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the discard-this cost did not put the Panther in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("landcycling did not search")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{basic}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(basic) {
		t.Error("the basic did not reach the hand")
	}
}

// --- Toxin Analysis ------------------------------------------------------------

func TestMkmSTToxinAnalysisGrantsDeathtouchAndLifelinkAndInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Toxin Analysis", "Instant", mkmSTToxinOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	for _, kw := range []string{"deathtouch", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the target did not gain %s", kw)
		}
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

// --- Trostani, Three Whispers --------------------------------------------------

func TestMkmSTTrostaniGrantsEachKeywordToAnyCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := mkmSTOpponent(g)
	effectstest.AdvanceToMain(t, g)
	trostani := effectstest.B12Push(g, me.ID, "Trostani, Three Whispers", "Legendary Creature — Dryad", mkmSTTrostaniOracle, 4, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	for i, tc := range []struct{ mana, keyword string }{
		{"{C}{G}", "deathtouch"},
		{"{G}", "vigilance"},
		{"{C}{C}{W}", "double strike"},
	} {
		effectstest.MkmAC2AddMana(t, g, me, tc.mana)
		effectstest.B16Activate(t, g, me.ID, trostani, i, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
		})
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasEffectiveKeyword(t, g, theirs, tc.keyword) {
			t.Errorf("row %d: the target did not gain %s", i, tc.keyword)
		}
	}
}

// --- Torch the Witness ---------------------------------------------------------

func TestMkmSTTorchTheWitnessInvestigatesOnlyOnExcess(t *testing.T) {
	for _, tc := range []struct {
		name         string
		toughness, x int
		survives     bool
		wantClues    int
	}{
		{"exact lethal", 4, 2, false, 0},
		{"excess", 3, 2, false, 1},
		{"not lethal", 5, 2, true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := mkmSTOpponent(g)
			victim := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, tc.toughness)
			effectstest.CastXSpell(t, g, "Torch the Witness", "Sorcery", mkmSTTorchOracle, "{X}{R}", tc.x,
				[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
			effectstest.PassPriorityAroundTable(t, g)
			if got := g.Battlefield.Contains(victim); got != tc.survives {
				t.Errorf("on the battlefield = %v, want %v", got, tc.survives)
			}
			if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != tc.wantClues {
				t.Errorf("%d Clues, want %d", got, tc.wantClues)
			}
		})
	}
}
