package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b576_helpers.go — shared shapes of S82 OLD batch 576 (Ravnica: City
// of Guilds, #5600): Spawnbroker's power-bounded second target.

// PowerNoGreaterThanTarget is "<creature> with power less than or equal to
// that creature's power": the clause's candidates must have a power no greater
// than the earlier target in `slot` (Spawnbroker's "target creature with power
// less than or equal to that creature's power an opponent controls"). Judged
// as the ability is put on the stack (CR 601.2c) and again as it resolves
// (CR 608.2b); the anchor is read for the power it has, wherever it is.
func PowerNoGreaterThanTarget(spec *game.TargetSpec, slot int) *game.TargetSpec {
	spec.Relation = &game.TargetRelation{Slot: slot, Kind: game.TargetPowerNoGreater}
	return spec
}
