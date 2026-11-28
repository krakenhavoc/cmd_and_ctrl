package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Trade Caravan — Creature — Human Nomad {W}, 1/1 (Homelands, S82 OLD b822,
// #5635):
//
//	"At the beginning of your upkeep, put a currency counter on this creature.
//	Remove two currency counters from this creature: Untap target basic land.
//	Activate only during an opponent's upkeep."
//
// A counter per upkeep (PutACounterOnThis) and Earthcraft's untap, paid with two
// counters removed as a cost, in the upkeep of a player who is not the
// Caravan's controller.
//
// No simplification.
func init() {
	const currency = "currency"
	effects.Register(effects.Spec{
		OracleID:     "8faccd56-dfc2-4cfc-880c-db40b935336a",
		Name:         "Trade Caravan",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Trade Caravan — put a currency counter on it", effects.PutACounterOnThis(currency)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove two currency counters from this creature: Untap target basic land. Activate only during an opponent's upkeep.",
			Cost:    effects.RemoveCountersFromThis(currency, 2),
			Targets: effects.TargetPermanent("target basic land", effects.OldB096BasicLandMatch),
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return g.Turn.Step == game.StepUpkeep && !effects.IsYourTurn(g, controller)
			},
			Effect: effects.UntapTheTarget,
		}},
	})
}
