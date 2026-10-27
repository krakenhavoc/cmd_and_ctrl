package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_03_restore_test.go — #4253's closure sweep, batch 03 (goblin_psychopath
// through magister_of_worth): the prompts that carried a Go closure are keyed
// now, so a table waiting on any of them is a restore point. Each test opens
// the prompt, takes a STRICT restore through JSON (restoreRoundTrip with
// strict, which fails on a table that is not a restore point) and answers on
// the restored table.

// sweep03OptionKey returns the open option pick of `chooser`, failing the
// test when it names a key other than `want`.
func sweep03OptionKey(t *testing.T, g *game.Game, chooser uuid.UUID, want string) *game.PendingChoice {
	t.Helper()
	c := latestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatalf("no option pick for %s: %+v", chooser, g.PendingChoices)
	}
	if c.PickThen != want {
		t.Fatalf("the option pick names key %q, want %q", c.PickThen, want)
	}
	return c
}

// Grapple with the Past (and every card that rides ReturnChosenFromGraveyard):
// the hand return is a keyed pick whose destination is carried data, and the
// milled creature comes back on the restored table.
func TestSweep03ReturnChosenFromGraveyardToHandIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	for i := 0; i < 5; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	}
	milled := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: milled, Name: "Milled Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Grapple with the Past", "Instant", b20GrappleWithThePastOracle, nil)
	passPriorityAroundTable(t, g)
	sweep00PickKey(t, g, me.ID, "choose-cards/return-chosen-from-graveyard")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep00PickKey(t, g, me.ID, "choose-cards/return-chosen-from-graveyard")
	if !hasID(c.ChooseCards, milled) {
		t.Fatal("the restored pick does not offer the creature the spell milled")
	}
	answerChooseCards(t, g, me.ID, milled)
	if !me.Hand.Contains(milled) || me.Graveyard.Contains(milled) {
		t.Error("the milled creature did not return to hand on the restored table")
	}
}

// The same key, the other destination: a return to the battlefield, tapped
// when the carried number says so.
func TestSweep03ReturnChosenFromGraveyardToBattlefieldTappedIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b38Graveyard(g, me.ID, "Dead Bear", "Creature — Bear", "{1}{G}", 2, 2)
	g.WithWriteLock(func() {
		item := &game.StackItem{Kind: game.StackItemSpell, Controller: me.ID, Owner: me.ID}
		err := (ReturnChosenFromGraveyard{Question: "Return a creature", Match: game.Card.IsCreature, Min: 1, Max: 1, Tapped: true}).Apply(NewContext(g, item))
		if err != nil {
			t.Fatal(err)
		}
	})
	sweep00PickKey(t, g, me.ID, "choose-cards/return-chosen-from-graveyard")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep00PickKey(t, g, me.ID, "choose-cards/return-chosen-from-graveyard")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	if !g.Battlefield.Contains(bear) || !b25Tapped(t, g, bear) {
		t.Error("the creature did not return tapped on the restored table")
	}
}

// Rickety Gazebo: the "up to two from among them" pick is keyed, and the
// picks reach the hand on the restored table.
func TestSweep03RicketyGazeboPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	ids := seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Ring", TypeLine: "Artifact"},
		game.Card{Name: "Deep", TypeLine: "Sorcery"},
	)
	forest, ring := ids[0], ids[3]
	me.Hand.PushTop(greenhouseCard(me.ID))
	c := me.Hand.Cards[len(me.Hand.Cards)-1]
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatal(err)
	}
	roomsDSettle(t, g, me.ID)
	sweep00PickKey(t, g, me.ID, "choose-cards/rooms-d-take-up-to-from-milled")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	answerChooseCards(t, g, me.ID, forest, ring)
	if !me.Hand.Contains(forest) || !me.Hand.Contains(ring) {
		t.Error("the two picked permanents did not reach the hand on the restored table")
	}
}

