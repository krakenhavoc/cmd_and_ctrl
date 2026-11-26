package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of Oblivion — Enchantment — Aura Curse {3}{B}:
//
//	"Enchant player
//	 At the beginning of enchanted player's upkeep, that player exiles two
//	 cards from their graveyard."
//
// The enchanted player chooses which two (all of them with fewer than two).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10b16b2b-7d7c-4f6f-932a-1f3bd30d33f1",
		Name:         "Curse of Oblivion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.OldB444EnchantedPlayersUpkeepTrigger("Curse of Oblivion — that player exiles two cards from their graveyard",
				effects.OldB444ExileTwoFromYourGraveyard),
		},
	})
}
