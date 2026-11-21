package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Predatory Rampage — Sorcery {3}{G}{G} (Commander Masters, #5948):
//
//	"Creatures you control get +3/+3 until end of turn. Each creature your
//	 opponents control blocks this turn if able."
//
// Overrun's pump over the creatures you control as it resolves (CR 611.2c),
// then Culling Mark's block requirement on each creature your opponents
// control as it resolves: only those creatures, not ones that arrive later.
// A tapped creature, or one that can't block, is not forced (CR 509.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee6d271c-0728-4080-85bb-a6af05f8b213",
		Name:         "Predatory Rampage",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 3, Toughness: 3, Label: "Predatory Rampage — +3/+3"}).Apply(ctx); err != nil {
				return err
			}
			for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.OpponentControls())) {
				if err := (effects.BlockRequirementUntilEOT{
					Target: c.InstanceID,
					Kind:   game.BlockRequirementBlocks,
					Label:  "Predatory Rampage — blocks this turn if able",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
