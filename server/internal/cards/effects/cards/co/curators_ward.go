package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curator's Ward — Enchantment — Aura {2}{U} (Dominaria, S82 OLD, #5530):
//
//	"Enchant permanent
//	 Enchanted permanent has hexproof.
//	 When enchanted permanent leaves the battlefield, if it was historic, draw
//	 two cards. (Artifacts, legendaries, and Sagas are historic.)"
//
// A layer 6 hexproof grant on the widest enchant clause. The leave trigger
// watches any departure of the enchanted permanent (Valor of the Worthy's
// EnchantedPermanentLeft) and is gated on its last-known types, supertypes and
// subtypes (CR 603.10a, 700.6): an artifact, a legendary or a Saga as it left.
// That is an intervening "if" about past information, so it needs no recheck as
// the trigger resolves (CR 603.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a89ad91e-2bc2-45f4-b32c-fb2b67a2019e",
		Name:         "Curator's Ward",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPermanent(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 2},
		Static:       []game.StaticAbility{effects.GrantToAttached("hexproof")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.S82OldB212EnchantedHistoricPermanentLeft,
				"Curator's Ward — enchanted permanent was historic: draw two cards",
				effects.Do(effects.DrawCards{N: 2})),
		},
	})
}
