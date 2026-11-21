package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ardent Plea — Enchantment {1}{W}{U} (Alara Reborn, #5586):
//
//	"Exalted (Whenever a creature you control attacks alone, that creature
//	 gets +1/+1 until end of turn.)
//	 Cascade (When you cast this spell, exile cards from the top of your
//	 library until you exile a nonland card that costs less. You may cast
//	 it without paying its mana cost. Put the exiled cards on the bottom in
//	 a random order.)"
//
// Exalted is the canonical keyword, on an enchantment as on Angelic
// Benediction (CR 702.83a); cascade is the catalog keyword (CR 702.85a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "49a8ca54-30d3-42d3-b122-334d24805e1b",
		Name:            "Ardent Plea",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordExalted},
		Triggered:       []game.TriggeredAbility{effects.Cascade()},
	})
}
