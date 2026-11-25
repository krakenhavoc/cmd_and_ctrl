package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sentinel of the Pearl Trident — Creature — Merfolk Soldier {4}{U}, 3/3
// (Dominaria, S82 OLD batch 214, #5530):
//
//	"Flash
//	 When this creature enters, you may exile target historic permanent you
//	 control. If you do, return that card to the battlefield under its owner's
//	 control at the beginning of the next end step. (Artifacts, legendaries,
//	 and Sagas are historic.)"
//
// Roon of the Hidden Realm's exile-and-return behind an optional enters
// trigger: the target is chosen as the trigger goes on the stack, the "may"
// is answered as it resolves, and a target that is no longer a legal one
// (CR 608.2b) is not exiled. The return is a CR 603.7 delayed trigger at the
// next end step; the card comes back as a new object (CR 400.7), and a token
// exiled this way is gone (CR 111.8). Historic is read through the layers
// (Jhoira's artifact, legendary or Saga test); the Sentinel itself is a
// legal target when it is legendary or an artifact through some effect.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2bddbd7d-9b6b-4166-88de-368d91c7fe4f",
		Name:            "Sentinel of the Pearl Trident",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.WhenThisEnters("Sentinel of the Pearl Trident — exile target historic permanent you control; return it at the beginning of the next end step",
						func(g *game.Game, item *game.StackItem) error {
							return effects.ExileTargetsThenScheduleReturn(effects.NewContext(g, item),
								"Sentinel of the Pearl Trident — return the exiled permanent")
						}),
					"Sentinel of the Pearl Trident — exile target historic permanent you control?"),
				effects.TargetPermanent("target historic permanent you control", effects.YouControl(), effects.B14Historic())),
		},
	})
}
