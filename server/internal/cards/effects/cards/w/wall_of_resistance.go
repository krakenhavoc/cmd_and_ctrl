package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wall of Resistance — Creature — Wall {1}{W}, 0/3 (Mirage, S82 OLD b805):
//
//	"Defender
//	 Flying
//	 At the beginning of each end step, if this creature was dealt damage this
//	 turn, put a +0/+1 counter on it."
//
// The intervening "if" (CR 603.4) reads the turn's damage record (every
// damage event dealt to this object, combat or not, after prevention) when
// the end step begins and again as the trigger resolves, so damage that was
// healed or prevented leaves nothing, and a Wall that came back as a new
// object (CR 400.7) was dealt none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e6f5f80c-2620-4558-8a5f-8fc36cea4793",
		Name:            "Wall of Resistance",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.OldB805InterveningIfSource("this creature was dealt damage this turn", func(g *game.Game, source, _ uuid.UUID) bool {
				return effects.OnBattlefield(g, source) && g.TotalDamageDealtToObjectThisTurnForEffect(source) > 0
			}, effects.On(game.EventBeginEndStep, effects.AnyPlayer,
				"Wall of Resistance — put a +0/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: "+0/+1", N: 1}))),
		},
	})
}
