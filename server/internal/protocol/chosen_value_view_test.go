package protocol

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// chosen_value_view_test.go — #781. A colour or creature type chosen
// as a permanent enters is announced at the table (CR 105.4) and is
// what makes the card's other abilities readable at all: CR 607.2d
// links "choose a color" to "creatures you control of the chosen
// color", and nobody can compute that set without the answer. Before
// this the engine stored the answer and told no one — not even the
// player who gave it.
//
// The redaction half (a non-knower gets neither, because "Elf" names
// Cavern of Souls as loudly as loyalty says "planeswalker") is pinned
// by the allowlist table in face_down_view_test.go, which fails if a
// new CardView field escapes it. What is below is the other half: the
// answer reaches EVERY seat, not just the one that gave it.

func TestChosenValuesReachEverySeat(t *testing.T) {
	g := buildActiveGame(t)
	owner, opponent := g.Seats[0], g.Seats[1]
	heartID, automatonID, plainID, inHandID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UnixNano()
	// Battlefield cards are public, but a card pushed straight onto a
	// zone has an empty knower set and FilterViewFor redacts for a
	// non-knower.
	seen := map[uuid.UUID]bool{owner.ID: true, opponent.ID: true}

	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: heartID, Name: "Coldsteel Heart", TypeLine: "Artifact",
			Owner: owner.ID, Controller: owner.ID, EnteredBattlefieldAt: now,
			ChosenColor: "G", KnownBy: seen,
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: automatonID, Name: "Adaptive Automaton", TypeLine: "Artifact Creature — Construct",
			Owner: owner.ID, Controller: owner.ID, EnteredBattlefieldAt: now,
			NamedTribe: "Elf", KnownBy: seen,
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: plainID, Name: "Sol Ring", TypeLine: "Artifact",
			Owner: owner.ID, Controller: owner.ID, EnteredBattlefieldAt: now, KnownBy: seen,
		})
		// CR 614.12 fires on each entry and zone exit clears the
		// answer, so a Cavern in a hand has named nothing. The plain
		// copy in viewOfCard is honest only because the engine keeps
		// that true; this is the assertion that says so.
		owner.Hand.PushTop(game.Card{
			InstanceID: inHandID, Name: "Cavern of Souls", TypeLine: "Land",
			Owner: owner.ID, Controller: owner.ID, KnownBy: seen,
		})
	})

	find := func(v GameView, id uuid.UUID) CardView {
		t.Helper()
		for _, c := range v.Battlefield.Cards {
			if c.InstanceID == id.String() {
				return c
			}
		}
		for _, s := range v.Seats {
			for _, c := range s.Hand.Cards {
				if c.InstanceID == id.String() {
					return c
				}
			}
		}
		t.Fatalf("card %s missing from the view", id)
		return CardView{}
	}

	// Both seats, not just the chooser. An opponent who cannot see
	// which creatures a Heraldic Banner is pumping is the bug.
	for _, seat := range []struct {
		label string
		id    uuid.UUID
	}{{"the controller", owner.ID}, {"an opponent", opponent.ID}} {
		v := ViewOfGameFor(g, seat.id.String())
		if got := find(v, heartID).ChosenColor; got != "G" {
			t.Errorf("%s reads chosen_color %q on Coldsteel Heart, want G", seat.label, got)
		}
		if got := find(v, automatonID).NamedTribe; got != "Elf" {
			t.Errorf("%s reads named_tribe %q on Adaptive Automaton, want Elf", seat.label, got)
		}
		if got := find(v, plainID); got.ChosenColor != "" || got.NamedTribe != "" {
			t.Errorf("%s reads chosen_color %q / named_tribe %q on a permanent that chose nothing, want both absent",
				seat.label, got.ChosenColor, got.NamedTribe)
		}
	}

	// One field each, never both, and never the other card's answer.
	v := ViewOfGameFor(g, owner.ID.String())
	if got := find(v, heartID).NamedTribe; got != "" {
		t.Errorf("Coldsteel Heart projects named_tribe %q, want absent", got)
	}
	if got := find(v, automatonID).ChosenColor; got != "" {
		t.Errorf("Adaptive Automaton projects chosen_color %q, want absent", got)
	}
	if got := find(v, inHandID); got.ChosenColor != "" || got.NamedTribe != "" {
		t.Errorf("a card in hand projects chosen_color %q / named_tribe %q, want both absent", got.ChosenColor, got.NamedTribe)
	}
}

// TestChosenValuesSurviveTheRealEntryChoice is the end-to-end shape:
// the engine's own "as this enters, choose a color" answer, read back
// off the wire rather than off game.Card. viewOfCard copying the field
// is only worth anything if it is the field the choice actually
// writes.
func TestChosenValuesSurviveTheRealEntryChoice(t *testing.T) {
	g := buildActiveGame(t)
	owner, opponent := g.Seats[0], g.Seats[1]
	id := uuid.New()
	seen := map[uuid.UUID]bool{owner.ID: true, opponent.ID: true}

	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: "Coldsteel Heart", TypeLine: "Artifact",
			Owner: owner.ID, Controller: owner.ID,
			EnteredBattlefieldAt: time.Now().UnixNano(), KnownBy: seen,
		})
	})
	if got := g.ChosenColorOf(id); got != "" {
		t.Fatalf("a permanent nobody has answered for reports %q, want empty", got)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].ChosenColor = "R"
			}
		}
	})

	for _, viewer := range []uuid.UUID{owner.ID, opponent.ID} {
		v := ViewOfGameFor(g, viewer.String())
		var got string
		for _, c := range v.Battlefield.Cards {
			if c.InstanceID == id.String() {
				got = c.ChosenColor
			}
		}
		if got != "R" {
			t.Errorf("viewer %s reads chosen_color %q after the choice landed, want R", viewer, got)
		}
	}
}
