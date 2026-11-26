package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gnathosaur — Creature — Dinosaur {4}{R}{R}, 5/4 (Mirrodin Besieged, S82 OLD
// b465, #5576):
//
//	"Sacrifice an artifact: This creature gains trample until end of turn."
//
// A free activation whose only cost is one artifact the controller picks at
// announce (Atog's cost), then the shared self keyword grant until end of
// turn. An activation whose creature has left does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f52993c-a20a-48fe-b463-2a33873b489a",
		Name:         "Gnathosaur",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice an artifact: This creature gains trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.SacrificeN(1, "an artifact", effects.Artifact()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.GrantKeywordUntilEOT{
					Target: item.SourceCardID, Keywords: []string{"trample"},
					Label: "Gnathosaur — trample until end of turn",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
