package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Teferi's Curse — Enchantment — Aura {1}{U} (Mirage, S82 OLD batch 811, #5633):
//
//	"Enchant artifact or creature
//	 Enchanted permanent has phasing. (It phases in or out before its
//	 controller untaps during each of their untap steps. While it's phased
//	 out, it's treated as though it doesn't exist.)"
//
// A layer 6 grant of phasing (CR 702.26a) to whatever the Aura is attached to,
// which the engine honours on its own. The Aura's legality ("artifact or
// creature") is the clause rerun every time state-based actions look
// (CR 704.5m).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e422d5e-d9f5-492d-b82a-31f7274d4153",
		Name:         "Teferi's Curse",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPermanent(effects.Or(effects.Artifact(), effects.Creature())),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.GrantToAttached(game.KeywordPhasing)},
	})
}
