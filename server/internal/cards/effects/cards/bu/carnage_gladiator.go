package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Carnage Gladiator — Creature — Skeleton Warrior {2}{B}{R}, 4/2 (Conspiracy:
// Take the Crown, S82 older sets batch 273, #5540):
//
//	"Whenever a creature blocks, that creature's controller loses 1 life.
//	 {1}{B}{R}: Regenerate this creature."
//
// One trigger per blocking creature, whoever controls it (CR 509.3a: the
// first of its blocks, so a creature that blocks several attackers triggers
// once), losing the life of the player who controls that blocker as it
// resolves, or as it last was if it has gone. The regeneration is the shared
// shield row (CR 701.19a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d8e20e5-6456-43f5-9188-263b54c0a923",
		Name:         "Carnage Gladiator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBlock && ev.Amount <= 1
			}, "Carnage Gladiator — that creature's controller loses 1 life",
				func(g *game.Game, item *game.StackItem) error {
					ev := item.Trigger.Event
					player := ev.Actor
					if c, ok := g.LookupCardForEffect(ev.CardID); ok {
						player = c.Controller
					}
					if p := g.PlayerByIDForEffect(player); p == nil || p.Eliminated {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, player, -1)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.RegenerateThisRow("{1}{B}{R}")},
	})
}
