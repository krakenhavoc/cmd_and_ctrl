package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// choose_cards_s7b_restore_test.go — #4361 slice S7b. The last card-set
// picks whose continuation was a closure frame (PendingChoice's
// chooseCardsResume) are data now: the fire-and-forget discard is a keyed
// leg, a run's reveal pick and permanent pick settle from their own fields,
// and an entry card choice reads its zone off its kind. Each is captured
// with its prompt open, restored strictly through JSON and answered in the
// restored game.

// A fire-and-forget discard ("target player discards a card", nothing
// after it) is a restore point, and the restored prompt discards the pick.
func TestS7bAFireAndForgetDiscardSurvivesARestore(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[1]
	pick := oneCardHand(t, g, p, "Pitch")
	if id := queueEffectDiscard(g, DiscardPrompt{Player: p.ID, N: 1}); id == uuid.Nil {
		t.Fatal("no discard prompt queued")
	}

	r := strictRestore(t, g)
	rp := r.playerByIDLocked(p.ID)
	c := discardPromptFor(r, p.ID)
	if c == nil {
		t.Fatal("the restored table has no discard prompt")
	}
	if c.liveZone() != ZoneHand {
		t.Errorf("the restored discard re-checks %q, want the hand", c.liveZone())
	}
	if err := r.ResolveChooseCards(c.ID, p.ID, []uuid.UUID{pick}); err != nil {
		t.Fatalf("answer the restored discard: %v", err)
	}
	if !rp.Graveyard.Contains(pick) || rp.Hand.Contains(pick) {
		t.Error("the restored discard did not put the card into the graveyard")
	}
	if len(r.PendingChoices) != 0 {
		t.Errorf("%d prompts left after the discard", len(r.PendingChoices))
	}
}

// A reveal pick nothing waits on (RevealPickThenForEffect with no key) is a
// restore point: it settles from the prompt's own fields, narrates the pick
// and leaves nothing behind.
func TestS7bARunsRevealPickSurvivesARestore(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner, opp := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, owner.ID, "Fact or Fiction")
	cards := pushFodderForTest(g, owner.ID, 3)
	g.WithWriteLock(func() {
		n, err := g.RevealPickThenForEffect(RevealPickPrompt{
			Chooser: opp.ID, Owner: owner.ID, Source: source,
			Question: "Separate the cards", Cards: cards, Min: 0, Max: 3,
		})
		if err != nil || n != 1 {
			t.Fatalf("RevealPickThenForEffect: %d prompts, %v", n, err)
		}
	})
	c := pendingChoiceOfKind(g, PendingChoiceRevealPick)
	if c == nil || c.PickThen != "" {
		t.Fatalf("the reveal pick is %+v, want a pick with no key", c)
	}

	r := strictRestore(t, g)
	rc := pendingChoiceOfKind(r, PendingChoiceRevealPick)
	if rc == nil || rc.promptRun != c.promptRun {
		t.Fatalf("the restored reveal pick is %+v, want its run link %v", rc, c.promptRun)
	}
	before := countEvents(r, EventCardsChosen)
	if err := r.ResolveRevealPick(rc.ID, opp.ID, cards[:2]); err != nil {
		t.Fatalf("answer the restored reveal pick: %v", err)
	}
	if got := countEvents(r, EventCardsChosen) - before; got != 1 {
		t.Errorf("%d cards_chosen events, want the leg to narrate its pick once", got)
	}
	if len(r.PendingChoices) != 0 || len(r.promptRuns) != 0 {
		t.Errorf("%d prompts and %d runs left, want the leg settled and the run gone", len(r.PendingChoices), len(r.promptRuns))
	}
}

var (
	// s7bPermanentPicks is what the permanent-pick run below was handed.
	s7bPermanentPicks PromptedPicks
	// s7bExactlyOneArtifact is a keyed permanent-pick run whose set rule
	// wants exactly one artifact.
	s7bExactlyOneArtifact = RegisterPromptRunThenRule("test/4361-s7b-exactly-one-artifact",
		func(picked []Card) bool { return len(picked) == 1 && picked[0].IsArtifact() },
		func(_ *Game, r RunDone) error {
			s7bPermanentPicks = r.Picks()
			return nil
		})
)

// A keyed permanent pick's leg holds no frame: its battlefield re-check is
// its kind's and its set rule is its run's, so both survive a restore.
func TestS7bAPermanentPickKeepsItsZoneAndRuleAcrossARestore(t *testing.T) {
	s7bPermanentPicks = nil
	g := newFourPlayerActiveGame(t)
	chooser, victim := g.Seats[0], g.Seats[1]
	fodder := pushFodderForTest(g, victim.ID, 3)
	g.WithWriteLock(func() {
		if _, err := g.PermanentsPickedThenKeyedForEffect(PermanentPickPrompt{
			Chooser: chooser.ID, Question: "choose an artifact",
			Of:         []uuid.UUID{victim.ID},
			Candidates: func(*Game, uuid.UUID) ([]uuid.UUID, int, int) { return fodder, 1, 2 },
		}, s7bExactlyOneArtifact, RunData{}); err != nil {
			t.Fatalf("PermanentsPickedThenKeyedForEffect: %v", err)
		}
	})

	r := strictRestore(t, g)
	c := pendingChoiceOfKind(r, PendingChoiceTheirPermanents)
	if c == nil || c.liveZone() != ZoneBattlefield {
		t.Fatalf("the restored pick is %+v, want its battlefield re-check", c)
	}
	if err := r.ResolveTheirPermanents(c.ID, chooser.ID, fodder[:2]); !errors.Is(err, ErrChoiceSetRejected) {
		t.Fatalf("two artifacts after a restore: %v, want ErrChoiceSetRejected", err)
	}
	if err := r.ResolveTheirPermanents(c.ID, chooser.ID, fodder[:1]); err != nil {
		t.Fatalf("one artifact after a restore: %v", err)
	}
	if by := s7bPermanentPicks.By(victim.ID); len(by) != 1 || by[0] != fodder[0] {
		t.Errorf("the run was handed %v from the victim's board, want %v", by, fodder[:1])
	}
}
