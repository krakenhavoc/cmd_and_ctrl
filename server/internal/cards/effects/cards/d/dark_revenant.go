package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dark Revenant — Creature — Spirit {3}{B}, 2/2 (Return to Ravnica, #5566):
//
//	"Flying
//	 When this creature dies, put it on top of its owner's library."
//
// Murderous Rider's dies trigger with the top of the library for the bottom:
// the card moves only while it is still the same card in its owner's graveyard
// (CR 400.7), through the tuck route, so a commander's owner is offered the
// command zone (CR 903.9) and a token, which ceased to exist, is not found.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7191f765-6e66-4bc4-8aa8-85280ac61d1e",
		Name:            "Dark Revenant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Dark Revenant — put it on top of its owner's library", effects.PutThisDiedCardOnTopOfItsOwnersLibrary),
		},
	})
}
