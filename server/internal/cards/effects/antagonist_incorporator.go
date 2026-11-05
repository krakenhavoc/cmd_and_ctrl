package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Antagonist Incorporator — Artifact Creature — Borg {1}{W}, 2/3
// (Star Trek, #4973):
//
//	"Artifact creatures you control have "Whenever this creature deals
//	 combat damage to a player, you may pay {1}. If you do, create a 1/1
//	 colorless Borg artifact creature token.""
//
// A real layer-6 grant of a trigger bundle (ADR 0093): the trigger is
// each artifact creature's own, so "this creature" and "you" are the
// host and its controller, an artifact creature that loses its
// abilities loses this one, and the Incorporator (an artifact creature
// itself) has it too. The grant reads post-layer types, so an artifact
// that is a creature only for now (a crewed Vehicle) has it while it is
// one.
//
// "You may pay {1}. If you do" is a MayPay asked of the host's
// controller as the trigger resolves; declining, or a pool that cannot
// pay, makes no token.
//
// No simplification.
const antagonistIncorporatorGrant = "antagonist-incorporator/borg-on-combat-damage"

func init() {
	Register(Spec{
		OracleID:     "6f0b814b-94e6-48f2-9ccf-c376ca03bba3",
		Name:         "Antagonist Incorporator",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: antagonistIncorporatorGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisDealsCombatDamageToAPlayer("Antagonist Incorporator — you may pay {1} to create a Borg token",
					func(g *game.Game, item *game.StackItem) error {
						return MayPay{
							Chooser:  item.Controller,
							Cost:     "{1}",
							Question: "Pay {1} to create a 1/1 colorless Borg artifact creature token?",
							Then:     trekB1IncorporatorPaid,
						}.Apply(NewContext(g, item))
					}),
			},
			Text: "Whenever this creature deals combat damage to a player, you may pay {1}. If you do, create a 1/1 colorless Borg artifact creature token.",
		}},
		Static: []game.StaticAbility{
			GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsArtifact() && target.IsCreature() && target.Controller == source.Controller
			}, antagonistIncorporatorGrant),
		},
	})
}

var trekB1IncorporatorPaid = PayUnlessBranches("pay-unless/antagonist-incorporator-borg",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 colorless Borg artifact"), N: 1}.Apply(ctx)
	}, nil)
