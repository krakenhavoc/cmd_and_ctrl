package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ControlledByOpponentsOfTargetPlayer is "target <permanents> their
// opponents control", where they are the earlier target player in `slot`
// (Portal Manipulator's "choose target player and any number of target
// attacking creatures their opponents control"). Build the permanent clause
// with TargetCreature and its count, then wrap it:
//
//	Clauses(
//	    TargetPlayer("target player"),
//	    ControlledByOpponentsOfTargetPlayer(
//	        TargetCreature("any number of target attacking creatures their opponents control", AttackingCreature()).WithCount(0, 0), 0),
//	)
//
// Each pick must be controlled by a player other than the anchor as the
// ability is put on the stack (CR 601.2c) and is judged again on its own as
// it resolves (CR 608.2b): a permanent that changed controller to the
// player is dropped and the rest stay.
func ControlledByOpponentsOfTargetPlayer(spec *game.TargetSpec, slot int) *game.TargetSpec {
	spec.Relation = &game.TargetRelation{Slot: slot, Kind: game.TargetControlledByOpponentOfPlayer}
	return spec
}

// checkOpponentsControlRelation is Register's gate for
// ControlledByOpponentsOfTargetPlayer: the anchor is a player-only clause,
// the dependent clause targets only permanents on the battlefield, and it
// carries no rule about its set of picks (the copy and retarget walks refill
// its picks one at a time).
func checkOpponentsControlRelation(name string, i int, anchor, clause *game.TargetClause) {
	if !anchor.Players || anchor.CardOK != nil || len(anchor.Zones) > 0 {
		panic(fmt.Sprintf("effects.Register: %q target clause %d is controlled by a target player's opponents, but clause %d is not a player-only clause", name, i, clause.Relation.Slot))
	}
	if clause.Players || len(clause.Zones) != 1 || clause.Zones[0] != game.ZoneBattlefield {
		panic(fmt.Sprintf("effects.Register: %q target clause %d is controlled by a target player's opponents, so it must target only permanents on the battlefield", name, i))
	}
	if clause.Different != nil || clause.Same != nil || clause.Total != nil || clause.Divide != nil {
		panic(fmt.Sprintf("effects.Register: %q target clause %d is controlled by a target player's opponents and cannot also carry a rule about its set of picks", name, i))
	}
}
