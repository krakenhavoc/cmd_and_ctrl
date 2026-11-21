package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sun Quan, Lord of Wu — Legendary Creature — Human Soldier {4}{U}{U}, 4/4
// (Commander Legends: Battle for Baldur's Gate):
//
//	"Creatures you control have horsemanship. (They can't be blocked except by
//	 creatures with horsemanship.)"
//
// A keyword grant in layer 6 (CR 613.1f) to every creature its controller
// controls, itself included: Herald of Hoofbeats' grant with no type
// restriction. Horsemanship is an engine keyword (the block check refuses any
// blocker without it).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "70503280-6bdf-431d-8213-367a9783463a",
		Name:         "Sun Quan, Lord of Wu",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{YoursOnly: true}, "horsemanship"),
		},
	})
}
