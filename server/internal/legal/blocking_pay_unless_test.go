package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// blocking_pay_unless_test.go — #567, rebuilt on #997's door. A
// pay_unless prompt can stop the table even though its KIND does not:
// the upkeep pay-or-else (cumulative upkeep's "sacrifice this unless
// you pay", CR 702.24; Stasis; Pact of Negation) is owed before the
// step it was asked in can end.
//
// #794's lesson is that the engine and the enumerator must answer
// "does this stop the table" from ONE function. They now both read
// game.ChoicePromptBlocksTable, and this pins that they agree about a
// per-prompt narrowing as well as about a kind. The prompt is queued to
// a NON-active seat both times, so what is being measured is whether
// the seat holding priority can still act.
func TestPayUnlessBlocksTheTableOnlyWhenTheContractSaysSo(t *testing.T) {
	for _, tc := range []struct {
		name     string
		blocking bool
	}{
		{"an upkeep pay-or-else blocks", true},
		{"a Rhystic tax does not", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newTable(t)
			active, payer := g.Seats[0], g.Seats[1]
			g.WithWriteLock(func() {
				var err error
				if tc.blocking {
					err = g.QueueUpkeepPayUnlessForEffect(game.UpkeepPayUnlessPrompt{
						Chooser:   payer.ID,
						Source:    payer.ID,
						Cost:      "{1}",
						Question:  "pay {1}?",
						OnDecline: func(*game.Game) error { return nil },
					})
				} else {
					err = g.QueuePayUnlessForEffect(payer.ID, payer.ID, "{1}", "pay {1}?",
						func(*game.Game) error { return nil })
				}
				if err != nil {
					t.Fatalf("queue: %v", err)
				}
			})

			// The payer is asked their own question either way, and
			// only that — a seat owing a prompt is offered nothing
			// else, blocking or not (#794).
			mine := legal.EnumerateFor(g, payer.ID)
			if len(mine) == 0 {
				t.Fatal("the seat owing the prompt was offered nothing")
			}
			for _, m := range mine {
				if m.Type != legal.TypeResolveChoice {
					t.Errorf("offered %q (%s) while owing a pay_unless", m.Label, m.Type)
				}
			}
			dispatchAll(t, g, payer.ID, mine)

			// The seat holding priority is the one the block is about.
			moves := legal.EnumerateFor(g, active.ID)
			if tc.blocking && len(moves) != 0 {
				t.Errorf("the active seat was offered %v while a blocking pay_unless is open", labels(moves))
			}
			if !tc.blocking && len(moves) == 0 {
				t.Error("an ordinary pay_unless stopped the active seat — ADR 0018 §6 says it must not")
			}
		})
	}
}
