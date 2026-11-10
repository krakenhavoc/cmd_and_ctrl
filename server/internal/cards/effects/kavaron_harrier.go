package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kavaron Harrier — Artifact Creature — Robot Soldier {R}, 2/1 (Edge of
// Eternities, #4461):
//
//	"Whenever this creature attacks, you may pay {2}. If you do, create
//	 a 2/2 colorless Robot artifact creature token that's tapped and
//	 attacking. Sacrifice that token at end of combat."
//
// The optional cost is MayPay, held to the declare attackers step so the
// token can't arrive after combat has moved on, and the paid branch runs
// only when the {2} was really paid. The token is Dorothea's
// Retribution's: put onto the battlefield tapped and attacking, never
// declared (CR 508.4), with its controller choosing what it attacks (a
// bot sends it where the Harrier went), and a delayed trigger at the
// beginning of this combat's end of combat step sacrifices exactly the
// tokens that landed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1972e0db-e595-4ec2-b97c-902561bb275b",
		Name:         "Kavaron Harrier",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Kavaron Harrier — you may pay {2} for a tapped and attacking 2/2 Robot",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:    item.Controller,
						Cost:       "{2}",
						Question:   "Kavaron Harrier — pay {2} to create a 2/2 Robot that's tapped and attacking?",
						InThisStep: true,
						Then:       kavaronHarrierPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// kavaronHarrierPaid is "if you do": the attacking Robot, keyed. An
// on-disk key.
var kavaronHarrierPaid = PayUnlessBranches("pay-unless/kavaron-harrier-robot",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return CreateTokensAttackingYourChoice{
			Controller: ctx.Controller(),
			Template:   TokenCard("2/2 colorless Robot artifact"),
			N:          1,
			Tapped:     true,
			Prefer:     triggerAttackTarget(ctx.Item),
			Label:      "Kavaron Harrier",
			Tail:       game.KeyedTail{Ref: kavaronHarrierSacrifice},
		}.Apply(ctx)
	}, nil)

// kavaronHarrierSacrifice is the Robot's tail, keyed: the end of combat
// trigger that sacrifices it. An on-disk key.
var kavaronHarrierSacrifice = AttackingTokensThen("attacking-tokens/kavaron-harrier-sacrifice-the-robot",
	func(ctx *Context, created []uuid.UUID, _ game.KeyedTail) error {
		return sacrificeAtEndOfCombat(ctx, created, "Kavaron Harrier — sacrifice the Robot")
	})
