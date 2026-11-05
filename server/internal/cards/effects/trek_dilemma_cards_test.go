package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

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
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// trekDilFace makes the active player face one dilemma: a modal spell
// cast with one mode chosen, its triggers left on the stack.
func trekDilFace(t *testing.T, g *game.Game) {
	t.Helper()
	stockLibraryFor(g, g.Seats[g.Turn.ActiveSeat], 6)
	castCatalogSpellWithModes(t, g, "Jeska's Will", "Sorcery", jeskasWillOracle, []int{1})
}

// --- the per-turn count ------------------------------------------------------

func TestTrekDilemmaCountIsPerPlayerAndResetsEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	trekDilFace(t, g)
	passPriorityAroundTable(t, g)
	trekDilFace(t, g)
	passPriorityAroundTable(t, g)
	castWithCost(t, g, "Plain Sorcery", "Sorcery", "{1}", "")
	passPriorityAroundTable(t, g)
	if n := g.DilemmasFacedThisTurnFor(me.ID); n != 2 {
		t.Fatalf("dilemmas faced %d, want 2 (a plain spell is none)", n)
	}
	if n := g.DilemmasFacedThisTurnFor(opp.ID); n != 0 {
		t.Errorf("the opponent faced %d dilemmas, want 0", n)
	}
	advanceToEndStepOf(t, g, (seat+1)%4)
	if n := g.DilemmasFacedThisTurnFor(me.ID); n != 0 {
		t.Errorf("the count survived into the next turn: %d", n)
	}
}

// --- Melkotian Adjudicator ---------------------------------------------------

func TestTrekDilemmaMelkotianGrowsOnYourDilemmaOnly(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	mine := b27Push(g, me.ID, "Melkotian Adjudicator", "Creature — Melkotian Advisor", trekDilMelkotianOracle, "{2}{U}", 2, 1, "U")
	theirs := b27Push(g, opp.ID, "Melkotian Adjudicator", "Creature — Melkotian Advisor", trekDilMelkotianOracle, "{2}{U}", 2, 1, "U")
	if !hasEffectiveKeyword(t, g, mine, "flying") || !hasEffectiveKeyword(t, g, mine, "vigilance") {
		t.Error("Melkotian Adjudicator has flying and vigilance")
	}
	trekDilFace(t, g)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("your Adjudicator has %d counters, want 1", got)
	}
	if got := counterOn(g, theirs, game.CounterPlusOne); got != 0 {
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
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			b27Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, "{3}{B}", 3, 3, "B")
			before := map[uuid.UUID]int{}
			for _, p := range g.Seats {
				before[p.ID] = lifeOf(g, p.ID)
			}
			trekDilFace(t, g)
			passPriorityAroundTable(t, g)
			for _, p := range g.Seats {
				want := before[p.ID] - tc.n
				if p.ID == me.ID {
					want = before[p.ID] + tc.n
				}
				if got := lifeOf(g, p.ID); got != want {
					t.Errorf("seat %s life %d, want %d", p.ID, got, want)
				}
			}
		})
	}
}

func TestTrekDilemmaGarakHasMenaceAndIgnoresPlainSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	garak := b27Push(g, me.ID, "Elim Garak, Simple Tailor", "Legendary Creature — Cardassian Rogue", trekDilGarakOracle, "{2}{B}", 3, 3, "B")
	if !hasEffectiveKeyword(t, g, garak, "menace") {
		t.Error("Elim Garak has menace")
	}
	castWithCost(t, g, "Plain Sorcery", "Sorcery", "{1}", "")
	if n := triggersOnStackFrom(g, garak); n != 0 {
		t.Errorf("%d Garak triggers off a non-modal spell", n)
	}
	passPriorityAroundTable(t, g)
}

// --- Seven of Nine -----------------------------------------------------------

func TestTrekDilemmaSevenOfNineDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seven := b27Push(g, me.ID, "Seven of Nine", "Legendary Artifact Creature — Human Scientist", trekDilSevenOracle, "{2}{U}{B}", 2, 4, "U", "B")
	trekDilFace(t, g)
	if n := triggersOnStackFrom(g, seven); n != 1 {
		t.Fatalf("%d Seven of Nine triggers, want 1", n)
	}
	hand := handSize(me)
	// Resolve Seven's trigger alone (Jeska's Will is under it).
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3 && triggersOnStackFrom(g, seven) > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if handSize(me) != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, handSize(me))
	}
	passPriorityAroundTable(t, g)
}

// --- Gabriel Lorca, Mirror Menace --------------------------------------------

