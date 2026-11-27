package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rushing-Tide Zubera — Creature — Zubera Spirit {2}{U}{U}, 3/3 (Saviors of
// Kamigawa, #5601):
//
//	"When this creature dies, if 4 or more damage was dealt to it this turn,
//	 draw three cards."
//
// An intervening "if" (CR 603.4): checked as the creature dies and again as
// the trigger resolves. The damage is the turn's total dealt to this
// object from any source, combat or not, after prevention
// (Game.TotalDamageDealtToObjectThisTurnForEffect), read through the object
// it last was now that it is in the graveyard (CR 603.10a).
//
// No simplification.
func init() {
	const label = "Rushing-Tide Zubera — draw three cards"
	effects.Register(effects.Spec{
		OracleID:     "d99b44bc-375a-4cf1-9148-ef6fe615cb10",
		Name:         "Rushing-Tide Zubera",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.OldB585DamageDealtToThisAtLeast(4),
				effects.WhenThisDies(label, effects.Do(effects.DrawCards{N: 3}))),
		},
	})
}
