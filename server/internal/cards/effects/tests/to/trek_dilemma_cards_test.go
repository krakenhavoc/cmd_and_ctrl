package teststo

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_dilemma_cards_test.go — Star Trek's "face a dilemma" cards (#4973).
// Each test plays the real catalog entry through the engine: a dilemma is
// a real modal cast (Jeska's Will, modes announced with the cast).

const (
	trekDilSiskoOracle     = "f997ad8a-6903-47b1-83af-837b3b55a3d4"
	trekDilCaretakerOracle = "f77ea393-4142-4c06-8bb3-589ca74a7338"
	trekDilMelkotianOracle = "1fdcdb1c-a140-40ee-b711-5ea539eb1084"
	trekDilGarakOracle     = "c83465bd-f95f-4c57-b9ce-925a85b45276"
	trekDilLorcaOracle     = "6349eecb-dabd-4dcb-a8fb-c8ee66f61461"
	trekDilOperativeOracle = "7145e1b7-b991-4651-b21d-0620cb017090"
	trekDilTellMeOracle    = "abfdfae5-61aa-42d6-9c8e-83fda59de233"
	trekDilSevenOracle     = "3c037865-d38c-48da-aee9-308975e12638"
)

func TestTrekDilemmaEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekDilSiskoOracle:     "Captain Benjamin Sisko",
		trekDilCaretakerOracle: "The Caretaker",
		trekDilMelkotianOracle: "Melkotian Adjudicator",
		trekDilGarakOracle:     "Elim Garak, Simple Tailor",
		trekDilLorcaOracle:     "Gabriel Lorca, Mirror Menace",
		trekDilOperativeOracle: "Obsidian Order Operative",
		trekDilTellMeOracle:    "Tell Me What You See",
		trekDilSevenOracle:     "Seven of Nine",
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

// trekDilFace makes the active player face one dilemma: a modal spell
// cast with one mode chosen, its triggers left on the stack.
func trekDilFace(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.StockLibraryFor(g, g.Seats[g.Turn.ActiveSeat], 6)
	effectstest.CastCatalogSpellWithModes(t, g, "Jeska's Will", "Sorcery", effectstest.JeskasWillOracle, []int{1})
}

// --- the per-turn count ------------------------------------------------------

func TestTrekDilemmaCountIsPerPlayerAndResetsEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	trekDilFace(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	trekDilFace(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastWithCost(t, g, "Plain Sorcery", "Sorcery", "{1}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := g.DilemmasFacedThisTurnFor(me.ID); n != 2 {
		t.Fatalf("dilemmas faced %d, want 2 (a plain spell is none)", n)
	}
	if n := g.DilemmasFacedThisTurnFor(opp.ID); n != 0 {
		t.Errorf("the opponent faced %d dilemmas, want 0", n)
	}
	effectstest.AdvanceToEndStepOf(t, g, (seat+1)%4)
	if n := g.DilemmasFacedThisTurnFor(me.ID); n != 0 {
		t.Errorf("the count survived into the next turn: %d", n)
	}
}

// --- Melkotian Adjudicator ---------------------------------------------------

func TestTrekDilemmaMelkotianGrowsOnYourDilemmaOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	mine := effectstest.B27Push(g, me.ID, "Melkotian Adjudicator", "Creature — Melkotian Advisor", trekDilMelkotianOracle, "{2}{U}", 2, 1, "U")
	theirs := effectstest.B27Push(g, opp.ID, "Melkotian Adjudicator", "Creature — Melkotian Advisor", trekDilMelkotianOracle, "{2}{U}", 2, 1, "U")
	if !effectstest.HasEffectiveKeyword(t, g, mine, "flying") || !effectstest.HasEffectiveKeyword(t, g, mine, "vigilance") {
		t.Error("Melkotian Adjudicator has flying and vigilance")
	}
	trekDilFace(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("your Adjudicator has %d counters, want 1", got)
	}
	if got := effectstest.CounterOn(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's Adjudicator grew off your dilemma: %d", got)
	}
}

// --- Elim Garak and Obsidian Order Operative ---------------------------------

