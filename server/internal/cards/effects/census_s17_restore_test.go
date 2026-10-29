package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// census_s17_restore_test.go — #4361 S17 (owner decision D4): an entry
// whose "as this enters, discard or sacrifice" choice has been answered
// is frozen as data (GameSnapshot.FrozenEntries) while its discard or
// sacrifice settles, and the action's continuation is a registered tail
// (engine/entry-choice-acted). So a CR 616 prompt in the middle of that
// action is a restore point. Each test takes a STRICT restore (the boot
// path) while the prompt is open, answers on the restored table, and
// checks the entry finished. Beside them: Obstinate Baloth's discard and
// Library of Leng's top order, now that every discard is the keyed batch,
// and a replaced move's life sequence (ADR 0178), now written.

// s17Snapshot captures g and fails unless it is a restore point carrying
// `frozen` frozen entries.
func s17Snapshot(t *testing.T, g *game.Game, frozen int) {
	t.Helper()
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("the table is not a restore point: %+v", snap.Continuations)
	}
	if len(snap.FrozenEntries) != frozen {
		t.Fatalf("the snapshot carries %d frozen entries, want %d", len(snap.FrozenEntries), frozen)
	}
}

// --- an as-enters sacrifice --------------------------------------------------

// Lotus Vale over a Plains and a land named Needle: the Needle's sacrifice
// asks for the order of two graveyard probes, the restored table has the
// Plains in the graveyard and Lotus Vale not yet on the battlefield, and
// the answer finishes the sacrifice and puts Lotus Vale onto the
// battlefield.
func TestAnAsEntersSacrificeIsARestorePointMidSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plains := pushSacrificeCandidate(g, me.ID, "Plains", "Basic Land — Plains", false)
	needle := pushSacrificeCandidate(g, me.ID, "Needle", "Basic Land — Island", false)
	seedOrderProbes(g, me.ID, graveyardOrderProbeOracles, "Graveyard Order Probe")

	vale := playLandFromHand(t, g, "Lotus Vale", sacrificeInsteadLandOracleIDs["Lotus Vale"])
	answerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, plains, needle)
	s11cOnlyPrompt(t, g, game.PendingChoiceReplacementOrder)
	s17Snapshot(t, g, 1)

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	me = seatByID(g, me.ID)
	if !me.Graveyard.Contains(plains) {
		t.Fatal("the Plains is not in the graveyard on the restored table")
	}
	if _, ok := battlefieldCard(g, needle); !ok {
		t.Fatal("the Needle left the battlefield before the restored answer")
	}
	if _, ok := battlefieldCard(g, vale); ok {
		t.Fatal("Lotus Vale entered before its sacrifice finished")
	}
	answerExileOrder(t, g)
	if !me.Graveyard.Contains(needle) {
		t.Fatal("the Needle was not sacrificed after the restored answer")
	}
	if _, ok := battlefieldCard(g, vale); !ok {
		t.Fatal("Lotus Vale did not enter after the restored answer")
	}
	if n := sacrificeEventsFor(g, needle); n != 1 {
		t.Errorf("%d EventSacrifice for the Needle, want 1", n)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("prompts left open: %+v", g.PendingChoices)
	}
}

// --- an as-enters discard ----------------------------------------------------

