package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b80_cards_test.go — S82 GAP batch 80 (Ravnica Remastered
// instants, planeswalkers and sorceries, #5948). Every test plays the real
// catalog entry through the engine.

const (
	gapB80Ghostway  = "c8981af1-b0aa-4c19-8520-a7fb59018bc8"
	gapB80Judgment  = "ae198ce9-097b-4204-b599-12bdb36f5195"
	gapB80Bolas     = "581eedc4-efe4-4ae4-baf2-88071e7e6f6d"
	gapB80Domri     = "0c4a2c42-0874-40b4-8d12-47eeadc3e976"
	gapB80Ral       = "75ed6946-e7b1-4bf8-a05c-62a2ba5a6c79"
	gapB80Chill     = "dc0d4e58-db39-4a05-b27c-452180c86d47"
	gapB80Tutor     = "68a203a3-2283-4e29-b177-b3be43f76899"
	gapB80Revival   = "836ecd67-6487-4f3d-9e08-013746be829a"
	gapB80OpenGates = "f38e43b6-9a8c-4f6a-bd1c-d5c4913a28a1"
)

func TestGapB80EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		gapB80Ghostway:       "Ghostway",
		gapB80Judgment:       "Summary Judgment",
		gapB80Bolas:          "Nicol Bolas, Dragon-God",
		gapB80Domri:          "Domri Rade",
		gapB80Ral:            "Ral Zarek",
		gapB80Chill:          "Creeping Chill",
		gapB80Tutor:          "Infernal Tutor",
		gapB80Revival:        "Revival",
		gapB80Revival + "#1": "Revenge",
		gapB80OpenGates:      "Open the Gates",
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

func TestGapB80GhostwayExilesYourCreaturesAndReturnsThemAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "Mine A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Mine B", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Ghostway", "Instant", gapB80Ghostway, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, a) || effectstest.BattlefieldHas(g, b) {
		t.Fatal("my creatures are still on the battlefield")
	}
	if !effectstest.BattlefieldHas(g, theirs) {
		t.Fatal("their creature was exiled")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.BattlefieldIDsNamed(g, "Mine A")) + len(effectstest.BattlefieldIDsNamed(g, "Mine B")); n != 2 {
		t.Errorf("%d of my creatures came back, want 2", n)
	}
}

func TestGapB80SummaryJudgmentDealsFiveInYourMainPhaseAndThreeOtherwise(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	untapped := effectstest.B12Creature(g, opp.ID, "Untapped", "Creature — Bear", 1, 9)
	tappedOne := effectstest.B12Creature(g, opp.ID, "Tapped One", "Creature — Bear", 1, 9)
	tappedTwo := effectstest.B12Creature(g, opp.ID, "Tapped Two", "Creature — Bear", 1, 9)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, tappedOne).Tapped = true
		effectstest.FindBattlefieldCardForTest(g, tappedTwo).Tapped = true
	})
	effectstest.AdvanceToMain(t, g)
	if err := gapB79CastNowB80(g, "Summary Judgment", gapB80Judgment, effectstest.CardTarget(untapped), true); err == nil {
		t.Fatal("targeted an untapped creature")
	}
	if err := gapB79CastNowB80(g, "Summary Judgment", gapB80Judgment, effectstest.CardTarget(tappedOne), false); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, tappedOne).DamageMarked; got != 5 {
		t.Errorf("in my main phase it dealt %d, want 5", got)
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := gapB79CastNowB80(g, "Summary Judgment", gapB80Judgment, effectstest.CardTarget(tappedTwo), false); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, tappedTwo).DamageMarked; got != 3 {
		t.Errorf("outside a main phase it dealt %d, want 3", got)
	}
}

// gapB79CastNowB80 casts a card for the active seat in whatever step it is.
func gapB79CastNowB80(g *game.Game, name, oracle string, targets []game.TargetRef, strict bool) error {
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", OracleID: oracle, Owner: me.ID, Controller: me.ID})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets, Strict: strict})
}

func TestGapB80CreepingChillDrainsAsASpellAndFromTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life, oppLife := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Creeping Chill", "Sorcery", gapB80Chill, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 || opp.Life != oppLife-3 {
		t.Fatalf("life %d/%d, want %d/%d", me.Life, opp.Life, life+3, oppLife-3)
	}
	chill := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: chill, Name: "Creeping Chill", TypeLine: "Sorcery", OracleID: gapB80Chill, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if err := g.MillNForEffect(me.ID, 1); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+6 || opp.Life != oppLife-6 {
		t.Errorf("after the mill trigger life %d/%d, want %d/%d", me.Life, opp.Life, life+6, oppLife-6)
	}
	if me.Graveyard.Contains(chill) {
		t.Error("Creeping Chill stayed in the graveyard")
	}
}

