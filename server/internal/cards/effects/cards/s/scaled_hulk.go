package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scaled Hulk — Creature — Spirit {5}{G}, 4/4 (Betrayers of Kamigawa, S82
// OLD, #5602):
//
//	"Whenever you cast a Spirit or Arcane spell, this creature gets +2/+2
//	 until end of turn."
//
// Kami of the Hunt's cast trigger with a +2/+2 self-pump, once per spell;
// each bonus ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	const label = "Scaled Hulk — +2/+2 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "ce4570eb-f637-45d1-afbc-f11357c09bf9",
		Name:         "Scaled Hulk",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.ChkB1SpiritOrArcane(), label, effects.ThisGetsUntilEndOfTurn(2, 2, label)),
		},
	})
}
