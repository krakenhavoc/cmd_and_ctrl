package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vapor Snare — Enchantment — Aura {4}{U} (Worldwake, S82 OLD batch 489, #5582):
//
//	"Enchant creature
//	 You control enchanted creature.
//	 At the beginning of your upkeep, sacrifice this Aura unless you return a
//	 land you control to its owner's hand."
//
// Control Magic's layer-2 control change, and an upkeep trigger that is
// returnLandsOrSacrifice with one land of any kind: the choice is made as it
// resolves, declining or having no land sacrifices the Aura, and it stays
// only if the land really reached a hand. "You" is the Aura's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29ae5540-5fdf-48be-86ba-b4db0fe19f6c",
		Name:         "Vapor Snare",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.ControlAttachedBySource()},
		Triggered: []game.TriggeredAbility{
			vaporSnareReturn.UpkeepTrigger("Vapor Snare — sacrifice it unless you return a land you control to its owner's hand"),
		},
	})
}

var vaporSnareReturn = effects.NewReturnLandsOrSacrifice(
	"Vapor Snare — return a land you control to its owner's hand (or sacrifice Vapor Snare)", 1,
	"choose-cards/vapor-snare-return-a-land", "wwk/vapor-snare-after-return",
	effects.S82OldB489AnyLand)
