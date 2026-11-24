package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maelstrom Muse — Creature — Djinn Wizard {1}{U}{U/R}{R}, 2/4 (Strixhaven:
// School of Mages, #5502):
//
//	"Flying
//	 Whenever this creature attacks, the next instant or sorcery spell you
//	 cast this turn costs {X} less to cast, where X is this creature's power as
//	 this ability resolves."
//
// Kaza, Roil Chaser's next-spell promise with the Muse's power read as the
// ability resolves (CR 608.2h: its last-known power if it has left the
// battlefield) and stored as a number, which is what the printed text says. A
// power of zero or less promises nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "30b03dda-409f-4615-b392-477d66eb49f4",
		Name:            "Maelstrom Muse",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.ThisAttacked,
				"Maelstrom Muse — the next instant or sorcery spell you cast this turn costs {X} less, X its power",
				func(g *game.Game, item *game.StackItem) error {
					x := effects.B31CurrentPowerOf(g, item.SourceCardID)
					if x <= 0 {
						return nil
					}
					return effects.GrantNextSpellPromise{From: "Maelstrom Muse", Promise: game.NextSpellPromise{
						Filter: game.PermissionFilter{InstantOrSorceryOnly: true},
						Reduce: x,
						Text:   "The next instant or sorcery spell you cast this turn costs {X} less to cast.",
					}}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
