package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// read_ahead_test.go — read ahead (CR 702.155, #2123), through the real
// read-ahead Sagas of the catalog.
//
//	CR 702.155a  Chapter abilities of this Saga can't trigger the turn it
//	             entered the battlefield unless it has exactly the number
//	             of lore counters on it specified in the chapter symbol.
//	CR 702.155b  As this Saga enters, choose a number between one and
//	             this Saga's final chapter number; it enters with that
//	             many lore counters.

const (
	crueltyOfGixOracle    = "2b5c62af-af5f-4beb-9688-19604ca46918"
	loveSongOracle        = "9b87b4cd-4fb9-4bb3-b4c0-de1fcc6f44a2"
	elderDragonWarOracle  = "4945f03b-ff99-4923-b62d-3cdb60275ef1"
	readAheadSagaTypeLine = "Enchantment — Saga"
)

func readAheadPrompt(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceEntryReadAhead {
			return c
		}
	}
	return nil
}

// castReadAheadSaga casts a read-ahead Saga from the active seat and
// passes priority until its entry asks.
func castReadAheadSaga(t *testing.T, g *game.Game, name, oracle string) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, name, readAheadSagaTypeLine, oracle, nil)
	passPriorityAroundTable(t, g)
	if readAheadPrompt(g) == nil {
		t.Fatalf("%s resolved without asking for its starting chapter", name)
	}
	return id
}

// chooseChapter answers the open read ahead prompt with chapter n.
func chooseChapter(t *testing.T, g *game.Game, n int) {
	t.Helper()
	c := readAheadPrompt(g)
	if c == nil {
		t.Fatal("no entry_read_ahead prompt is open")
	}
	if err := g.ResolveEntryReadAhead(c.ID, c.Chooser, n-1); err != nil {
		t.Fatalf("ResolveEntryReadAhead(chapter %d): %v", n, err)
	}
}

// chaptersFired lists the chapter numbers the Saga has triggered, in
// the order they fired.
func chaptersFired(g *game.Game, saga uuid.UUID) []int {
	var out []int
	for _, ev := range g.Events {
		if ev.Kind == game.EventSagaChapter && ev.CardID == saga {
			out = append(out, ev.Amount)
		}
	}
	return out
}

// The question is asked of the caster before the Saga is on the
// battlefield (CR 614.12a), offering every chapter from I to the final
// one (CR 702.155b).
func TestReadAheadAsksForTheStartingChapterBeforeTheSagaEnters(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	id := castReadAheadSaga(t, g, "The Elder Dragon War", elderDragonWarOracle)

	c := readAheadPrompt(g)
	if c.Chooser != caster.ID {
		t.Fatalf("chooser = %s, want the caster", c.Chooser)
	}
	if c.Source != id {
		t.Fatalf("source = %s, want the entering Saga", c.Source)
	}
	var labels []string
	for _, o := range c.PickOptions {
		labels = append(labels, o.Label)
	}
	if want := []string{"Chapter I", "Chapter II", "Chapter III"}; !slices.Equal(labels, want) {
		t.Fatalf("options = %v, want %v", labels, want)
	}
	if _, onBF := battlefieldCardByID(g, id); onBF {
		t.Fatal("the Saga is on the battlefield before its chapter was chosen")
	}
	if err := g.PassPriority(); err == nil {
		t.Fatal("the table moved on with the read ahead question open")
	}
}

// Chapter I is the ordinary lifecycle: one lore counter, chapter I
// fires as the Saga enters, and chapter II on the next precombat main.
func TestReadAheadChapterOneIsTheOrdinaryLifecycle(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	meLife, oppLife := me.Life, opp.Life

	id := castReadAheadSaga(t, g, "The Elder Dragon War", elderDragonWarOracle)
	chooseChapter(t, g, 1)
	passPriorityAroundTable(t, g)

	if got := loreCountersOn(g, id); got != 1 {
		t.Fatalf("lore counters = %d, want 1", got)
	}
	if got := chaptersFired(g, id); !slices.Equal(got, []int{1}) {
		t.Fatalf("chapters fired = %v, want [1]", got)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("chapter I did not deal 2 damage to the creature")
	}
	if lifeOf(g, opp.ID) != oppLife-2 {
		t.Errorf("opponent life = %d, want %d", lifeOf(g, opp.ID), oppLife-2)
	}
	if lifeOf(g, me.ID) != meLife {
		t.Errorf("the Saga's controller took damage: life %d, want %d", lifeOf(g, me.ID), meLife)
	}

	advanceToPrecombatMainOf(t, g, seat)
	if got := loreCountersOn(g, id); got != 2 {
		t.Fatalf("lore counters a turn later = %d, want 2", got)
	}
	if got := chaptersFired(g, id); !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("chapters fired a turn later = %v, want [1 2]", got)
	}
}

