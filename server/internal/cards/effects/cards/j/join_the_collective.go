package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Join the Collective — Sorcery {5}{W}{W} (Star Trek Commander, #4974):
//
//	"Destroy all creatures. Create X 1/1 colorless Borg artifact creature
//	 tokens, where X is the number of nonartifact creatures destroyed this
//	 way."
//
// The batched sweep (DestroyAllMatching): every creature is destroyed
// at once, and the continuation is told the pre-move copies of
// exactly the ones that were (CR 603.10). X counts those that were not
// artifacts as they died, through the layers; an indestructible creature
// was not destroyed and counts for nothing, and a commander its owner
// then moves to the command zone (CR 903.9a) was still destroyed and
// counts. The tokens are the set's printed Borg.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b1a6d794-0e48-4bbe-b3a3-d9957ed1bfbf",
		Name:         "Join the Collective",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Creature(), ThenTail: ctx.ItemTail(joinTheCollectiveBorg)}.Apply(ctx)
		},
	})
}

// joinTheCollectiveBorg is "create X Borg, where X is the number of
// nonartifact creatures destroyed this way". An on-disk key.
var joinTheCollectiveBorg = effects.MassTail("mass/join-the-collective-borg", func(ctx *effects.Context, swept []game.Card, _ int) error {
	n := 0
	for _, c := range swept {
		if !c.IsArtifact() {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("1/1 colorless Borg artifact"), N: n}.Apply(ctx)
})
