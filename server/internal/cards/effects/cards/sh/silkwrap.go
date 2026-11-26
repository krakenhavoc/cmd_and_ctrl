package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silkwrap — Enchantment {1}{W} (Dragons of Tarkir):
//
//	"When this enchantment enters, exile target creature with mana value 3 or
//	 less an opponent controls until this enchantment leaves the battlefield.
//	 (That creature returns under its owner's control.)"
//
// The Oblivion Ring family's shared tail (ExileChosenTargetUntilThisLeaves): the
// card returns under its owner's control when Silkwrap leaves (CR 610.3c); if
// Silkwrap has already left when the trigger resolves nothing is exiled
// (CR 610.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "909becbc-1d84-4e94-b589-e58438569403",
		Name:         "Silkwrap",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Silkwrap — exile target creature with mana value 3 or less an opponent controls until this enchantment leaves the battlefield",
				effects.ExileChosenTargetUntilThisLeaves("Silkwrap — the exiled creature returns when Silkwrap leaves the battlefield")),
				effects.TargetCreature("target creature with mana value 3 or less an opponent controls", effects.And(effects.OpponentControls(), effects.ManaValueLE(3)))),
		},
	})
}
