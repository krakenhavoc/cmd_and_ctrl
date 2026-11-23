package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b35AnswerOption answers the open option pick with the option labelled `label`.
func B35AnswerOption(t *testing.T, g *game.Game, chooser uuid.UUID, label string) {
	t.Helper()
	c := LatestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatal("no option prompt is open")
	}
	for i, opt := range c.PickOptions {
		if opt.Label == label {
			if err := g.ResolveOptionPick(c.ID, chooser, i); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			return
		}
	}
	t.Fatalf("%q is not offered; options are %+v", label, c.PickOptions)
}

// b35SettleTriggerOrders answers any CR 603.3b ordering prompt as it is
// asked and passes priority until the stack is empty.
func B35SettleTriggerOrders(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	for i := 0; i < 48; i++ {
		if p := TriggerOrderPromptFor(g, who); p != nil {
			order := append([]uuid.UUID(nil), p.TriggerOrderIDs...)
			if err := g.ResolveTriggerOrder(p.ID, who, order); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
}
