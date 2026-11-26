package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Punishing Fire — Instant {1}{R} (Commander 2011, #5574):
//
//	"Punishing Fire deals 2 damage to any target.
//	 Whenever an opponent gains life, you may pay {R}. If you do, return this
//	 card from your graveyard to your hand."
//
// The spell is Shock. The second paragraph is a triggered ability that works
// only from the graveyard (InGraveyard), where "you" is the card's owner
// (CR 108.4): it fires once per life-gain event of any other player, and the
// {R} is asked as the ability resolves; the card comes back only if it is
// still the very object that triggered (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1890ef1b-38e5-432d-b040-f33cea7814e7",
		Name:         "Punishing Fire",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(2),
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.On(game.EventChangeLife, anOpponentGainedLife,
				"Punishing Fire — you may pay {R} to return it to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPayForGraveyardReturnToHand(g, item, "{R}", "Punishing Fire — pay {R} to return it to your hand?")
				})),
		},
	})
}

// anOpponentGainedLife — a life-gain event whose player is not the source's
// controller (every other player is an opponent).
func anOpponentGainedLife(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventChangeLife && ev.Amount > 0 && ev.Target != source.Controller
}
