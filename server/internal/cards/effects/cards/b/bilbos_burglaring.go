package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bilbo's Burglaring — Sorcery {4}{U}{U} (The Hobbit Commander):
//
//	"For each opponent, gain control of up to one target artifact that
//	 player controls."
//
// Blatant Thievery's per-opponent shape with an "up to one" floor of zero
// and a narrower noun: one clause per opponent seat in turn order, each
// bound to that opponent (perOpponentClauses, Celebrate the Mountain-king's
// helper), so the CR 608.2b re-check asks whether THAT player still controls
// the pick. The control change has no stated duration (CR 611.2a), so it
// lasts until the game ends or the artifact stops being the object it was.
// Each pick resolves on its own: a target that left is skipped and the rest
// are still taken.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b98f6ff6-780f-420e-af22-67c9834db2a9",
		Name:         "Bilbo's Burglaring",
		Completeness: effects.CompletenessFull,
		Targets:      effects.PerOpponentClauses(game.MaxPlayers-1, "artifact", effects.Artifact()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.GainControl{
					Target:     t.ID,
					Controller: item.Controller,
					Duration:   game.IndefiniteDuration(),
					Label:      "Bilbo's Burglaring — gain control (no stated duration)",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