// Choosing the final chapter fires only that chapter (CR 702.155a), and
// the Saga is sacrificed once it has resolved (CR 714.4).
func TestReadAheadFinalChapterFiresOnlyItThenTheSagaIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	oppLife := opp.Life

	id := castReadAheadSaga(t, g, "The Elder Dragon War", elderDragonWarOracle)
	chooseChapter(t, g, 3)
	if got := loreCountersOn(g, id); got != 3 {
		t.Fatalf("lore counters = %d, want 3", got)
	}
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Saga was sacrificed before its final chapter resolved")
	}
	passPriorityAroundTable(t, g)

	if got := chaptersFired(g, id); !slices.Equal(got, []int{3}) {
		t.Fatalf("chapters fired = %v, want only [3]", got)
	}
	if lifeOf(g, opp.ID) != oppLife {
		t.Error("the skipped chapter I dealt damage")
	}
	if got := countBattlefieldByName(g, "Dragon"); got != 1 {
		t.Errorf("Dragons = %d, want 1 from chapter III", got)
	}
	if g.Battlefield.Contains(id) || !inGraveyardOf(g, me.ID, id) {
		t.Error("the Saga was not sacrificed after its final chapter resolved")
	}
}

// A middle chapter fires and the chapters before it do not.
func TestReadAheadMiddleChapterSkipsTheEarlierOnes(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	hand := len(me.Hand.Cards)

	id := castReadAheadSaga(t, g, "Love Song of Night and Day", loveSongOracle)
	chooseChapter(t, g, 2)
	passPriorityAroundTable(t, g)

	if got := chaptersFired(g, id); !slices.Equal(got, []int{2}) {
		t.Fatalf("chapters fired = %v, want only [2]", got)
	}
	if got := countBattlefieldByName(g, "Bird"); got != 1 {
		t.Errorf("Birds = %d, want 1 from chapter II", got)
	}
	// The cast card was put into the hand and cast from it, so the hand
	// is back where it was; the skipped chapter I would have drawn two.
	if got := len(me.Hand.Cards); got != hand {
		t.Errorf("hand = %d, want %d: the skipped chapter I drew", got, hand)
	}
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Saga left with chapter III still to come")
	}

	advanceToPrecombatMainOf(t, g, seat)
	if got := chaptersFired(g, id); !slices.Equal(got, []int{2, 3}) {
		t.Fatalf("chapters fired a turn later = %v, want [2 3]", got)
	}
}

// A proliferate the turn the Saga entered fires the chapter it lands on
// exactly (CR 702.155a), and only that one.
func TestReadAheadProliferateTheSameTurnFiresTheChapterItLandsOn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]

	id := castReadAheadSaga(t, g, "Love Song of Night and Day", loveSongOracle)
	chooseChapter(t, g, 1)
	passPriorityAroundTable(t, g)
	pickPlayer(t, g, me.ID, g.Seats[(seat+1)%4].ID)
	passPriorityAroundTable(t, g)
	if got := chaptersFired(g, id); !slices.Equal(got, []int{1}) {
		t.Fatalf("chapters fired on entry = %v, want [1]", got)
	}

	g.WithWriteLock(func() {
		if err := g.ProliferateForEffect(me.ID, uuid.Nil, []uuid.UUID{id}, nil); err != nil {
			t.Fatalf("ProliferateForEffect: %v", err)
		}
	})
	passPriorityAroundTable(t, g)

	if got := loreCountersOn(g, id); got != 2 {
		t.Fatalf("lore counters after proliferate = %d, want 2", got)
	}
	if got := chaptersFired(g, id); !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("chapters fired after proliferate = %v, want [1 2]", got)
	}
	if got := countBattlefieldByName(g, "Bird"); got != 1 {
		t.Errorf("Birds = %d, want 1 from the proliferated chapter II", got)
	}
}

// Doubling Season doubles the chosen count (the counters are put on by
// the entry, an effect), and a Saga that enters past every chapter
// number has no chapter with exactly that many lore counters: nothing
// fires, and it is sacrificed (CR 714.4).
func TestReadAheadDoubledPastTheFinalChapterFiresNothing(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Doubling Season",
		OracleID:   doublingSeasonSagaOracl,
		TypeLine:   "Enchantment",
		Owner:      me.ID,
		Controller: me.ID,
	})

	id := castReadAheadSaga(t, g, "The Elder Dragon War", elderDragonWarOracle)
	chooseChapter(t, g, 2)
	passPriorityAroundTable(t, g)

	if got := chaptersFired(g, id); len(got) != 0 {
		t.Fatalf("chapters fired = %v, want none (four lore counters, no chapter IV)", got)
	}
	if g.Battlefield.Contains(id) || !inGraveyardOf(g, me.ID, id) {
		t.Error("the Saga past its final chapter was not sacrificed")
	}
}

// The choice is the permanent's ability, not the spell's (CR 702.155b):
// a read-ahead Saga put onto the battlefield by an effect asks too.
func TestReadAheadAsksWhenTheSagaIsPutOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: id,
		Name:       "The Elder Dragon War",
		TypeLine:   readAheadSagaTypeLine,
		OracleID:   elderDragonWarOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardForEffect(id, game.ZoneBattlefield); err != nil {
			t.Fatalf("ReturnFromGraveyardForEffect: %v", err)
		}
	})
	if readAheadPrompt(g) == nil {
		t.Fatal("a read-ahead Saga put onto the battlefield was not asked for its chapter")
	}
	chooseChapter(t, g, 3)
	passPriorityAroundTable(t, g)
	if got := chaptersFired(g, id); !slices.Equal(got, []int{3}) {
		t.Fatalf("chapters fired = %v, want [3]", got)
	}
}

