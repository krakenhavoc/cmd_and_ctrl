package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Forsaken Wastes — World Enchantment {2}{B}:
//
//	"Players can't gain life.
//	 At the beginning of each player's upkeep, that player loses 1 life.
//	 Whenever this enchantment becomes the target of a spell, that
//	 spell's controller loses 5 life."
//
// "Players can't gain life" is ADR 0107 §5's battlefield static
// (CR 119.7, #1880). The upkeep loss fires on every seat's upkeep, the
// controller's included, and "that player" is the active player the
// event names. The targeting punisher watches EventBecomesTarget for a
// SPELL only (SelfTargetedByASpell, Bonecrusher Giant's predicate), and
// "that spell's controller" is the event's actor, captured when the
// ability triggered — the spell may be countered before it resolves.
// All three are life LOSS, not damage.
//
// A world permanent: the world rule (CR 704.5k, ADR 0109 §8) puts it
// into its owner's graveyard when a newer one enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b9e61e68-9dc8-4295-95dc-dd66a0907c8c",
		Name:         "Forsaken Wastes",
		Completeness: CompletenessFull,
		CantGainLife: PlayersCantGainLife(),
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil
			}, "Forsaken Wastes — that player loses 1 life", func(g *game.Game, item *game.StackItem) error {
				if p := triggeringActor(item); p != uuid.Nil {
					return g.ChangePlayerLifeForEffect(item.SourceCardID, p, -1)
				}
				return nil
			}),
			On(game.EventBecomesTarget, SelfTargetedByASpell, "Forsaken Wastes — that spell's controller loses 5 life",
				func(g *game.Game, item *game.StackItem) error {
					if p := triggeringActor(item); p != uuid.Nil {
						return g.ChangePlayerLifeForEffect(item.SourceCardID, p, -5)
					}
					return nil
				}),
		},
	})
}
