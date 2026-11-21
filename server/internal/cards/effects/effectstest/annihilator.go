package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pushAnnihilatorCreature puts a ready Eldrazi with the given keyword
// tokens printed on it onto the battlefield.
func PushAnnihilatorCreature(g *game.Game, controller uuid.UUID, keywords ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Eldrazi", TypeLine: "Creature — Eldrazi",
		Power: 5, Toughness: 5, Keywords: keywords,
		Owner: controller, Controller: controller,
	})
}

// pushAnnFodder puts n plain artifacts under owner's control.
func PushAnnFodder(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Fodder", TypeLine: "Artifact",
			Owner: owner, Controller: owner,
		}))
	}
	return out
}

// annOnBattlefield reports whether the permanent is on the battlefield.
func AnnOnBattlefield(g *game.Game, id uuid.UUID) bool {
	_, ok := BattlefieldCard(g, id)
	return ok
}

// annCountOn counts the permanents a player controls.
func AnnCountOn(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller {
			n++
		}
	}
	return n
}

// annSettle resolves the stack, answering trigger order in the order
// offered and every own_permanents prompt with `pick` (nil: the first
// ChooseMax candidates). It records each prompt it answered.
func AnnSettle(t *testing.T, g *game.Game, pick func(c *game.PendingChoice) []uuid.UUID) []*game.PendingChoice {
	t.Helper()
	var asked []*game.PendingChoice
	for i := 0; i < 64; i++ {
		answered := false
		for _, c := range g.PendingChoices {
			if c == nil {
				continue
			}
			switch c.Kind {
			case game.PendingChoiceTriggerOrder:
				if err := g.ResolveTriggerOrder(c.ID, c.Chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
					t.Fatalf("ResolveTriggerOrder: %v", err)
				}
				answered = true
			case game.PendingChoiceOwnPermanents:
				cp := *c
				asked = append(asked, &cp)
				var ids []uuid.UUID
				if pick != nil {
					ids = pick(c)
				} else {
					ids = append(ids, c.ChooseCards[:c.ChooseMax]...)
				}
				if err := g.ResolveOwnPermanents(c.ID, c.Chooser, ids); err != nil {
					t.Fatalf("ResolveOwnPermanents: %v", err)
				}
				answered = true
			}
			if answered {
				break
			}
		}
		if answered {
			continue
		}
		if StackFullyEmpty(g) {
			return asked
		}
		PassPriorityAroundTable(t, g)
	}
	t.Fatal("the stack did not settle")
	return asked
}
