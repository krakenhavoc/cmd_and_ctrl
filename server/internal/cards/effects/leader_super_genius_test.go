package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// leader_super_genius_test.go — "if a creature you control would connive,
// instead you draw a card, then that creature connives" (#4589, CR 614.1a,
// CR 701.50a). Connive is an event (game.ConniveForEffect); the replacement
// adds a card drawn first.

const leaderOracle = "bec686dd-a9ea-4db5-bc20-8e40420b9a9b"

// leaderBoard is a table where seat 0 controls Leader and a Bear
// and seat 1 a Bear. Both hands are empty, so the hand size is the draws.
type leaderBoard struct {
	g            *game.Game
	me, opp      *game.Player
	leader, bear uuid.UUID
	theirBear    uuid.UUID
}

func newLeaderBoard(t *testing.T, withLeader bool) leaderBoard {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b := leaderBoard{g: g, me: me, opp: opp}
	if withLeader {
		b.leader = b12Push(g, me.ID, "Leader, Super-Genius", "Legendary Creature — Gamma Scientist Villain", leaderOracle, 1, 3)
	}
	b.bear = b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b.theirBear = b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { me.Hand.Cards = nil; opp.Hand.Cards = nil })
	return b
}

// connive makes `creatures` connive for `controller`, as an effect does.
func (b leaderBoard) connive(t *testing.T, controller uuid.UUID, creatures ...uuid.UUID) {
	t.Helper()
	var err error
	b.g.WithWriteLock(func() {
		err = ConniveEach{Controller: controller, Creatures: creatures}.Apply(NewContext(b.g, &game.StackItem{
			Kind: game.StackItemTriggered, Controller: controller, SourceCardID: creatures[0],
		}))
	})
	if err != nil {
		t.Fatalf("ConniveEach: %v", err)
	}
}

func (b leaderBoard) handSize(p *game.Player) int {
	return p.Hand.Size()
}

func TestLeaderIsRegisteredWhole(t *testing.T) {
	spec, ok := Lookup(leaderOracle)
	if !ok {
		t.Fatal("Leader, Super-Genius is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %s, caveats %v", spec.Completeness, spec.Caveats)
	}
	if len(spec.Replacements) != 1 || len(spec.Triggered) != 1 {
		t.Errorf("%d replacements and %d triggers, want one each", len(spec.Replacements), len(spec.Triggered))
	}
}

// The control: with no Leader a connive draws one card.
func TestConniveWithoutLeaderDrawsOneCard(t *testing.T) {
	b := newLeaderBoard(t, false)
	b.connive(t, b.me.ID, b.bear)
	if got := b.handSize(b.me); got != 1 {
		t.Errorf("hand = %d, want 1", got)
	}
}

// Your creature connives: you draw a card, then it connives, so two cards in
// hand at the discard, one discarded, and the counter for a nonland card.
func TestLeaderDrawsACardThenYourCreatureConnives(t *testing.T) {
	b := newLeaderBoard(t, true)
	nonland := handCardFull(b.me, "Nonland Card", "Artifact", "", "", nil)
	library := b.me.Library.Size()
	b.connive(t, b.me.ID, b.bear)
	if got := b.handSize(b.me); got != 3 {
		t.Fatalf("hand = %d at the discard, want 3 (one held, one for Leader, one for the connive)", got)
	}
	if got := library - b.me.Library.Size(); got != 2 {
		t.Errorf("%d cards left the library, want 2", got)
	}
	answerDiscard(t, b.g, b.me.ID, nonland)
	passPriorityAroundTable(t, b.g)
	if got := b.handSize(b.me); got != 2 {
		t.Errorf("hand = %d after the discard, want 2", got)
	}
	if got := counterOn(b.g, b.bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1: it still connived", got)
	}
}

// The discard of a land still ends the connive with no counter (CR 701.50a),
// and the connive is announced once, however many cards were drawn.
func TestLeaderConniveIsOneConniveEvent(t *testing.T) {
	b := newLeaderBoard(t, true)
	land := handCardFull(b.me, "Forest", "Basic Land — Forest", "", "", nil)
	b.connive(t, b.me.ID, b.bear)
	answerDiscard(t, b.g, b.me.ID, land)
	passPriorityAroundTable(t, b.g)
	if got := len(eventsOfKind(b.g, game.EventConnived)); got != 1 {
		t.Errorf("%d connive events, want 1", got)
	}
	if got := counterOn(b.g, b.bear, game.CounterPlusOne); got != 0 {
		t.Errorf("the Bear has %d counters after discarding a land, want 0", got)
	}
}

// An opponent's creature is not yours: their connive draws one card and
// yours nothing.
func TestLeaderDoesNotReplaceAnOpponentsConnive(t *testing.T) {
	b := newLeaderBoard(t, true)
	b.connive(t, b.opp.ID, b.theirBear)
	if got := b.handSize(b.opp); got != 1 {
		t.Errorf("their hand = %d, want 1", got)
	}
	if got := b.handSize(b.me); got != 0 {
		t.Errorf("my hand = %d, want 0", got)
	}
}

