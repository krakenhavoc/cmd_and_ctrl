package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harvester Troll — Creature — Troll {3}{G}, 2/3 (Oath of the Gatewatch,
// #5545):
//
//	"When this creature enters, you may sacrifice a creature or land. If you
//	 do, put two +1/+1 counters on this creature."
//
// The sacrifice is chosen as the trigger resolves (Provisions Merchant's
// MaySacrificeOne), over the creatures and lands its controller has then,
// the Troll itself included; with nothing to sacrifice there is no question.
// The counters go on only for a permanent that really left, and not at all if
// the Troll sacrificed itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "28c44f88-1d47-4894-86ce-4089bd761360",
		Name:         "Harvester Troll",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Harvester Troll — you may sacrifice a creature or land; if you do, put two +1/+1 counters on it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Harvester Troll — sacrifice a creature or land to put two +1/+1 counters on it?",
						effects.Or(effects.Creature(), effects.Land()), false, harvesterTrollCounters)
				}),
		},
	})
}

// harvesterTrollCounters is the "if you do": two +1/+1 counters on the Troll,
// if it is still on the battlefield. An on-disk key: never renamed, never
// reused.
var harvesterTrollCounters = effects.RegisterSacrificedOne("prompt-run/harvester-troll-sacrifice-counters", func(ctx *effects.Context, _ uuid.UUID) error {
	if !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
})
