package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quagmire Lamprey — Creature — Fish {2}{B}, 1/1 (Mercadian Masques, S82 OLD
// batch 698, #5619):
//
//	"Whenever this creature becomes blocked by a creature, put a -1/-1 counter
//	 on that creature."
//
// Somberwald Vigilante's per-blocker trigger (CR 509.3d): one for each creature
// that blocks it, and the counter lands on that blocker only if it is still the
// same object on the battlefield as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5b374d8-bf22-44f7-a18a-18c096a84640",
		Name:         "Quagmire Lamprey",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, effects.ViciousBattleragerBlocked,
				"Quagmire Lamprey — put a -1/-1 counter on the creature blocking it",
				effects.S82OldB698MinusCounterOnBlocker),
		},
	})
}
