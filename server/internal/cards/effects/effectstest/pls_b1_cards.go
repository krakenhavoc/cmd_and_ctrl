package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// plsB1Cast casts a catalog card with a mana cost from the active seat's
// hand, without resolving it.
func PlsB1Cast(t *testing.T, g *game.Game, name, typeLine, oracle, manaCost string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID,
	})
	AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// plsB1Settle resolves the stack, ordering triggers as offered and
// answering every target pick with pickTarget and every own-permanent
// pick with pickOwn. It returns the own-permanent prompts it answered.
func PlsB1Settle(t *testing.T, g *game.Game,
	pickTarget func(c *game.PendingChoice) game.TargetRef,
	pickOwn func(c *game.PendingChoice) []uuid.UUID,
) []game.PendingChoice {
	t.Helper()
	var own []game.PendingChoice
	for i := 0; i < 32; i++ {
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
			case game.PendingChoicePickTarget:
				if pickTarget == nil {
					t.Fatal("an unexpected target pick")
				}
				if err := g.ResolvePickTarget(c.ID, c.Chooser, pickTarget(c)); err != nil {
					t.Fatalf("ResolvePickTarget: %v", err)
				}
			case game.PendingChoiceOwnPermanents:
				if pickOwn == nil {
					t.Fatal("an unexpected own-permanent pick")
				}
				own = append(own, *c)
				if err := g.ResolveOwnPermanents(c.ID, c.Chooser, pickOwn(c)); err != nil {
					t.Fatalf("ResolveOwnPermanents: %v", err)
				}
			default:
				continue
			}
			answered = true
			break
		}
		if answered {
			continue
		}
		if StackFullyEmpty(g) {
			return own
		}
		PassPriorityAroundTable(t, g)
	}
	t.Fatal("the table did not settle")
	return own
}