func TestTrekDilemmaLorcaPaysToReturnFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lorca := seedGraveyardCard(t, g, "Gabriel Lorca, Mirror Menace", "Legendary Creature — Human Rogue Officer", trekDilLorcaOracle)
	trekDilFace(t, g)
	passPriorityAroundTable(t, g)
	b06AddMana(me, "B")
	if p := answerPayUnless(t, g, me.ID, true); p.PayCost != "{B}" {
		t.Errorf("payment %q, want {B}", p.PayCost)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(lorca) || me.Graveyard.Contains(lorca) {
		t.Errorf("paying {B} did not return Lorca to hand (hand=%t graveyard=%t)", me.Hand.Contains(lorca), me.Graveyard.Contains(lorca))
	}
}

func TestTrekDilemmaLorcaDeclinedOrUnpaidStays(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lorca := seedGraveyardCard(t, g, "Gabriel Lorca, Mirror Menace", "Legendary Creature — Human Rogue Officer", trekDilLorcaOracle)
	trekDilFace(t, g)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(lorca) {
		t.Fatal("declining moved Lorca")
	}
	// No {B} to spend: a yes pays nothing and returns nothing.
	trekDilFace(t, g)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		answerPayUnless(t, g, me.ID, true)
		passPriorityAroundTable(t, g)
	}
	if !me.Graveyard.Contains(lorca) {
		t.Error("Lorca came back without {B} being paid")
	}
}

func TestTrekDilemmaLorcaDoesNothingOnTheBattlefieldOrForAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	onField := b27Push(g, me.ID, "Gabriel Lorca, Mirror Menace", "Legendary Creature — Human Rogue Officer", trekDilLorcaOracle, "{4}{B}", 5, 4, "B")
	if !hasEffectiveKeyword(t, g, onField, "menace") {
		t.Error("Lorca has menace")
	}
	theirs := uuid.New()
	g.WithWriteLock(func() {
		opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Gabriel Lorca, Mirror Menace",
			TypeLine: "Legendary Creature — Human Rogue Officer", OracleID: trekDilLorcaOracle, Owner: opp.ID, Controller: opp.ID})
	})
	trekDilFace(t, g)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) || hasPayUnlessFor(g, opp.ID) {
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
	id := castCatalogSpell(t, g, "Tell Me What You See", "Enchantment", trekDilTellMeOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTarget(latestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func TestTrekDilemmaTellMeOpponentLetsYouDraw(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	stockLibraryFor(g, me, 5)
	life := lifeOf(g, opp.ID)
	trekDilTellMeEnters(t, g, opp.ID)
	hand := handSize(me)
	answerMayChoice(t, g, opp.ID, true)
	passPriorityAroundTable(t, g)
	if handSize(me) != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, handSize(me))
	}
	if lifeOf(g, opp.ID) != life || surveilChoiceFor(g, me.ID) != nil {
		t.Error("a yes is only the draw")
	}
}

func TestTrekDilemmaTellMeOpponentRefusesLosesFourAndYouSurveil(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	stockLibraryFor(g, me, 5)
	life := lifeOf(g, opp.ID)
	trekDilTellMeEnters(t, g, opp.ID)
	hand := handSize(me)
	answerMayChoice(t, g, opp.ID, false)
	if got := lifeOf(g, opp.ID); got != life-4 {
		t.Errorf("opponent life %d, want %d", got, life-4)
	}
	if surveilChoiceFor(g, me.ID) == nil {
		t.Error("a no makes you surveil 1")
	}
	if handSize(me) != hand {
		t.Error("a no draws nothing")
	}
}

func TestTrekDilemmaTellMeMayReturnToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tell := b27Push(g, me.ID, "Tell Me What You See", "Enchantment", trekDilTellMeOracle, "{1}{B}", 0, 0, "B")
	trekDilFace(t, g)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, tell); !ok {
		t.Fatal("declining returned it anyway")
	}
	trekDilFace(t, g)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, tell); ok || !me.Hand.Contains(tell) {
		t.Errorf("yes: still on the battlefield=%t, in hand=%t", ok, me.Hand.Contains(tell))
	}
}

// --- The Caretaker -----------------------------------------------------------

func trekDilCaretaker(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b27Push(g, owner, "The Caretaker", "Legendary Creature — Elemental", trekDilCaretakerOracle, "{5}{U}", 5, 6, "U")
}

