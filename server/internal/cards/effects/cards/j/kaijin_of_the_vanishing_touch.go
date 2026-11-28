package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kaijin of the Vanishing Touch — Creature — Spirit {1}{U}, 0/3 (Betrayers of
// Kamigawa, #5602):
//
//	"Defender
//	 Whenever this creature blocks a creature, return that creature to its
//	 owner's hand at end of combat. (Return it only if it's on the
//	 battlefield.)"
//
// A block trigger once per attacker blocked (CR 509.3b), stamped with that
// attacker's object as it was blocked, that schedules a delayed trigger at the
// beginning of this combat's end of combat step (CR 511.2, 603.7) naming it.
// The delayed trigger is Wall of Junk's (ReturnObjectToHandBody): it returns
// the object only if it is still that object and still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cd8bab2c-b04f-43a2-9786-df5f7411bf16",
		Name:            "Kaijin of the Vanishing Touch",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered:       []game.TriggeredAbility{effects.OldB749WheneverThisBlocksReturnItAtEndOfCombat("Kaijin of the Vanishing Touch")},
	})
}
