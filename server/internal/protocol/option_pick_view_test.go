package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// option_pick_view_test.go — #568. The redaction pass on a prompt
// addressed to a seat over cards that are NOT that seat's.
//
// pending_choice_visibility_test.go covers the older rule: a viewer
// who is not the chooser loses the options they are not a knower of.
// The rule this file pins is the one an opponent-facing prompt needed:
// the CHOOSER loses them too, when the pool belongs to somebody else.
//
// Fact or Fiction is why. An opponent separates five cards off the top
// of your library, and the only thing that entitles them to look is
// that the card revealed them first (#549). Without the reveal the
// prompt would be handing one seat five stable instance IDs out of
// another seat's library — PR #513's leak, in a new field.
//
// It also pins the new field itself: an option's pile goes through the
// same pass as PendingChoiceView.Options, because a second card list
// the filter did not know about would reach every seat raw.

// seedOptionPickOverLibrary queues an option pick addressed to
// `chooser` whose two options carry the top `n` cards of `owner`'s
// library, optionally revealing them first. Returns the card IDs,
// top-first.
func seedOptionPickOverLibrary(t *testing.T, g *game.Game, chooser, owner *game.Player, n int, reveal bool) []uuid.UUID {
	t.Helper()
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		if reveal {
			ids = g.RevealTopOfLibraryForEffect(owner.ID, uuid.Nil, n, "test — reveal")
		} else {
			size := owner.Library.Size()
			for i := 0; i < n && i < size; i++ {
				ids = append(ids, owner.Library.Cards[size-1-i].InstanceID)
			}
		}
		if len(ids) < 2 {
			t.Fatalf("need at least two cards, got %d", len(ids))
		}
		g.QueueOptionPickForEffect(game.OptionPickPrompt{
			Chooser:    chooser.ID,
			FromPlayer: owner.ID,
			Question:   "test — take a pile",
			Options: []game.ChoiceOption{
				{Label: "Pile 1", Cards: ids[:1]},
				{Label: "Pile 2", Cards: ids[1:]},
			},
			Then: func(*game.Game, int) error { return nil },
		})
	})
	return ids
}

// TestOptionPickShowsTheChooserOnlyTheRevealedCards is the #568
// redaction rule: the prompt is addressed to a seat that does not own
// the pool, and it shows them exactly what was revealed.
func TestOptionPickShowsTheChooserOnlyTheRevealedCards(t *testing.T) {
	g := buildThreeSeatGame(t)
	owner, chooser := g.Seats[0], g.Seats[1]
	ids := seedOptionPickOverLibrary(t, g, chooser, owner, 3, true)

	got := FilterViewFor(ViewOfGame(g), chooser.ID.String())
	c := choiceFor(t, got, string(game.PendingChoiceOptionPick))
	if len(c.PickOptions) != 2 {
		t.Fatalf("two piles on the wire, got %d", len(c.PickOptions))
	}
	seen := pileCardIDs(c)
	if len(seen) != len(ids) {
		t.Fatalf("the splitter sees the %d revealed cards, got %d: %v", len(ids), len(seen), seen)
	}
	for _, opt := range c.PickOptions {
		for _, card := range opt.Cards {
			if !card.KnownByYou || card.Name == "" {
				t.Errorf("a revealed card keeps its identity: %+v", card)
			}
		}
	}

	// The owner is not the chooser, but the reveal made them a knower
	// too, so they see the piles as well — which is right: the table
	// watched the same five cards.
	ownerView := FilterViewFor(ViewOfGame(g), owner.ID.String())
	if len(pileCardIDs(choiceFor(t, ownerView, string(game.PendingChoiceOptionPick)))) != len(ids) {
		t.Error("a reveal is public — the owner sees the piles too")
	}
	// And so does the third seat, for the same reason.
	watcherView := FilterViewFor(ViewOfGame(g), g.Seats[2].ID.String())
	if len(pileCardIDs(choiceFor(t, watcherView, string(game.PendingChoiceOptionPick)))) != len(ids) {
		t.Error("a reveal is public to every seat, chooser or not")
	}
}

// TestOptionPickHidesAnUnrevealedPoolFromItsChooser is the other half
// of the same rule, and the one that makes it a GUARD rather than a
// coincidence: without the reveal, the seat being asked about somebody
// else's library is shown nothing at all.
//
// The engine therefore also owes the effect a contract — a cross-seat
// prompt must reveal what it asks about — and that is the right
// failure, because the alternative leaks a hidden zone.
func TestOptionPickHidesAnUnrevealedPoolFromItsChooser(t *testing.T) {
	g := buildThreeSeatGame(t)
	owner, chooser := g.Seats[0], g.Seats[1]
	seedOptionPickOverLibrary(t, g, chooser, owner, 3, false)

	got := FilterViewFor(ViewOfGame(g), chooser.ID.String())
	c := choiceFor(t, got, string(game.PendingChoiceOptionPick))
	if seen := pileCardIDs(c); len(seen) != 0 {
		t.Errorf("an unrevealed library is not shown to another seat: %v", seen)
	}
	// The prompt is still there, with its labels — a viewer learns
	// that a question is open and whose it is, and nothing about a
	// zone they cannot see.
	if len(c.PickOptions) != 2 || c.PickOptions[0].Label != "Pile 1" {
		t.Errorf("the option labels survive redaction: %+v", c.PickOptions)
	}
}

