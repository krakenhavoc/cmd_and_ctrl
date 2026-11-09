package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eternal Taskmaster — Creature — Zombie {1}{B}, 2/3 (Foundations
// Commander):
//
//	"This creature enters tapped.
//	 Whenever this creature attacks, you may pay {2}{B}. If you do, return
//	 target creature card from your graveyard to your hand."
//
// The entry is the CR 614.1d self replacement. The attack trigger is
// Veinwitch Coven's shape: the target is chosen as the trigger goes on the
// stack (CR 603.3d), and the {2}{B} is an optional cost paid during
// resolution, once only (printed ruling). A graveyard with no creature card
// puts no trigger on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7c000a8-1864-4bb2-80fa-a61d74dbd966",
		Name:         "Eternal Taskmaster",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Eternal Taskmaster — pay {2}{B} to return a creature card", func(g *game.Game, item *game.StackItem) error {
					if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
						return nil
					}
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}{B}",
						Question: "Eternal Taskmaster — pay {2}{B} to return a creature card from your graveyard to your hand?",
						Then:     eternalTaskmasterPaid,
					}.Apply(NewContext(g, item))
				}),
				TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
			),
		},
	})
}

// eternalTaskmasterPaid is the paid half, keyed: the card is the asking
// item's target.
var eternalTaskmasterPaid = PayUnlessBranches("pay-unless/eternal-taskmaster-return",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		if len(ctx.Item.Targets) == 0 || ctx.Item.Targets[0].Kind != game.TargetCard {
			return nil
		}
		return ReturnFromGraveyard{Target: ctx.Item.Targets[0].ID, Dest: game.ZoneHand}.Apply(ctx)
	}, nil)
