package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crystal Golem — Artifact Creature — Golem {4}, 3/3 (Mirage, #5633):
//
//	"At the beginning of your end step, this creature phases out. (While
//	 it's phased out, it's treated as though it doesn't exist. It phases
//	 in before you untap during your next untap step.)"
//
// The end-step trigger phases the Golem itself out (CR 702.26): no zone
// change, so counters, damage and Auras stay, and it phases back in
// before its controller untaps (CR 502.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "82246ef2-4267-4daf-9463-908c2e142e7d",
		Name:         "Crystal Golem",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Crystal Golem — this creature phases out", func(g *game.Game, item *game.StackItem) error {
				if !g.Battlefield.Contains(item.SourceCardID) {
					return nil
				}
				return g.PhaseOutForEffect(item.SourceCardID, item.SourceCardID)
			}),
		},
	})
}
