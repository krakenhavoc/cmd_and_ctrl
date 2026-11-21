package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// color_choice_cards_test.go — the catalog cards #742's "choose a
// color" prompt and one-pick-N-mana amount unblocked.

// --- helpers --------------------------------------------------------

// pendingOfKind returns the first open choice of `kind`, or nil.
func PendingOfKind(g *game.Game, kind game.PendingChoiceKind) *game.PendingChoice {
	var out *game.PendingChoice
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == kind {
				out = c
				return
			}
		}
	})
	return out
}

// answerColor answers the open choose_color prompt owed by `chooser`.
func AnswerColor(t *testing.T, g *game.Game, chooser uuid.UUID, color string) *game.PendingChoice {
	t.Helper()
	c := PendingOfKind(g, game.PendingChoiceColor)
	if c == nil {
		t.Fatalf("no choose_color prompt is open (answering %s)", color)
	}
	if c.Chooser != chooser {
		t.Fatalf("the open choose_color prompt belongs to %s, not %s", c.Chooser, chooser)
	}
	if err := g.ResolveColorChoice(c.ID, chooser, color); err != nil {
		t.Fatalf("ResolveColorChoice(%s): %v", color, err)
	}
	return c
}

// pushChosenColorPermanent seeds a catalog permanent and answers its
// "as this enters, choose a color" prompt, the way
// pushNamedTribePermanent does for creature types.
func PushChosenColorPermanent(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine, oracleID, color string) uuid.UUID {
	t.Helper()
	id := PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracleID,
		Owner: owner, Controller: owner,
	})
	g.WithWriteLock(func() { g.QueueColorChoiceForEffect(owner, id, name, nil, game.ColorForMana) })
	AnswerColor(t, g, owner, color)
	return id
}

func Untap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
}
