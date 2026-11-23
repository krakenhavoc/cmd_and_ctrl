package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Agency Coroner — Creature — Ogre Cleric {4}{B}:
//
//	"{2}{B}, Sacrifice another creature: Draw a card. If the sacrificed
//	 creature was suspected, draw two cards instead."
//
// "Was suspected" is a fact about the creature as it left, and the
// designation is cleared the moment it does (CR 400.7), so the effect
// reads the cost's record of it — the sacrificed permanent's last-known
// information (CR 608.2h), which carries Suspected — rather than the
// card now in the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3686fd97-49ff-4a92-9cb2-7d8e9438d945",
		Name:         "Agency Coroner",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice another creature: Draw a card. If the sacrificed creature was suspected, draw two cards instead.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				n := 1
				if info, ok := ctx.SacrificedPermanent(); ok && info.Suspected {
					n = 2
				}
				return effects.DrawCards{Player: item.Controller, N: n}.Apply(ctx)
			},
		}},
	})
}
