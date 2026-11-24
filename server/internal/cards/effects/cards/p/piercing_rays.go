package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Piercing Rays — Sorcery {1}{W} (Modern Horizons 2):
//
//	"Exile target tapped creature.
//	 Forecast — {2}{W}, Reveal this card from your hand: Tap target untapped
//	 creature. (Activate this ability only during your upkeep and only once each
//	 turn.)"
//
// The spell is Expel's clause (a creature that untaps in response is no longer
// a legal target). Forecast (CR 702.57a) is an activated ability that works from
// the hand: the card is revealed as a cost and stays in the hand, and it is
// limited to your upkeep and once each turn, counted per object (Sky Hussar's
// shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8960a213-cb72-4d67-bb4e-bcb1f657b944",
		Name:         "Piercing Rays",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target tapped creature", effects.TappedPermanent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstLegalCardTarget(ctx.Game, item)
		},
		Activated: []effects.ActivatedAbility{{
			Label:     effects.OldB035ForecastTapLabel,
			Purpose:   game.Purpose{Answers: game.AnswerRemove},
			Zones:     []game.ZoneKind{game.ZoneHand},
			Cost:      effects.Plus(effects.ManaCost("{2}{W}"), game.AbilityCost{RevealThis: true}),
			Targets:   effects.TargetCreature("target untapped creature", effects.Untapped()),
			Condition: effects.AllConditions(effects.DuringYourUpkeep(), effects.OncePerTurnActivation(effects.OldB035ForecastTapLabel)),
			Effect:    effects.TapChosenPermanent,
		}},
	})
}
