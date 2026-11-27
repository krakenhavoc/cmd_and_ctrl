package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kami of Tattered Shoji — Creature — Spirit {4}{W}, 2/5 (Betrayers of Kamigawa, S82 OLD
// b595):
//
//	"Whenever you cast a Spirit or Arcane spell, this creature gains flying until end of
//	 turn."
//
// Kami of the Hunt's cast trigger with a keyword instead of a pump: once per spell, and the
// flying ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	const label = "Kami of Tattered Shoji — gains flying until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "a7cb1da6-e56a-42b6-8087-cf822783373f",
		Name:         "Kami of Tattered Shoji",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.ChkB1SpiritOrArcane(), label, effects.ThisGainsKeywordUntilEndOfTurn("flying", label)),
		},
	})
}
