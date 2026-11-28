package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Márton Stromgald — Legendary Creature — Human Knight {2}{R}{R}, 1/1 (Ice Age,
// S82 OLD, #5636):
//
//	"Whenever Márton Stromgald attacks, other attacking creatures get +1/+1 until
//	 end of turn for each attacking creature other than Márton Stromgald.
//	 Whenever Márton Stromgald blocks, other blocking creatures get +1/+1 until
//	 end of turn for each blocking creature other than Márton Stromgald."
//
// Each trigger counts the other attackers (blockers) as it resolves and gives each
// of them that many +1/+1 (CR 611.2c); Márton himself is left out of both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8364cf93-c323-46f4-854f-ae872e484438",
		Name:         "Márton Stromgald",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Márton Stromgald — other attacking creatures get +1/+1 for each attacking creature other than Márton",
				effects.OldB834OtherCombatantsGetPerOther(effects.AttackingCreature(), "Márton Stromgald — other attacking creatures get +1/+1 for each other attacker")),
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B28SelfBlocked(ev, source)
			}, "Márton Stromgald — other blocking creatures get +1/+1 for each blocking creature other than Márton",
				effects.OldB834OtherCombatantsGetPerOther(effects.BlockingCreature(), "Márton Stromgald — other blocking creatures get +1/+1 for each other blocker")),
		},
	})
}
