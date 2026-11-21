package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ee6d271c-0728-4080-85bb-a6af05f8b213",
		Name:         "Predatory Rampage",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 3, Toughness: 3, Label: "Predatory Rampage — +3/+3"}).Apply(ctx); err != nil {
				return err
			}
			for _, c := range MatchingBattlefield(ctx, And(Creature(), OpponentControls())) {
				if err := (BlockRequirementUntilEOT{
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
