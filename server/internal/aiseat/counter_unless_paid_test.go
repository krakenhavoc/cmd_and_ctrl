package aiseat_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_unless_paid_test.go — the bot half of #951.
//
// The fix makes a ward tax stop the table, and a prompt that stops the
// table is only as good as the seat it is addressed to. If that seat is
// a bot and the bot cannot answer, the halt is a wedge: nobody may act,
// the runner sleeps, and the humans at the table are stuck behind a
// question no one will ever answer (#544, and the whole of sprint S36).
//
// So this drives the real heuristic policy through the real enumerator
// against a real halt, the way chained_choice_test.go does for chains.
// Deliberately NOT behind AISEAT_GAME_TESTS: it plays no whole game and
// runs in milliseconds, and the gate on the whole-game tests is exactly
// why the wedges this guards against reached production before.

// wardTaxAgainstTheBot puts a spell the bot controls on the stack and
// queues the ward tax that guards it, addressed to the bot.
func wardTaxAgainstTheBot(t *testing.T, g *game.Game, bot *game.Player) uuid.UUID {
	t.Helper()
	spell := uuid.New()
	g.WithWriteLock(func() {
		g.Stack.PushTop(game.Card{
			InstanceID: spell, Name: "Doom Blade", TypeLine: "Instant",
			Owner: bot.ID, Controller: bot.ID,
		})
		if g.StackMeta == nil {
			g.StackMeta = make(map[uuid.UUID]*game.StackItem)
		}
		g.StackMeta[spell] = &game.StackItem{
			ID: spell, Kind: game.StackItemSpell,
			Controller: bot.ID, Owner: bot.ID, SourceCardID: spell,
		}
		if err := g.QueueCounterUnlessPaidForEffect(game.CounterUnlessPaidPrompt{
			StackItem: spell,
			Source:    uuid.New(),
			Cost:      "{2}",
			Question:  "Ward — pay {2} or the spell is countered",
		}); err != nil {
			t.Fatalf("QueueCounterUnlessPaidForEffect: %v", err)
		}
	})
	return spell
}

// TestTheBotSeatAnswersAWardTaxAndFreesTheTable.
func TestTheBotSeatAnswersAWardTaxAndFreesTheTable(t *testing.T) {
	g := newSettledTable(t, 17)
	bot := g.Seats[1]
	spell := wardTaxAgainstTheBot(t, g, bot)

	// The halt is real before the bot touches it.
	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("PassPriority = %v, want ErrChoicePending while the ward tax is open", err)
	}

	took := driveChoices(t, g, heuristic.New(), bot.ID, 4)
	if len(took) != 1 {
		t.Errorf("the bot took %d answers, want exactly 1: %v", len(took), took)
	}
	if owesChoice(g, bot.ID) {
		t.Fatalf("the bot still owes a prompt after answering: %s", openChoices(g, bot.ID))
	}
	// Answered, so the table plays on — and the spell's fate was
	// settled by the answer rather than by the table walking past it.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority after the bot answered: %v", err)
	}
	// The bot has no mana on a freshly settled table, so the
	// enumerator offers it only the decline (it will not offer a "pay"
	// it cannot fund) and the spell is countered. That is the outcome
	// #951 says the table must not be able to skip.
	if g.Stack.Contains(spell) {
		t.Error("the bot could not pay, so the guarded spell should have been countered")
	}
}
