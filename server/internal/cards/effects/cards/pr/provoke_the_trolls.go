package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Provoke the Trolls — Instant {3}{R} (Starter Commander Decks):
//
//	"Provoke the Trolls deals 3 damage to any target. If a creature is dealt
//	 damage this way, it gets +5/+0 until end of turn."
//
// Silver Bolt's keyed tail: the +5/+0 is applied to the creature that was
// actually dealt damage, told how much it took, so damage that is prevented
// entirely gives nothing. A player or planeswalker is dealt damage as usual
// and gets no bonus.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac844303-487b-49e9-9170-02892153ca6a",
		Name:         "Provoke the Trolls",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			t := ts[0]
			if c, ok := ctx.Game.LookupCardForEffect(t.ID); ok && t.Kind == game.TargetCard && c.IsCreature() && ctx.Game.Battlefield.Contains(t.ID) {
				return ctx.Game.DealDamageToCreatureKeyedForEffect(ctx.Source(), t.ID, 3, ctx.Tail(provokeTheTrollsPump))
			}
			return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}.Apply(ctx)
		},
	})
}

// provokeTheTrollsPump is "if a creature is dealt damage this way, it gets
// +5/+0 until end of turn", told the permanent and the damage it was dealt.
// An on-disk key.
var provokeTheTrollsPump = effects.CardTail("then/provoke-the-trolls-pump", func(ctx *effects.Context, o game.TailOutcome) error {
	if o.Amount <= 0 || len(o.IDs) == 0 {
		return nil
	}
	if !effects.OnBattlefield(ctx.Game, o.IDs[0]) {
		return nil
	}
	return effects.BoostUntilEOT{Target: o.IDs[0], Power: 5, Label: "Provoke the Trolls — +5/+0"}.Apply(ctx)
})
