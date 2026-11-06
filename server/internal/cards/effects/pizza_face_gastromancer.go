package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pizza Face, Gastromancer — {3}{B}{G} Legendary Artifact Creature — Food
// Mutant 2/4 (Teenage Mutant Ninja Turtles, #4457):
//
//	"When Pizza Face enters, create a Food token.
//	 Disappear — At the beginning of your end step, if a permanent left
//	 the battlefield under your control this turn, put three +1/+1
//	 counters on up to one other target artifact or creature. If it isn't
//	 a creature, it becomes a 0/0 Mutant creature in addition to its other
//	 types.
//	 {10}, {T}, Sacrifice Pizza Face: You gain 15 life."
//
// Disappear is an ability word (CR 207.2c) for revolt's condition: the
// end-step trigger is AtYourEndStepIfRevolt's condition written as an
// InterveningIf, so its steps can wait for a prompt. It is checked as the
// step begins and again as the trigger resolves (CR 603.4). The target is
// "other": never Pizza Face itself (Another).
//
// The counters go on first, as printed, and only then is the target
// checked: a noncreature artifact becomes a 0/0 Mutant creature in
// addition to its other types with no duration, Tezzeret, Cruel Captain's
// indefinite type change pinned to that object (CR 611.2a), so the three
// counters make it a 3/3. Because the counters land while it is still a
// noncreature, an effect that changes counters put on a creature does not
// see them, as printed.
//
// The last line is a Food's sacrifice with a bigger cost and payoff.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e71479c9-bb38-4112-bf8a-41b39a3ade51",
		Name:         "Pizza Face, Gastromancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Pizza Face, Gastromancer — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
			Targeting(
				InterveningIf(CondBoard{
					Name: "a permanent left the battlefield under your control this turn",
					Test: func(ctx *Context) bool { return Revolt(ctx.Game, ctx.Controller()) },
				}, On(game.EventBeginEndStep, AllOf(ByYou, YouHaveRevolt),
					"Disappear — put three +1/+1 counters on up to one other target artifact or creature", pizzaFaceDisappear)),
				Another(TargetPermanent("up to one other target artifact or creature", Or(Artifact(), Creature())).WithCount(0, 1))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{10}, {T}, Sacrifice Pizza Face: You gain 15 life.",
			Cost:    Plus(ManaCost("{10}"), TapCost(), SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GainLife{Amount: 15}.Apply(NewContext(g, item))
			},
		}},
	})
}

// pizzaFaceDisappear is the end-step trigger's steps: the counters on the
// target, then, if it is not a creature, the type change. The second step
// waits for any question the counters ask (CR 608.2c, #4155).
var pizzaFaceDisappear = Do(
	Instruction(func(ctx *Context) error {
		target := FirstLegalBattlefieldTarget(ctx)
		if target == uuid.Nil {
			return nil
		}
		return AddCounter{Target: target, Kind: game.CounterPlusOne, N: 3}.Apply(ctx)
	}),
	Instruction(func(ctx *Context) error {
		target := FirstLegalBattlefieldTarget(ctx)
		if target == uuid.Nil {
			return nil
		}
		c, ok := ctx.Game.LookupCardForEffect(target)
		if !ok || c.IsCreature() {
			return nil
		}
		return ScopedEffectFor{
			Target: target,
			Mods: []game.Mod{
				game.AddTypesMod("Creature"),
				game.AddSubtypesMod(ctx.TextSubtypes("Mutant")...),
				game.SetBasePowerMod(0),
				game.SetBaseToughnessMod(0),
			},
			Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), target),
			Label:    "Pizza Face, Gastromancer — becomes a 0/0 Mutant creature",
		}.Apply(ctx)
	}),
)
