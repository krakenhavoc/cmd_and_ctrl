package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestFraMiscACardsAreRegistered(t *testing.T) {
	for oracle, want := range map[string]struct {
		name         string
		completeness effects.Completeness
	}{
		effectstest.RfmaEnlightenedConfidant: {"Enlightened Confidant", effects.CompletenessFull},
		effectstest.RfmaEyeOfJace:            {"Eye of Jace", effects.CompletenessFull},
		effectstest.RfmaProctorOfPotential:   {"Proctor of Potential", effects.CompletenessFull},
		effectstest.RfmaSurveillancePhantasm: {"Surveillance Phantasm", effects.CompletenessFull},
		effectstest.RfmaYuriko:               {"Yuriko, Hope from the Shadows", effects.CompletenessFull},
		effectstest.RfmaRoilingCanopy:        {"Roiling Canopy", effects.CompletenessFull},
		effectstest.RfmaRoomOfRefuge:         {"Room of Refuge", effects.CompletenessFull},
		effectstest.RfmaTurbulentCrater:      {"Turbulent Crater", effects.CompletenessFull},
		effectstest.RfmaTurbulentShore:       {"Turbulent Shore", effects.CompletenessFull},
		effectstest.RfmaTurbulentWetlands:    {"Turbulent Wetlands", effects.CompletenessFull},
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != want.name || spec.Completeness != want.completeness {
			t.Errorf("%s: registered=%v name=%q completeness=%v, want %q %v",
				oracle, ok, spec.Name, spec.Completeness, want.name, want.completeness)
		}
	}
}

// --- Turbulent Crater / Shore / Wetlands ---------------------------------

func TestFraMiscATurbulentDualsEnterUntappedOnlyAgainstEightOpposingLands(t *testing.T) {
	for _, row := range []struct {
		name, typeLine, oracle, a, b string
	}{
		{"Turbulent Crater", "Land — Swamp Mountain", effectstest.RfmaTurbulentCrater, "B", "R"},
		{"Turbulent Shore", "Land — Plains Island", effectstest.RfmaTurbulentShore, "W", "U"},
		{"Turbulent Wetlands", "Land — Island Swamp", effectstest.RfmaTurbulentWetlands, "U", "B"},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
		effectstest.B36Lands(g, opp.ID, 4)
		effectstest.B36Lands(g, other.ID, 3)
		effectstest.B36Lands(g, me.ID, 6) // your own lands never count
		tapped := effectstest.B12PlayFromHand(t, g, row.name, row.typeLine, row.oracle, game.CastSpellParams{})
		if !effectstest.B16Tapped(t, g, tapped) {
			t.Fatalf("%s: seven opposing lands is not eight — it enters tapped", row.name)
		}
		effectstest.B36Lands(g, other.ID, 1)
		effectstest.AdvanceToMainOf(t, g, 0)
		untapped := effectstest.B12PlayFromHand(t, g, row.name, row.typeLine, row.oracle, game.CastSpellParams{})
		if effectstest.B16Tapped(t, g, untapped) {
			t.Fatalf("%s: eight opposing lands — it enters untapped", row.name)
		}
		effectstest.B28TapForMana(t, g, me.ID, untapped, row.b)
		if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != row.b {
			t.Errorf("%s: tapped for {%s}: pool %v", row.name, row.b, got)
		}
	}
}

// --- Eye of Jace --------------------------------------------------------

func TestEyeOfJaceStaysBelowSevenCardsInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	eye := effectstest.RfmaPush(g, me.ID, "Eye of Jace", "Artifact", effectstest.RfmaEyeOfJace, 0, 0)
	for i := 0; i < 5; i++ {
		effectstest.Batch01GraveyardCard(me, "Fodder", "Sorcery")
	}
	ids := effectstest.SeedLibrary(me, "Top", "Next")
	before := me.Life
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RfmaAnswerSurveil(t, g, me, ids[0]) // 5 + 1 = 6 cards
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(eye) {
		t.Error("six cards in the graveyard: the Eye stays")
	}
	if me.Life != before {
		t.Errorf("life = %d, want %d", me.Life, before)
	}
}

func TestEyeOfJaceCountsTheCardItJustBinned(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eye := effectstest.RfmaPush(g, me.ID, "Eye of Jace", "Artifact", effectstest.RfmaEyeOfJace, 0, 0)
	for i := 0; i < 6; i++ {
		effectstest.Batch01GraveyardCard(me, "Fodder", "Sorcery")
	}
	ids := effectstest.SeedLibrary(me, "Top", "Next")
	lifeMe, lifeOpp := me.Life, opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RfmaAnswerSurveil(t, g, me, ids[0]) // the seventh card
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(eye) {
		t.Error("the Eye should have been sacrificed")
	}
	if me.Life != lifeMe+2 {
		t.Errorf("life = %d, want %d", me.Life, lifeMe+2)
	}
	if opp.Life != lifeOpp-2 {
		t.Errorf("opponent life = %d, want %d", opp.Life, lifeOpp-2)
	}
	if !me.Graveyard.Contains(eye) {
		t.Error("the sacrificed Eye should be in the graveyard")
	}
}

