package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lord of Atlantis — Creature — Merfolk, {U}{U}, 2/2:
//
//	"Other Merfolk get +1/+1 and have islandwalk."
//
// The first S16 Layer-6 catalog card, and since S26 the canonical
// example of the shared lord builders in tribal.go.
//
// FIXED IN S26 — the anthem used to be restricted to the Lord's own
// controller. The printed card has no "you control" clause: Lord of
// Atlantis buffs every Merfolk on the battlefield, an opponent's
// included, which is exactly the sort of symmetrical drawback a 1993
// lord has and a modern one does not. Goblin King and Elvish Champion
// are in the same family and are written the same way; Goblin
// Chieftain and Elvish Archdruid really do say "you control" and set
// YoursOnly. (The "only Merfolk you control" caveat outlived that fix
// by a year; tribal_test.go asserts the opponent's Merfolk is pumped.)
//
// LIVE SINCE #705. The islandwalk grant was inert from S16 until the
// block-legality check could see the defending player's lands
// (game/landwalk.go, ADR 0045 addendum Decision 10). The grant is a
// layer-6 keyword, so it stops the moment the Lord leaves, and it is
// symmetrical like the anthem: an opponent's Merfolk attacking you
// has islandwalk too, and it bites if YOU control an Island.
//
// No simplification.
func init() {
	otherMerfolk := TribeFilter{Tribes: []string{"Merfolk"}, Others: true}
	Register(Spec{
		OracleID:     "cc7f290f-ca00-4285-9bdb-4b4402444f30",
		Name:         "Lord of Atlantis",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(otherMerfolk, 1, 1),
			TribalKeywordGrant(otherMerfolk, "islandwalk"),
		},
	})
}
