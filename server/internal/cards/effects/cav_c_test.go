package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cav_c_test.go — #3172 slice cav-c: attacking grants, tokens that enter
// attacking, copy triggers.

// passAndSettleAttackSplit is passPriorityAroundTable for a trigger that
// makes tokens "tapped and attacking": with more than one thing to
// attack the controller is asked how many go at each (CR 508.4), so the
// settle loop answers every such question with its goal (a bot's and a
// lazy player's answer: all of them at the thing the creature attacked),
// then keeps passing until the stack is empty.
func passAndSettleAttackSplit(t *testing.T, g *game.Game, controller uuid.UUID) {
	t.Helper()
	for i := 0; i < 8; i++ {
		passPriorityAroundTable(t, g)
		c := payAmountPrompt(g, controller)
		if c == nil {
			return
		}
		if err := g.ResolvePayAmount(c.ID, controller, c.PayAmount.Goal); err != nil {
			t.Fatalf("answering the attack split with its goal %d: %v", c.PayAmount.Goal, err)
		}
	}
	t.Fatal("the attack split kept asking")
}
