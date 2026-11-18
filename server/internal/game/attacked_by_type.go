package game

import "github.com/google/uuid"

// AttackedByCreatureTypeThisTurn reports whether `defender` was attacked
// this turn by a creature of `subtype` that `controller` controlled as it
// was declared as an attacker (S82, Ramses, Assassin Lord: "if they were
// attacked this turn by an Assassin you controlled"). Who controlled the
// attacker and what it was are read from the declaration (AttackRecord),
// not from the battlefield, so an Assassin that has since died or changed
// hands still counts (CR 608.2h). A changeling is every creature type
// (CR 702.73a). An attack on a planeswalker or a battle is not an attack
// on a player (CR 506.2).
//
// Caller must hold g.mu.
func (g *Game) AttackedByCreatureTypeThisTurn(defender, controller uuid.UUID, subtype string) bool {
	if defender == uuid.Nil || controller == uuid.Nil || subtype == "" {
		return false
	}
	for _, a := range g.TurnTally.Attacks {
		if a.Defender != defender || a.Controller != controller || a.Battle {
			continue
		}
		if a.AllTypes && IsCreatureType(subtype) {
			return true
		}
		for _, s := range a.Subtypes {
			if s == subtype {
				return true
			}
		}
	}
	return false
}
