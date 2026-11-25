package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slate Street Ruffian — Creature — Human Warrior {2}{B}, 2/2 (Jumpstart,
// #5510):
//
//	"Whenever this creature becomes blocked, defending player discards a
//	 card."
//
// Abyssal Nightstalker's discard prompt (the discarding player chooses the
// card, CR 701.9a) on the once-per-attacker "becomes blocked" event, which
// names the defending player as its actor (CR 509.3c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c74d72ee-4795-41ca-b676-4d5a05bd12b7",
		Name:         "Slate Street Ruffian",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBecomesBlocked && ev.CardID == source.InstanceID
			}, "Slate Street Ruffian — defending player discards a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DefendingPlayerDiscardsACard(g, item, item.Trigger.Event.Actor)
				}),
		},
	})
}
