package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// InGraveyardOfTargetsController is "in that player's graveyard",
// where slot names the earlier target permanent controlled by that player.
func InGraveyardOfTargetsController(spec *game.TargetSpec, slot int) *game.TargetSpec {
	spec.Relation = &game.TargetRelation{Slot: slot, Kind: game.TargetInControllersGraveyard}
	return spec
}

// AttachedToTarget is "attached to that creature".
func AttachedToTarget(spec *game.TargetSpec, slot int) *game.TargetSpec {
	spec.Relation = &game.TargetRelation{Slot: slot, Kind: game.TargetAttachedTo}
	return spec
}
