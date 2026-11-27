package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tel-Jilad Stylus — Artifact {1} (Mirrodin, S82 OLD, #5606):
//
//	"{T}: Put target permanent you own on the bottom of your library."
//
// The target is any permanent its controller owns, the Stylus itself
// included (the {T} is paid first, and it is still on the battlefield as a
// legal target while the ability is on the stack). It is put on the bottom
// as the ability resolves, if it is still a legal target (CR 608.2b), and
// the owner of a commander in it is offered the command zone first (CR 903.9).
// Nothing is shuffled or revealed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75a7fdd4-046a-433b-b16f-a0424a05c904",
		Name:         "Tel-Jilad Stylus",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Put target permanent you own on the bottom of your library.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target permanent you own", effects.YouOwn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				id, ok := effects.B16FirstLegalTargetCard(effects.NewContext(g, item))
				if !ok {
					return nil
				}
				return g.TuckToLibraryForEffect(id, true)
			},
		}},
	})
}
