package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Takenuma Bleeder — {2}{B} Creature — Ogre Shaman 3/3 (Jumpstart 2022, #5948):
//
//	"Whenever this creature attacks or blocks, you lose 1 life if you don't
//	 control a Demon."
//
// Hamlet Captain's attacks-or-blocks trigger with an intervening "if"
// (CR 603.4): no Demon when it triggers, and none as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee397a87-6933-4d2d-aac6-8dc4e96d524b",
		Name:         "Takenuma Bleeder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return (effects.AttackDeclared(ev, source) || effects.SelfBlocksOnce(ev, source)) && !takenumaBleederControlsDemon(g, source.Controller)
				},
				"Takenuma Bleeder — you lose 1 life if you don't control a Demon",
				func(g *game.Game, item *game.StackItem) error {
					if takenumaBleederControlsDemon(g, item.Controller) {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -1)
				}),
		},
	})
}

func takenumaBleederControlsDemon(g *game.Game, you uuid.UUID) bool {
	return len(effects.PermanentsControlledByMatching(g, you, effects.Subtype("Demon"))) > 0
}
