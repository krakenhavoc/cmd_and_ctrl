package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// face_down_piles_view_test.go — #2147. What each seat's wire carries
// at both steps of "look at the top four cards, separate them into a
// face-down pile and a face-up pile" (Sauron's Ransom).
//
// Seat 0 is the caster and the owner of the library, seat 1 the
// opponent who separates, seat 2 a bystander. The face-down pile is
// the one card the separator hides.

func seedFaceDownSplit(t *testing.T, g *game.Game) (looked []uuid.UUID, splitID uuid.UUID) {
	t.Helper()
	caster, splitter := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		looked = g.LookAtTopOfPlayersLibraryForEffect(splitter.ID, caster.ID, 4)
		g.QueuePileSplitForEffect(game.PileSplitPrompt{
			Splitter:      splitter.ID,
			Chooser:       caster.ID,
			Owner:         caster.ID,
			SplitQuestion: "test — choose the face-down pile",
			PickQuestion:  "test — take a pile",
			Cards:         looked,
			FaceDown:      true,
			Then:          func(*game.Game, []uuid.UUID, []uuid.UUID) error { return nil },
		})
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceRevealPick {
				splitID = c.ID
			}
		}
	})
	if len(looked) != 4 || splitID == uuid.Nil {
		t.Fatalf("setup: looked %d, split %v", len(looked), splitID)
	}
	return looked, splitID
}

func splitOptionIDs(c PendingChoiceView) []string {
	var out []string
	for _, card := range c.Options {
		out = append(out, card.InstanceID)
	}
	return out
}

func TestFaceDownSplitPromptIsTheSeparatorsAlone(t *testing.T) {
	g := buildThreeSeatGame(t)
	looked, _ := seedFaceDownSplit(t, g)

	sep := choiceFor(t, FilterViewFor(ViewOfGame(g), g.Seats[1].ID.String()), string(game.PendingChoiceRevealPick))
	if got := splitOptionIDs(sep); len(got) != 4 {
		t.Fatalf("the separator sees all four cards, got %d", len(got))
	}
	for _, card := range sep.Options {
		if !card.KnownByYou || card.Name == "" {
			t.Errorf("the separator looked at %s, so it has a face: %+v", card.InstanceID, card)
		}
	}
	for _, seat := range []int{0, 2} {
		c := choiceFor(t, FilterViewFor(ViewOfGame(g), g.Seats[seat].ID.String()), string(game.PendingChoiceRevealPick))
		if got := splitOptionIDs(c); len(got) != 0 {
			t.Errorf("seat %d sees %d of the looked-at cards: %v (looked %v)", seat, len(got), got, looked)
		}
	}
}

