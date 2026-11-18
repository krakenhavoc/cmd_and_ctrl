package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anafenza, Kin-Tree Spirit — Legendary Creature — Spirit Soldier {W}{W},
// 2/2:
//
//	"Whenever another nontoken creature you control enters, bolster 1."
//
// Bolster is effects.Bolster (CR 701.39): the creature with the least
// toughness among those you control gets the counter, chosen as the
// trigger resolves, so the entering creature may itself be the one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f7fa6ac0-70db-4586-b881-5a19db644dfb",
		Name:         "Anafenza, Kin-Tree Spirit",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, anotherNontokenCreatureEnteredUnderYourControl,
				"Anafenza, Kin-Tree Spirit — bolster 1",
				Do(Bolster{N: 1})),
		},
	})
}
