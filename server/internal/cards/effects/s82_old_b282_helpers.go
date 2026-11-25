package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b282_helpers.go — shared pieces of S82 older-set batch 282
// (Shadows over Innistrad: delirium, #5543).

// S82OldB282DamageUpkeepPlayer is "this creature deals N damage to that
// player" on a trigger for an upkeep: the player is the upkeep's player,
// read off the event's actor, and the damage comes from the item's source
// (as it last existed if it has left, CR 608.2h).
func S82OldB282DamageUpkeepPlayer(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		player := ctx.Trigger().Event.Actor
		if player == uuid.Nil {
			return nil
		}
		return DealDamage{Source: item.SourceCardID, Target: player, Amount: n}.Apply(ctx)
	}
}

// S82OldB282CreatureOfUpkeepPlayer is the TargetsFrom clause "target
// creature that player controls" on a trigger for an upkeep: the player is
// the upkeep's player, read off the event's actor.
func S82OldB282CreatureOfUpkeepPlayer(tc game.TriggerContext, _ *game.Card, g *game.Game) *game.TargetSpec {
	player := tc.Event.Actor
	if player == uuid.Nil || g.PlayerByIDForEffect(player) == nil {
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