// Mox Diamond discarding a land under Library of Leng: Leng's "may" is a
// restore point with the Mox's entry frozen behind it, and either answer
// on the restored table finishes the entry.
func TestAnAsEntersDiscardIsARestorePointAtLibraryOfLeng(t *testing.T) {
	for _, apply := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		pushTokenReplacementCard(g, libraryOfLengOracle, "Library of Leng", "Artifact", me.ID)
		g.WithWriteLock(func() { me.Hand.Cards = nil })
		forest := handCard(me, "Forest", "Basic Land — Forest")

		mox := castMox(t, g)
		answerEntryChoice(t, g, game.PendingChoiceEntryDiscardFromHand, me.ID, forest)
		s17Snapshot(t, g, 1)

		g = strictRestoreWithFrame(t, g, game.ReplacementResumeDiscard)
		me = seatByID(g, me.ID)
		if _, ok := battlefieldCard(g, mox); ok {
			t.Fatal("the Mox entered before Leng was answered on the restored table")
		}
		if !me.Hand.Contains(forest) {
			t.Fatal("the land left the hand before Leng was answered")
		}
		answerLengPrompt(t, g, me.ID, apply)

		if _, ok := battlefieldCard(g, mox); !ok {
			t.Fatalf("apply %v: the Mox did not enter after the restored answer", apply)
		}
		if g.Stack.Contains(mox) {
			t.Errorf("apply %v: the Mox was left on the stack", apply)
		}
		if apply && !me.Library.Contains(forest) {
			t.Error("Leng applied, but the land is not in the library")
		}
		if !apply && !me.Graveyard.Contains(forest) {
			t.Error("Leng declined, but the land is not in the graveyard")
		}
	}
}

// --- Obstinate Baloth --------------------------------------------------------

// An opponent's Mind Rot over a filler and an Obstinate Baloth, under the
// victim's Library of Leng: the Baloth's discard asks for the order of
// Leng and the Baloth's own replacement, the restored table has the Baloth
// in hand, and taking the Baloth's first puts it onto the battlefield
// through the keyed discard entry (the closure branch is gone).
func TestObstinateBalothsDiscardIsARestorePointAtItsOrder(t *testing.T) {
	d := newDBOTable(t)
	leng := pushTokenReplacementCard(d.g, libraryOfLengOracle, "Library of Leng", "Artifact", d.victim.ID)
	filler := d.filler(d.victim)
	baloth := d.handCard(d.victim, "Obstinate Baloth", "Creature — Beast", "{2}{G}{G}", obstinateBalothOracle, 4, 4)
	castCatalogSpell(t, d.g, "Mind Rot", "Sorcery", mindRotOracleID, []game.TargetRef{{Kind: game.TargetPlayer, ID: d.victim.ID}})
	passPriorityAroundTable(t, d.g)
	answerDiscard(t, d.g, d.victim.ID, baloth, filler)
	order := s11cOnlyPrompt(t, d.g, game.PendingChoiceReplacementOrder)
	if len(order.ReplacementEffectIDs) != 2 {
		t.Fatalf("the order offers %d effects, want Leng and the Baloth", len(order.ReplacementEffectIDs))
	}

	g := strictRestoreWithFrame(t, d.g, game.ReplacementResumeDiscard)
	victim := seatByID(g, d.victim.ID)
	if !victim.Hand.Contains(baloth) || !victim.Hand.Contains(filler) {
		t.Fatal("the restored table discarded before the order was answered")
	}
	mark := len(g.Events)
	// The Baloth's own replacement first, Leng's last.
	p := s11cOnlyPrompt(t, g, game.PendingChoiceReplacementOrder)
	last := replacementIDForSource(t, g, p.ReplacementEffectIDs, leng)
	var answer []game.ReplacementEffectID
	for _, id := range p.ReplacementEffectIDs {
		if id != last {
			answer = append(answer, id)
		}
	}
	answer = append(answer, last)
	if err := g.ResolveReplacementOrder(p.ID, p.Chooser, answer); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	// CR 616.1e: Leng still applies to the modified discard, and asks; a
	// "no" lets the Baloth's battlefield stand. That "may" is a restore
	// point too.
	g = strictRestoreWithFrame(t, g, game.ReplacementResumeDiscard)
	victim = seatByID(g, d.victim.ID)
	answerLengPrompt(t, g, victim.ID, false)
	if _, ok := findInZone(g.Battlefield, baloth); !ok {
		t.Fatal("the Baloth is not on the battlefield after the restored answers")
	}
	logged := false
	for _, ev := range g.Events[mark:] {
		if ev.Kind == game.EventDiscardCard && ev.CardID == baloth && ev.NewZone == game.ZoneBattlefield {
			logged = true
		}
	}
	if !logged {
		t.Error("the Baloth's discard was not logged on the battlefield")
	}
	// The rest of the batch went on: Leng offers its "may" for the filler.
	if len(g.PendingChoices) == 1 && g.PendingChoices[0].Kind == game.PendingChoiceOptionalReplacement {
		answerLengPrompt(t, g, victim.ID, false)
	}
	if !victim.Graveyard.Contains(filler) {
		t.Error("the filler was not discarded after the Baloth")
	}
}

