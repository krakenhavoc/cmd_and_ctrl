package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silumgar Spell-Eater — Creature — Snake Wizard {2}{U}, 2/3 (Dragons of Tarkir,
// #5551):
//
//	"Megamorph {4}{U} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its megamorph cost and put a +1/+1 counter on it.)
//	 When this creature is turned face up, counter target spell unless its controller pays {3}."
//
// Megamorph is the shared constructor (morph.go): the face-down cast for
// {3}, and the +1/+1 counter placed by the turn-face-up special action
// (CR 702.37b).
// The face-up trigger is Complicate's cycling trigger: the spell's
// controller is asked to pay {3} as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "34d90f98-7af5-4c67-8fb3-620237d94d14",
		Name:             "Silumgar Spell-Eater",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Megamorph("{4}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisIsTurnedFaceUp("Silumgar Spell-Eater — counter target spell unless its controller pays {3}",
				effects.CounterTheTargetSpellUnlessPaid("Silumgar Spell-Eater", "{3}")), effects.TargetSpell("target spell")),
		},
	})
}
