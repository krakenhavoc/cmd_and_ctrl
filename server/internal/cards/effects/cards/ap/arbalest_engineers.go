package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arbalest Engineers — Creature — Human Artificer {1}{R}{G}, 2/2 (The
// Brothers' War):
//
//	"When this creature enters, choose one —
//	 • This creature deals 1 damage to any target.
//	 • Put a +1/+1 counter on target creature. It gains trample and haste
//	   until end of turn.
//	 • Create a tapped Powerstone token."
//
// A modal enters trigger (Apothecary Stomper's shape): the mode and its
// target are chosen as the trigger goes on the stack and re-checked at
// resolution (CR 608.2b). The counter and the keywords go to the same
// target, in that order.
//
// No simplification.
func init() {
	trigger := effects.WhenThisEnters("Arbalest Engineers — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = effects.ChooseOne(
		effects.ModeDoing("This creature deals 1 damage to any target.", effects.TargetAny(), effects.SocB20DamageToTheModeTarget(1)),
		effects.ModeDoing("Put a +1/+1 counter on target creature. It gains trample and haste until end of turn.",
			effects.TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample", "haste"}, Label: "Arbalest Engineers"}.Apply(ctx)
			}),
		effects.ModeDoing("Create a tapped Powerstone token.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: 1}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "06bea0b2-1d0f-4d64-a13e-723d08042e83",
		Name:         "Arbalest Engineers",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
