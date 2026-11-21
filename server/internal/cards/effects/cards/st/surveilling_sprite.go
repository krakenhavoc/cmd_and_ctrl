package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Surveilling Sprite — Creature — Faerie Rogue {1}{U}, 1/1 (Ravnica: City of
// Guilds, #5600):
//
//	"Flying
//	 When this creature dies, you may draw a card."
//
// Solemn Simulacrum's optional dies draw: only a trip to the graveyard
// from the battlefield triggers it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "449142ff-7190-471b-99a5-f77095a827dd",
		Name:            "Surveilling Sprite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.WhenThisDies("Surveilling Sprite — draw a card", effects.Do(effects.DrawCards{N: 1})), "Surveilling Sprite — draw a card?", 1),
		},
	})
}
