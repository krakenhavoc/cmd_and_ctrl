package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Garruk, Savage Herald — Legendary Planeswalker — Garruk {4}{G}{G}, loyalty 5
// (Core Set 2021, #5511):
//
//	"+1: Reveal the top card of your library. If it's a creature card, put it
//	 into your hand. Otherwise, put it on the bottom of your library.
//	 −2: Target creature you control deals damage equal to its power to another
//	 target creature.
//	 −7: Until end of turn, creatures you control gain "You may have this
//	 creature assign its combat damage as though it weren't blocked.""
//
// The −2 is a one-sided bite over two clauses, the second a different object
// from the first (Distinct); nothing is dealt unless both are still legal
// (CR 608.2b). The −7 locks the creatures you control as it resolves
// (CR 611.2c) and gives each the option the combat damage step asks about
// (CR 510.1c) until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fc762e40-84c1-4762-82f7-a0fc4f6fd522",
		Name:            "Garruk, Savage Herald",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+1: Reveal the top card of your library. If it's a creature card, put it into your hand. Otherwise, put it on the bottom of your library.",
				Cost:   effects.LoyaltyCost(1),
				Effect: effects.Do(effects.Instruction(garrukSavageHeraldReveal)),
			},
			{
				Label: "−2: Target creature you control deals damage equal to its power to another target creature.",
				Cost:  effects.LoyaltyCost(-2),
				Targets: effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.Distinct(effects.TargetCreature("another target creature"))),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error { return effects.OneSidedBite(nil, ctx) })),
			},
			{
				Label: "−7: Until end of turn, creatures you control gain \"You may have this creature assign its combat damage as though it weren't blocked.\"",
				Cost:  effects.LoyaltyCost(-7),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.UntilEndOfTurn(ctx, uuid.Nil, effects.And(effects.Creature(), effects.YouControl()),
						"Garruk, Savage Herald — assign combat damage as though unblocked", game.AssignsAsThoughUnblockedMod())
				})),
			},
		},
	})
}

// garrukSavageHeraldReveal is the +1: the top card goes to your hand if it is
// a creature card, to the bottom of your library otherwise.
func garrukSavageHeraldReveal(ctx *effects.Context) error {
	var revealed []uuid.UUID
	if err := (effects.RevealTopOfLibrary{Player: ctx.Controller(), N: 1, Reason: "Garruk, Savage Herald — reveal the top card of your library", Revealed: &revealed}).Apply(ctx); err != nil {
		return err
	}
	if len(revealed) != 1 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(revealed[0])
	if !ok {
		return nil
	}
	if c.IsCreature() {
		return effects.BounceToHand{Target: revealed[0]}.Apply(ctx)
	}
	return ctx.Game.PutOnBottomInRandomOrderForEffect(ctx.Controller(), game.ZoneLibrary, revealed)
}
