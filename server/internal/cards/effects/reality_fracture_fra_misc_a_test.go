package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_fra_misc_a_test.go — the fra-misc-a slice: surveil
// cards and lands.

const (
	rfmaEnlightenedConfidant = "a139eb7f-3853-490c-a839-35d0e5aed7b5"
	rfmaEyeOfJace            = "c8f802ad-4621-4f71-b52d-fbc55853dc88"
	rfmaProctorOfPotential   = "c612ba0a-5134-45fa-a762-f4621bda4f83"
	rfmaSurveillancePhantasm = "b97648b7-ae98-45c9-8f0f-a52d6faf2d2d"
	rfmaYuriko               = "983f4fde-ab04-44f0-a294-c3b472de0b19"
	rfmaRoilingCanopy        = "c6e87760-4cd0-4281-82d4-7a377ea960ad"
	rfmaRoomOfRefuge         = "27d5b2eb-94fe-45fb-806c-354044de7ba5"
	rfmaTurbulentCrater      = "e86a0b64-fe4a-4ea8-99e8-35860dbac765"
	rfmaTurbulentShore       = "c13c5f07-dac5-47d0-a088-e24b9fbec3c5"
	rfmaTurbulentWetlands    = "b51c8659-ec4f-4213-beb1-49a7acd7366c"
)

func TestFraMiscACardsAreRegistered(t *testing.T) {
	for oracle, want := range map[string]struct {
		name         string
		completeness Completeness
	}{
		rfmaEnlightenedConfidant: {"Enlightened Confidant", CompletenessFull},
		rfmaEyeOfJace:            {"Eye of Jace", CompletenessFull},
		rfmaProctorOfPotential:   {"Proctor of Potential", CompletenessFull},
		rfmaSurveillancePhantasm: {"Surveillance Phantasm", CompletenessCaveats},
		rfmaYuriko:               {"Yuriko, Hope from the Shadows", CompletenessFull},
		rfmaRoilingCanopy:        {"Roiling Canopy", CompletenessCaveats},
		rfmaRoomOfRefuge:         {"Room of Refuge", CompletenessFull},
		rfmaTurbulentCrater:      {"Turbulent Crater", CompletenessFull},
		rfmaTurbulentShore:       {"Turbulent Shore", CompletenessFull},
		rfmaTurbulentWetlands:    {"Turbulent Wetlands", CompletenessFull},
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != want.name || spec.Completeness != want.completeness {
			t.Errorf("%s: registered=%v name=%q completeness=%v, want %q %v",
				oracle, ok, spec.Name, spec.Completeness, want.name, want.completeness)
		}
	}
}

// rfmaPush puts a catalog permanent on seat 0's side of the battlefield.
func rfmaPush(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tgh int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tgh, Owner: owner, Controller: owner,
	})
}

// rfmaAnswerSurveil answers the open surveil prompt, putting `bin` in the
// graveyard and leaving the rest of the looked-at cards on top.
func rfmaAnswerSurveil(t *testing.T, g *game.Game, p *game.Player, bin ...uuid.UUID) {
	t.Helper()
	ch := surveilChoiceFor(g, p.ID)
	if ch == nil {
		t.Fatal("no surveil prompt")
	}
	binned := map[uuid.UUID]bool{}
	for _, id := range bin {
		binned[id] = true
	}
	var top []uuid.UUID
	for _, id := range ch.ScryCards {
		if !binned[id] {
			top = append(top, id)
		}
	}
	if err := g.ResolveSurveil(ch.ID, p.ID, bin, top); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
}

// --- Turbulent Crater / Shore / Wetlands ---------------------------------

func TestFraMiscATurbulentDualsEnterUntappedOnlyAgainstEightOpposingLands(t *testing.T) {
	for _, row := range []struct {
		name, typeLine, oracle, a, b string
	}{
		{"Turbulent Crater", "Land — Swamp Mountain", rfmaTurbulentCrater, "B", "R"},
		{"Turbulent Shore", "Land — Plains Island", rfmaTurbulentShore, "W", "U"},
		{"Turbulent Wetlands", "Land — Island Swamp", rfmaTurbulentWetlands, "U", "B"},
	} {
		g := newCatalogGame(t)
		me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
		b36Lands(g, opp.ID, 4)
		b36Lands(g, other.ID, 3)
		b36Lands(g, me.ID, 6) // your own lands never count
		tapped := b12PlayFromHand(t, g, row.name, row.typeLine, row.oracle, game.CastSpellParams{})
		if !b16Tapped(t, g, tapped) {
			t.Fatalf("%s: seven opposing lands is not eight — it enters tapped", row.name)
		}
		b36Lands(g, other.ID, 1)
		advanceToMainOf(t, g, 0)
		untapped := b12PlayFromHand(t, g, row.name, row.typeLine, row.oracle, game.CastSpellParams{})
		if b16Tapped(t, g, untapped) {
			t.Fatalf("%s: eight opposing lands — it enters untapped", row.name)
		}
		b28TapForMana(t, g, me.ID, untapped, row.b)
		if got := poolColors(me); len(got) != 1 || got[0] != row.b {
			t.Errorf("%s: tapped for {%s}: pool %v", row.name, row.b, got)
		}
	}
}

