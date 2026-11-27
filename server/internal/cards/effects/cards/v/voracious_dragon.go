package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voracious Dragon — Creature — Dragon {3}{R}{R}, 4/4 (Conflux):
//
//	"Flying
//	 Devour 1 (As this creature enters, you may sacrifice any number of
//	 creatures. It enters with that many +1/+1 counters on it.)
//	 When this creature enters, it deals damage to any target equal to twice
//	 the number of Goblins it devoured."
//
// Devour 1 is the shared as-enters replacement (devour.go, CR 702.82a), told
// to tally the devoured creatures that were Goblins (EntryCardChoice.
// DevourTally, #5587); the tally is stamped on the permanent as it enters
// (Card.DevouredTally) and read when the enters trigger goes on the stack, so
// a Dragon that has left by the time it resolves still deals the damage its
// count asks for (CR 608.2h). The damage is dealt by the Dragon (its
// last-known self), to the target chosen as the trigger goes on the stack,
// and nothing happens when no Goblin was devoured.
//
// No simplification.
func init() {
	const label = "Voracious Dragon — it deals damage to any target equal to twice the number of Goblins it devoured"
	enters := effects.Targeting(effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		n := item.Params.Amount
		if n <= 0 {
			return nil
		}
		t, ok := ctx.ClauseTarget(0)
		if !ok {
			return nil
		}
		return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
	}), effects.TargetAny())
	enters.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = 2 * g.DevouredTallyBy(source.InstanceID)
		return item
	}
	effects.Register(effects.Spec{
		OracleID:        "7f39bbae-6c84-4cb3-9a58-d746a723ad35",
		Name:            "Voracious Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{effects.OldB507DevourTallyingGoblins("Voracious Dragon")},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
