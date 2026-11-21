package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// saga_test.go covers the S27 Saga lifecycle end-to-end through a
// real catalog card: the entry lore counter (CR 714.3), the
// precombat-main advance, chapter dispatch in printed order, and the
// CR 704.5s sacrifice — including the half of that rule that is easy
// to get wrong, which is that the final chapter has to RESOLVE first.

const (
	HistoryOfBenaliaOracle  = "c15bb7eb-aaaa-4468-9641-8f706d6137e8"
	FirstIroanGamesOracle   = "58934f6d-1aa2-414c-85c6-955a1e26d675"
	DoublingSeasonSagaOracl = "01546b7d-a233-4176-8843-d732074dc5b6"
)

// advanceToPrecombatMainOf walks the turn engine to the named seat's
// precombat main phase — the step whose entry hook performs the
// CR 714.3 lore-counter turn-based action.
// Always advances at least one step, so calling it while already
// parked at the target seat's precombat main walks a full turn cycle
// round to the NEXT one rather than returning immediately. Every
// caller here means "the next advance", and the tests cast their Saga
// during precombat main.
func AdvanceToPrecombatMainOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward precombat main of seat %d: %v", seat, err)
		}
		if g.Turn.Step == game.StepPrecombatMain && g.Turn.ActiveSeat == seat {
			return
		}
	}
	t.Fatalf("never reached precombat main of seat %d", seat)
}

// loreCountersOn returns the lore-counter count on a battlefield
// card, or -1 when the card is no longer on the battlefield.
func LoreCountersOn(g *game.Game, cardID uuid.UUID) int {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == cardID {
			return c.Counters[game.CounterLore]
		}
	}
	return -1
}

func CountBattlefieldByName(g *game.Game, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			n++
		}
	}
	return n
}

func InGraveyardOf(g *game.Game, playerID, cardID uuid.UUID) bool {
	for _, p := range g.Seats {
		if p.ID != playerID {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if c.InstanceID == cardID {
				return true
			}
		}
	}
	return false
}

// passPriorityUntilTriggerOrderPrompt passes priority until a
// CR 603.3b ordering prompt appears (or the stack settles). Distinct
// from passPriorityAroundTable, which fatals when a prompt stops the
// passes.
func PassPriorityUntilTriggerOrderPrompt(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 32; i++ {
		if TriggerOrderPrompt(g) != nil || StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatal("no trigger_order prompt and the stack never settled")
}

func TriggerOrderPrompt(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerOrder {
			return c
		}
	}
	return nil
}

func AnswerTriggerOrderInOfferedOrder(t *testing.T, g *game.Game) {
	t.Helper()
	ch := TriggerOrderPrompt(g)
	if ch == nil {
		t.Fatal("no trigger_order prompt to answer")
	}
	if err := g.ResolveTriggerOrder(ch.ID, ch.Chooser, ch.TriggerOrderIDs); err != nil {
		t.Fatalf("ResolveTriggerOrder: %v", err)
	}
}

// answerAnyPendingTargetPrompts clicks through the pick_target
// prompts a targeted chapter queues, choosing the first legal option.
// The First Iroan Games' chapter II targets a creature you control,
// and its own Human Soldier is the only one on an empty board.
func AnswerAnyPendingTargetPrompts(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8; i++ {
		var ch *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoicePickTarget {
				ch = c
				break
			}
		}
		if ch == nil {
			return
		}
		if len(ch.PickTargetCards) == 0 {
			t.Fatalf("pick_target prompt %q offered no card options", ch.Reason)
		}
		pick := game.TargetRef{Kind: game.TargetCard, ID: ch.PickTargetCards[0]}
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, pick); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
}
