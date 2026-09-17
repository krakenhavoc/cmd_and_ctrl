package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Master of the Pearl Trident — Creature — Merfolk, {U}{U}, 2/2:
//
//	"Other Merfolk creatures you control get +1/+1 and have islandwalk."
//
// Lord of Atlantis's modern reprint, and the difference is the one the
// tribal builders make a field of: this one says "you control", so an
// opponent's Merfolk get neither the +1/+1 nor the islandwalk. The
// islandwalk is enforced by the block check since #705
// (game/landwalk.go): it bites when the defending player controls an
// Island, including a land an effect has made one.
//
// No simplification.
func init() {
	yourOtherMerfolk := TribeFilter{Tribes: []string{"Merfolk"}, Others: true, YoursOnly: true}
	Register(Spec{
		OracleID:     "9f6ea9fe-eee5-4dca-a34e-c66d3e218716",
		Name:         "Master of the Pearl Trident",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(yourOtherMerfolk, 1, 1),
			TribalKeywordGrant(yourOtherMerfolk, "islandwalk"),
		},
	})
}
