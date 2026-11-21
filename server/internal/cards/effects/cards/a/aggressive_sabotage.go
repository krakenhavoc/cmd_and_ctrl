package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aggressive Sabotage — {2}{B} Sorcery (Dominaria United):
//
//	"Kicker {R} (You may pay an additional {R} as you cast this spell.)
//	 Target player discards two cards. If this spell was kicked, it deals
//	 3 damage to that player."
//
// CR 702.33's Kicker, read back with ctx.WasKicked. The two sentences are
// two steps (ResolveSteps), so the damage waits for the discard choice
// (CR 608.2c). The damage is dealt even when the player holds fewer than
// two cards. A target player who is gone does nothing (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:      "a47ff11e-dbcf-45b1-a13d-fff677ea42c9",
		Name:          "Aggressive Sabotage",
		Completeness:  effects.CompletenessFull,
		Purpose:       effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 2}),
		Targets:       effects.TargetPlayer("target player"),
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{R}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetPlayer {
				return nil
			}
			prompt := game.DiscardPrompt{Player: t.ID, Source: item.SourceCardID, N: 2}
			if !ctx.WasKicked() {
				ctx.Game.QueueDiscardChoiceForEffect(prompt)
				return nil
			}
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(prompt, aggressiveSabotageBurn, nil)
		},
	})
}

// aggressiveSabotageBurn is the kicked sentence, run once the discard is
// settled: 3 damage to that player, whatever they discarded.
var aggressiveSabotageBurn = game.RegisterDiscardThen("choose-cards/aggressive-sabotage-burn", func(g *game.Game, r game.DiscardsDone) error {
	return g.DealDamageToPlayerForEffect(r.Source, r.Player, 3)
})