// A token copy of a read-ahead Saga has read ahead (it is copiable
// text), so its entry asks too.
func TestReadAheadTokenCopyAsks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := castReadAheadSaga(t, g, "The Elder Dragon War", elderDragonWarOracle)
	chooseChapter(t, g, 3)
	// Leave the original's chapter III on the stack: its source has to
	// still be on the battlefield to be copied.
	if !g.Battlefield.Contains(id) {
		t.Fatal("the original left before it could be copied")
	}
	g.WithWriteLock(func() {
		tmpl, ok := TokenCopyTemplate(g, id)
		if !ok {
			t.Fatal("TokenCopyTemplate failed")
		}
		if err := g.CreateTokenForEffect(me.ID, tmpl, 1); err != nil {
			t.Fatalf("CreateTokenForEffect: %v", err)
		}
	})
	c := readAheadPrompt(g)
	if c == nil {
		t.Fatal("the token copy did not ask for its starting chapter")
	}
	if len(c.PickOptions) != 3 {
		t.Fatalf("token copy offered %d chapters, want 3", len(c.PickOptions))
	}
}

// A table with the question open is not a restore point: the paused
// entry is a continuation, counted in the census. Once it is answered,
// the restore point carries the Saga and the turn it entered, so the
// chapter rule still holds after a restore.
func TestReadAheadRestoreWithTheChoicePending(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	id := castReadAheadSaga(t, g, "Love Song of Night and Day", loveSongOracle)

	snap := g.CaptureSnapshot()
	if snap.Restorable() {
		t.Fatal("a table with the read ahead question open reports itself restorable")
	}
	if snap.Continuations.ChoiceResumeFrames == 0 {
		t.Fatal("the census does not count the paused entry")
	}

	chooseChapter(t, g, 2)
	passPriorityAroundTable(t, g)
	snap = g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("the settled table is not restorable: %+v", snap.Continuations)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	// Still the turn it entered: a proliferate to III fires III.
	restored.WithWriteLock(func() {
		if err := restored.ProliferateForEffect(me.ID, uuid.Nil, []uuid.UUID{id}, nil); err != nil {
			t.Fatalf("ProliferateForEffect: %v", err)
		}
	})
	if got := chaptersFired(restored, id); !slices.Equal(got[len(got)-1:], []int{3}) {
		t.Fatalf("chapters fired after the restore = %v, want chapter III last", got)
	}
}

// The bot's half (#544's rule): the enumerator offers one answer per
// chapter and nothing else while the prompt blocks the table, and every
// offered answer is one the dispatcher accepts.
func TestReadAheadEnumeratorOffersEachChapterAndDispatches(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	id := castReadAheadSaga(t, g, "The Elder Dragon War", elderDragonWarOracle)
	var answers []legal.Move
	for _, m := range legal.EnumerateFor(g, caster.ID) {
		if m.Type != legal.TypeResolveChoice {
			t.Fatalf("the chooser was offered %q while the read ahead prompt blocks the table", m.Type)
		}
		answers = append(answers, m)
	}
	if len(answers) != 3 {
		t.Fatalf("enumerated %d answers, want one per chapter (3)", len(answers))
	}
	last := answers[2]
	if err := actions.Dispatch(g, actions.Action{
		Type: actions.Type(last.Type), Player: last.Player, Caller: last.Player, Params: last.Params,
	}); err != nil {
		t.Fatalf("dispatch the enumerated answer: %v", err)
	}
	if got := loreCountersOn(g, id); got != 3 {
		t.Fatalf("lore counters = %d, want 3 from the third answer", got)
	}
}

// The Cruelty of Gix started on chapter III reanimates a creature card
// from an opponent's graveyard under its controller, and skips the
// discard and the search.
func TestTheCrueltyOfGixStartedOnChapterThree(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	corpse := uuid.New()
	opp.Graveyard.PushTop(game.Card{
		InstanceID: corpse, Name: "Corpse", TypeLine: "Creature — Zombie",
		Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID,
	})
	meLife := me.Life

	id := castReadAheadSaga(t, g, "The Cruelty of Gix", crueltyOfGixOracle)
	chooseChapter(t, g, 3)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, corpse)
	passPriorityAroundTable(t, g)

	card, ok := battlefieldCardByID(g, corpse)
	if !ok {
		t.Fatal("chapter III did not put the creature card onto the battlefield")
	}
	if card.Controller != me.ID {
		t.Errorf("reanimated creature's controller = %s, want the Saga's controller", card.Controller)
	}
	if got := chaptersFired(g, id); !slices.Equal(got, []int{3}) {
		t.Fatalf("chapters fired = %v, want [3]", got)
	}
	if me.Life != meLife {
		t.Error("the skipped chapter II cost life")
	}
	if g.Battlefield.Contains(id) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}
