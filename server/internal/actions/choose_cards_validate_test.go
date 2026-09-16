package actions

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestDispatchResolveChoiceChooseCardsRejectsASetTheRuleRefuses — #624
// at the wire's front door. A resolve_choice whose card_ids break a
// choose-cards prompt's set-level Validate comes back from Dispatch as
// game.ErrChoiceSetRejected (which the ws classifier turns into a
// player-facing sentence), the prompt stays queued so the client's modal
// stays open, and the continuation does not run. A set the rule
// accepts, sent the same way, resolves.
func TestDispatchResolveChoiceChooseCardsRejectsASetTheRuleRefuses(t *testing.T) {
	g := newGame(t)
	p := g.Seats[0]
	if p.Hand.Size() < 2 {
		t.Fatalf("setup: want 2 cards in hand, have %d", p.Hand.Size())
	}
	var land, bear uuid.UUID
	ran := false
	g.WithWriteLock(func() {
		p.Hand.Cards[0].TypeLine = "Land"
		p.Hand.Cards[1].TypeLine = "Creature — Bear"
		land, bear = p.Hand.Cards[0].InstanceID, p.Hand.Cards[1].InstanceID
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  p.ID,
			Question: "Discard two cards unless you discard a creature card",
			Cards:    []uuid.UUID{land, bear},
			Min:      1,
			Max:      2,
			Zone:     game.ZoneHand,
			Validate: func(picked []game.Card) bool {
				return len(picked) == 2 || picked[0].IsCreature()
			},
			Then: func(*game.Game, []uuid.UUID) error {
				ran = true
				return nil
			},
		})
	})
	choice := pendingChoiceOfKind(g, game.PendingChoiceChooseCards)
	if choice == nil {
		t.Fatal("no choose-cards prompt queued")
	}

	send := func(ids ...uuid.UUID) error {
		t.Helper()
		cardIDs := make([]string, 0, len(ids))
		for _, id := range ids {
			cardIDs = append(cardIDs, id.String())
		}
		a, err := Decode(string(TypeResolveChoice), p.ID.String(), params(t, map[string]any{
			"choice_id": choice.ID.String(),
			"card_ids":  cardIDs,
		}))
		if err != nil {
			t.Fatalf("Decode: %v", err)
		}
		return Dispatch(g, a)
	}

	if err := send(land); !errors.Is(err, game.ErrChoiceSetRejected) {
		t.Fatalf("a lone land: Dispatch error = %v, want game.ErrChoiceSetRejected", err)
	}
	if pendingChoiceOfKind(g, game.PendingChoiceChooseCards) == nil {
		t.Fatal("the rejected answer dropped the prompt; the client's modal has nothing to retry")
	}
	if ran {
		t.Fatal("the continuation ran for a rejected set")
	}

	if err := send(bear); err != nil {
		t.Fatalf("a lone creature: Dispatch error = %v", err)
	}
	if !ran {
		t.Error("the accepted set did not run the continuation")
	}
	if pendingChoiceOfKind(g, game.PendingChoiceChooseCards) != nil {
		t.Error("the accepted set left the prompt queued")
	}
}
