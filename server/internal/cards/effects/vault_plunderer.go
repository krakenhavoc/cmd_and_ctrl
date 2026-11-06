package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vault Plunderer — Creature — Human Rogue {2}{B}, 3/1 (Outlaws of
// Thunder Junction, #4468):
//
//	"When this creature enters, target player draws a card and loses 1
//	 life."
//
// Blood Pact's body (b32TargetPlayerDrawsAndLosesLife) on an enters
// trigger: the player is any player, you included, and an opponent makes
// it a crime (CR 700.13). The life is lost, not dealt as damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "80e5d768-cc1f-48f8-b0fe-ed974e4d1a56",
		Name:         "Vault Plunderer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Vault Plunderer — target player draws a card and loses 1 life",
				func(g *game.Game, item *game.StackItem) error {
					return b32TargetPlayerDrawsAndLosesLife(item, NewContext(g, item), 1, 1)
				}), TargetPlayer("target player")),
				ForTargets(game.TargetPurpose{Slot: 0, Draws: 1, LifeLoss: 1})),
		},
	})
}
