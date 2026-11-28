package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chain Stasis — Instant {U} (Homelands, S82 OLD b827, #5635):
//
//	"You may tap or untap target creature. Then that creature's controller may
//	 pay {2}{U}. If the player does, they may copy this spell and may choose a
//	 new target for that copy."
//
// The Chain cycle's self-copy (chain_of_vapor.go, CR 707.10) priced in mana,
// with Rejoinder's tap-or-untap pick in front. The caster picks tap, untap or
// neither as the spell resolves; the creature's controller, read before
// anything happens to it, is then asked whether to pay {2}{U} (a controller
// who cannot pay is not asked a second question), and only a paid offer leads
// to the copy, which is theirs (CR 707.10b) with a new target of their choice
// (CR 707.10c). A target that is illegal on resolution takes the whole spell
// with it (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ec365f8-bfb8-416f-b4db-868a8f5d042a",
		Name:         "Chain Stasis",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRestrict},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			c, found := ctx.Game.LookupCardForEffect(t.ID)
			if !found {
				return nil
			}
			return effects.PickOption{
				Question: "Chain Stasis — you may tap or untap " + c.Name,
				Options: []game.ChoiceOption{
					{Label: "Tap it", Cards: []uuid.UUID{t.ID}},
					{Label: "Untap it", Cards: []uuid.UUID{t.ID}},
					{Label: "Leave it as it is", Cards: []uuid.UUID{t.ID}},
				},
				ThenKey:  chainStasisPicked,
				Carry:    []uuid.UUID{t.ID, c.Controller},
				WithItem: true,
			}.Apply(ctx)
		},
	})
}

var (
	// chainStasisPicked applies the pick, then asks the creature's controller
	// (Carry[1], read before the pick) whether to pay {2}{U}. An on-disk key:
	// never renamed, never reused.
	chainStasisPicked = effects.OptionPickThen("option-pick/old-b827-chain-stasis-tap-or-untap", func(ctx *effects.Context, r game.OptionPicked) error {
		if len(r.Carry) < 2 {
			return nil
		}
		switch r.Index {
		case 0:
			if err := (effects.TapTarget{Target: r.Carry[0]}).Apply(ctx); err != nil {
				return err
			}
		case 1:
			if err := (effects.UntapTarget{Target: r.Carry[0]}).Apply(ctx); err != nil {
				return err
			}
		}
		return effects.MayPay{
			Chooser:  r.Carry[1],
			Cost:     "{2}{U}",
			Question: "Chain Stasis — pay {2}{U}? (if you do, you may copy this spell)",
			Then:     chainStasisPaid,
		}.Apply(ctx)
	})
	// chainStasisPaid is the paid branch: the payer may copy the spell. An
	// on-disk key: never renamed, never reused.
	chainStasisPaid = effects.PayUnlessBranches("pay-unless/old-b827-chain-stasis-copy",
		func(ctx *effects.Context, r game.PayUnlessAnswer) error {
			if ctx.Item == nil {
				return nil
			}
			return effects.ChainOfferCopy(ctx.Game, ctx.Item, r.Chooser, "Chain Stasis")
		}, nil)
)
