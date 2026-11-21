package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keldon Strike Team — Creature — Human Warrior {2}{R}, 3/1 (Dominaria
// United):
//
//	"Kicker {1}{W}
//	 When this creature enters, if it was kicked, create two 1/1 white
//	 Soldier creature tokens.
//	 As long as this creature entered this turn, creatures you control have
//	 haste."
//
// Two printed abilities. The kicked enters trigger is Kor Sanctifiers' (an
// intervening "if", CR 603.4) making two Soldiers. The static is a layer-6
// keyword grant to every creature its controller controls, itself included,
// that applies only while this object entered this turn: it reads the turn
// tally's entry count (Game.EnteredThisTurn), not summoning sickness, so it
// ends with the turn (and, through a blink, starts again with the new object).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "6465f93d-e6bf-4a7b-86f3-5a2376f8b127",
		Name:          "Keldon Strike Team",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Tokens: 2, CreatureTokens: 2},
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}{W}")},
		Static: []game.StaticAbility{
			effects.Labelled(effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					g.EnteredThisTurn(source.InstanceID)
			}, "haste"), "As long as this creature entered this turn, creatures you control have haste."),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
				"Keldon Strike Team — kicked, create two 1/1 white Soldier creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 2})),
		},
	})
}
