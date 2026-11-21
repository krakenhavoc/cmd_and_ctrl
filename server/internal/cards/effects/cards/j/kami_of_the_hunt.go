package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kami of the Hunt — Creature — Spirit {2}{G}, 2/2 (Champions of
// Kamigawa, #5603):
//
//	"Whenever you cast a Spirit or Arcane spell, this creature gets
//	 +1/+1 until end of turn."
//
// The ki cycle's cast trigger with bushido's self-pump, once per spell;
// each bonus ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	const label = "Kami of the Hunt — +1/+1 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "96e68c7a-f187-4aa6-998c-238e7c833809",
		Name:         "Kami of the Hunt",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.ChkB1SpiritOrArcane(), label, effects.ThisGetsUntilEndOfTurn(1, 1, label)),
		},
	})
}
