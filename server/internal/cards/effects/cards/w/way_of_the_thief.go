package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Way of the Thief — Enchantment — Aura {3}{U} (Gatecrash, S82 OLD b407,
// #5565):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2.
//	 Enchanted creature can't be blocked as long as you control a Gate."
//
// Become the Pilot's conditional "can't be blocked" (CantBeBlockedWhile): the
// condition is read as the block is checked, for the Aura's controller ("you"),
// so a Gate that leaves lets the creature be blocked again, and a creature
// someone else controls is unblockable for the Aura's controller's Gates.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b43ba23b-ff92-4f52-a4f4-2d0825abb7ed",
		Name:         "Way of the Thief",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedWhile(effects.OnAttached(), func(g *game.Game, you uuid.UUID, _ game.Card) bool {
				return effects.ControlsA("Gate")(g, you)
			}),
		},
	})
}