func TestTrekDilemmaCaretakerCastsTheOpponentsNonlandFree(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	ct := trekDilCaretaker(g, me.ID)
	if !hasEffectiveKeyword(t, g, ct, "vigilance") {
		t.Error("The Caretaker has vigilance")
	}
	opp.Library.Cards = nil
	land := pushLibraryCardForTest(opp, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	spell := pushLibraryCardForTest(opp, game.Card{Name: "Seven Drops", TypeLine: "Artifact", ManaCost: "{7}"})
	under := pushLibraryCardForTest(opp, game.Card{Name: "Under", TypeLine: "Artifact", ManaCost: "{1}"})

	trekDilFace(t, g)
	passPriorityAroundTable(t, g)
	advanceToEndStep(t, g)
	if n := triggersOnStackFrom(g, ct) + len(g.PendingChoices); n == 0 {
		t.Fatal("no Caretaker trigger at your end step after a dilemma")
	}
	b04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTarget(latestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !inExile(g, land) || !inExile(g, spell) || inExile(g, under) {
		t.Fatalf("exiled until the first nonland: land=%t spell=%t under=%t", inExile(g, land), inExile(g, spell), inExile(g, under))
	}
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, spell)
	if err := castThroughResolutionCast(t, g, me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the exiled card for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	got, ok := battlefieldCard(g, spell)
	if !ok || got.Controller != me.ID {
		t.Errorf("the opponent's card should resolve under your control (on battlefield=%t)", ok)
	}
}

// "Each end step": a dilemma you face on an opponent's turn (a modal
// instant) triggers The Caretaker at that opponent's end step.
func TestTrekDilemmaCaretakerTriggersAtAnOpponentsEndStep(t *testing.T) {
	g := newCatalogGame(t)
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
			OracleID: rakdosCharmOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, charm, game.CastSpellParams{Modes: []int{2}}); err != nil {
		t.Fatalf("cast Rakdos Charm on the opponent's turn: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := g.DilemmasFacedThisTurnFor(me.ID); n != 1 {
		t.Fatalf("dilemmas faced %d, want 1", n)
	}
	advanceToEndStepOf(t, g, theirSeat)
	b04WaitForPick(t, g, me.ID)
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("The Caretaker did not trigger at the opponent's end step")
	}
}

func TestTrekDilemmaCaretakerHasWardTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ct := trekDilCaretaker(g, me.ID)
	advanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	castAtWardedCreature(t, g, g.Seats[1], ct)
	for i := 0; i < 8 && !hasPayUnlessFor(g, g.Seats[1].ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	p := latestChoiceOfKindFor(g, game.PendingChoicePayUnless, g.Seats[1].ID)
	if p == nil || p.PayCost != "{2}" {
		t.Errorf("an opponent targeting The Caretaker is asked to pay ward {2}: %+v", p)
	}
}

func TestTrekDilemmaCaretakerNeedsADilemmaThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	ct := trekDilCaretaker(g, me.ID)
	advanceToEndStep(t, g)
	if n := triggersOnStackFrom(g, ct); n != 0 || latestPickTarget(g, me.ID) != nil {
		t.Errorf("The Caretaker triggered with no dilemma faced this turn (%d on the stack)", n)
	}
}

// --- Captain Benjamin Sisko --------------------------------------------------

func trekDilSisko(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b27Push(g, owner, "Captain Benjamin Sisko", "Legendary Creature — Human Officer", trekDilSiskoOracle, "{1}{W}", 2, 2, "W")
}

func TestTrekDilemmaSiskoMakesAnOfficerPerDilemma(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	trekDilSisko(g, me.ID)
	trekDilFace(t, g)
	passPriorityAroundTable(t, g)
	trekDilFace(t, g)
	passPriorityAroundTable(t, g)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	trekDilSisko(g, me.ID)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if n := len(trekB7TokensNamed(t, g, me.ID, "Officer")); n != 0 {
		t.Errorf("%d Officer tokens with no dilemma", n)
	}
}

func TestTrekDilemmaSiskoCastsSpellsUpToTheOfficerCount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sisko := trekDilSisko(g, me.ID)
	// Sisko is an Officer himself; one more makes two.
	b27Push(g, me.ID, "Ensign", "Creature — Human Officer", "", "{W}", 1, 1, "W")
	b27Push(g, me.ID, "Not an Officer", "Creature — Human", "", "{W}", 1, 1, "W")
	me.Library.Cards = nil
	two := pushLibraryCardForTest(me, game.Card{Name: "Two", TypeLine: "Artifact", ManaCost: "{2}"})
	three := pushLibraryCardForTest(me, game.Card{Name: "Three", TypeLine: "Artifact", ManaCost: "{3}"})
	one := pushLibraryCardForTest(me, game.Card{Name: "One", TypeLine: "Artifact", ManaCost: "{1}"})
	forest := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	zero := pushLibraryCardForTest(me, game.Card{Name: "Zero", TypeLine: "Artifact", ManaCost: "{0}"})
	sixth := pushLibraryCardForTest(me, game.Card{Name: "Sixth", TypeLine: "Artifact", ManaCost: "{1}"})

	advanceToMain(t, g)
	if err := activateOn(t, g, me.ID, sisko, 0, uuid.Nil); err == nil {
		t.Fatal("activated without {2}{W}{U}{B}{R}{G}")
	}
	b06AddMana(me, "C", "C", "W", "U", "B", "R", "G")
	if err := activateOn(t, g, me.ID, sisko, 0, uuid.Nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{two, three, one, forest, zero} {
		if !inExile(g, id) {
			t.Fatalf("%v was among the top five and is not in exile", id)
		}
	}
	if inExile(g, sixth) {
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
	if !layeredCard(t, g, sisko).Tapped {
		t.Error("Sisko was not tapped for the ability")
	}
}
