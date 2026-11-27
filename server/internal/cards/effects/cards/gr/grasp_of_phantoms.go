package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grasp of Phantoms — {3}{U} Sorcery (Innistrad, S82 OLD b448, #5572):
//
//	"Put target creature on top of its owner's library.
//	 Flashback {7}{U} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// The tuck is the ordinary one (a commander's owner is offered the command zone,
// CR 903.9), to the top of the OWNER's library. Flashback is the shared
// graveyard cast, exiled as it leaves the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "0d96e4bc-ff4e-4f37-89b6-8a4e3c1069a0",
		Name:             "Grasp of Phantoms",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerRemove},
		Targets:          effects.TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{7}{U}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return ctx.Game.TuckToLibraryForEffect(t.ID, false)
				}
			}
			return nil
		},
	})
}
