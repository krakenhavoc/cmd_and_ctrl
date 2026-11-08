package effects

import "github.com/google/uuid"

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
	Register(Spec{
		OracleID:        "1d5ff280-7f4d-4801-aefd-9b6e6b1f2818",
		Name:            "Ajani, Caller of the Pride",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Put a +1/+1 counter on up to one target creature.",
				Cost:    LoyaltyCost(1),
				Targets: TargetCreature("up to one target creature").WithCount(0, 1),
				Effect:  PlusOneCounterOnTheTarget,
			},
			{
				Label:   "−3: Target creature gains flying and double strike until end of turn.",
				Cost:    LoyaltyCost(-3),
				Targets: TargetCreature("target creature"),
				Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "flying", "double strike")),
				Effect: Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
					return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying", "double strike"},
						Label: "Ajani, Caller of the Pride — flying and double strike"}.Apply(ctx)
				})),
			},
			{
				Label: "−8: Create X 2/2 white Cat creature tokens, where X is your life total.",
				Cost:  LoyaltyCost(-8),
				Effect: Do(Instruction(func(ctx *Context) error {
					p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
					if p == nil || p.Life <= 0 {
						return nil
					}
					return CreateToken{Template: ctx.TextToken(TokenCard("2/2 white Cat")), N: p.Life}.Apply(ctx)
				})),
			},
		},
	})
}
