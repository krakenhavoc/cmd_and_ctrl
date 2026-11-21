package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drider — Creature — Elf Spider {4}{B}, 4/3 (Adventures in the Forgotten
// Realms, S82 batch 15, #5499):
//
//	"Reach
//	 Whenever this creature deals combat damage to a player, create a 2/1
//	 black Spider creature token with reach and menace."
//
// One trigger per combat damage event to a player: a Drider that is
// double-blocked or removed after damage still triggers (the event is the
// damage). The Spider is the printed token, in tokens_table.go.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1db76752-31ed-4a16-986a-d0852ba7d812",
		Name:            "Drider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Drider — create a 2/1 black Spider creature token with reach and menace",
				effects.Do(effects.CreateToken{Template: effects.TokenCard(effects.AfrB15SpiderToken), N: 1})),
		},
	})
}
