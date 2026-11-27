package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skeleton Scavengers — Creature — Skeleton {2}{B}, 0/0 (Stronghold, S82 OLD
// batch 748, #5627):
//
//	"This creature enters with a +1/+1 counter on it.
//	 Pay {1} for each +1/+1 counter on this creature: Regenerate this creature.
//	 When it regenerates this way, put a +1/+1 counter on it."
//
// The entry counter is a self-replacement (CR 614.1c). The cost is generic mana
// equal to the +1/+1 counters on the Scavengers as the ability is announced
// (game.AbilityCost.ManaPerCounter, CR 602.2b), {0} with none. The shield is
// an ordinary regeneration shield (CR 701.19a) that carries the rider
// (Game.RegenerateWithCounterRiderForEffect): when it is used up, the engine
// announces the regeneration with Amount 1 and the printed triggered ability
// puts a +1/+1 counter on it. With several shields up, one that carries the
// rider is the one used up first (the controller may choose which replacement
// applies, CR 616.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88782913-bfba-4300-b26c-405e6a45537c",
		Name:         "Skeleton Scavengers",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.OldB031SelfEntersWithCounters(game.CounterPlusOne, 1)},
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay {1} for each +1/+1 counter on this creature: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.PayOnePerCounterOnThis(game.CounterPlusOne),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.NothingIfGone(g.RegenerateWithCounterRiderForEffect(item.SourceCardID))
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventRegenerated, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventRegenerated && ev.CardID == source.InstanceID && ev.Amount > 0
			}, "Skeleton Scavengers — put a +1/+1 counter on it", effects.PutCounterOnSelf),
		},
	})
}
