package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deadeye Tormentor — Creature — Human Pirate {2}{B}, 2/2 (Ixalan):
//
//	"Raid — When this creature enters, if you attacked this turn, target
//	 opponent discards a card."
//
// Wingmate Roc's raid trigger: an intervening "if" (CR 603.4) read from the
// turn's attack record as the creature enters and again as the ability
// resolves, with Fell Specter's discard (the opponent chooses what goes).
// Cast before combat, it triggers nothing and picks no target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab735456-f831-4988-84c6-50ca14351990",
		Name:         "Deadeye Tormentor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.InterveningIf(effects.AvatarB11CondYouAttacked{},
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.TmcB6YouAttackedThisTurn),
					"Deadeye Tormentor — raid: target opponent discards a card", effects.TargetOpponentDiscards(1))),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
