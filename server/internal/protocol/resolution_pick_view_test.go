package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_pick_view_test.go — #1214 on the wire. What each seat sees
// of the three resolution-time picks, and what the log says afterwards.
//
// The interesting one is reveal_pick, because it is the kind whose
// whole justification is a visibility difference: a choose_cards prompt
// withholds its candidates AND its bounds from every non-chooser,
// because its pool is usually a hand. A reveal_pick's pool is public by
// construction — the card revealed it — so hiding it would hide from
// the table a fact it watched happen.

// seedRevealPick queues a reveal pick addressed to `chooser` over the
// top `n` cards of `owner`'s library, optionally revealed first.
func seedRevealPick(t *testing.T, g *game.Game, chooser, owner *game.Player, n int, reveal bool) []uuid.UUID {
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
		if _, err := g.RevealPickThenForEffect(game.RevealPickPrompt{
			Chooser:  chooser.ID,
			Owner:    owner.ID,
			Question: "test — choose one of those cards",
			Cards:    ids,
			Min:      1,
			Max:      1,
		}, func(*game.Game, []uuid.UUID, []uuid.UUID) error { return nil }); err != nil {
			t.Fatalf("RevealPickThenForEffect: %v", err)
		}
	})
	return ids
}

// TestRevealPickShowsTheChooserExactlyWhatWasRevealed — the chooser is
// asked about cards out of a library that is not theirs, and sees
// precisely what the reveal made public.
func TestRevealPickShowsTheChooserExactlyWhatWasRevealed(t *testing.T) {
	g := buildThreeSeatGame(t)
	owner, chooser := g.Seats[0], g.Seats[1]
	ids := seedRevealPick(t, g, chooser, owner, 3, true)

	got := FilterViewFor(ViewOfGame(g), chooser.ID.String())
	c := choiceFor(t, got, string(game.PendingChoiceRevealPick))
	if len(c.Options) != len(ids) {
		t.Fatalf("the chooser sees %d of %d revealed cards", len(c.Options), len(ids))
	}
	for _, o := range c.Options {
		if !o.KnownByYou || o.Name == "" {
			t.Errorf("a revealed card reaches its chooser face up: %+v", o)
		}
	}
	if c.ChooseMin != 1 || c.ChooseMax != 1 {
		t.Errorf("the chooser is told the bounds: %d..%d", c.ChooseMin, c.ChooseMax)
	}
}

// TestRevealPickHidesAnUnrevealedPoolFromItsChooser — the contract
// stated the other way: a prompt over another seat's cards that were
// never revealed reaches its chooser EMPTY rather than as a handle on
// somebody's library. That is the right failure; an engine that leaked
// the library instead would be the bug.
func TestRevealPickHidesAnUnrevealedPoolFromItsChooser(t *testing.T) {
	g := buildThreeSeatGame(t)
	owner, chooser := g.Seats[0], g.Seats[1]
	seedRevealPick(t, g, chooser, owner, 3, false)

	got := FilterViewFor(ViewOfGame(g), chooser.ID.String())
	c := choiceFor(t, got, string(game.PendingChoiceRevealPick))
	if len(c.Options) != 0 {
		t.Errorf("an unrevealed pool reached its chooser: %d cards", len(c.Options))
	}
}

// TestRevealPickBystanderSeesTheRevealedSet — the other half of the
// kind's justification. A bystander is not the chooser, and a
// choose_cards prompt would show them nothing at all, not even a
// count. A reveal is public, so they see it.
func TestRevealPickBystanderSeesTheRevealedSet(t *testing.T) {
	g := buildThreeSeatGame(t)
	owner, chooser, bystander := g.Seats[0], g.Seats[1], g.Seats[2]
	ids := seedRevealPick(t, g, chooser, owner, 3, true)

	got := FilterViewFor(ViewOfGame(g), bystander.ID.String())
	c := choiceFor(t, got, string(game.PendingChoiceRevealPick))
	if c.Chooser != chooser.ID.String() {
		t.Errorf("the bystander is told whose question it is: %s", c.Chooser)
	}
	if len(c.Options) != len(ids) {
		t.Errorf("a bystander sees %d of %d revealed cards", len(c.Options), len(ids))
	}
	if c.ChooseMax != 1 {
		t.Errorf("the bounds over a public set are public: %d..%d", c.ChooseMin, c.ChooseMax)
	}
}

