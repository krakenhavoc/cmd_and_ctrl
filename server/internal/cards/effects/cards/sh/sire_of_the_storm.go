package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sire of the Storm — Creature — Spirit {4}{U}{U}, 3/3 (Champions of
// Kamigawa, #5603):
//
//	"Flying
//	 Whenever you cast a Spirit or Arcane spell, you may draw a card."
//
// The ki cycle's cast trigger with an optional draw (CR 603.5), asked as
// it resolves. It triggers on the cast, so a countered spell still drew.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "68f1db12-82fb-4bf1-908d-db38fd67efe5",
		Name:            "Sire of the Storm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.WheneverYouCast(effects.ChkB1SpiritOrArcane(), "Sire of the Storm — draw a card", effects.Do(effects.DrawCards{N: 1})),
				"Sire of the Storm — draw a card?", 1),
		},
	})
}
