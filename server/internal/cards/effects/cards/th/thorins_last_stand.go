package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thorin's Last Stand — Instant {2}{W}{W}:
//
//	"Choose one —
//	 • Creatures you control get +2/+1 until end of turn.
//	 • Destroy target artifact or enchantment. You gain 2 life."
//
// Roar of the Kha's anthem bullet (the creatures are fixed as the spell
// resolves, CR 611.2c) and Pyrrhic Strike's destroy bullet. The second
// bullet's only target is the artifact or enchantment, so when it is gone
// the whole mode does nothing and no life is gained (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c4da169-1e50-4790-abdc-214e7b8a2ad5",
		Name:         "Thorin's Last Stand",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Creatures you control get +2/+1 until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Toughness: 1, Label: "Thorin's Last Stand"}.Apply(ctx)
				}),
			effects.ModeDoing("Destroy target artifact or enchantment. You gain 2 life.",
				effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					if _, ok := effects.ModeTarget(ctx, occ); !ok {
						return nil
					}
					if err := effects.DestroyTheModesTarget(item, ctx, occ); err != nil {
						return err
					}
					return effects.GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
				}),
		),
	})
}
