package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ajani, Caller of the Pride — Legendary Planeswalker — Ajani {1}{W}{W},
// loyalty 4 (Foundations, #4465):
//
//	"+1: Put a +1/+1 counter on up to one target creature.
//	 −3: Target creature gains flying and double strike until end of turn.
//	 −8: Create X 2/2 white Cat creature tokens, where X is your life
//	     total."
//
// +1 is Basri Ket's counter without the keyword: one counter on the target
// if it is still legal, and nothing (but the loyalty) with no target
// chosen. −3 grants both keywords until end of turn (layer 6). −8 reads
// your life total as it resolves, so a life change in response changes X;
// at 0 life or less it makes no Cats. The Cat's colour and type are words
// of the card's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1d5ff280-7f4d-4801-aefd-9b6e6b1f2818",
		Name:            "Ajani, Caller of the Pride",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Put a +1/+1 counter on up to one target creature.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1),
				Effect:  effects.PlusOneCounterOnTheTarget,
			},
			{
				Label:   "−3: Target creature gains flying and double strike until end of turn.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying", "double strike")),
				Effect: effects.Do(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying", "double strike"},
						Label: "Ajani, Caller of the Pride — flying and double strike"}.Apply(ctx)
				})),
			},
			{
				Label: "−8: Create X 2/2 white Cat creature tokens, where X is your life total.",
				Cost:  effects.LoyaltyCost(-8),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
					if p == nil || p.Life <= 0 {
						return nil
					}
					return effects.CreateToken{Template: ctx.TextToken(effects.TokenCard("2/2 white Cat")), N: p.Life}.Apply(ctx)
				})),
			},
		},
	})
}