func TestFaceDownPickShowsEachSeatItsOwnHalf(t *testing.T) {
	g := buildThreeSeatGame(t)
	looked, splitID := seedFaceDownSplit(t, g)
	caster, splitter := g.Seats[0], g.Seats[1]

	// The separator hides the first two as the face-down pile.
	down := looked[:2]
	up := looked[2:]
	if err := g.ResolveRevealPick(splitID, splitter.ID, down); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}

	pileView := func(seat int) (upCards, downCards []CardView, c PendingChoiceView) {
		c = choiceFor(t, FilterViewFor(ViewOfGame(g), g.Seats[seat].ID.String()), string(game.PendingChoiceOptionPick))
		if len(c.PickOptions) != 2 {
			t.Fatalf("seat %d: two piles on the wire, got %d", seat, len(c.PickOptions))
		}
		return c.PickOptions[0].Cards, c.PickOptions[1].Cards, c
	}
	idsOf := func(cs []CardView) map[string]bool {
		m := map[string]bool{}
		for _, c := range cs {
			m[c.InstanceID] = true
		}
		return m
	}

	// The caster (the chooser): the face-up pile in full, the face-down
	// pile as two backs they cannot read.
	upCards, downCards, c := pileView(0)
	if len(upCards) != 2 || len(downCards) != 2 {
		t.Fatalf("the caster sees 2 face-up cards and 2 backs, got %d and %d", len(upCards), len(downCards))
	}
	for _, card := range upCards {
		if !card.KnownByYou || card.Name == "" {
			t.Errorf("the face-up pile is revealed: %+v", card)
		}
	}
	for _, card := range downCards {
		if card.KnownByYou || card.Name != "" {
			t.Errorf("the caster must not read the face-down pile: %+v", card)
		}
	}
	if c.Chooser != caster.ID.String() {
		t.Errorf("the pick is the caster's: %q", c.Chooser)
	}
	if got := c.PickOptions[1].Label; got != "Take the face-down pile (2 cards)" {
		t.Errorf("the face-down pile's size is on its label: %q", got)
	}

	// The separator still reads everything.
	upCards, downCards, _ = pileView(1)
	if len(upCards) != 2 || len(downCards) != 2 {
		t.Fatalf("the separator sees both piles, got %d and %d", len(upCards), len(downCards))
	}
	for _, card := range downCards {
		if !card.KnownByYou || card.Name == "" {
			t.Errorf("the separator hid these, so it knows them: %+v", card)
		}
	}

	// A bystander sees the face-up pile and the face-down pile's size,
	// and not one of its cards, not even as a back with an ID.
	upCards, downCards, c = pileView(2)
	if len(upCards) != 2 {
		t.Errorf("the table sees the face-up pile, got %d", len(upCards))
	}
	if len(downCards) != 0 {
		t.Errorf("a bystander sees %d face-down cards", len(downCards))
	}
	seen := idsOf(upCards)
	for _, id := range up {
		if !seen[id.String()] {
			t.Errorf("face-up card %s is missing from the bystander's view", id)
		}
	}
	for _, id := range down {
		if seen[id.String()] {
			t.Errorf("face-down card %s leaked to the bystander", id)
		}
	}
	if got := c.PickOptions[1].Label; got != "Take the face-down pile (2 cards)" {
		t.Errorf("the bystander sees the size: %q", got)
	}
}

// TestFaceDownPickOverTheCastersCardsHidesThemFromAnOpponentChooser is
// the mirror (Riddles in the Dark): the caster separates their own
// cards and an opponent chooses without seeing the face-down pile.
func TestFaceDownPickOverTheCastersCardsHidesThemFromAnOpponentChooser(t *testing.T) {
	g := buildThreeSeatGame(t)
	caster, chooser := g.Seats[0], g.Seats[1]
	var looked []uuid.UUID
	var splitID uuid.UUID
	g.WithWriteLock(func() {
		looked = g.LookAtTopOfPlayersLibraryForEffect(caster.ID, caster.ID, 4)
		g.QueuePileSplitForEffect(game.PileSplitPrompt{
			Splitter: caster.ID, Chooser: chooser.ID, Owner: caster.ID,
			SplitQuestion: "test — face-down pile", PickQuestion: "test — choose",
			Cards: looked, FaceDown: true,
			Then: func(*game.Game, []uuid.UUID, []uuid.UUID) error { return nil },
		})
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceRevealPick {
				splitID = c.ID
			}
		}
	})
	if err := g.ResolveRevealPick(splitID, caster.ID, looked[:3]); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	c := choiceFor(t, FilterViewFor(ViewOfGame(g), chooser.ID.String()), string(game.PendingChoiceOptionPick))
	if n := len(c.PickOptions[0].Cards); n != 1 {
		t.Errorf("the opponent sees the one face-up card, got %d", n)
	}
	if n := len(c.PickOptions[1].Cards); n != 0 {
		t.Errorf("the opponent sees %d face-down cards", n)
	}
	if got := c.PickOptions[1].Label; got != "Take the face-down pile (3 cards)" {
		t.Errorf("the size is public: %q", got)
	}
	owner := choiceFor(t, FilterViewFor(ViewOfGame(g), caster.ID.String()), string(game.PendingChoiceOptionPick))
	if n := len(owner.PickOptions[1].Cards); n != 3 {
		t.Errorf("the caster separated them and sees all three, got %d", n)
	}
}
