package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f409e869-ea1a-4e08-b19e-949436d01cc1",
		Name:         "Excalbian Gamemaster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Excalbian Gamemaster — strange new worlds: create a 1/1 red Officer for each differently named land you control",
				Do(Instruction(func(ctx *Context) error {
					return CreateToken{Template: TokenCard("1/1 red Officer"), N: socB23DifferentlyNamedLands(ctx)}.Apply(ctx)
				}))),
			TriggerWithPurpose(Targeting(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b16SelfOrAnotherCreatureYouControlDied(ev, source, g)
			}, "Excalbian Gamemaster — put a +1/+1 counter on target creature you control", PlusOneCounterOnTheTarget),
				TargetCreature("target creature you control", YouControl())), game.Purpose{DeathPayoff: true}),
		},
	})
}
