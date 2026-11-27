package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fiddlehead Kami — Creature — Spirit {4}{G}, 3/3 (Saviors of Kamigawa,
// #5601):
//
//	"Whenever you cast a Spirit or Arcane spell, regenerate this
//	 creature."
//
// The ki cycle's cast trigger (ChkB1SpiritOrArcane) with a regeneration
// shield on the source (CR 701.19a); a Kami that has left gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7392c58b-1ba9-4cf3-8449-aa5de3c42130",
		Name:         "Fiddlehead Kami",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.ChkB1SpiritOrArcane(), "Fiddlehead Kami — regenerate this creature", effects.RegenerateThisSource),
		},
	})
}
