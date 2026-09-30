package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dead Drop — Sorcery {9}{B}:
//
//	"Delve. Target player sacrifices two creatures of their choice."
//
// Delve (CR 702.66, ADR 0100) is `Delve: true` and nothing else: the
// engine prices it, validates the graveyard cards named on
// `delve_ids` and exiles them at CR 601.2h with the spell on the
// stack.
//
// Lotus Field's shape: ONE choice of two creatures, made by the target
// player, then one simultaneous sacrifice (CR 701.21a). A player with
// fewer than two creatures sacrifices what they have. The choice is not
// a target, so hexproof creatures are fair game. No simplification.
func init() {
	Register(Spec{
		OracleID:     "09f3d60c-34ee-41ec-a047-fe2140b11950",
		Name:         "Dead Drop",
		Completeness: CompletenessFull,
		Delve:        true,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			return ChoosePermanents{
				Player:     item.Targets[0].ID,
				Question:   "Dead Drop — sacrifice two creatures",
				Candidates: deadDropCreatures,
				Sacrifice:  true,
			}.Apply(ctx)
		},
	})
}

// deadDropCreatures offers every creature the chooser controls, exactly
// two to be picked (clamped by the engine when there are fewer).
//
// Caller holds g.mu.
func deadDropCreatures(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == of && c.IsCreature() {
			out = append(out, c.InstanceID)
		}
	}
	return out, 2, 2
}