// An opponent's Leader does not replace my connive.
func TestAnOpponentsLeaderDoesNotReplaceMyConnive(t *testing.T) {
	b := newLeaderBoard(t, false)
	b12Push(b.g, b.opp.ID, "Leader, Super-Genius", "Legendary Creature — Gamma Scientist Villain", leaderOracle, 1, 3)
	b.connive(t, b.me.ID, b.bear)
	if got := b.handSize(b.me); got != 1 {
		t.Errorf("hand = %d, want 1", got)
	}
}

// Leader itself conniving is a creature you control: it is replaced too.
func TestLeaderReplacesItsOwnConnive(t *testing.T) {
	b := newLeaderBoard(t, true)
	b.connive(t, b.me.ID, b.leader)
	if got := b.handSize(b.me); got != 2 {
		t.Errorf("hand = %d, want 2", got)
	}
}

// Two creatures connive together (CR 701.50c): each connive is its own
// event and is replaced on its own.
func TestLeaderReplacesEachConniveOfAChain(t *testing.T) {
	b := newLeaderBoard(t, true)
	other := b12Creature(b.g, b.me.ID, "Other Bear", "Creature — Bear", 2, 2)
	land := handCardFull(b.me, "Forest", "Basic Land — Forest", "", "", nil)
	b.connive(t, b.me.ID, b.bear, other)
	if got := b.handSize(b.me); got != 3 {
		t.Fatalf("hand = %d at the first discard, want 3", got)
	}
	answerDiscard(t, b.g, b.me.ID, land)
	if got := b.handSize(b.me); got != 4 {
		t.Fatalf("hand = %d at the second discard, want 4 (2 left + 2 drawn)", got)
	}
	answerDiscard(t, b.g, b.me.ID, b.me.Hand.Cards[0].InstanceID)
	passPriorityAroundTable(t, b.g)
	if got := len(eventsOfKind(b.g, game.EventConnived)); got != 2 {
		t.Errorf("%d connive events, want 2", got)
	}
}

// Two Leaders are two replacement effects, each with its one opportunity
// (CR 614.5, 616.1f): two cards before the connive's own.
func TestTwoLeadersDrawTwoCards(t *testing.T) {
	b := newLeaderBoard(t, true)
	b12Push(b.g, b.me.ID, "Leader, Super-Genius", "Legendary Creature — Gamma Scientist Villain", leaderOracle, 1, 3)
	b.connive(t, b.me.ID, b.bear)
	if got := b.handSize(b.me); got != 3 {
		t.Errorf("hand = %d, want 3", got)
	}
}

// Connive X keeps its X under Leader: a card is drawn first, then the
// creature connives X as it would have (CR 701.50d), drawing and discarding X.
func TestLeaderKeepsConniveX(t *testing.T) {
	b := newLeaderBoard(t, true)
	var err error
	b.g.WithWriteLock(func() {
		err = ConniveX{Controller: b.me.ID, Creature: b.bear, X: 3}.Apply(NewContext(b.g, &game.StackItem{
			Kind: game.StackItemTriggered, Controller: b.me.ID, SourceCardID: b.bear,
		}))
	})
	if err != nil {
		t.Fatalf("ConniveX: %v", err)
	}
	if got := b.handSize(b.me); got != 4 {
		t.Fatalf("hand = %d, want 4 (Leader's card and a connive 3)", got)
	}
	c := discardChoiceFor(b.g, b.me.ID)
	if c == nil {
		t.Fatal("no discard prompt")
	}
	if c.ChooseMin != 3 || c.ChooseMax != 3 {
		t.Errorf("discard %d..%d, want exactly 3", c.ChooseMin, c.ChooseMax)
	}
}

// Without a Leader connive X draws and discards X, unchanged.
func TestConniveXWithoutLeaderIsUnchanged(t *testing.T) {
	b := newLeaderBoard(t, false)
	var err error
	b.g.WithWriteLock(func() {
		err = ConniveX{Controller: b.me.ID, Creature: b.bear, X: 3}.Apply(NewContext(b.g, &game.StackItem{
			Kind: game.StackItemTriggered, Controller: b.me.ID, SourceCardID: b.bear,
		}))
	})
	if err != nil {
		t.Fatalf("ConniveX: %v", err)
	}
	if got := b.handSize(b.me); got != 3 {
		t.Errorf("hand = %d, want 3", got)
	}
	if c := discardChoiceFor(b.g, b.me.ID); c == nil || c.ChooseMax != 3 {
		t.Errorf("discard prompt %+v, want 3 cards", c)
	}
}

// A conniver Leader's controller loses the creature before its turn comes
// (CR 701.50b): the replacement still draws for its last controller.
func TestLeaderStillDrawsWhenTheCreatureHasLeft(t *testing.T) {
	b := newLeaderBoard(t, true)
	gone := uuid.New()
	b.connive(t, b.me.ID, gone)
	if got := b.handSize(b.me); got != 2 {
		t.Errorf("hand = %d, want 2", got)
	}
}

