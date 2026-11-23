package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prevention_choice_now_integration_test.go — #2040 against the two
// seams that landed beside it: CR 615.13's "damage was prevented"
// announcements (#2047, Selfless Squire), which wait in the same
// follow-up queue a question holds, and #2056's optional redirection
// (Blood of the Martyr), which asks its own question inside a damage
// event.

// squireTriggersLocked is squireTriggers for a caller that already holds
// the game's lock: a damage continuation.

// The damage continuations these tests hear, as registered tails (#4253
// retired DealDamageToPlayerThenForEffect's closure): testSquireSeenTail
// records how many Selfless Squire triggers (its carried ID) are waiting
// when the second instruction runs, testDamageRanTail that it ran at all.
var (
	TestSquireSeen     = -1
	TestSquireSeenTail = game.RegisterTail("test/4253-squire-seen", func(g *game.Game, o game.TailOutcome) error {
		if len(o.Tail.Carry) > 0 {
			TestSquireSeen = squireTriggersLocked(g, o.Tail.Carry[0])
		}
		return nil
	})
	TestDamageTailRan bool
	TestDamageRanTail = game.RegisterTail("test/4253-damage-ran", func(*game.Game, game.TailOutcome) error {
		TestDamageTailRan = true
		return nil
	})
)

func squireTriggersLocked(g *game.Game, squire uuid.UUID) int {
	n := 0
	for _, it := range g.PendingTriggers {
		if it != nil && it.SourceCardID == squire {
			n++
		}
	}
	for _, it := range g.StackMeta {
		if it != nil && it.Kind == game.StackItemTriggered && it.SourceCardID == squire {
			n++
		}
	}
	return n
}

// settleOrderingTriggers passes priority until the stack is empty,
// answering each trigger-order prompt in the order it offers (Khârn's
// Sigil and the Squire's trigger are one controller's, CR 603.3b).
func SettleOrderingTriggers(t *testing.T, g *game.Game) {
	t.Helper()
	for range 8 {
		PassPriorityAroundTable(t, g)
		ch := TriggerOrderPrompt(g)
		if ch == nil {
			return
		}
		if err := g.ResolveTriggerOrder(ch.ID, ch.Chooser, ch.TriggerOrderIDs); err != nil {
			t.Fatal(err)
		}
	}
	t.Fatal("the triggers never settled")
}
