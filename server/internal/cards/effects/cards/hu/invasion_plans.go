package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invasion Plans — Enchantment {2}{R} (Secret Lair Drop, #5462):
//
//	"All creatures block each combat if able.
//	 The attacking player chooses how each creature blocks each combat."
//
// The first line is Grand Melee's block half (BlocksEachCombatWhere over every
// creature, CR 509.1c). The second is the engine's AttackingPlayerChoosesBlocks
// (game/declaration_control.go): while Invasion Plans is on the battlefield,
// in every declare blockers step the attacking player declares each other
// defending player's blocks, which creatures block and what they block
// (CR 509.1a), the Master Warcraft projection. Every block restriction and
// requirement still applies to the chosen blocks, and the defenders keep every
// other decision.
//
// No simplification.
func init() {
	allCreatures := func(_ *game.Card, _ *game.Game, _ *game.Card) bool { return true }
	effects.Register(effects.Spec{
		OracleID:                     "db9c0d94-9fa5-4f07-be48-23e19d2f3d68",
		Name:                         "Invasion Plans",
		Completeness:                 effects.CompletenessFull,
		AttackingPlayerChoosesBlocks: true,
		Static: []game.StaticAbility{
			effects.BlocksEachCombatWhere(allCreatures),
		},
	})
}
