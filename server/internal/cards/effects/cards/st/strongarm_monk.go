package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Strongarm Monk — Creature — Human Monk {4}{W}, 3/3 (Dragons of Tarkir, #5551):
//
//	"Whenever you cast a noncreature spell, creatures you control get +1/+1 until end of turn."
//
// Monastery Mentor's cast trigger; the creatures pumped are the ones you
// control as the ability resolves (CR 611.2c).
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "8711b732-8dab-41d4-a85b-b10edcd25e97",
		Name:         "Strongarm Monk",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Strongarm Monk — creatures you control get +1/+1 until end of turn",
				effects.Do(effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1, Label: "Strongarm Monk — creatures you control get +1/+1"})),
		},
	})
}
