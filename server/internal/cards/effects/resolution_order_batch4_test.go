package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_order_batch4_test.go — #4155 (CR 608.2c): the fourth batch of
// hand-written bodies written as steps. Batches 1 to 3 covered a spell's
// ResolveSteps, a row's Do, ModeSteps, and (with #4168) continuation steps and
// the step record. These add what batch 4 introduced, each with a real card:
//
//   - a spell's ResolveSteps on a CR 616 counter order, the later sentences
//     waiting (Fleeting Flight);
//   - a spell's ResolveSteps whose middle step opens a prompt of its own, so the
//     last waits behind that (Stern Lesson);
//   - a mode's ResolveSteps guarded by its mode, on a dredge offer (Abzan Charm);
//   - a mode's ModeSteps carrying a count from the discard to the draw, on a
//     replacement prompt a discard asks (Collective Defiance, Library of Leng);
//   - a continuation run for one of several players answering side by side,
//     parked on a discard's prompt and resumed for the right player (Ruin
//     Grinder, ThenSteps.RunAt);
//   - a continuation that asks the next player from its last step, on a dredge
//     offer (Snort).
//
// A dredge offer and a CR 616 order resume through closures, so a table that
// holds one is not a restore point; an undo (a clone) stands in for the restore
// there, as in resolution_order_batch1_test.go.

const (
	abzanCharmOracle     = "4137a22c-f793-4e27-a9b2-72740a6e2121"
	fleetingFlightOracle = "6336401f-4e2d-4ebe-8c5b-24aa6f516abf"
	snortOracle          = "ab94985c-5a3e-42c7-be10-29c5b638346b"
)

// Fleeting Flight — "Put a +1/+1 counter on target creature. It gains flying
// until end of turn. Prevent all combat damage that would be dealt to it this
// turn." With Doubling Season and Hardened Scales out the counter waits on the
// CR 616 order, and the keyword waits behind it.
func TestFleetingFlightKeywordWaitsForTheCounterOrder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hs := seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me.ID)
	ds := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	castCatalogSpell(t, g, "Fleeting Flight", "Instant", fleetingFlightOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("want the one CR 616 prompt open, got %d choices", len(g.PendingChoices))
	}
	if hasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Fatal("the creature has flying with the order open: the keyword was granted before the counter")
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		answerOrderBySource(t, table, ds, hs)
		if got := b12Counter(t, table, bear, "+1/+1"); got != 3 {
			t.Fatalf("+1/+1 counters %d, want 3 ((1×2)+1, Doubling Season first)", got)
		}
		if !hasKeywordOnBattlefield(t, table, bear, "flying") {
			t.Fatal("the creature did not gain flying after the order")
		}
	}
}

// Stern Lesson — "Draw two cards, then discard a card. Create a tapped
// Powerstone token." The discard waits for the dredge offer the first draw
// makes, and the Powerstone waits for the discard.
func TestSternLessonDiscardAndTokenWaitInOrder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	awbPadLibrary(me)
	gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	powerstones := func(table *game.Game) int {
		n := 0
		for _, c := range table.Battlefield.Cards {
			if c.Name == "Powerstone" {
				n++
			}
		}
		return n
	}
	castCatalogSpell(t, g, "Stern Lesson", "Instant", sternLessonOracle, nil)
	passPriorityAroundTable(t, g)
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if latestChooseCardsFor(g, me.ID) != nil {
		t.Fatal("the discard prompt is open with dredge open: it ran before the answer")
	}
	if n := powerstones(g); n != 0 {
		t.Fatalf("%d Powerstones with dredge open, want 0", n)
	}
	for optionalOffer(g, me) != nil {
		answerOptional(t, g, me, false)
	}
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the discard prompt did not open after the draws")
	}
	if n := powerstones(g); n != 0 {
		t.Fatalf("%d Powerstones with the discard open, want 0: the token was made before the discard", n)
	}
	answerChooseCards(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	g.SettleResolution()
	if n := powerstones(g); n != 1 {
		t.Fatalf("%d Powerstones after the discard, want 1", n)
	}
}

