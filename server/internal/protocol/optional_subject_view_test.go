package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// optional_subject_view_test.go — #4877: a trigger_prompt about one card
// carries that card as `subject`, and only while every seat may see it.

func subjectPrompt(g *game.Game, chooser, subject uuid.UUID) {
	g.PendingChoices = append(g.PendingChoices, &game.PendingChoice{
		ID: uuid.New(), Kind: game.PendingChoiceTriggerPrompt, Chooser: chooser, Count: 1,
		Reason:       "Lazav, Dimir Mastermind — become a copy of that card?",
		OptionalItem: uuid.New(), OptionalSubject: subject,
	})
}

func onlyPrompt(t *testing.T, v GameView) PendingChoiceView {
	t.Helper()
	if len(v.PendingChoices) != 1 {
		t.Fatalf("got %d pending choices, want 1", len(v.PendingChoices))
	}
	return v.PendingChoices[0]
}

// TestOptionalPromptCarriesAPublicSubject — a creature card in a
// graveyard is public: the asked player and everyone else see it.
func TestOptionalPromptCarriesAPublicSubject(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card := uuid.New()
	g.WithWriteLock(func() {
		opp.Graveyard.PushTop(game.Card{InstanceID: card, Name: "Craterhoof Behemoth",
			TypeLine: "Creature — Beast", Owner: opp.ID, Controller: opp.ID})
		subjectPrompt(g, me.ID, card)
	})
	full := ViewOfGame(g)
	for _, viewer := range []string{me.ID.String(), opp.ID.String(), SpectatorViewerID} {
		p := onlyPrompt(t, FilterViewFor(full, viewer))
		if p.Subject == nil || p.Subject.InstanceID != card.String() || p.Subject.Name != "Craterhoof Behemoth" {
			t.Errorf("viewer %s: subject = %+v, want the Craterhoof", viewer, p.Subject)
		}
	}
}

// TestOptionalPromptHidesAHiddenSubject — a card in a hand or a library,
// or a face-down permanent, is not shown: the prompt reads as one with
// no subject, for every viewer, the asked player included.
func TestOptionalPromptHidesAHiddenSubject(t *testing.T) {
	for _, where := range []string{"hand", "library", "face-down"} {
		t.Run(where, func(t *testing.T) {
			g := buildActiveGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			card := uuid.New()
			c := game.Card{InstanceID: card, Name: "Craterhoof Behemoth",
				TypeLine: "Creature — Beast", Owner: opp.ID, Controller: opp.ID}
			g.WithWriteLock(func() {
				switch where {
				case "hand":
					opp.Hand.PushTop(c)
				case "library":
					opp.Library.PushTop(c)
				case "face-down":
					c.FaceDown = true
					g.Battlefield.PushTop(c)
				}
				subjectPrompt(g, me.ID, card)
			})
			full := ViewOfGame(g)
			for _, viewer := range []string{"", me.ID.String(), opp.ID.String()} {
				if p := onlyPrompt(t, FilterViewFor(full, viewer)); p.Subject != nil {
					t.Errorf("viewer %q: a %s card reached the wire as the subject: %+v", viewer, where, p.Subject)
				}
			}
		})
	}
}

// TestOptionalPromptWithoutASubjectIsUnchanged — a prompt that names no
// subject carries none.
func TestOptionalPromptWithoutASubjectIsUnchanged(t *testing.T) {
	g := buildActiveGame(t)
	g.WithWriteLock(func() { subjectPrompt(g, g.Seats[0].ID, uuid.Nil) })
	if p := onlyPrompt(t, ViewOfGame(g)); p.Subject != nil {
		t.Errorf("subject = %+v, want none", p.Subject)
	}
}
