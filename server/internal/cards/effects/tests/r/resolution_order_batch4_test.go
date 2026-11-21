package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hs := effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me.ID)
	ds := effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	effectstest.CastCatalogSpell(t, g, "Fleeting Flight", "Instant", fleetingFlightOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("want the one CR 616 prompt open, got %d choices", len(g.PendingChoices))
	}
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Fatal("the creature has flying with the order open: the keyword was granted before the counter")
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		effectstest.AnswerOrderBySource(t, table, ds, hs)
		if got := effectstest.B12Counter(t, table, bear, "+1/+1"); got != 3 {
			t.Fatalf("+1/+1 counters %d, want 3 ((1×2)+1, Doubling Season first)", got)
		}
		if !effectstest.HasKeywordOnBattlefield(t, table, bear, "flying") {
			t.Fatal("the creature did not gain flying after the order")
		}
	}
}

// Stern Lesson — "Draw two cards, then discard a card. Create a tapped
// Powerstone token." The discard waits for the dredge offer the first draw
// makes, and the Powerstone waits for the discard.
func TestSternLessonDiscardAndTokenWaitInOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AwbPadLibrary(me)
	effectstest.GyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	powerstones := func(table *game.Game) int {
		n := 0
		for _, c := range table.Battlefield.Cards {
			if c.Name == "Powerstone" {
				n++
			}
		}
		return n
	}
	effectstest.CastCatalogSpell(t, g, "Stern Lesson", "Instant", effectstest.SternLessonOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OptionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Fatal("the discard prompt is open with dredge open: it ran before the answer")
	}
	if n := powerstones(g); n != 0 {
		t.Fatalf("%d Powerstones with dredge open, want 0", n)
	}
	for effectstest.OptionalOffer(g, me) != nil {
		effectstest.AnswerOptional(t, g, me, false)
	}
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the discard prompt did not open after the draws")
	}
	if n := powerstones(g); n != 0 {
		t.Fatalf("%d Powerstones with the discard open, want 0: the token was made before the discard", n)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	g.SettleResolution()
	if n := powerstones(g); n != 1 {
		t.Fatalf("%d Powerstones after the discard, want 1", n)
	}
}

