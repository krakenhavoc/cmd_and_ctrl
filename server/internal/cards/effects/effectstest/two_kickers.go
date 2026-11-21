package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// settleAnsweringPicks resolves everything on the stack, answering
// each of `me`'s pick_target prompts on the way: one that offers
// players is aimed at `victim`, one that offers only cards takes the
// first. Two enters abilities that trigger together are ordered as
// offered (CR 603.3b). Returns how many target prompts were answered —
// one per targeted trigger that fired.
func SettleAnsweringPicks(t *testing.T, g *game.Game, me, victim uuid.UUID) int {
	t.Helper()
	asked := 0
	for i := 0; i < 16; i++ {
		PassPriorityAroundTable(t, g)
		if c := OpenTriggerOrder(g, me); c != nil {
			if err := g.ResolveTriggerOrder(c.ID, me, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		p := LatestPickTarget(g, me)
		if p == nil {
			if StackFullyEmpty(g) {
				return asked
			}
			continue
		}
		pick := game.TargetRef{Kind: game.TargetPlayer, ID: victim}
		if len(p.PickTargetPlayers) == 0 {
			if len(p.PickTargetCards) == 0 {
				t.Fatalf("pick_target prompt %q offers nothing", p.Reason)
			}
			pick = game.TargetRef{Kind: game.TargetCard, ID: p.PickTargetCards[0]}
		}
		if err := g.ResolvePickTarget(p.ID, me, pick); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
		asked++
	}
	t.Fatal("the stack did not settle")
	return asked
}

func OpenTriggerOrder(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerOrder && c.Chooser == chooser {
			return c
		}
	}
	return nil
}
