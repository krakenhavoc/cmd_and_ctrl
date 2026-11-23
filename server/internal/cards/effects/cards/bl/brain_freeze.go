package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brain Freeze — Instant {1}{U}:
//
//	"Target player mills three cards.
//	 Storm (When you cast this spell, copy it for each spell cast
//	 before it this turn. You may choose new targets for the copies.)"
//
// Storm on a mill, and the reason the storm count is TABLE-WIDE
// rather than per player (CR 702.40a, "each other spell"): Brain
// Freeze is cast in response to the opponent's storm turn and counts
// THEIR spells. A per-player reading would make this card a blank,
// which is the sharpest test there is of the count being right.
//
// Each copy re-picks its target, so a storm count of three can mill
// four different players three cards each rather than twelve off one
// library.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "464c0150-3dbc-403b-9ada-fef25ab1f29d",
		Name:         "Brain Freeze",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Triggered:    []game.TriggeredAbility{effects.Storm()},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			return effects.MillCards{Player: item.Targets[0].ID, N: 3}.Apply(ctx)
		},
	})
}
