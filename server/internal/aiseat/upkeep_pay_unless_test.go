package aiseat_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// upkeep_pay_unless_test.go — the bot half of #997.
//
// The fix makes Stasis's and Pact of Negation's upkeep question stop
// the table, and a prompt that stops the table is only as good as the
// seat it is addressed to. If that seat is a bot and the bot cannot
// answer, the halt is a wedge: nobody may act, the runner sleeps, and
// the humans at the table are stuck behind a question no one will ever
// answer (#544, and the whole of sprint S36).
//
// So this drives the real heuristic policy through the real enumerator
// against a real halt, the way counter_unless_paid_test.go does for the
// ward tax. Deliberately NOT behind AISEAT_GAME_TESTS: it plays no
// whole game and runs in milliseconds, and the gate on the whole-game
// tests is exactly why the wedges this guards against reached
// production before.

// upkeepTaxAgainstTheBot queues the "pay or sacrifice" a bot seat owes
// in its own upkeep, and reports whether the or-else ever ran.
func upkeepTaxAgainstTheBot(t *testing.T, g *game.Game, bot *game.Player, declined *int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.QueueUpkeepPayUnlessForEffect(game.UpkeepPayUnlessPrompt{
			Chooser:  bot.ID,
			Source:   uuid.New(),
			Cost:     "{U}",
			Question: "Stasis — pay {U} or sacrifice Stasis?",
			OnDecline: func(*game.Game) error {
				*declined++
				return nil
			},
		}); err != nil {
			t.Fatalf("QueueUpkeepPayUnlessForEffect: %v", err)
		}
	})
}

// TestTheBotSeatAnswersAnUpkeepTaxAndFreesTheTable.
func TestTheBotSeatAnswersAnUpkeepTaxAndFreesTheTable(t *testing.T) {
	g := newSettledTable(t, 19)
	bot := g.Seats[1]
	declined := 0
	upkeepTaxAgainstTheBot(t, g, bot, &declined)

	// The halt is real before the bot touches it.
	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("PassPriority = %v, want ErrChoicePending while the upkeep tax is open", err)
	}

	took := driveChoices(t, g, heuristic.New(), bot.ID, 4)
	if len(took) != 1 {
		t.Errorf("the bot took %d answers, want exactly 1: %v", len(took), took)
	}
	if owesChoice(g, bot.ID) {
		t.Fatalf("the bot still owes a prompt after answering: %s", openChoices(g, bot.ID))
	}
	// Answered, so the table plays on — and the permanent's fate was
	// settled by the answer rather than by the table walking past it.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority after the bot answered: %v", err)
	}
	// The bot has no mana on a freshly settled table, so the enumerator
	// offers it only the decline (it will not offer a "pay" it cannot
	// fund) and the or-else runs. That is the outcome #997 says the
	// table must not be able to skip.
	if declined != 1 {
		t.Errorf("the or-else ran %d times, want 1 — the bot could not pay", declined)
	}
}