func TestGapB80OpenTheGatesFindsABasicLandOrAGate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gate := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: gate, Name: "Test Gate", TypeLine: "Land — Gate", Owner: me.ID, Controller: me.ID})
	nonbasic := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: nonbasic, Name: "Wastes Cousin", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Open the Gates", "Sorcery", gapB80OpenGates, nil)
	effectstest.PassPriorityAroundTable(t, g)
	search := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID)
	if search == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{nonbasic}); err == nil {
		t.Fatal("a nonbasic non-Gate land was accepted")
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{gate}); err != nil {
		t.Fatalf("search: %v", err)
	}
	g.SettleResolution()
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(gate) {
		t.Error("the Gate is not in hand")
	}
}

func TestGapB80RevivalReturnsASmallCreatureAndRevengeHalvesLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: small, Name: "Small", TypeLine: "Creature — Bear", ManaCost: "{2}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	big := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: big, Name: "Big", TypeLine: "Creature — Giant", ManaCost: "{5}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	gapB80CastSplit(t, g, 0, effectstest.CardTarget(big), true)
	gapB80CastSplit(t, g, 0, effectstest.CardTarget(small), false)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, small) {
		t.Error("the small creature did not return")
	}
	me.Life, opp.Life = 7, 15
	gapB80CastSplit(t, g, 1, []game.TargetRef{effectstest.PlayerTarget(opp.ID)}, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 14 {
		t.Errorf("my life is %d, want 14", me.Life)
	}
	if opp.Life != 7 {
		t.Errorf("their life is %d, want 7 (lost half of 15, rounded up = 8)", opp.Life)
	}
}

func gapB80CastSplit(t *testing.T, g *game.Game, face int, targets []game.TargetRef, strict bool) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	card := game.Card{
		InstanceID: uuid.New(), OracleID: gapB80Revival, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Revival", TypeLine: "Sorcery", ManaCost: "{0}"},
			{Name: "Revenge", TypeLine: "Sorcery", ManaCost: "{0}"},
		},
	}
	card.SetFace(0)
	me.Hand.PushTop(card)
	err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: face, Targets: targets, Strict: strict})
	if strict && err == nil {
		t.Fatal("the cast was legal, want a refusal")
	}
	if !strict && err != nil {
		t.Fatalf("cast: %v", err)
	}
}

func TestGapB80RalZarekPlusOneTapsThenUntapsAndMinusTwoShocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	ral := effectstest.PushWalkerForTest(g, me.ID, "Ral Zarek", gapB80Ral, 4)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, b).Tapped = true })
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, ral, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(a, a)}); err == nil {
		t.Fatal("the same permanent was both targets")
	}
	if err := g.ActivateCatalogAbility(me.ID, ral, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a, Slot: 0}, {Kind: game.TargetCard, ID: b, Slot: 1}}}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.FindBattlefieldCardForTest(g, a).Tapped || effectstest.FindBattlefieldCardForTest(g, b).Tapped {
		t.Error("expected A tapped and B untapped")
	}
	if got := effectstest.LoyaltyOf(g, ral); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestGapB80RalZarekMinusTwoDealsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	ral := effectstest.PushWalkerForTest(g, me.ID, "Ral Zarek", gapB80Ral, 4)
	effectstest.AdvanceToMainOf(t, g, seat)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, ral, 1, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("-2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d", opp.Life, life-3)
	}
}

func TestGapB80RalZarekMinusSevenFlipsFiveCoins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	ral := effectstest.PushWalkerForTest(g, me.ID, "Ral Zarek", gapB80Ral, 7)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, ral, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("-7: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	flips := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventFlipCoin {
			flips++
		}
	}
	if flips != 5 {
		t.Errorf("%d coin flips, want 5", flips)
	}
}

func TestGapB80DomriRadePlusOneTakesOnlyACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	domri := effectstest.PushWalkerForTest(g, me.ID, "Domri Rade", gapB80Domri, 3)
	land := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, domri, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(land) {
		t.Error("a land was put into hand")
	}
	if cards := me.Library.Cards; len(cards) == 0 || cards[len(cards)-1].InstanceID != land {
		t.Error("the land is not still on top")
	}
}

func TestGapB80DomriRadeMinusTwoFightsAndMinusSevenGrantsKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	domri := effectstest.PushWalkerForTest(g, me.ID, "Domri Rade", gapB80Domri, 9)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, domri, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(mine, mine)}); err == nil {
		t.Fatal("the creature was both targets")
	}
	if err := g.ActivateCatalogAbility(me.ID, domri, 1, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1}}}); err != nil {
		t.Fatalf("-2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, theirs) {
		t.Error("their creature survived the fight")
	}
	if got := effectstest.FindCardAnywhere(t, g, mine).DamageMarked; got != 2 {
		t.Errorf("my creature took %d, want 2", got)
	}
}

func TestGapB80DomriRadeEmblemGrantsFourKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	domri := effectstest.PushWalkerForTest(g, me.ID, "Domri Rade", gapB80Domri, 7)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, domri, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("-7: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"double strike", "trample", "hexproof", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, mine, kw) {
			t.Errorf("my creature lacks %s", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, theirs, kw) {
			t.Errorf("their creature has %s", kw)
		}
	}
}