// --- Eye of Jace --------------------------------------------------------

func TestEyeOfJaceStaysBelowSevenCardsInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	eye := rfmaPush(g, me.ID, "Eye of Jace", "Artifact", rfmaEyeOfJace, 0, 0)
	for i := 0; i < 5; i++ {
		batch01GraveyardCard(me, "Fodder", "Sorcery")
	}
	ids := seedLibrary(me, "Top", "Next")
	before := me.Life
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	rfmaAnswerSurveil(t, g, me, ids[0]) // 5 + 1 = 6 cards
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(eye) {
		t.Error("six cards in the graveyard: the Eye stays")
	}
	if me.Life != before {
		t.Errorf("life = %d, want %d", me.Life, before)
	}
}

func TestEyeOfJaceCountsTheCardItJustBinned(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eye := rfmaPush(g, me.ID, "Eye of Jace", "Artifact", rfmaEyeOfJace, 0, 0)
	for i := 0; i < 6; i++ {
		batch01GraveyardCard(me, "Fodder", "Sorcery")
	}
	ids := seedLibrary(me, "Top", "Next")
	lifeMe, lifeOpp := me.Life, opp.Life
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	rfmaAnswerSurveil(t, g, me, ids[0]) // the seventh card
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	eye := rfmaPush(g, me.ID, "Eye of Jace", "Artifact", rfmaEyeOfJace, 0, 0)
	for i := 0; i < 6; i++ {
		batch01GraveyardCard(me, "Fodder", "Sorcery")
	}
	seedLibrary(me, "Top", "Next")
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	rfmaAnswerSurveil(t, g, me) // keep it on top
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(eye) {
		t.Error("only six cards in the graveyard: the Eye stays")
	}
}

// --- Enlightened Confidant ---------------------------------------------

func rfmaConfidantEndStep(t *testing.T, g *game.Game, me *game.Player, gain int, topCost string) uuid.UUID {
	t.Helper()
	rfmaPush(g, me.ID, "Enlightened Confidant", "Creature — Kor Cleric", rfmaEnlightenedConfidant, 2, 1)
	top := seedLibraryTop(me, "Top Card", "Sorcery", topCost)
	advanceToMainOf(t, g, 0)
	if gain > 0 {
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, gain) })
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	return top
}

func TestEnlightenedConfidantReturnsACheapEnoughBinnedCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	top := rfmaConfidantEndStep(t, g, me, 3, "{1}{1}")
	rfmaAnswerSurveil(t, g, me, top)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) {
		t.Error("a mana value 2 card with 3 life gained goes to hand")
	}
}

func TestEnlightenedConfidantLeavesATooExpensiveCardInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	top := rfmaConfidantEndStep(t, g, me, 3, "{2}{2}")
	rfmaAnswerSurveil(t, g, me, top)
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(top) || !me.Graveyard.Contains(top) {
		t.Error("a mana value 4 card with 3 life gained stays in the graveyard")
	}
}

func TestEnlightenedConfidantKeepingTheCardReturnsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	top := rfmaConfidantEndStep(t, g, me, 3, "{1}")
	rfmaAnswerSurveil(t, g, me)
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(top) || me.Graveyard.Contains(top) {
		t.Error("a card kept on top is neither binned nor returned")
	}
}

func TestEnlightenedConfidantDoesNothingWithoutLifeGain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rfmaConfidantEndStep(t, g, me, 0, "{1}")
	if surveilChoiceFor(g, me.ID) != nil {
		t.Error("no life gained this turn: no surveil")
	}
}

// --- Proctor of Potential ----------------------------------------------

func TestProctorSurveilsWhenACreatureYouControlEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rfmaPush(g, me.ID, "Proctor of Potential", "Creature — Human Cleric", rfmaProctorOfPotential, 3, 1)
	seedLibrary(me, "Top", "Next")
	castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) == nil {
		t.Error("another creature entering should surveil 1")
	}
}

func TestProctorReturnsWithFinalityOnlyAfterScryingOrSurveilling(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToStepInTurn(t, g, game.StepPrecombatMain)
	id := pushCatalogGraveyardCard(me, "Proctor of Potential", "Creature — Human Cleric", rfmaProctorOfPotential, 3, 1)
	seedLibrary(me, "Top", "Next")
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
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Proctor did not return")
	}
	if got := allCountersOn(t, g, id)[rfCreatureBFinalityCounter]; got != 1 {
		t.Fatalf("finality counters = %d, want 1", got)
	}
	if surveilChoiceFor(g, me.ID) == nil {
		t.Error("the returning Proctor is a creature entering: it surveils")
	}
	destroy(t, g, id)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(id) {
		t.Error("a creature with a finality counter is exiled instead of dying")
	}
}

