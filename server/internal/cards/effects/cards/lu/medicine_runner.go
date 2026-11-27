package cardslu

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Medicine Runner — Creature — Elf Cleric {1}{G/W}, 2/1 (Shadowmoor, #5591):
//
//	"When this creature enters, you may remove a counter from target
//	 permanent."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d). As it
// resolves the controller picks which counter to take, of whatever kinds
// the permanent carries, or none: the same per-kind question Render Inert
// asks, with a ceiling of one, which is the "you may". A target gone by
// then gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c0a8412-af9f-4a22-b462-e9b2d9bd04c6",
		Name:         "Medicine Runner",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Medicine Runner — you may remove a counter from target permanent",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						return medicineRunnerRemoves.Choose(ctx, t.ID, nil, 1, "Medicine Runner — remove a counter")
					}
					return nil
				}), effects.TargetPermanent("target permanent")),
		},
	})
}

// medicineRunnerRemoves takes the counter chosen off the permanent. A bot
// removes -1/-1 counters from its own permanent and everything else from
// another's. An on-disk key: never renamed, never reused.
var medicineRunnerRemoves = effects.RegisterCounterChoice("number/medicine-runner-remove-counter",
	func(ctx *effects.Context, target uuid.UUID, kind string, most int) int {
		mine := false
		if c, ok := ctx.Game.LookupCardForEffect(target); ok {
			mine = c.Controller == ctx.Controller()
		}
		if (kind == "-1/-1") == mine {
			return most
		}
		return 0
	},
	func(ctx *effects.Context, target uuid.UUID, _ []uuid.UUID, picked map[string]int) error {
		kinds := make([]string, 0, len(picked))
		for kind := range picked {
			kinds = append(kinds, kind)
		}
		sort.Strings(kinds)
		for _, kind := range kinds {
			if err := (effects.AddCounter{Target: target, Kind: kind, N: -picked[kind]}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