// TestOptionPickBystanderSeesThePromptButNoHiddenCards — a seat that
// is neither the chooser nor a knower.
func TestOptionPickBystanderSeesThePromptButNoHiddenCards(t *testing.T) {
	g := buildThreeSeatGame(t)
	owner, chooser, bystander := g.Seats[0], g.Seats[1], g.Seats[2]
	seedOptionPickOverLibrary(t, g, chooser, owner, 3, false)

	got := FilterViewFor(ViewOfGame(g), bystander.ID.String())
	c := choiceFor(t, got, string(game.PendingChoiceOptionPick))
	if c.Chooser != chooser.ID.String() {
		t.Errorf("the prompt names its chooser: %q", c.Chooser)
	}
	if seen := pileCardIDs(c); len(seen) != 0 {
		t.Errorf("a bystander sees no cards from a hidden zone: %v", seen)
	}
}

// TestOptionPickOverYourOwnPoolKeepsBacksForTheChooser — the older
// rule is unchanged where it applies: a chooser picking among their
// OWN cards keeps every option, redacted to a back when they are not
// a knower, because picking a back is still a legal answer.
func TestOptionPickOverYourOwnPoolKeepsBacksForTheChooser(t *testing.T) {
	g := buildThreeSeatGame(t)
	owner := g.Seats[0]
	seedOptionPickOverLibrary(t, g, owner, owner, 3, false)

	got := FilterViewFor(ViewOfGame(g), owner.ID.String())
	c := choiceFor(t, got, string(game.PendingChoiceOptionPick))
	seen := pileCardIDs(c)
	if len(seen) != 3 {
		t.Fatalf("the chooser keeps their own pool: got %d", len(seen))
	}
	for _, opt := range c.PickOptions {
		for _, card := range opt.Cards {
			if card.KnownByYou || card.Name != "" {
				t.Errorf("an unknown card of your own is a back, not a face: %+v", card)
			}
		}
	}
}

// pileCardIDs flattens the card IDs across an option pick's piles.
func pileCardIDs(c PendingChoiceView) []string {
	var out []string
	for _, opt := range c.PickOptions {
		for _, card := range opt.Cards {
			out = append(out, card.InstanceID)
		}
	}
	return out
}

// TestSplitPromptOverAnotherSeatsLibraryShowsOnlyTheRevealedCards is
// rule 2 on the Options path rather than the PickOptions one: Fact or
// Fiction's FIRST prompt is an ordinary choose_cards addressed to the
// splitter over the controller's library, and it is redacted the same
// way.
func TestSplitPromptOverAnotherSeatsLibraryShowsOnlyTheRevealedCards(t *testing.T) {
	for _, tc := range []struct {
		name   string
		reveal bool
		want   int
	}{
		{"revealed", true, 3},
		{"not revealed", false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := buildThreeSeatGame(t)
			owner, splitter := g.Seats[0], g.Seats[1]
			var ids []uuid.UUID
			g.WithWriteLock(func() {
				if tc.reveal {
					ids = g.RevealTopOfLibraryForEffect(owner.ID, uuid.Nil, 3, "test — reveal")
				} else {
					size := owner.Library.Size()
					for i := 0; i < 3; i++ {
						ids = append(ids, owner.Library.Cards[size-1-i].InstanceID)
					}
				}
				g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
					Chooser:    splitter.ID,
					FromPlayer: owner.ID,
					Question:   "test — separate them",
					Cards:      ids,
					Min:        0,
					Max:        len(ids),
					Then:       func(*game.Game, []uuid.UUID) error { return nil },
				})
			})
			got := FilterViewFor(ViewOfGame(g), splitter.ID.String())
			c := choiceFor(t, got, string(game.PendingChoiceChooseCards))
			if len(c.Options) != tc.want {
				t.Errorf("the splitter sees %d cards, want %d", len(c.Options), tc.want)
			}
		})
	}
}

// TestCoerciveDiscardStillHandsTheChooserBacks pins the exception in
// redactChoiceCards, beside the rule it is an exception to: a hand's
// size is public, so a cross-seat discard prompt stays answerable.
func TestCoerciveDiscardStillHandsTheChooserBacks(t *testing.T) {
	g := buildThreeSeatGame(t)
	chooser, victim := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind:       game.PendingChoiceDiscardFromHand,
			Chooser:    chooser.ID,
			FromPlayer: victim.ID,
			Count:      1,
			Reason:     "Coercive discard",
		})
	})
	got := FilterViewFor(ViewOfGame(g), chooser.ID.String())
	c := choiceFor(t, got, string(game.PendingChoiceDiscardFromHand))
	if len(c.Options) != len(victim.Hand.Cards) {
		t.Fatalf("the chooser got %d of %d backs", len(c.Options), len(victim.Hand.Cards))
	}
	for _, o := range c.Options {
		if o.KnownByYou || o.Name != "" {
			t.Errorf("a back, not a face: %+v", o)
		}
	}
}
