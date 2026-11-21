package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Morkrut Necropod — Creature — Slug Horror {5}{B}, 7/7 (Foundations
// Jumpstart):
//
//	"Menace
//	 Whenever this creature attacks or blocks, sacrifice another
//	 creature or land."
//
// One trigger on both events (Burning Sun Cavalry's OnAny). The
// sacrifice is its controller's choice among the other creatures and
// lands they control, asked as the trigger resolves; with none, nothing
// is sacrificed (CR 701.21a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ed3990cd-fa67-427b-bbe4-a2b1942e9be2",
		Name:            "Morkrut Necropod",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclared(ev, source) || effects.SelfBlocksOnce(ev, source)
				},
				"Morkrut Necropod — sacrifice another creature or land",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SacrificeChoice{
						Player:     item.Controller,
						Candidates: effects.PermanentsControlledByMatching(g, item.Controller, effects.And(effects.Or(effects.Creature(), effects.Land()), effects.NotSelf(item.SourceCardID))),
						Question:   "Morkrut Necropod — sacrifice another creature or land",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