// --- Surveillance Phantasm ---------------------------------------------

func TestSurveillancePhantasmCanAttackOnceYouHaveSurveilled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ph := rfmaPush(g, me.ID, "Surveillance Phantasm", "Creature — Bird Illusion", rfmaSurveillancePhantasm, 2, 3)
	seedLibrary(me, "Top", "Next")
	advanceToStepInTurn(t, g, game.StepPrecombatMain)
	if !hasKeyword(effectiveAbilities(t, g, ph), "defender") {
		t.Fatal("before any scry or surveil it has defender")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{U}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, ph, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("surveil ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	rfmaAnswerSurveil(t, g, me)
	passPriorityAroundTable(t, g)
	abilities := effectiveAbilities(t, g, ph)
	if hasKeyword(abilities, "defender") {
		t.Error("after surveilling it can attack as though it had no defender")
	}
	if !hasKeyword(abilities, "flying") || !hasKeyword(abilities, "vigilance") {
		t.Errorf("it keeps flying and vigilance: %v", abilities)
	}
}

// --- Yuriko, Hope from the Shadows -------------------------------------

func rfmaYurikoEnters(t *testing.T, g *game.Game, me *game.Player) *game.PendingChoice {
	t.Helper()
	castAndResolveCreature(t, g, "Yuriko, Hope from the Shadows", "Legendary Creature — Human Ninja", rfmaYuriko)
	mode := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatalf("no mode prompt from the enters trigger: %+v", g.PendingChoices)
	}
	return mode
}

func TestYurikoShrinksPowerByYourGraveyardSize(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := rfmaPush(g, opp.ID, "Their Ogre", "Creature — Ogre", "", 5, 5)
	for i := 0; i < 3; i++ {
		batch01GraveyardCard(me, "Fodder", "Sorcery")
	}
	mode := rfmaYurikoEnters(t, g, me)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("choosing the first bullet: %v", err)
	}
	if err := sgAnswerTargets(t, g, theirs); err != nil {
		t.Fatalf("target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, th := rfPT(t, g, theirs); p != 2 || th != 5 {
		t.Errorf("P/T = %d/%d, want 2/5 (-3/-0)", p, th)
	}
}

func TestYurikoSurveilsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "One", "Two", "Three")
	mode := rfmaYurikoEnters(t, g, me)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("choosing surveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	ch := surveilChoiceFor(g, me.ID)
	if ch == nil || len(ch.ScryCards) != 2 {
		t.Fatalf("want a surveil prompt over 2 cards, got %+v", ch)
	}
}

// --- Roiling Canopy -----------------------------------------------------

func rfmaForests(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		b12Permanent(g, owner, "Forest", "Basic Land — Forest")
	}
}

func TestRoilingCanopyPumpsWhenAForestEntersWithFiveOthers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := rfmaPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	canopy := b12PlayFromHand(t, g, "Roiling Canopy", "Land", rfmaRoilingCanopy, game.CastSpellParams{})
	if !b16Tapped(t, g, canopy) {
		t.Error("the Canopy enters tapped")
	}
	rfmaForests(g, me.ID, 5)
	advanceToMainOf(t, g, 0)
	b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	if err := sgAnswerTargets(t, g, bear); err != nil {
		t.Fatalf("target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, th := rfPT(t, g, bear); p != 5 || th != 5 {
		t.Errorf("P/T = %d/%d, want 5/5", p, th)
	}
}

func TestRoilingCanopyNeedsFiveOtherForests(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rfmaPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	b12PlayFromHand(t, g, "Roiling Canopy", "Land", rfmaRoilingCanopy, game.CastSpellParams{})
	rfmaForests(g, me.ID, 4)
	advanceToMainOf(t, g, 0)
	b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("four other Forests is not five: no trigger")
	}
}

// --- Room of Refuge -----------------------------------------------------

func TestRoomOfRefugeEntersTappedAndTapsForTheChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	room := b12PlayFromHand(t, g, "Room of Refuge", "Land", rfmaRoomOfRefuge, game.CastSpellParams{})
	if !b16Tapped(t, g, room) {
		t.Error("Room of Refuge enters tapped")
	}
	answerColor(t, g, me.ID, "G")
	untap(g, room)
	if err := g.ActivateManaAbility(me.ID, room, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool = %v, want [G]", got)
	}
}

func TestRoomOfRefugeSacrificesForTwoCountersAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := rfmaPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	room := rfmaPush(g, me.ID, "Room of Refuge", "Land", rfmaRoomOfRefuge, 0, 0)
	advanceToStepInTurn(t, g, game.StepBeginCombat)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, room, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("outside a main phase the ability is refused")
	}
	advanceToStepInTurn(t, g, game.StepPostcombatMain)
	if err := g.ActivateCatalogAbility(me.ID, room, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if g.Battlefield.Contains(room) {
		t.Error("the Room is sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	if got := rfcbCounters(t, g, bear); got != 2 {
		t.Errorf("counters = %d, want 2", got)
	}
}
