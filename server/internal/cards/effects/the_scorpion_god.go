package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Scorpion God — Legendary Creature — God {3}{B}{R}, 6/5 (Lorwyn
// Eclipsed Commander, #5368):
//
//	"Whenever a creature with a -1/-1 counter on it dies, draw a card.
//	 {1}{B}{R}: Put a -1/-1 counter on another target creature.
//	 When The Scorpion God dies, return it to its owner's hand at the
//	 beginning of the next end step."
//
// The first trigger is any creature's and The Scorpion God's own, read from
// the counters the creature last had (b29CreatureWithMinusCounterDied). The
// ability's target is any creature but the God. The return is a CR 603.7
// delayed trigger scheduled as the dies trigger resolves, carrying the dead
// card; at the next end step, any player's, it returns only if the card is
// still in a graveyard (Liesa's returnListedGraveyardToHandBody), so one that
// was exiled or reanimated meanwhile is a new object and is left alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aff223d4-6338-460d-a4d5-d45f522864dc",
		Name:         "The Scorpion God",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b29CreatureWithMinusCounterDied(ev, g)
			}, "The Scorpion God — draw a card", Do(DrawCards{N: 1})),
			WhenThisDies("The Scorpion God — return it to its owner's hand at the beginning of the next end step",
				func(g *game.Game, item *game.StackItem) error {
					return ScheduleDelayedTrigger{
						Label: "The Scorpion God — return it to its owner's hand",
						Cards: []uuid.UUID{item.Trigger.Event.CardID},
						Body:  returnListedGraveyardToHandBody,
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}{R}: Put a -1/-1 counter on another target creature.",
			Cost:    ManaCost("{1}{B}{R}"),
			Targets: Harmful(Another(TargetCreature("another target creature"))),
			Effect:  cmdr2B25CountersOnTarget(1),
		}},
	})
}
