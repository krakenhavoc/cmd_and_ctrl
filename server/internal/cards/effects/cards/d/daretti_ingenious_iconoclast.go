package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Daretti, Ingenious Iconoclast — Legendary Planeswalker — Daretti {1}{B}{R},
// loyalty 3:
//
//	"+1: Create a 1/1 colorless Construct artifact creature token with
//	 defender.
//	 −1: You may sacrifice an artifact. If you do, destroy target artifact or
//	 creature.
//	 −6: Choose target artifact card in a graveyard or artifact on the
//	 battlefield. Create three tokens that are copies of it."
//
// −1 chooses its target as the ability goes on the stack (CR 601.2c), and the
// sacrifice is asked as it resolves; the destroy happens only if an artifact
// really went (CR 608.2c). Sacrificing the target itself leaves nothing to
// destroy. −6's one target is an artifact card in any graveyard or an
// artifact permanent; the copies are tokens (CR 707.2) whatever zone it is in.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "add0f7e7-8990-4fd1-8ac7-8115f7802912",
		Name:            "Daretti, Ingenious Iconoclast",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Create a 1/1 colorless Construct artifact creature token with defender.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Tokens: 1},
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Construct artifact with defender"), N: 1}),
			},
			{
				Label:   "−1: You may sacrifice an artifact. If you do, destroy target artifact or creature.",
				Cost:    effects.LoyaltyCost(-1),
				Targets: effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Daretti — sacrifice an artifact to destroy the target artifact or creature?",
						effects.Artifact(), false, darettiIngeniousIconoclastSacrificed)
				},
			},
			{
				Label:   "−6: Choose target artifact card in a graveyard or artifact on the battlefield. Create three tokens that are copies of it.",
				Cost:    effects.LoyaltyCost(-6),
				Targets: darettiIngeniousIconoclastCopyTarget(),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return effects.CreateTokenCopy{Controller: ctx.Controller(), Copy: t.ID, N: 3}.Apply(ctx)
						}
					}
					return nil
				})),
			},
		},
	})
}

// darettiIngeniousIconoclastSacrificed is the "if you do" of the −1: destroy
// the target, which is judged again as it resolves (CR 608.2b).
var darettiIngeniousIconoclastSacrificed = effects.RegisterSacrificedOne("prompt-run/daretti-ingenious-iconoclast-sacrifice",
	func(ctx *effects.Context, _ uuid.UUID) error {
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
			}
		}
		return nil
	})

func darettiIngeniousIconoclastCopyTarget() *game.TargetSpec {
	spec := effects.TargetPermanent("target artifact card in a graveyard or artifact on the battlefield", effects.Artifact())
	spec.Zones = []game.ZoneKind{game.ZoneBattlefield, game.ZoneGraveyard}
	return spec
}
