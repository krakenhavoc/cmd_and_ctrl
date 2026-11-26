package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Havengul Skaab — Creature — Zombie Horror {5}{U}, 4/5 (Avacyn Restored,
// S82 OLD, #5570):
//
//	"Whenever this creature attacks, return another creature you control to
//	 its owner's hand."
//
// Roaring Primadox's mandatory bounce on an attack trigger, with the source
// left out of the candidates ("another" is the object). The pick is made as
// the ability resolves; with no other creature nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8fda6c20-04dc-4b60-9313-896a032ee5cb",
		Name:         "Havengul Skaab",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Havengul Skaab — return another creature you control to its owner's hand",
				effects.Do(effects.S82OldB437ReturnAnotherCreatureYouControl{Question: "Havengul Skaab — return another creature you control to its owner's hand"})),
		},
	})
}