// --- Library of Leng's top order ---------------------------------------------

// Two cards Library of Leng saves: the owner's CR 401.4 order is a restore
// point, and the restored answer is the library's top with the discard's
// "then" after it.
func TestLibraryOfLengsTopOrderIsARestorePoint(t *testing.T) {
	g, me, cards := lengBoard(t, 2)
	ran := 0
	lengDiscardThen(t, g, me, &ran, cards...)
	answerLengPrompt(t, g, me.ID, true)
	// The second card's "may" is a restore point too.
	g = restoreRoundTrip(t, g, true)
	me = seatByID(g, me.ID)
	answerLengPrompt(t, g, me.ID, true)
	if putInLibraryPrompt(g) == nil {
		t.Fatal("no ordering prompt for the two saved cards")
	}

	g = restoreRoundTrip(t, g, true)
	me = seatByID(g, me.ID)
	prompt := putInLibraryPrompt(g)
	if prompt == nil {
		t.Fatal("the restored table lost the ordering prompt")
	}
	if ran != 0 {
		t.Fatal("the discard's then ran before the order")
	}
	want := append([]uuid.UUID(nil), cards...)
	if err := g.ResolvePutInLibrary(prompt.ID, me.ID, nil, want); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := lengLibraryTop(me, 2); !sameIDs(got, want) {
		t.Errorf("library top %v, want the chosen %v", got, want)
	}
	if ran != 1 {
		t.Errorf("the discard's then ran %d times, want once after the order", ran)
	}
}

// --- a replaced move's life sequence (ADR 0178) ------------------------------

// s17CounterOnCarry puts a +1/+1 counter on Carry[0]: a destroy's "then"
// a test can read back.
var s17CounterOnCarry = CardTail("test/4361-s17-counter-on-carry", func(ctx *Context, o game.TailOutcome) error {
	return ctx.Game.AddCounterForEffect(o.Tail.Carry[0], game.CounterPlusOne, 1)
})

// The Darkness Crystal exiles an opponent's dying creature and gains its
// controller 2 life; under Rhox Faithmender and Alhammarret's Archive the
// gain asks for an order. The rest of the destroy waits on it
// (lifeTail.move). The restored order gains 8 and the destroy's "then"
// runs after it.
func TestAReplacedMovesLifeChangeIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kbCrystal(g, me.ID)
	faith := seedReplacementPermanent(g, rhoxFaithmenderOracle, "Rhox Faithmender", me.ID)
	archive := seedReplacementPermanent(g, alhammarretsArchiveOracle, "Alhammarret's Archive", me.ID)
	mine := seedCreature(g, "Bears", me.ID)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	life := me.Life
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentsKeyedForEffect([]uuid.UUID{theirs},
			game.KeyedTail{Ref: s17CounterOnCarry, Player: me.ID, Carry: []uuid.UUID{mine}}); err != nil {
			t.Fatalf("DestroyPermanentsKeyedForEffect: %v", err)
		}
	})
	s11cOnlyPrompt(t, g, game.PendingChoiceReplacementOrder)
	if !g.Exile.Contains(theirs) {
		t.Fatal("the creature was not exiled before the gain asked")
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeLifeTail)
	me = seatByID(g, me.ID)
	if got := countersOn(g, mine, "+1/+1"); got != 0 {
		t.Fatal("the destroy's then ran before the gain was answered")
	}
	answerOrderBySource(t, g, archive, faith)
	if me.Life != life+8 {
		t.Errorf("life %d -> %d, want +8", life, me.Life)
	}
	if got := countersOn(g, mine, "+1/+1"); got != 1 {
		t.Errorf("the destroy's then ran %d times, want once after the gain", got)
	}
}
