package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// tmt_b6_helpers.go — shared bodies for Teenage Mutant Ninja Turtles,
// batch 6 (#4457).

// tmtB6CounterThenGrant is "Put a +1/+1 counter on target creature you
// control. It gains <keywords> until end of turn." as a spell's whole
// OnResolve (Saved by the Shell; Snakeskin Veil's shape with more
// keywords). Each sentence is a step of its own, so the grant waits for a
// prompt the counter asks (CR 608.2c). A target that is no longer legal
// gets neither (CR 608.2b).
func tmtB6CounterThenGrant(label string, keywords ...string) func(*game.StackItem, *Context) error {
	return ResolveSteps(
		Instruction(func(ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}
			}
			return nil
		}),
		Instruction(func(ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return GrantKeywordUntilEOT{Target: t.ID, Keywords: keywords, Label: label}.Apply(ctx)
				}
			}
			return nil
		}),
	)
}

// tmtB6DamageAnyTargetAndYou is "it deals <n> damage to any target and
// <self> damage to you" as an enters trigger's body (Spicy Oatmeal Pizza;
// Daredevil's Billy Club's shape with other amounts). The source deals both
// at once, as one damage instance.
func tmtB6DamageAnyTargetAndYou(n, self int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		return g.DamageInstanceForEffect(func() error {
			for _, t := range ctx.LegalTargets() {
				if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}).Apply(ctx); err != nil {
					return err
				}
			}
			return DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: self}.Apply(ctx)
		})
	}
}

// tmtB6ModeTargetDraws is "target player draws <n> cards" as one step of a
// modal bullet built with ModeSteps (Shredder's Revenge).
func tmtB6ModeTargetDraws(n int) Applier {
	return Instruction(func(ctx *Context) error {
		t, ok := ModeTarget(ctx, ctx.Occurrence())
		if !ok || t.Kind != game.TargetPlayer {
			return nil
		}
		return DrawCards{Player: t.ID, N: n}.Apply(ctx)
	})
}

// tmtB6ModeTargetLosesLife is "and loses <n> life" as the next step of the
// same bullet: a step of its own, so it waits for a prompt the draws ask
// (CR 608.2c).
func tmtB6ModeTargetLosesLife(n int) Applier {
	return Instruction(func(ctx *Context) error {
		t, ok := ModeTarget(ctx, ctx.Occurrence())
		if !ok || t.Kind != game.TargetPlayer {
			return nil
		}
		return GainLife{Player: t.ID, Amount: -n}.Apply(ctx)
	})
}