func TestTrekDilemmaGarakAndOperativeDrain(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		n                      int
	}{
		{"Elim Garak, Simple Tailor", "Legendary Creature — Cardassian Rogue", trekDilGarakOracle, 2},
		{"Obsidian Order Operative", "Creature — Cardassian Officer Spy", trekDilOperativeOracle, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.B27Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, "{3}{B}", 3, 3, "B")
			before := map[uuid.UUID]int{}
			for _, p := range g.Seats {
				before[p.ID] = effectstest.LifeOf(g, p.ID)
			}
			trekDilFace(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			for _, p := range g.Seats {
				want := before[p.ID] - tc.n
				if p.ID == me.ID {
					want = before[p.ID] + tc.n
				}
				if got := effectstest.LifeOf(g, p.ID); got != want {
					t.Errorf("seat %s life %d, want %d", p.ID, got, want)
				}
			}
		})
	}
}

func TestTrekDilemmaGarakHasMenaceAndIgnoresPlainSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	garak := effectstest.B27Push(g, me.ID, "Elim Garak, Simple Tailor", "Legendary Creature — Cardassian Rogue", trekDilGarakOracle, "{2}{B}", 3, 3, "B")
	if !effectstest.HasEffectiveKeyword(t, g, garak, "menace") {
		t.Error("Elim Garak has menace")
	}
	effectstest.CastWithCost(t, g, "Plain Sorcery", "Sorcery", "{1}", "")
	if n := effectstest.TriggersOnStackFrom(g, garak); n != 0 {
		t.Errorf("%d Garak triggers off a non-modal spell", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Seven of Nine -----------------------------------------------------------

func TestTrekDilemmaSevenOfNineDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seven := effectstest.B27Push(g, me.ID, "Seven of Nine", "Legendary Artifact Creature — Human Scientist", trekDilSevenOracle, "{2}{U}{B}", 2, 4, "U", "B")
	trekDilFace(t, g)
	if n := effectstest.TriggersOnStackFrom(g, seven); n != 1 {
		t.Fatalf("%d Seven of Nine triggers, want 1", n)
	}
	hand := effectstest.HandSize(me)
	// Resolve Seven's trigger alone (Jeska's Will is under it).
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3 && effectstest.TriggersOnStackFrom(g, seven) > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if effectstest.HandSize(me) != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, effectstest.HandSize(me))
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Gabriel Lorca, Mirror Menace --------------------------------------------

func TestTrekDilemmaLorcaPaysToReturnFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lorca := effectstest.SeedGraveyardCard(t, g, "Gabriel Lorca, Mirror Menace", "Legendary Creature — Human Rogue Officer", trekDilLorcaOracle)
	trekDilFace(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "B")
	if p := effectstest.AnswerPayUnless(t, g, me.ID, true); p.PayCost != "{B}" {
		t.Errorf("payment %q, want {B}", p.PayCost)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(lorca) || me.Graveyard.Contains(lorca) {
		t.Errorf("paying {B} did not return Lorca to hand (hand=%t graveyard=%t)", me.Hand.Contains(lorca), me.Graveyard.Contains(lorca))
	}
}

func TestTrekDilemmaLorcaDeclinedOrUnpaidStays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lorca := effectstest.SeedGraveyardCard(t, g, "Gabriel Lorca, Mirror Menace", "Legendary Creature — Human Rogue Officer", trekDilLorcaOracle)
	trekDilFace(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(lorca) {
		t.Fatal("declining moved Lorca")
	}
	// No {B} to spend: a yes pays nothing and returns nothing.
	trekDilFace(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		effectstest.AnswerPayUnless(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !me.Graveyard.Contains(lorca) {
		t.Error("Lorca came back without {B} being paid")
	}
}

func TestTrekDilemmaLorcaDoesNothingOnTheBattlefieldOrForAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	onField := effectstest.B27Push(g, me.ID, "Gabriel Lorca, Mirror Menace", "Legendary Creature — Human Rogue Officer", trekDilLorcaOracle, "{4}{B}", 5, 4, "B")
	if !effectstest.HasEffectiveKeyword(t, g, onField, "menace") {
		t.Error("Lorca has menace")
	}
	theirs := uuid.New()
	g.WithWriteLock(func() {
		opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Gabriel Lorca, Mirror Menace",
			TypeLine: "Legendary Creature — Human Rogue Officer", OracleID: trekDilLorcaOracle, Owner: opp.ID, Controller: opp.ID})
	})
	trekDilFace(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) || effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Error("Lorca asked to return: on the battlefield, or for an opponent's dilemma")
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("the opponent's Lorca moved")
	}
}

// --- Tell Me What You See ----------------------------------------------------

