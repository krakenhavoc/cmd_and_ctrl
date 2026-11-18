package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kirtar's Wrath — Sorcery {4}{W}{W}:
//
//	"Destroy all creatures. They can't be regenerated.
//	 Threshold — If there are seven or more cards in your graveyard, instead
//	 destroy all creatures, then create two 1/1 white Spirit creature tokens
//	 with flying. Creatures destroyed this way can't be regenerated."
//
// Threshold is an ability word with no rules of its own, read as the spell
// resolves. The wipe is the shared
// simultaneous destroy with the can't-be-regenerated rider; the Spirits are
// made after it, so they are not among the creatures destroyed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f66d82a-492f-4638-9f77-190d4a33ad7f",
		Name:         "Kirtar's Wrath",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := wrathDestroyAllCreaturesNoRegen(item, ctx); err != nil {
				return err
			}
			if !fraThresholdMet(ctx.Game, ctx.Controller()) {
				return nil
			}
			return CreateToken{Template: TokenCard("1/1 white Spirit with flying"), N: 2}.Apply(ctx)
		},
	})
}
