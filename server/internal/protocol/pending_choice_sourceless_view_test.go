package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pending_choice_sourceless_view_test.go is #4148's wire test.
//
// A prompt that names no source player carries FromPlayer =
// uuid.Nil, which viewOfPendingChoices formats as the zero UUID
// "00000000-0000-0000-0000-000000000000", not "". redactChoiceCards
// used to ask `FromPlayer == ""`, so a sourceless prompt never counted as
// the chooser's own pool, and every candidate the chooser was not a
// stamped knower of was dropped from their own picker (the failure shape
// that hid Sakashima's options, #4027).
//
// The audit of every pending-choice kind that inlines a card list on
// the wire (the table below is that list) found no engine site that
// queues one with FromPlayer unset today: each names its chooser, or
// the owner of the pool, and copy_target (the one that does not) lists
// a public zone. These tests pin the contract for the next prompt that
// does not, with both halves of it:
//
//   - own cards: the chooser keeps every candidate they own, even one
//     they are not a knower of, and no other seat and no spectator gets
//     any of them;
//   - a hidden zone that is not theirs: a missing FromPlayer is no
//     licence. The chooser is NOT handed an opponent's library cards
//     they were never made a knower of, as backs or otherwise.

// sourcelessKinds is every kind viewOfPendingChoices projects a card
// list for through redactChoiceCards, minus the ones that always name a
// source (discard_from_hand, revealed_hand_pick, controlled_play read
// FromPlayer's hand).
var sourcelessKinds = []struct {
	kind game.PendingChoiceKind
	// field names where the candidate ids go.
	field string
}{
	{game.PendingChoiceScry, "scry"},
	{game.PendingChoiceSurveil, "scry"},
	{game.PendingChoiceLookAtTop, "scry"},
	{game.PendingChoicePutInLibrary, "scry"},
	{game.PendingChoiceSearchLibrary, "search"},
	{game.PendingChoiceSacrifice, "sacrifice"},
	{game.PendingChoiceChooseCards, "choose"},
	{game.PendingChoiceUntapChoice, "choose"},
	{game.PendingChoiceEntryRevealFromHand, "choose"},
	{game.PendingChoiceEntryDiscardFromHand, "choose"},
	{game.PendingChoiceEntrySacrifice, "choose"},
	{game.PendingChoiceEntryChoosePermanents, "choose"},
	{game.PendingChoiceChooseSource, "choose"},
	{game.PendingChoiceRingBearer, "choose"},
	{game.PendingChoiceProliferate, "choose"},
	{game.PendingChoiceRevealPick, "choose"},
	{game.PendingChoiceTheirPermanents, "choose"},
	{game.PendingChoiceOwnPermanents, "choose"},
}

// queueSourceless queues a prompt of `kind` for `chooser` over `ids`,
// with FromPlayer left unset.
func queueSourceless(g *game.Game, kind game.PendingChoiceKind, field string, chooser uuid.UUID, ids []uuid.UUID) {
	c := game.PendingChoice{
		Kind:    kind,
		Chooser: chooser,
		Count:   1,
		Reason:  "sourceless " + string(kind),
	}
	switch field {
	case "scry":
		c.ScryCards = ids
	case "search":
		c.SearchCards = ids
		c.SearchMax = 1
	case "sacrifice":
		c.SacrificeOptions = ids
	default:
		c.ChooseCards = ids
		c.ChooseMin, c.ChooseMax = 1, 1
	}
	g.QueueChoiceForEffect(c)
}

// topOfLibrary is the ids of the top two cards of p's library, none of
// them marked known to anybody.
func topOfLibrary(g *game.Game, p *game.Player) []uuid.UUID {
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		lib := p.Library.Cards
		for i := 0; i < 2; i++ {
			ids = append(ids, lib[len(lib)-1-i].InstanceID)
		}
	})
	return ids
}

func optionIDs(c PendingChoiceView) map[string]bool {
	out := map[string]bool{}
	for _, o := range c.Options {
		out[o.InstanceID] = true
	}
	return out
}

