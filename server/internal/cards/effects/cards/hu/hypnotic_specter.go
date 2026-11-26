package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hypnotic Specter — Creature — Specter {1}{B}{B}, 2/2 (Magic 2010, #5585):
//
//	"Flying
//	 Whenever this creature deals damage to an opponent, that player
//	 discards a card at random."
//
// Skullknocker Ogre's trigger (any damage to an opponent) without the draw:
// the damaged opponent discards through the engine's random discard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "759af941-f6a3-4726-91f2-9b1e4e55ea71",
		Name:            "Hypnotic Specter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsDamageToAnOpponent("Hypnotic Specter — that player discards a card at random", func(g *game.Game, item *game.StackItem) error {
				return g.DiscardRandomKeyedForEffect(item.Trigger.Event.Target, 1, game.KeyedTail{})
			}),
		},
	})
}
