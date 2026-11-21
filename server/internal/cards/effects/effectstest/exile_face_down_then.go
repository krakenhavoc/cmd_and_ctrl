package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4259 (CR 616.1, CR 406.3): a face-down exile that stops on an ordering
// prompt reports nothing until the prompt is answered, so a clause that reads
// "that card" has to run from the answer. Necropotence is that clause.

// orderExileOf registers two mandatory replacements over `card`'s move to
// exile, so CR 616.1 asks the affected player which goes first and the move
// waits for the answer.
func OrderExileOf(g *game.Game, card uuid.UUID) {
	g.WithWriteLock(func() {
		for _, label := range []string{"first order", "second order"} {
			g.RegisterReplacementForTest(game.ReplacementEffect{
				Watches: []game.EventKind{game.EventZoneMove},
				AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
					return ev.Kind == game.RepEventMove && ev.CardID == card && ev.NewZone == game.ZoneExile
				},
				Replace: func(*game.ReplacementEvent, *game.Game, *game.Card) error { return nil },
				Label:   label,
			})
		}
	})
}

// answerExileOrder answers the one open ordering prompt.
func AnswerExileOrder(t *testing.T, g *game.Game) {
	t.Helper()
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("pending choices %+v, want the one ordering prompt", g.PendingChoices)
	}
	prompt := g.PendingChoices[0]
	if err := g.ResolveReplacementOrder(prompt.ID, prompt.Chooser, prompt.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
}

func DeliverAtOwnEndStep(t *testing.T, g *game.Game, me *game.Player, needle uuid.UUID) {
	t.Helper()
	WalkSteps(t, g, 40, func() bool {
		return g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepEnd
	})
	PassPriorityAroundTable(t, g)
	count := 0
	for _, c := range me.Hand.Cards {
		if c.InstanceID == needle {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("the exiled card is in the hand %d times after the end step, want once", count)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers left after delivery", len(g.DelayedTriggers))
	}
}