// trekDilTellMeEnters casts the enchantment and points its enters trigger
// at `opp`, leaving the trigger resolved up to the opponent's question.
func trekDilTellMeEnters(t *testing.T, g *game.Game, opp uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.CastCatalogSpell(t, g, "Tell Me What You See", "Enchantment", trekDilTellMeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestTrekDilemmaTellMeOpponentLetsYouDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	effectstest.StockLibraryFor(g, me, 5)
	life := effectstest.LifeOf(g, opp.ID)
	trekDilTellMeEnters(t, g, opp.ID)
	hand := effectstest.HandSize(me)
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandSize(me) != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, effectstest.HandSize(me))
	}
	if effectstest.LifeOf(g, opp.ID) != life || effectstest.SurveilChoiceFor(g, me.ID) != nil {
		t.Error("a yes is only the draw")
	}
}

func TestTrekDilemmaTellMeOpponentRefusesLosesFourAndYouSurveil(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	effectstest.StockLibraryFor(g, me, 5)
	life := effectstest.LifeOf(g, opp.ID)
	trekDilTellMeEnters(t, g, opp.ID)
	hand := effectstest.HandSize(me)
	effectstest.AnswerMayChoice(t, g, opp.ID, false)
	if got := effectstest.LifeOf(g, opp.ID); got != life-4 {
		t.Errorf("opponent life %d, want %d", got, life-4)
	}
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Error("a no makes you surveil 1")
	}
	if effectstest.HandSize(me) != hand {
		t.Error("a no draws nothing")
	}
}

func TestTrekDilemmaTellMeMayReturnToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tell := effectstest.B27Push(g, me.ID, "Tell Me What You See", "Enchantment", trekDilTellMeOracle, "{1}{B}", 0, 0, "B")
	trekDilFace(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, tell); !ok {
		t.Fatal("declining returned it anyway")
	}
	trekDilFace(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, tell); ok || !me.Hand.Contains(tell) {
		t.Errorf("yes: still on the battlefield=%t, in hand=%t", ok, me.Hand.Contains(tell))
	}
}

// --- The Caretaker -----------------------------------------------------------

func trekDilCaretaker(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B27Push(g, owner, "The Caretaker", "Legendary Creature — Elemental", trekDilCaretakerOracle, "{5}{U}", 5, 6, "U")
}

func TestTrekDilemmaCaretakerCastsTheOpponentsNonlandFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	ct := trekDilCaretaker(g, me.ID)
	if !effectstest.HasEffectiveKeyword(t, g, ct, "vigilance") {
		t.Error("The Caretaker has vigilance")
	}
	opp.Library.Cards = nil
	land := effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	spell := effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Seven Drops", TypeLine: "Artifact", ManaCost: "{7}"})
	under := effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Under", TypeLine: "Artifact", ManaCost: "{1}"})

	trekDilFace(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStep(t, g)
	if n := effectstest.TriggersOnStackFrom(g, ct) + len(g.PendingChoices); n == 0 {
		t.Fatal("no Caretaker trigger at your end step after a dilemma")
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, land) || !effectstest.InExile(g, spell) || effectstest.InExile(g, under) {
		t.Fatalf("exiled until the first nonland: land=%t spell=%t under=%t", effectstest.InExile(g, land), effectstest.InExile(g, spell), effectstest.InExile(g, under))
	}
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, spell)
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the exiled card for free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	got, ok := effectstest.BattlefieldCard(g, spell)
	if !ok || got.Controller != me.ID {
		t.Errorf("the opponent's card should resolve under your control (on battlefield=%t)", ok)
	}
}

// "Each end step": a dilemma you face on an opponent's turn (a modal
// instant) triggers The Caretaker at that opponent's end step.
func TestTrekDilemmaCaretakerTriggersAtAnOpponentsEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	theirSeat := (seat + 3) % 4 // the seat before yours: you are next in priority
	trekDilCaretaker(g, me.ID)
	for i := 0; !(g.Turn.ActiveSeat == theirSeat && g.Turn.Step == game.StepPrecombatMain); i++ {
		if i > 200 {
			t.Fatal("never reached the opponent's main phase")
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	charm := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: charm, Name: "Rakdos Charm", TypeLine: "Instant",
			OracleID: effectstest.RakdosCharmOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, charm, game.CastSpellParams{Modes: []int{2}}); err != nil {
		t.Fatalf("cast Rakdos Charm on the opponent's turn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := g.DilemmasFacedThisTurnFor(me.ID); n != 1 {
		t.Fatalf("dilemmas faced %d, want 1", n)
	}
	effectstest.AdvanceToEndStepOf(t, g, theirSeat)
	effectstest.B04WaitForPick(t, g, me.ID)
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("The Caretaker did not trigger at the opponent's end step")
	}
}

func TestTrekDilemmaCaretakerHasWardTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ct := trekDilCaretaker(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.CastAtWardedCreature(t, g, g.Seats[1], ct)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, g.Seats[1].ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePayUnless, g.Seats[1].ID)
	if p == nil || p.PayCost != "{2}" {
		t.Errorf("an opponent targeting The Caretaker is asked to pay ward {2}: %+v", p)
	}
}

func TestTrekDilemmaCaretakerNeedsADilemmaThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	ct := trekDilCaretaker(g, me.ID)
	effectstest.AdvanceToEndStep(t, g)
	if n := effectstest.TriggersOnStackFrom(g, ct); n != 0 || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Errorf("The Caretaker triggered with no dilemma faced this turn (%d on the stack)", n)
	}
}

// --- Captain Benjamin Sisko --------------------------------------------------

func trekDilSisko(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B27Push(g, owner, "Captain Benjamin Sisko", "Legendary Creature — Human Officer", trekDilSiskoOracle, "{1}{W}", 2, 2, "W")
}

func TestTrekDilemmaSiskoMakesAnOfficerPerDilemma(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	trekDilSisko(g, me.ID)
	trekDilFace(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	trekDilFace(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	officers := trekB7TokensNamed(t, g, me.ID, "Officer")
	if len(officers) != 2 {
		t.Fatalf("%d Officer tokens, want 2", len(officers))
	}
	o := officers[0]
	if o.CurrentPower() != 1 || o.CurrentToughness() != 1 || !o.HasColor("R") || !o.HasSubtype("Officer") {
		t.Errorf("token %+v is not a 1/1 red Officer", o)
	}
}

func TestTrekDilemmaSiskoMakesNoOfficerWithoutADilemma(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	trekDilSisko(g, me.ID)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(trekB7TokensNamed(t, g, me.ID, "Officer")); n != 0 {
		t.Errorf("%d Officer tokens with no dilemma", n)
	}
}

func TestTrekDilemmaSiskoCastsSpellsUpToTheOfficerCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sisko := trekDilSisko(g, me.ID)
	// Sisko is an Officer himself; one more makes two.
	effectstest.B27Push(g, me.ID, "Ensign", "Creature — Human Officer", "", "{W}", 1, 1, "W")
	effectstest.B27Push(g, me.ID, "Not an Officer", "Creature — Human", "", "{W}", 1, 1, "W")
	me.Library.Cards = nil
	two := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Two", TypeLine: "Artifact", ManaCost: "{2}"})
	three := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Three", TypeLine: "Artifact", ManaCost: "{3}"})
	one := effectstest.PushLibraryCardForTest(me, game.Card{Name: "One", TypeLine: "Artifact", ManaCost: "{1}"})
	forest := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	zero := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Zero", TypeLine: "Artifact", ManaCost: "{0}"})
	sixth := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Sixth", TypeLine: "Artifact", ManaCost: "{1}"})

	effectstest.AdvanceToMain(t, g)
	if err := effectstest.ActivateOn(t, g, me.ID, sisko, 0, uuid.Nil); err == nil {
		t.Fatal("activated without {2}{W}{U}{B}{R}{G}")
	}
	effectstest.B06AddMana(me, "C", "C", "W", "U", "B", "R", "G")
	if err := effectstest.ActivateOn(t, g, me.ID, sisko, 0, uuid.Nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{two, three, one, forest, zero} {
		if !effectstest.InExile(g, id) {
			t.Fatalf("%v was among the top five and is not in exile", id)
		}
	}
	if effectstest.InExile(g, sixth) {
		t.Fatal("a sixth card was exiled")
	}
	cast := func(id uuid.UUID) error {
		return g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile"})
	}
	if err := cast(three); !errors.Is(err, game.ErrSpellManaValueTooHigh) {
		t.Fatalf("cast mana value 3 with two Officers: err = %v, want ErrSpellManaValueTooHigh", err)
	}
	for _, id := range []uuid.UUID{two, one, zero} {
		if err := cast(id); err != nil {
			t.Fatalf("cast a spell under the cap: %v", err)
		}
	}
	if err := cast(forest); err == nil {
		t.Error("a land was cast")
	}
	if !effectstest.LayeredCard(t, g, sisko).Tapped {
		t.Error("Sisko was not tapped for the ability")
	}
}
