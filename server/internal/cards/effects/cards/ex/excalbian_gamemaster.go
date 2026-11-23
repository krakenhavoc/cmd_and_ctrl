package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Excalbian Gamemaster — {5}{W} Creature — Excalbian 6/6 (Star Trek,
// #4973):
//
//	"Strange new worlds — When this creature enters, create X 1/1 red
//	 Officer creature tokens, where X is the number of differently named
//	 lands you control.
//	 Whenever this creature or another creature you control dies, put a
//	 +1/+1 counter on target creature you control."
//
// Strange new worlds is an ability word with no rules meaning (CR
// 207.2c). X is counted as the trigger resolves, each land name once
// (socB23DifferentlyNamedLands, Emil's count), so five Forests are one.
// The dies trigger is Vengeful Bloodwitch's "this or another creature
// you control" condition, judged on the creature as it last existed on
// the battlefield (CR 603.10a); its target is chosen as it goes on the
// stack, so with no creature left to target it is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f409e869-ea1a-4e08-b19e-949436d01cc1",
		Name:         "Excalbian Gamemaster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Excalbian Gamemaster — strange new worlds: create a 1/1 red Officer for each differently named land you control",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.CreateToken{Template: effects.TokenCard("1/1 red Officer"), N: effects.SocB23DifferentlyNamedLands(ctx)}.Apply(ctx)
				}))),
			effects.TriggerWithPurpose(effects.Targeting(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B16SelfOrAnotherCreatureYouControlDied(ev, source, g)
			}, "Excalbian Gamemaster — put a +1/+1 counter on target creature you control", effects.PlusOneCounterOnTheTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())), game.Purpose{DeathPayoff: true}),
		},
	})
}
