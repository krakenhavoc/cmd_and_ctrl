package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// In the Eye of Chaos — World Enchantment {2}{U}:
//
//	"Whenever a player casts an instant spell, counter it unless that
//	 player pays {X}, where X is its mana value."
//
// Nether Void's trigger narrowed to instants, with the price read off
// the spell as the trigger resolves: its mana value on the stack, an
// announced {X} included (CR 202.3e). A mana value of 0 is a free
// payment the caster is still asked to make.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6a0da9f3-cb06-42b1-ae73-b142c0eedaff",
		Name:         "In the Eye of Chaos",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverAPlayerCastsCounterUnlessPays("In the Eye of Chaos — counter it unless that player pays {X}, where X is its mana value",
				func(_ *game.Game, spell game.Card) bool { return spell.HasCardType("instant") },
				effects.PayTheSpellsManaValue),
		},
	})
}