// A Leader that has left the battlefield replaces nothing.
func TestAConniveAfterLeaderLeavesDrawsOneCard(t *testing.T) {
	b := newLeaderBoard(t, true)
	b.g.WithWriteLock(func() {
		if _, err := b.g.Battlefield.Remove(b.leader); err != nil {
			t.Fatalf("remove Leader: %v", err)
		}
	})
	b.connive(t, b.me.ID, b.bear)
	if got := b.handSize(b.me); got != 1 {
		t.Errorf("hand = %d, want 1", got)
	}
}

// "Whenever a creature you control connives" still triggers exactly once on
// the replaced connive.
func TestLeaderLetsIronMongerTriggerOnce(t *testing.T) {
	b := newLeaderBoard(t, true)
	monger := b12Push(b.g, b.me.ID, "Iron Monger, Sadistic Tycoon", "Legendary Artifact Creature — Human Villain", ironMongerOracle, 2, 2)
	land := handCardFull(b.me, "Forest", "Basic Land — Forest", "", "", nil)
	b.connive(t, b.me.ID, b.bear)
	answerDiscard(t, b.g, b.me.ID, land)
	passPriorityAroundTable(t, b.g)
	if got := counterOn(b.g, monger, game.CounterPlusOne); got != 1 {
		t.Errorf("Iron Monger has %d counters, want 1 (one connive)", got)
	}
	if got := counterOn(b.g, b.leader, game.CounterPlusOne); got != 1 {
		t.Errorf("Leader (a Villain) has %d counters, want 1", got)
	}
}

// The table waiting at a replaced connive's discard is a restore point, and
// the counter still lands after the restore.
func TestLeaderConniveDiscardIsARestorePoint(t *testing.T) {
	b := newLeaderBoard(t, true)
	nonland := handCardFull(b.me, "Nonland Card", "Artifact", "", "", nil)
	b.connive(t, b.me.ID, b.bear)
	g := restoreRoundTrip(t, b.g, true)
	me := g.Seats[0]
	if got := me.Hand.Size(); got != 3 {
		t.Fatalf("restored hand = %d, want 3", got)
	}
	answerDiscard(t, g, me.ID, nonland)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, b.bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bear has %d counters on the restored table, want 1", got)
	}
	if got := len(eventsOfKind(g, game.EventConnived)); got != 1 {
		t.Errorf("%d connive events, want 1", got)
	}
}

// At the beginning of combat on your turn, the target creature connives, and
// Leader replaces that connive too: it draws two cards, discards one.
func TestLeaderConnivesATargetAtBeginningOfCombat(t *testing.T) {
	b := newLeaderBoard(t, true)
	advanceTo(t, b.g, game.StepBeginCombat)
	p := latestPickTarget(b.g, b.me.ID)
	if p == nil {
		t.Fatal("no target prompt at the beginning of combat")
	}
	for _, offered := range p.PickTargetCards {
		if offered == b.theirBear {
			t.Error("an opponent's creature is offered")
		}
	}
	before := b.handSize(b.me)
	pickCard(t, b.g, b.me.ID, b.bear)
	passPriorityAroundTable(t, b.g)
	if got := b.handSize(b.me) - before; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
	if discardChoiceFor(b.g, b.me.ID) == nil {
		t.Error("no discard prompt after the connive")
	}
}

// Ledger Shredder and Cyclonus connive through the same window.
func TestLeaderReplacesLedgerShredderAndCyclonus(t *testing.T) {
	b := newLeaderBoard(t, true)
	shredder := b12Push(b.g, b.me.ID, "Ledger Shredder", "Creature — Bird Advisor", ledgerShredderOracle, 1, 3)
	var err error
	b.g.WithWriteLock(func() {
		err = ledgerShredderConnive(b.g, &game.StackItem{
			Kind: game.StackItemTriggered, Controller: b.me.ID, SourceCardID: shredder,
		})
	})
	if err != nil {
		t.Fatalf("Ledger Shredder: %v", err)
	}
	if got := b.handSize(b.me); got != 2 {
		t.Fatalf("Ledger Shredder: hand = %d, want 2", got)
	}
	answerDiscard(t, b.g, b.me.ID, b.me.Hand.Cards[0].InstanceID)
	passPriorityAroundTable(t, b.g)

	cyclonus := b12Push(b.g, b.me.ID, "Cyclonus, the Saboteur", "Legendary Artifact Creature — Robot", cyclonusOracleID, 2, 5)
	before := b.handSize(b.me)
	b.g.WithWriteLock(func() {
		err = cyclonusConnive(b.g, &game.StackItem{
			Kind: game.StackItemTriggered, Controller: b.me.ID, SourceCardID: cyclonus,
		})
	})
	if err != nil {
		t.Fatalf("Cyclonus: %v", err)
	}
	if got := b.handSize(b.me) - before; got != 2 {
		t.Errorf("Cyclonus: drew %d cards, want 2", got)
	}
}
