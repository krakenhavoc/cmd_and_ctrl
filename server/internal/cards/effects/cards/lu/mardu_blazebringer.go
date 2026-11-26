package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mardu Blazebringer — Creature — Ogre Warrior {2}{R}, 4/4 (Khans of Tarkir):
//
//	"When this creature attacks or blocks, sacrifice it at end of combat."
//
// Runaway Carriage's attack-or-block trigger over the delayed end-of-combat
// sacrifice (CR 511.2): one printed ability, one registered trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "352844c5-2160-4787-97f1-9cd0fed060eb",
		Name:         "Mardu Blazebringer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclared(ev, source) || effects.SelfBlocksOnce(ev, source)
			}, "Mardu Blazebringer — sacrifice it at end of combat",
				effects.SacrificeThisAtEndOfCombat("Mardu Blazebringer — sacrifice it")),
		},
	})
}
