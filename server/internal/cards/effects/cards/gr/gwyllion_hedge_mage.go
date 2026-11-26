package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gwyllion Hedge-Mage — Creature — Hag Wizard {2}{W/B}, 2/2 (Commander 2011,
// #5574):
//
//	"When this creature enters, if you control two or more Plains, you may
//	 create a 1/1 white Kithkin Soldier creature token.
//	 When this creature enters, if you control two or more Swamps, you may put
//	 a -1/-1 counter on target creature."
//
// Two separate triggers, each with its own intervening "if" (CR 603.4): the
// land count is checked as it triggers and again as it resolves, before the
// "you may" is asked. The Plains and Swamps are lands with that subtype, so a
// nonbasic one counts. The counter trigger targets as it goes on the stack and
// is not put there at all unless a creature can be targeted (CR 603.3d), and
// not even then without two Swamps.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b89b285d-62d3-4013-b881-23f0f236337d",
		Name:         "Gwyllion Hedge-Mage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.InterveningIf(effects.CondControlTwoOrMoreOfSubtype("Plains"),
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ControlTwoOrMoreOfSubtype("Plains")),
					"Gwyllion Hedge-Mage — if you control two or more Plains, you may create a 1/1 white Kithkin Soldier creature token",
					effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Kithkin Soldier"), N: 1}))),
				"Gwyllion Hedge-Mage — create a 1/1 white Kithkin Soldier creature token?"),
			effects.Optional(effects.Targeting(effects.InterveningIf(effects.CondControlTwoOrMoreOfSubtype("Swamp"),
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ControlTwoOrMoreOfSubtype("Swamp")),
					"Gwyllion Hedge-Mage — if you control two or more Swamps, you may put a -1/-1 counter on target creature",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						target, ok := ctx.ClauseTarget(0)
						if !ok || target.Kind != game.TargetCard {
							return nil
						}
						return effects.AddCounter{Target: target.ID, Kind: game.CounterMinusOne, N: 1}.Apply(ctx)
					})),
				effects.TargetCreature("target creature")),
				"Gwyllion Hedge-Mage — put a -1/-1 counter on target creature?"),
		},
	})
}
