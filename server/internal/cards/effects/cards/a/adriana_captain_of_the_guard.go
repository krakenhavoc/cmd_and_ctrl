package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Adriana, Captain of the Guard — Legendary Creature — Human Knight
// {3}{R}{W}, 4/4 (Assassin's Creed, reprinted in Commander Legends sets):
//
//	"Melee (Whenever this creature attacks, it gets +1/+1 until end of turn
//	 for each opponent you attacked this combat.)
//	 Other creatures you control have melee. (If a creature has multiple
//	 instances of melee, each triggers separately.)"
//
// Melee is the engine keyword (CR 702.121). The grant is a layer 6 keyword
// grant to every other creature its controller has (Drogskol
// Reinforcements' shape without the tribe).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "de23275e-e9c1-4151-aed8-f79ead03d5d4",
		Name:            "Adriana, Captain of the Guard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordMelee},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Others: true, YoursOnly: true}, game.KeywordMelee),
		},
	})
}