// TestSourcelessPromptKeepsTheChoosersOwnCandidates is the #4148
// regression: the chooser's own unstamped cards survive a prompt with
// no source player, for every kind that lists cards. Another seat and a
// spectator get none.
func TestSourcelessPromptKeepsTheChoosersOwnCandidates(t *testing.T) {
	for _, k := range sourcelessKinds {
		t.Run(string(k.kind), func(t *testing.T) {
			g := buildThreeSeatGame(t)
			chooser, other := g.Seats[0], g.Seats[1]
			ids := topOfLibrary(g, chooser)
			queueSourceless(g, k.kind, k.field, chooser.ID, ids)

			mine := choiceFor(t, ViewOfGameFor(g, chooser.ID.String()), string(k.kind))
			if mine.FromPlayer != uuid.Nil.String() {
				t.Fatalf("the fixture is not sourceless: from_player = %q", mine.FromPlayer)
			}
			got := optionIDs(mine)
			for _, id := range ids {
				if !got[id.String()] {
					t.Errorf("the chooser lost their own candidate %s: %+v", id, mine.Options)
				}
			}
			for _, o := range mine.Options {
				if o.KnownByYou || o.Name != "" {
					t.Errorf("a card the chooser was never made a knower of was shown face-up: %+v", o)
				}
			}

			for name, viewer := range map[string]string{"another seat": other.ID.String(), "a spectator": SpectatorViewerID} {
				theirs := choiceFor(t, ViewOfGameFor(g, viewer), string(k.kind))
				if len(theirs.Options) != 0 {
					t.Errorf("%s was handed %d of the chooser's candidates: %+v", name, len(theirs.Options), theirs.Options)
				}
			}
		})
	}
}

// TestSourcelessPromptNeverRevealsAnotherSeatsHiddenCards is the other
// half: no FromPlayer must not widen what the chooser may see into a
// hidden zone that is not theirs. The candidates are an opponent's top
// library cards, which the chooser was never made a knower of.
func TestSourcelessPromptNeverRevealsAnotherSeatsHiddenCards(t *testing.T) {
	for _, k := range sourcelessKinds {
		t.Run(string(k.kind), func(t *testing.T) {
			g := buildThreeSeatGame(t)
			chooser, victim, bystander := g.Seats[0], g.Seats[1], g.Seats[2]
			ids := topOfLibrary(g, victim)
			queueSourceless(g, k.kind, k.field, chooser.ID, ids)

			for name, viewer := range map[string]string{
				"the chooser": chooser.ID.String(),
				"the owner":   victim.ID.String(),
				"a bystander": bystander.ID.String(),
				"a spectator": SpectatorViewerID,
			} {
				c := choiceFor(t, ViewOfGameFor(g, viewer), string(k.kind))
				if len(c.Options) != 0 {
					t.Errorf("%s was handed %d cards from another seat's library: %+v", name, len(c.Options), c.Options)
				}
			}
		})
	}
}

// TestSourcelessPromptKeepsWhatTheChooserKnows pins that the owner
// check does not narrow the other rule: a candidate in another seat's
// zone that the chooser WAS made a knower of (a reveal) still reaches
// them, face-up, and only them.
func TestSourcelessPromptKeepsWhatTheChooserKnows(t *testing.T) {
	g := buildThreeSeatGame(t)
	chooser, victim, bystander := g.Seats[0], g.Seats[1], g.Seats[2]
	ids := topOfLibrary(g, victim)
	g.WithWriteLock(func() {
		lib := victim.Library.Cards
		for i := range 2 {
			lib[len(lib)-1-i].AddKnower(chooser.ID)
		}
	})
	queueSourceless(g, game.PendingChoiceChooseCards, "choose", chooser.ID, ids)

	mine := choiceFor(t, ViewOfGameFor(g, chooser.ID.String()), string(game.PendingChoiceChooseCards))
	got := optionIDs(mine)
	for _, id := range ids {
		if !got[id.String()] {
			t.Errorf("the chooser lost a card they were shown: %s", id)
		}
	}
	for _, o := range mine.Options {
		if !o.KnownByYou {
			t.Errorf("a card the chooser knows came back as a back: %+v", o)
		}
	}
	for _, viewer := range []string{bystander.ID.String(), SpectatorViewerID} {
		if c := choiceFor(t, ViewOfGameFor(g, viewer), string(game.PendingChoiceChooseCards)); len(c.Options) != 0 {
			t.Errorf("a viewer who was not shown them got %d of the cards: %+v", len(c.Options), c.Options)
		}
	}
}
