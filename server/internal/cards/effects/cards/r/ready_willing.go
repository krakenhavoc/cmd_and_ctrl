package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ready // Willing — split card, oracle 44c7c6bb-2592-4d7c-8485-1815093e70be
// (Dragon's Maze, S82 OLD batch 402, #5564):
//
//	Ready   — Instant {1}{G}{W}: "Creatures you control gain indestructible until
//	          end of turn. Untap each creature you control."
//	Willing — Instant {1}{W}{B}: "Creatures you control gain deathtouch and
//	          lifelink until end of turn."
//	          Fuse (You may cast one or both halves of this card from your hand.)
//
// Both halves are ADR 0034 faces (Ready under the bare oracle ID, Willing under
// "#1"); the fused cast is the engine's (CR 702.102). Each grant is one layer-6
// effect over the creatures you control as it resolves, a fixed set (CR 611.2c),
// until cleanup; Ready then untaps each of them.
//
// No simplification.
func init() {
	mine := effects.And(effects.Creature(), effects.YouControl())
	effects.Register(effects.Spec{
		OracleID:     "44c7c6bb-2592-4d7c-8485-1815093e70be",
		Name:         "Ready",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.GrantKeywordUntilEOT{Match: mine, Keywords: []string{"indestructible"}, Label: "Ready — indestructible"}).Apply(ctx); err != nil {
				return err
			}
			for _, c := range effects.MatchingBattlefield(ctx, mine) {
				if err := (effects.UntapTarget{Target: c.InstanceID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "44c7c6bb-2592-4d7c-8485-1815093e70be#1",
		Name:         "Willing",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(effects.GrantKeywordUntilEOT{
			Match: mine, Keywords: []string{"deathtouch", "lifelink"}, Label: "Willing — deathtouch and lifelink",
		}),
	})
}
