package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Structural Assault — Sorcery {3}{R}{R} (Streets of New Capenna, #5491):
//
//	"Destroy all artifacts, then Structural Assault deals damage to each
//	 creature equal to the number of artifacts that were put into graveyards
//	 from the battlefield this turn."
//
// The count is the whole turn's, any controller, tokens included: artifacts
// that went to a graveyard before this spell resolved count along with the
// ones it destroyed, and an indestructible artifact that survived does not
// (s82SncB14ArtifactsPutIntoGraveyardsThisTurn, read off the turn's
// leave-the-battlefield events with each permanent's types as it last
// existed, CR 603.10a). The destruction is its own step, so a prompt it asks
// (a regeneration or replacement order) is answered before the count is
// taken. With a count of 0 nothing is damaged.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "23cf22fe-483f-4841-882e-e1513bb1d495",
		Name:         "Structural Assault",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifacts, How: game.SweepDestroy}},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DestroyAllMatching{Match: effects.Artifact()}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DamageEachMatching(ctx, effects.Creature(), effects.S82SncB14ArtifactsPutIntoGraveyardsThisTurn(ctx.Game))
			}),
		),
	})
}