// Hidden Strings: one keyed option pick per target, the targets still to ask
// about carried as data, so the table is a restore point at each question.
func TestSweep03HiddenStringsQuestionsAreRestorePoints(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ciaSeats(t, g)
	theirs := seedBear(g, opp.ID)
	mine := seedBear(g, me.ID)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(mine) })
	castCatalogSpell(t, g, "Hidden Strings", "Sorcery", ciaHiddenStringsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}, {Kind: game.TargetCard, ID: mine}})
	if cipherResolveToPrompt(t, g, me.ID) == nil {
		t.Fatal("the first question")
	}
	sweep03OptionKey(t, g, me.ID, "option-pick/hidden-strings-tap-or-untap")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	first := sweep03OptionKey(t, g, me.ID, "option-pick/hidden-strings-tap-or-untap")
	if first.PickOptions[0].Cards[0] != theirs {
		t.Fatalf("the restored first question is about %v, want the first target", first.PickOptions[0].Cards)
	}
	answerOptionPick(t, g, me.ID, 0) // tap it
	if !ciaTapped(g, theirs) {
		t.Error("the first target is not tapped on the restored table")
	}

	// And restored again with the second question open.
	second := sweep03OptionKey(t, g, me.ID, "option-pick/hidden-strings-tap-or-untap")
	if second.PickOptions[0].Cards[0] != mine {
		t.Fatalf("the second question is about %v, want the second target", second.PickOptions[0].Cards)
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	sweep03OptionKey(t, g, me.ID, "option-pick/hidden-strings-tap-or-untap")
	answerOptionPick(t, g, me.ID, 1) // untap it
	if ciaTapped(g, mine) {
		t.Error("the second target is still tapped on the restored table")
	}
}

// Indulgent Tormentor: the opponent's answer is keyed. Declining draws a card
// for the Tormentor's controller on the restored table (the controller rides
// the prompt), and paying 3 life is the opponent's own life.
func TestSweep03IndulgentTormentorAnswerIsARestorePoint(t *testing.T) {
	for _, tc := range []struct {
		name   string
		answer int
		draws  bool
	}{{"decline draws", 0, true}, {"pay life stops the draw", 1, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[1], g.Seats[2]
			pushCatalogPermanent(g, me.ID, "Indulgent Tormentor", "Creature — Demon", indulgentTormentorOracle, false)
			advanceToUpkeepOf(t, g, 1)
			pickPlayer(t, g, me.ID, opp.ID)
			passPriorityAroundTable(t, g)
			sweep03OptionKey(t, g, opp.ID, "option-pick/indulgent-tormentor-answer")

			g = restoreRoundTrip(t, g, true)
			me, opp = g.Seats[1], g.Seats[2]
			sweep03OptionKey(t, g, opp.ID, "option-pick/indulgent-tormentor-answer")
			hand, life := me.Hand.Size(), opp.Life
			answerOptionPick(t, g, opp.ID, tc.answer)
			if tc.draws && me.Hand.Size() != hand+1 {
				t.Errorf("hand %d -> %d on the restored table, want +1", hand, me.Hand.Size())
			}
			if !tc.draws && (me.Hand.Size() != hand || opp.Life != life-3) {
				t.Errorf("hand %d -> %d, life %d -> %d on the restored table, want no draw and -3 life",
					hand, me.Hand.Size(), life, opp.Life)
			}
		})
	}
}

// Ledger Shredder: the connive discard is a keyed discard-then whose source is
// the Shredder, so a nonland discard puts the counter on it after a restore.
func TestSweep03LedgerShredderDiscardIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	shredder := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ledger Shredder", TypeLine: "Creature — Bird Advisor",
		OracleID: ledgerShredderOracle, Power: 1, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	handCardForTest(me, "Pitch Spell", "Sorcery", "")
	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	nonland := handCardForTest(me, "Nonland Card", "Artifact", "")
	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("no connive discard prompt")
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	answerDiscard(t, g, me.ID, nonland)
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, shredder)
	if got := c.CurrentPower(); got != 2 {
		t.Errorf("Ledger Shredder power = %d on the restored table, want 2 (a +1/+1 counter)", got)
	}
}

// Lich's Mastery: the exile pick is keyed, the pool, the count and the chooser
// all read back from the prompt, so the picks leave on the restored table.
func TestSweep03LichsMasteryExilePickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Lich's Mastery", "Legendary Enchantment", lichsMasteryOracle, false)
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	inHand := handCardForTest(me, "Forest", "Basic Land — Forest", "")
	_ = handCardForTest(me, "Island", "Basic Land — Island", "")
	inYard := pushGraveyardCardForTest(me, "Old Card")
	_ = pushGraveyardCardForTest(me, "Older Card")
	src := pr7Creature(g, opp.ID, "Source", 3, "R")
	pr7Hit(t, g, src, me.ID, 3)
	passPriorityAroundTable(t, g)
	sweep00PickKey(t, g, me.ID, "choose-cards/lichs-mastery-exile")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	pick := sweep00PickKey(t, g, me.ID, "choose-cards/lichs-mastery-exile")
	if pick.ChooseMin != 3 || pick.ChooseMax != 3 {
		t.Fatalf("restored bounds %d..%d, want exactly 3", pick.ChooseMin, pick.ChooseMax)
	}
	answerChooseCards(t, g, me.ID, bear, inHand, inYard)
	for _, id := range []uuid.UUID{bear, inHand, inYard} {
		g.ReadSnapshot(func() {
			if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
				t.Errorf("%v is not in exile on the restored table", id)
			}
		})
	}
}
