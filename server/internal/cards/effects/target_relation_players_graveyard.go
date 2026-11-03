package effects

import (
	"fmt"
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// FromTargetPlayersGraveyard is "target cards from their graveyard", where
// they are the earlier target player in `slot` (#4862): Quandrix Command's
// "target player shuffles up to three target cards from their graveyard
// into their library", and the Gaea's Blessing family after it. Build the
// card clause with TargetCardInGraveyard and its count, then wrap it:
//
//	Clauses(
//	    TargetPlayer("target player"),
//	    FromTargetPlayersGraveyard(
//	        TargetCardInGraveyard("up to three target cards from their graveyard").WithCount(0, 3), 0),
//	)
//
// Every pick must be in that player's graveyard as the spell is cast
// (CR 601.2c) and is judged again on its own as it resolves (CR 608.2b):
// a card that has left the graveyard is dropped, and the rest stay. The
// player staying legal is the card body's question, not the clause's: an
// illegal player performs no action, but its cards are still legal
// targets.
func FromTargetPlayersGraveyard(spec *game.TargetSpec, slot int) *game.TargetSpec {
	spec.Relation = &game.TargetRelation{Slot: slot, Kind: game.TargetInTargetPlayersGraveyard}
	return spec
}

// checkPlayersGraveyardRelation is Register's gate for
// FromTargetPlayersGraveyard. The anchor is a player and only a player
// ("target player", not "any target"), the dependent clause looks only in
// graveyards, and the dependent clause carries no rule about its set of
// picks: the copy and retarget walks refill its picks one at a time,
// which is exact only while each pick is judged against the anchor alone.
func checkPlayersGraveyardRelation(name string, i int, anchor, clause *game.TargetClause) {
	if !anchor.Players || anchor.CardOK != nil || len(anchor.Zones) > 0 {
		panic(fmt.Sprintf("effects.Register: %q target clause %d is from a target player's graveyard, but clause %d is not a player-only clause", name, i, clause.Relation.Slot))
	}
	if clause.Players || !slices.Equal(clause.Zones, []game.ZoneKind{game.ZoneGraveyard}) {
		panic(fmt.Sprintf("effects.Register: %q target clause %d is from a target player's graveyard, so it must target only cards in graveyards", name, i))
	}
	if clause.Different != nil || clause.Same != nil || clause.Total != nil || clause.Divide != nil {
		panic(fmt.Sprintf("effects.Register: %q target clause %d is from a target player's graveyard and cannot also carry a rule about its set of picks", name, i))
	}
}
