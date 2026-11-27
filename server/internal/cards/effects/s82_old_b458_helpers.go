package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b458_helpers.go — shared seams of New Phyrexia, batch 458 (#5575).

// S82OldB458TargetCreatureThatPlayerControls is "target creature that player
// controls" for a trigger on a damage event: a creature controlled by the player
// the damage was dealt to (the event's Target), judged when the trigger is put on
// the stack and again as it resolves. With no player named there is nothing to
// target. Mordant Dragon and Arm with Aether share it.
func S82OldB458TargetCreatureThatPlayerControls(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	player := tc.Event.Target
	if player == uuid.Nil {
		return nil
	}
	return &game.TargetSpec{
		Mode: "creature", Label: "target creature that player controls",
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsCreature() && c.Controller == player
		},
		Min: 1, Max: 1,
	}
}