func TestEyeOfJaceKeepingTheCardDoesNotCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	eye := effectstest.RfmaPush(g, me.ID, "Eye of Jace", "Artifact", effectstest.RfmaEyeOfJace, 0, 0)
	for i := 0; i < 6; i++ {
		effectstest.Batch01GraveyardCard(me, "Fodder", "Sorcery")
	}
	effectstest.SeedLibrary(me, "Top", "Next")
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RfmaAnswerSurveil(t, g, me) // keep it on top
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(eye) {
		t.Error("only six cards in the graveyard: the Eye stays")
	}
}

// --- Enlightened Confidant ---------------------------------------------

func rfmaConfidantEndStep(t *testing.T, g *game.Game, me *game.Player, gain int, topCost string) uuid.UUID {
	t.Helper()
	effectstest.RfmaPush(g, me.ID, "Enlightened Confidant", "Creature — Kor Cleric", effectstest.RfmaEnlightenedConfidant, 2, 1)
	top := effectstest.SeedLibraryTop(me, "Top Card", "Sorcery", topCost)
	effectstest.AdvanceToMainOf(t, g, 0)
	if gain > 0 {
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, gain) })
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	return top
}

func TestEnlightenedConfidantReturnsACheapEnoughBinnedCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	top := rfmaConfidantEndStep(t, g, me, 3, "{1}{1}")
	effectstest.RfmaAnswerSurveil(t, g, me, top)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) {
		t.Error("a mana value 2 card with 3 life gained goes to hand")
	}
}

func TestEnlightenedConfidantLeavesATooExpensiveCardInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	top := rfmaConfidantEndStep(t, g, me, 3, "{2}{2}")
	effectstest.RfmaAnswerSurveil(t, g, me, top)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(top) || !me.Graveyard.Contains(top) {
		t.Error("a mana value 4 card with 3 life gained stays in the graveyard")
	}
}

func TestEnlightenedConfidantKeepingTheCardReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	top := rfmaConfidantEndStep(t, g, me, 3, "{1}")
	effectstest.RfmaAnswerSurveil(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(top) || me.Graveyard.Contains(top) {
		t.Error("a card kept on top is neither binned nor returned")
	}
}

func TestEnlightenedConfidantDoesNothingWithoutLifeGain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rfmaConfidantEndStep(t, g, me, 0, "{1}")
	if effectstest.SurveilChoiceFor(g, me.ID) != nil {
		t.Error("no life gained this turn: no surveil")
	}
}

// --- Proctor of Potential ----------------------------------------------

func TestProctorSurveilsWhenACreatureYouControlEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.RfmaPush(g, me.ID, "Proctor of Potential", "Creature — Human Cleric", effectstest.RfmaProctorOfPotential, 3, 1)
	effectstest.SeedLibrary(me, "Top", "Next")
	effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Error("another creature entering should surveil 1")
	}
}

func TestProctorReturnsWithFinalityOnlyAfterScryingOrSurveilling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToStepInTurn(t, g, game.StepPrecombatMain)
	id := effectstest.PushCatalogGraveyardCard(me, "Proctor of Potential", "Creature — Human Cleric", effectstest.RfmaProctorOfPotential, 3, 1)
	effectstest.SeedLibrary(me, "Top", "Next")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{U}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activating before scrying or surveilling must be refused")
	}
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventSurveil, Actor: me.ID, LookedAt: 1})
	})
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("after surveilling: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Proctor did not return")
	}
	if got := effectstest.AllCountersOn(t, g, id)[effects.RfCreatureBFinalityCounter]; got != 1 {
		t.Fatalf("finality counters = %d, want 1", got)
	}
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Error("the returning Proctor is a creature entering: it surveils")
	}
	effectstest.Destroy(t, g, id)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(id) {
		t.Error("a creature with a finality counter is exiled instead of dying")
	}
}

// --- Surveillance Phantasm ---------------------------------------------

