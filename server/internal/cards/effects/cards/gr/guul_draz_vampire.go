package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guul Draz Vampire — Creature — Vampire Rogue {B}, 1/1 (Zendikar, S82 OLD b495):
//
//	"As long as an opponent has 10 or less life, this creature gets +2/+1 and has
//	 intimidate."
//
// Ruthless Cullblade's pump and Bloodghast's gated keyword, both on the same
// condition (SelfWhileAnOpponentsLifeAtMost): any one opponent still in the game
// at 10 or less is enough, and both statics follow the life total as it changes.
// Intimidate is granted, not printed.
//
// No simplification.
func init() {
	pump := effects.LifeGatedPump(effects.SelfWhileAnOpponentsLifeAtMost(10), 2, 1)
	pump.Label = "This creature gets +2/+1 as long as an opponent has 10 or less life."
	effects.Register(effects.Spec{
		OracleID:     "00d3fcdf-6959-4437-9bad-93a8625510a1",
		Name:         "Guul Draz Vampire",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			pump,
			effects.LifeGatedKeyword(effects.SelfWhileAnOpponentsLifeAtMost(10), "intimidate"),
		},
	})
}