// Abzan Charm's second mode — "You draw two cards and you lose 2 life." The
// OnResolve is ResolveSteps with each step guarded by its mode: the life loss
// waits for the dredge offer the first draw makes.
func TestAbzanCharmLifeLossWaitsForDredge(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	awbPadLibrary(me)
	imp := gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	life := me.Life
	castModal(t, g, "Abzan Charm", "Instant", abzanCharmOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if me.Life != life {
		t.Fatalf("life %d with dredge open, want %d: the loss ran before the answer", me.Life, life)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		answerOptional(t, table, p, true)
		if !p.Hand.Contains(imp) || p.Life != life-2 {
			t.Fatalf("Imp in hand %v, life %d; want true, %d", p.Hand.Contains(imp), p.Life, life-2)
		}
	}
}

// Collective Defiance's first mode — "Target player discards all the cards in
// their hand, then draws that many cards." The victim has an opponent's
// Library of Leng out, so each discard offers to put the card on top of the
// library instead. The draw waits for those answers and draws the number of
// cards the discard filed in the step record.
func TestCollectiveDefianceDrawWaitsForTheDiscardsPromptAndKeepsTheCount(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushTokenReplacementCard(g, libraryOfLengOracle, "Library of Leng", "Artifact", victim.ID)
	g.WithWriteLock(func() { victim.Hand.Cards = nil })
	awbPadLibrary(victim)
	var pitched []uuid.UUID
	for i := 0; i < 2; i++ {
		pitched = append(pitched, handCardFull(victim, "Pitch", "Sorcery", "{1}", "", nil))
	}
	lib := victim.Library.Size()
	caster.Hand.Cards = nil
	if _, err := castEscalate(t, g, "Collective Defiance", "Sorcery", collectiveDefianceOracle, "",
		game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{escPlayerRef(victim.ID, 0)}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if optionalOffer(g, victim) == nil {
		t.Fatal("Library of Leng's offer is not open")
	}
	if victim.Library.Size() != lib {
		t.Fatalf("library %d with the discard's offer open, want %d: the draw ran before the answer", victim.Library.Size(), lib)
	}
	lengSettleBatch(t, g, victim, pitched, map[uuid.UUID]bool{})
	g.SettleResolution()
	if victim.Hand.Size() != 2 || victim.Library.Size() != lib-2 {
		t.Fatalf("hand %d, library %d after the answers; want 2 cards drawn (hand 2, library %d)",
			victim.Hand.Size(), victim.Library.Size(), lib-2)
	}
}

// Ruin Grinder — "When this dies, each player may discard their hand and draw
// seven cards." Every player's question is open at once, and each answer runs
// the continuation for that player (ThenSteps.RunAt): each ends with seven cards.
func TestRuinGrinderEachPlayerWheelsTheirOwnHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i, p := range g.Seats {
		g.WithWriteLock(func() { p.Hand.Cards = nil })
		awbPadLibrary(p)
		for k := 0; k <= i; k++ {
			handCardFull(p, "Pitch", "Sorcery", "{1}", "", nil)
		}
	}
	grinder := pushCatalogPermanent(g, me.ID, "Ruin Grinder", "Creature — Zombie", deck3655GrinderOracle, false)
	killCreature(t, g, me.ID, grinder)
	passPriorityAroundTable(t, g)
	// Answer in reverse seat order, the last one asked first.
	for i := len(g.Seats) - 1; i >= 1; i-- {
		answerMayChoice(t, g, g.Seats[i].ID, true)
	}
	answerMayChoice(t, g, me.ID, false)
	for i, p := range g.Seats {
		want := 7
		if i == 0 {
			want = 1
		}
		if p.Hand.Size() != want {
			t.Fatalf("seat %d hand %d, want %d", i, p.Hand.Size(), want)
		}
	}
}

const runAtProbeOracle = "test-4155-run-at-probe"

// runAtProbeRan is the players the probe's last step ran for, in order.
var runAtProbeRan []uuid.UUID

// runAtProbeChosen is the number's keyed branch: it does nothing.
var runAtProbeChosen = NumberThen("test/run-at-probe-chosen",
	func(*Context, game.NumberAnswer) error { return nil })

// runAtProbeSteps is "choose a number, then note which player you are": the
// first step asks the player the occurrence names, the second runs after the
// answer and reads the occurrence again.
var runAtProbeSteps = RegisterThenSteps("test/run-at-probe",
	Instruction(func(ctx *Context) error {
		return ChooseNumber{
			Chooser: seatAtOccurrence(ctx), Min: 0, Max: 1, Question: "Run At Probe — how many?", Unit: game.PayAmountCards,
			Then: runAtProbeChosen,
		}.Apply(ctx)
	}),
	Instruction(func(ctx *Context) error {
		runAtProbeRan = append(runAtProbeRan, seatAtOccurrence(ctx))
		return nil
	}),
)

// runAtProbeGo is a player's yes, keyed (#4253 retired MayChoice's OnYes):
// the player rides the prompt as its Carry.
var runAtProbeGo = MayYes("test/4253-run-at-probe-go", func(c *Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	return runAtProbeSteps.RunAt(c, seatIndexOf(c.Game, r.Carry[0]))
})

func init() {
	Register(Spec{
		OracleID: runAtProbeOracle,
		Name:     "Run At Probe",
		Activated: []ActivatedAbility{{
			Label:   "{T}: Ask two other players, side by side.",
			Cost:    TapCost(),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: Do(Instruction(func(ctx *Context) error {
				for _, p := range ctx.Opponents()[:2] {
					player := p
					if err := (MayChoice{
						Player:   player,
						Question: "Run At Probe — go?",
						Then:     runAtProbeGo,
						Carry:    []uuid.UUID{player},
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})),
		}},
	})
}

// Two players' confirmations are open at once; each answer starts the
// continuation for that player, and its first step opens a number prompt that
// parks the rest. The players answer those in the opposite order, and each
// parked sequence resumes for the player who parked it (ctx.Occurrence()
// survives the park), not for the one that answered last.
func TestRunAtResumesEachParkedSequenceForItsOwnPlayer(t *testing.T) {
	g := newCatalogGame(t)
	runAtProbeRan = nil
	me := g.Seats[g.Turn.ActiveSeat]
	var opps []uuid.UUID
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps = append(opps, p.ID)
		}
	}
	first, second := opps[0], opps[1]
	probe := pushCatalogPermanent(g, me.ID, "Run At Probe", "Artifact", runAtProbeOracle, false)
	if err := activateOn(t, g, me.ID, probe, 0, uuid.Nil); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, first, true)
	answerMayChoice(t, g, second, true)
	if len(runAtProbeRan) != 0 {
		t.Fatalf("the last step ran for %v with the number prompts open", runAtProbeRan)
	}
	answerNumber(t, g, second, 0)
	answerNumber(t, g, first, 0)
	if len(runAtProbeRan) != 2 || runAtProbeRan[0] != second || runAtProbeRan[1] != first {
		t.Fatalf("the last step ran for %v, want [%s %s]: a parked sequence resumed for the wrong player",
			runAtProbeRan, second, first)
	}
}

// Snort — "Each player may discard their hand and draw five cards. Snort deals
// 5 damage to each opponent who discarded." The players are asked one after
// another: the first player's draw offers dredge, and the next player's question
// waits for the answer.
func TestSnortNextPlayerIsAskedAfterTheDredgeAnswer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	next := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	awbPadLibrary(me)
	gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	castCatalogSpell(t, g, "Snort", "Sorcery", snortOracle, nil)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered on the first draw")
	}
	if latestConfirmFor(g, next.ID) != nil {
		t.Fatal("the next player is asked with dredge open: the chain ran before the answer")
	}
	for optionalOffer(g, me) != nil {
		answerOptional(t, g, me, false)
	}
	if latestConfirmFor(g, next.ID) == nil {
		t.Fatal("the next player's question did not open after the draws")
	}
	if me.Hand.Size() != 5 {
		t.Fatalf("hand %d after the wheel, want 5", me.Hand.Size())
	}
}

// Starving Revenant — "When this enters, surveil 2. For each card put back on
// top, draw a card and lose 3 life." The chain draws, loses, and starts the
// next card from its last step: each loss waits for the dredge offer its draw
// makes, and the next draw waits for the loss.
func TestStarvingRevenantEachLossWaitsForItsDrawsDredgeOffer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	s58Library(me, "Deep", "Next", "Top")
	gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	life := me.Life
	s58CastRevenant(t, g, me)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered on the first draw")
	}
	if me.Life != life {
		t.Fatalf("life %d with the first offer open, want %d: the loss ran before the answer", me.Life, life)
	}
	answerOptional(t, g, me, false)
	g.SettleResolution()
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered on the second draw")
	}
	if me.Life != life-3 {
		t.Fatalf("life %d with the second offer open, want %d: one loss so far", me.Life, life-3)
	}
	answerOptional(t, g, me, false)
	g.SettleResolution()
	if me.Life != life-6 {
		t.Fatalf("life %d after both cards, want %d", me.Life, life-6)
	}
}