func TestSurveillancePhantasmCanAttackOnceYouHaveSurveilled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ph := effectstest.RfmaPush(g, me.ID, "Surveillance Phantasm", "Creature — Bird Illusion", effectstest.RfmaSurveillancePhantasm, 2, 3)
	effectstest.SeedLibrary(me, "Top", "Next")
	effectstest.AdvanceToStepInTurn(t, g, game.StepPrecombatMain)
	if !effectstest.HasKeyword(effectstest.EffectiveAbilities(t, g, ph), "defender") {
		t.Fatal("before any scry or surveil it has defender")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{U}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, ph, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("surveil ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RfmaAnswerSurveil(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	abilities := effectstest.EffectiveAbilities(t, g, ph)
	if effectstest.HasKeyword(abilities, "defender") {
		t.Error("after surveilling it can attack as though it had no defender")
	}
	if !effectstest.HasKeyword(abilities, "flying") || !effectstest.HasKeyword(abilities, "vigilance") {
		t.Errorf("it keeps flying and vigilance: %v", abilities)
	}
}

// --- Yuriko, Hope from the Shadows -------------------------------------

func rfmaYurikoEnters(t *testing.T, g *game.Game, me *game.Player) *game.PendingChoice {
	t.Helper()
	effectstest.CastAndResolveCreature(t, g, "Yuriko, Hope from the Shadows", "Legendary Creature — Human Ninja", effectstest.RfmaYuriko)
	mode := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatalf("no mode prompt from the enters trigger: %+v", g.PendingChoices)
	}
	return mode
}

func TestYurikoShrinksPowerByYourGraveyardSize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.RfmaPush(g, opp.ID, "Their Ogre", "Creature — Ogre", "", 5, 5)
	for i := 0; i < 3; i++ {
		effectstest.Batch01GraveyardCard(me, "Fodder", "Sorcery")
	}
	mode := rfmaYurikoEnters(t, g, me)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("choosing the first bullet: %v", err)
	}
	if err := effectstest.SgAnswerTargets(t, g, theirs); err != nil {
		t.Fatalf("target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.RfPT(t, g, theirs); p != 2 || th != 5 {
		t.Errorf("P/T = %d/%d, want 2/5 (-3/-0)", p, th)
	}
}

func TestYurikoSurveilsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	mode := rfmaYurikoEnters(t, g, me)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("choosing surveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.SurveilChoiceFor(g, me.ID)
	if ch == nil || len(ch.ScryCards) != 2 {
		t.Fatalf("want a surveil prompt over 2 cards, got %+v", ch)
	}
}

func TestRoilingCanopyPumpsWhenAForestEntersWithFiveOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.RfmaPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	canopy := effectstest.B12PlayFromHand(t, g, "Roiling Canopy", "Land", effectstest.RfmaRoilingCanopy, game.CastSpellParams{})
	if !effectstest.B16Tapped(t, g, canopy) {
		t.Error("the Canopy enters tapped")
	}
	effectstest.RfmaForests(g, me.ID, 5)
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.B12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	if err := effectstest.SgAnswerTargets(t, g, bear); err != nil {
		t.Fatalf("target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.RfPT(t, g, bear); p != 5 || th != 5 {
		t.Errorf("P/T = %d/%d, want 5/5", p, th)
	}
}

func TestRoilingCanopyNeedsFiveOtherForests(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.RfmaPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.B12PlayFromHand(t, g, "Roiling Canopy", "Land", effectstest.RfmaRoilingCanopy, game.CastSpellParams{})
	effectstest.RfmaForests(g, me.ID, 4)
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.B12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("four other Forests is not five: no trigger")
	}
}

// --- Room of Refuge -----------------------------------------------------

func TestRoomOfRefugeEntersTappedAndTapsForTheChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	room := effectstest.B12PlayFromHand(t, g, "Room of Refuge", "Land", effectstest.RfmaRoomOfRefuge, game.CastSpellParams{})
	if !effectstest.B16Tapped(t, g, room) {
		t.Error("Room of Refuge enters tapped")
	}
	effectstest.AnswerColor(t, g, me.ID, "G")
	effectstest.Untap(g, room)
	if err := g.ActivateManaAbility(me.ID, room, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool = %v, want [G]", got)
	}
}

func TestRoomOfRefugeSacrificesForTwoCountersAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.RfmaPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	room := effectstest.RfmaPush(g, me.ID, "Room of Refuge", "Land", effectstest.RfmaRoomOfRefuge, 0, 0)
	effectstest.AdvanceToStepInTurn(t, g, game.StepBeginCombat)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, room, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("outside a main phase the ability is refused")
	}
	effectstest.AdvanceToStepInTurn(t, g, game.StepPostcombatMain)
	if err := g.ActivateCatalogAbility(me.ID, room, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if g.Battlefield.Contains(room) {
		t.Error("the Room is sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := rfcbCounters(t, g, bear); got != 2 {
		t.Errorf("counters = %d, want 2", got)
	}
}
