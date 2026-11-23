package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blinkmoth Infusion — Instant {12}{U}{U} (Fifth Dawn, #5604):
//
//	"Affinity for artifacts (This spell costs {1} less to cast for each
//	 artifact you control.)
//	 Untap all artifacts."
//
// Affinity for artifacts, and Dramatic Reversal's untap over every tapped
// artifact on the battlefield, whoever controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8788f555-8c18-4842-9716-385afae2b9a4",
		Name:         "Blinkmoth Infusion",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for artifacts", effects.Artifact()),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var tapped []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsArtifact() && c.Tapped {
					tapped = append(tapped, c.InstanceID)
				}
			}
			for _, id := range tapped {
				if err := (effects.UntapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
