package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rogue Elephant — Creature — Elephant {G}, 3/3 (Weatherlight, S82 older sets batch 777, #5629):
//
//	"When this creature enters, sacrifice it unless you sacrifice a Forest."
//
// Lithophage's tax over Forests (SacrificeThisUnlessYouSacrificeA): a two-way choice when its
// controller has a Forest to give, and an unconditional sacrifice of the Elephant when they have none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bed7ac55-fe40-46d0-bc22-1c8d11f41459",
		Name:         "Rogue Elephant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rogue Elephant — sacrifice it unless you sacrifice a Forest",
				effects.SacrificeThisUnlessYouSacrificeA("Rogue Elephant", "a Forest", effects.HasSubtype("Forest"), effects.S82OldB777SacrificeAForestOrThisAnswer)),
		},
	})
}
