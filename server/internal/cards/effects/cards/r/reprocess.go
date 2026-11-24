package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reprocess — Sorcery {2}{B}{B} (Secret Lair Drop):
//
//	"Sacrifice any number of artifacts, creatures, and/or lands. Draw a card
//	 for each permanent sacrificed this way."
//
// The sacrifice is a choice made as the spell resolves (zero is legal), over
// the caster's artifacts, creatures and lands; the picked permanents leave
// together through the one sacrifice path, and the draw counts what was
// chosen once they have gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27177cb0-1670-488e-82be-4610da5a7de3",
		Name:         "Reprocess",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ChoosePermanents{
				Question: "Reprocess — sacrifice any number of artifacts, creatures, and/or lands",
				Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == of && (c.IsCreature() || c.IsLand() || c.IsArtifact()) {
							ids = append(ids, c.InstanceID)
						}
					}
					if len(ids) == 0 {
						return nil, 0, 0
					}
					return ids, 0, len(ids)
				},
				Keyed: reprocessPicked,
			}.Apply(ctx)
		},
	})
}

var reprocessPicked = effects.RegisterPermanentsPick("prompt-run/reprocess-sacrifice", effects.PermanentsDef{
	Sacrifice: true,
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		return effects.DrawCards{N: picked.Count()}.Apply(ctx)
	},
})