func TestGapB80InfernalTutorFindsTheRevealedCardsName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for me.Hand.Size() > 0 {
		_ = g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, me.Hand.Cards[0].InstanceID)
	}
	keep := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: keep, Name: "Twin", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	other := uuid.New()
	twin := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: other, Name: "Other", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: twin, Name: "Twin", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Infernal Tutor", "Sorcery", gapB80Tutor, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatalf("no reveal prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{keep}); err != nil {
		t.Fatalf("reveal: %v", err)
	}
	g.SettleResolution()
	search := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID)
	if search == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{other}); err == nil {
		t.Fatal("a card with another name was accepted")
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{twin}); err != nil {
		t.Fatalf("search: %v", err)
	}
	g.SettleResolution()
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(twin) {
		t.Error("the same-name card is not in hand")
	}
}

func TestGapB80InfernalTutorHellbentFindsAnyCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for me.Hand.Size() > 0 {
		_ = g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, me.Hand.Cards[0].InstanceID)
	}
	target := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: target, Name: "Anything", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Infernal Tutor", "Sorcery", gapB80Tutor, nil)
	effectstest.PassPriorityAroundTable(t, g)
	search := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID)
	if search == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{target}); err != nil {
		t.Fatalf("search: %v", err)
	}
	g.SettleResolution()
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(target) {
		t.Error("the card is not in hand")
	}
}

func TestGapB80NicolBolasDragonGod(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	bolas := effectstest.PushWalkerForTest(g, me.ID, "Nicol Bolas, Dragon-God", gapB80Bolas, 4)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, bolas, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(victim)}); err != nil {
		t.Fatalf("-3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, victim) {
		t.Error("the creature survived -3")
	}
}

func TestGapB80NicolBolasPlusOneDrawsAndEachOpponentExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	bolas := effectstest.PushWalkerForTest(g, me.ID, "Nicol Bolas, Dragon-God", gapB80Bolas, 4)
	effectstest.AdvanceToMainOf(t, g, seat)
	hand := effectstest.HandSize(me)
	handSizes := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		if s.ID != me.ID {
			handSizes[s.ID] = s.Hand.Size()
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, bolas, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandSize(me) != hand+1 {
		t.Errorf("hand %d, want %d", effectstest.HandSize(me), hand+1)
	}
	for _, s := range g.Seats {
		if s.ID == me.ID {
			continue
		}
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, s.ID)
		if pick == nil {
			t.Fatalf("%s has no exile prompt", s.Name)
		}
		if err := g.ResolveChooseCards(pick.ID, s.ID, []uuid.UUID{s.Hand.Cards[0].InstanceID}); err != nil {
			t.Fatalf("%s: %v", s.Name, err)
		}
		g.SettleResolution()
	}
	for _, s := range g.Seats {
		if s.ID != me.ID && s.Hand.Size() != handSizes[s.ID]-1 {
			t.Errorf("%s hand %d, want %d", s.Name, s.Hand.Size(), handSizes[s.ID]-1)
		}
	}
}

func TestGapB80NicolBolasMinusEightSparesLegendControllers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	bolas := effectstest.PushWalkerForTest(g, me.ID, "Nicol Bolas, Dragon-God", gapB80Bolas, 8)
	var spared *game.Player
	for _, s := range g.Seats {
		if s.ID != me.ID {
			spared = s
			break
		}
	}
	effectstest.B12Creature(g, spared.ID, "Legend", "Legendary Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, bolas, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("-8: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		if s.ID == me.ID || s.ID == spared.ID {
			if s.Eliminated {
				t.Errorf("%s lost", s.Name)
			}
			continue
		}
		if !s.Eliminated {
			t.Errorf("%s did not lose", s.Name)
		}
	}
}

func TestGapB80NicolBolasUsesAnotherPlaneswalkersLoyaltyAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	bolas := effectstest.PushWalkerForTest(g, me.ID, "Nicol Bolas, Dragon-God", gapB80Bolas, 4)
	effectstest.PushWalkerForTest(g, opp.ID, "Ral Zarek", gapB80Ral, 4)
	effectstest.AdvanceToMainOf(t, g, seat)
	g.BumpLayerVersionForTest()
	idx, origins := r35BlBorrowedActivated(t, g, bolas)
	borrowed := len(idx)
	if borrowed != 3 {
		t.Fatalf("%d borrowed rows, want Ral Zarek's three loyalty abilities", borrowed)
	}
	life := opp.Life
	// Ral Zarek's −2 (three damage to any target), paid with Bolas's loyalty.
	if err := g.ActivateCatalogAbility(me.ID, bolas, idx[1], game.ActivateAbilityParams{
		Ref: origins[1].Ref, Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("borrowed -2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d", opp.Life, life-3)
	}
	if got := effectstest.LoyaltyOf(g, bolas); got != 2 {
		t.Errorf("Bolas's loyalty is %d, want 2", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, bolas, idx[0], game.ActivateAbilityParams{Ref: origins[0].Ref, Strict: true}); err == nil {
		t.Error("a second loyalty ability in one turn was allowed")
	}
}
