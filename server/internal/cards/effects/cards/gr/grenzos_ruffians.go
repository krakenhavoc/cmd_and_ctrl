package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grenzo's Ruffians — Creature — Goblin {2}{R}{R}, 2/2 (Conspiracy: Take
// the Crown, S82 older sets batch 270, #5540):
//
//	"Melee (Whenever this creature attacks, it gets +1/+1 until end of turn
//	 for each opponent you attacked this combat.)
//	 Whenever this creature deals combat damage to an opponent, it deals
//	 that much damage to each other opponent."
//
// Melee is the engine's keyword (CR 702.121, game/melee.go). The trigger is
// Amarant Coral's: how much landed and on whom are facts of the damage
// event, and the damage to each other opponent is not combat damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3ed4ee5c-91da-41a3-894c-6181496ec955",
		Name:            "Grenzo's Ruffians",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordMelee},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisHitsAnOpponentThatMuchToEachOther("Grenzo's Ruffians — that much damage to each other opponent"),
		},
	})
}
