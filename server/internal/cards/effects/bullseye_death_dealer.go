package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bullseye, Death Dealer — Legendary Creature — Human Assassin Villain
// {2}{B/R}, 2/3:
//
//	"When Bullseye enters, you may sacrifice an artifact or discard a
//	 nonland card. When you do, Bullseye deals 2 damage to any target.
//	 {3}, {T}, Sacrifice an artifact or discard a nonland card: Bullseye
//	 deals 2 damage to any target."
//
// Both halves are #4566's either/or costs. The enters trigger is a may-pay
// whose payment has two branches (SacrificeAnArtifactOrDiscard): the
// player names one, a branch they can't pay is not offered, and paying
// creates the CR 603.12 reflexive trigger, which picks its target as it
// goes on the stack (CR 603.3d). The activated ability is one ability with
// an either/or cost (OrCost, CR 602.2b / 601.2b): the activator announces
// the branch with the activation, and with neither payable it can't be
// activated (CR 118.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "10db57b4-f895-4514-b610-2252ba676353",
		Name:         "Bullseye, Death Dealer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Bullseye, Death Dealer — you may sacrifice an artifact or discard a nonland card; when you do, 2 damage to any target",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MayPay{
						Chooser:  ctx.Controller(),
						Pay:      SacrificeAnArtifactOrDiscard(true),
						Question: "Bullseye, Death Dealer — sacrifice an artifact or discard a nonland card to deal 2 damage to any target?",
						Then:     bullseyePaid,
					}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{{
			Label: "{3}, {T}, Sacrifice an artifact or discard a nonland card: Bullseye deals 2 damage to any target.",
			Cost: Plus(ManaCost("{3}"), TapCost(), OrCost(
				SacrificeBranch(1, "an artifact", game.PermanentQuery{Types: []string{"artifact"}}),
				DiscardBranch(1, "a nonland card", true),
			)),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Effect:  sourceDealsDamageToEachLegalTarget(2),
		}},
	})
}

// bullseyePaid is the enters trigger's "when you do": the reflexive
// trigger that deals the 2 damage.
var bullseyePaid = PayUnlessBranches("pay-unless/bullseye-death-dealer-damage",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return ReflexiveTrigger{
			Label:  "Bullseye, Death Dealer — 2 damage to any target",
			Body:   bullseyeDamageBody,
			Params: game.EffectParams{Amount: 2},
		}.Apply(ctx)
	}, nil)
