package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Titan's Presence — Instant {3} (Battle for Zendikar; Commander Masters,
// #5948):
//
//	"As an additional cost to cast this spell, reveal a colorless creature
//	 card from your hand.
//	 Exile target creature if its power is less than or equal to the
//	 revealed card's power."
//
// The cost is Calamity of the Titans' reveal component (game.RevealCost
// {CardType, Colorless, Chosen}): the caster names the card on cast_spell
// as reveal_ids, it is shown to the table and moves nowhere, and the
// payment records it as it was revealed (game.ChosenPowerForEffect), so a
// card that has left the hand by resolution still counts for the power it
// was revealed with (CR 608.2h). The target's power is read as the spell
// resolves, through the layers; a greater power exiles nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b011d6b2-8930-4192-aa94-82869faa991b",
		Name:         "Titan's Presence",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		AdditionalCost: &game.AdditionalCost{
			Reveal: &game.RevealCost{CardType: "creature", Colorless: true, Chosen: true},
			Label:  "Reveal a colorless creature card from your hand",
		},
		Targets: effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			revealed, ok := ctx.Game.ChosenPowerForEffect(ctx.Item)
			if !ok {
				return nil
			}
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			ctx.Game.RecomputeLayersIfStaleLocked()
			c, found := ctx.Game.LookupCardForEffect(t.ID)
			if !found || c.PowerForComparison() > revealed {
				return nil
			}
			return effects.ExileTarget{Target: t.ID}.Apply(ctx)
		},
	})
}