// Abzan Charm's second mode — "You draw two cards and you lose 2 life." The
// OnResolve is ResolveSteps with each step guarded by its mode: the life loss
// waits for the dredge offer the first draw makes.
func TestAbzanCharmLifeLossWaitsForDredge(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AwbPadLibrary(me)
	imp := effectstest.GyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	life := me.Life
	effectstest.CastModal(t, g, "Abzan Charm", "Instant", abzanCharmOracle, []int{1}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OptionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if me.Life != life {
		t.Fatalf("life %d with dredge open, want %d: the loss ran before the answer", me.Life, life)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		effectstest.AnswerOptional(t, table, p, true)
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
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushTokenReplacementCard(g, effectstest.LibraryOfLengOracle, "Library of Leng", "Artifact", victim.ID)
	g.WithWriteLock(func() { victim.Hand.Cards = nil })
	effectstest.AwbPadLibrary(victim)
	var pitched []uuid.UUID
	for i := 0; i < 2; i++ {
		pitched = append(pitched, effectstest.HandCardFull(victim, "Pitch", "Sorcery", "{1}", "", nil))
	}
	lib := victim.Library.Size()
	caster.Hand.Cards = nil
	if _, err := effectstest.CastEscalate(t, g, "Collective Defiance", "Sorcery", effectstest.CollectiveDefianceOracle, "",
		game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{effectstest.EscPlayerRef(victim.ID, 0)}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OptionalOffer(g, victim) == nil {
		t.Fatal("Library of Leng's offer is not open")
	}
	if victim.Library.Size() != lib {
		t.Fatalf("library %d with the discard's offer open, want %d: the draw ran before the answer", victim.Library.Size(), lib)
	}
	effectstest.LengSettleBatch(t, g, victim, pitched, map[uuid.UUID]bool{})
	g.SettleResolution()
	if victim.Hand.Size() != 2 || victim.Library.Size() != lib-2 {
		t.Fatalf("hand %d, library %d after the answers; want 2 cards drawn (hand 2, library %d)",
			victim.Hand.Size(), victim.Library.Size(), lib-2)
	}
}

// Ruin Grinder — "When this dies, each player may discard their hand and draw
// seven cards." Every player chooses in APNAP order before anyone discards or
// draws (CR 101.4); then each player who said yes ends with seven new cards.
func TestRuinGrinderEachPlayerWheelsTheirOwnHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i, p := range g.Seats {
		g.WithWriteLock(func() { p.Hand.Cards = nil })
		effectstest.AwbPadLibrary(p)
		for k := 0; k <= i; k++ {
			effectstest.HandCardFull(p, "Pitch", "Sorcery", "{1}", "", nil)
		}
	}
	grinder := effectstest.PushCatalogPermanent(g, me.ID, "Ruin Grinder", "Creature — Zombie", effectstest.Deck3655GrinderOracle, false)
	effectstest.KillCreature(t, g, me.ID, grinder)
	effectstest.PassPriorityAroundTable(t, g)
	// APNAP: the active player (seat 0) says no, the rest say yes. No hand
	// changes until the last answer is in.
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	for i := 1; i < len(g.Seats); i++ {
		for j, p := range g.Seats {
			if p.Hand.Size() != j+1 {
				t.Fatalf("before seat %d answered, seat %d hand %d, want %d", i, j, p.Hand.Size(), j+1)
			}
		}
		effectstest.AnswerMayChoice(t, g, g.Seats[i].ID, true)
	}
	for i, p := range g.Seats {
		want := 7
		if i == 0 {
			want = 1
		}
		if p.Hand.Size() != want {
			t.Fatalf("seat %d hand %d, want %d", i, p.Hand.Size(), want)
		}
		if i > 0 && p.Graveyard.Size() != i+1 {
			t.Fatalf("seat %d graveyard %d, want %d", i, p.Graveyard.Size(), i+1)
		}
	}
}

// Two players' confirmations are open at once; each answer starts the
// continuation for that player, and its first step opens a number prompt that
// parks the rest. The players answer those in the opposite order, and each
// parked sequence resumes for the player who parked it (ctx.Occurrence()
// survives the park), not for the one that answered last.
func TestRunAtResumesEachParkedSequenceForItsOwnPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.RunAtProbeRan = nil
	me := g.Seats[g.Turn.ActiveSeat]
	var opps []uuid.UUID
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps = append(opps, p.ID)
		}
	}
	first, second := opps[0], opps[1]
	probe := effectstest.PushCatalogPermanent(g, me.ID, "Run At Probe", "Artifact", effectstest.RunAtProbeOracle, false)
	if err := effectstest.ActivateOn(t, g, me.ID, probe, 0, uuid.Nil); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, first, true)
	effectstest.AnswerMayChoice(t, g, second, true)
	if len(effectstest.RunAtProbeRan) != 0 {
		t.Fatalf("the last step ran for %v with the number prompts open", effectstest.RunAtProbeRan)
	}
	effectstest.AnswerNumber(t, g, second, 0)
	effectstest.AnswerNumber(t, g, first, 0)
	if len(effectstest.RunAtProbeRan) != 2 || effectstest.RunAtProbeRan[0] != second || effectstest.RunAtProbeRan[1] != first {
		t.Fatalf("the last step ran for %v, want [%s %s]: a parked sequence resumed for the wrong player",
			effectstest.RunAtProbeRan, second, first)
	}
}

// Snort — "Each player may discard their hand and draw five cards. Snort deals
// 5 damage to each opponent who discarded." The players are asked one after
// another: the first player's draw offers dredge, and the next player's question
// waits for the answer.
func TestSnortNextPlayerIsAskedAfterTheDredgeAnswer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	next := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AwbPadLibrary(me)
	effectstest.GyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	effectstest.CastCatalogSpell(t, g, "Snort", "Sorcery", snortOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if effectstest.OptionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered on the first draw")
	}
	if effectstest.LatestConfirmFor(g, next.ID) != nil {
		t.Fatal("the next player is asked with dredge open: the chain ran before the answer")
	}
	for effectstest.OptionalOffer(g, me) != nil {
		effectstest.AnswerOptional(t, g, me, false)
	}
	if effectstest.LatestConfirmFor(g, next.ID) == nil {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.S58Library(me, "Deep", "Next", "Top")
	effectstest.GyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	life := me.Life
	effectstest.S58CastRevenant(t, g, me)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OptionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered on the first draw")
	}
	if me.Life != life {
		t.Fatalf("life %d with the first offer open, want %d: the loss ran before the answer", me.Life, life)
	}
	effectstest.AnswerOptional(t, g, me, false)
	g.SettleResolution()
	if effectstest.OptionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered on the second draw")
	}
	if me.Life != life-3 {
		t.Fatalf("life %d with the second offer open, want %d: one loss so far", me.Life, life-3)
	}
	effectstest.AnswerOptional(t, g, me, false)
	g.SettleResolution()
	if me.Life != life-6 {
		t.Fatalf("life %d after both cards, want %d", me.Life, life-6)
	}
}
