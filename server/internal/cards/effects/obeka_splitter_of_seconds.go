package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Obeka, Splitter of Seconds — Legendary Creature — Ogre Warlock
// {1}{U}{B}{R}, 2/5 (Outlaws of Thunder Junction, #4468):
//
//	"Menace
//	 Whenever Obeka deals combat damage to a player, you get that many
//	 additional upkeep steps after this phase."
//
// #5405's seam (YouGetUpkeepStepsAfterThisPhase,
// Game.AddUpkeepStepsAfterThisPhaseForEffect). "That many" is the
// damage event's amount. Each upkeep is a beginning phase of its own
// after this combat phase with its untap and draw steps skipped
// (CR 500.10, the first ruling), the newest added phase first (CR 500.8,
// the last ruling), and then the turn goes on to its postcombat main
// phase. "At the beginning of your upkeep" triggers fire in every one
// (the third ruling), and the active player gets priority in each. The
// steps are added only on Obeka's controller's turn (CR 500.10a: "you
// get"). Menace rides PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1d15a06f-5471-48f2-8c4a-9b0886bef5af",
		Name:            "Obeka, Splitter of Seconds",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(
				"Obeka, Splitter of Seconds — you get that many additional upkeep steps after this phase",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return YouGetUpkeepStepsAfterThisPhase{N: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
				}),
		},
	})
}