// TestPermanentPicksArePublicToEverySeat — both permanent picks are
// over battlefield cards, which every seat can already see, so neither
// the candidates nor the bounds are redacted for anybody.
func TestPermanentPicksArePublicToEverySeat(t *testing.T) {
	g := buildThreeSeatGame(t)
	chooser, of := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		pushPublicPermanent(g, of.ID, "View Bear", "Creature — Bear")
		if _, err := g.PermanentsPickedThenForEffect(game.PermanentPickPrompt{
			Chooser:  chooser.ID,
			Question: "test — choose one of theirs",
			Of:       []uuid.UUID{of.ID},
			Candidates: func(g *game.Game, who uuid.UUID) ([]uuid.UUID, int, int) {
				var out []uuid.UUID
				for _, c := range g.Battlefield.Cards {
					if c.Controller == who {
						out = append(out, c.InstanceID)
					}
				}
				return out, 1, 1
			},
		}, func(*game.Game, game.PromptedPicks) error { return nil }); err != nil {
			t.Fatalf("PermanentsPickedThenForEffect: %v", err)
		}
	})

	for _, viewer := range g.Seats {
		got := FilterViewFor(ViewOfGame(g), viewer.ID.String())
		c := choiceFor(t, got, string(game.PendingChoiceTheirPermanents))
		if len(c.Options) == 0 {
			t.Errorf("seat %s sees no candidates on a battlefield prompt", viewer.Name)
			continue
		}
		if c.ChooseMin != 1 || c.ChooseMax != 1 {
			t.Errorf("seat %s is not told the bounds: %d..%d", viewer.Name, c.ChooseMin, c.ChooseMax)
		}
		if c.FromPlayer != of.ID.String() {
			t.Errorf("seat %s is not told whose board it is: %s", viewer.Name, c.FromPlayer)
		}
	}
}

// TestRevealPickAnswerIsNarrated — #1023's gate. The answer reaches the
// public log with the chooser's name, the card they chose, and the card
// that asked.
func TestRevealPickAnswerIsNarrated(t *testing.T) {
	g := buildThreeSeatGame(t)
	owner, chooser := g.Seats[0], g.Seats[1]
	var source uuid.UUID
	g.WithWriteLock(func() {
		source = pushPublicPermanent(g, owner.ID, "Intuition Stand-in", "Enchantment")
	})
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		ids = g.RevealTopOfLibraryForEffect(owner.ID, source, 3, "test — reveal")
		if _, err := g.RevealPickThenForEffect(game.RevealPickPrompt{
			Chooser: chooser.ID,
			Owner:   owner.ID,
			Source:  source,
			Cards:   ids,
			Min:     1,
			Max:     1,
		}, func(*game.Game, []uuid.UUID, []uuid.UUID) error { return nil }); err != nil {
			t.Fatalf("RevealPickThenForEffect: %v", err)
		}
	})
	var choiceID uuid.UUID
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceRevealPick {
			choiceID = c.ID
		}
	}
	if err := g.ResolveRevealPick(choiceID, chooser.ID, ids[:1]); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}

	v := FilterViewFor(ViewOfGame(g), chooser.ID.String())
	line := lastLogOfKind(t, v, LogChooseCards)
	if !strings.Contains(line.Text, chooser.Name) {
		t.Errorf("the line does not name the chooser: %q", line.Text)
	}
	if !strings.Contains(line.Text, "Intuition Stand-in") {
		t.Errorf("the line does not name the card that asked: %q", line.Text)
	}
	if line.Amount != 1 {
		t.Errorf("the line carries the count: %d", line.Amount)
	}

	// A seat that was never made a knower of the chosen card reads the
	// same line without its name. Here the reveal made everyone a
	// knower, so this pins the plural wording instead: a pick of more
	// than one names no card at all.
	if strings.Contains(line.Text, "cards") {
		t.Errorf("a one-card pick is singular: %q", line.Text)
	}
}

// pushPublicPermanent puts a permanent on the battlefield with every
// seat marked a knower, which is what a real battlefield entry does
// (markKnownInZoneLocked) and what a raw PushTop does not.
func pushPublicPermanent(g *game.Game, controller uuid.UUID, name, typeLine string) uuid.UUID {
	known := make(map[uuid.UUID]bool, len(g.Seats))
	for _, p := range g.Seats {
		known[p.ID] = true
	}
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		Owner:      controller,
		Controller: controller,
		KnownBy:    known,
	})
	return id
}

// lastLogOfKind returns the newest public log entry of a kind.
func lastLogOfKind(t *testing.T, v GameView, kind LogKind) LogEvent {
	t.Helper()
	for i := len(v.Log) - 1; i >= 0; i-- {
		if v.Log[i].Kind == kind {
			return v.Log[i]
		}
	}
	t.Fatalf("no %q entry in the log: %+v", kind, v.Log)
	return LogEvent{}
}
