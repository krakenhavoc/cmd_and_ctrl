package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Awaken the Maelstrom — Sorcery, the back face of Invasion of Alara
// (March of the Machine, S82 batch 23):
//
//	"Awaken the Maelstrom is all colors.
//	 Target player draws two cards. You may put an artifact card from your
//	 hand onto the battlefield. Create a token that's a copy of a permanent
//	 you control. Distribute three +1/+1 counters among one, two, or three
//	 creatures you control. Destroy target permanent an opponent controls."
//
// Cast transformed from exile, the face is a sorcery spell whose colours
// are all five (the face data carries them, as the first line says). Two
// targets, announced together (CR 601.2c): a player and a permanent an
// opponent controls; each half applies only if its target is still legal as
// the spell resolves (CR 608.2b). The five instructions run in printed order
// and each waits for the prompts of the one before (CR 608.2c): the draw;
// the optional put from the hand (Aether Vial's family of puts, any artifact
// card, the caster's own hand); the copy, chosen among every permanent the
// caster then controls, the artifact just put included, and not targeted;
// the three counters, chosen and split at resolution (Feast of the
// Victorious Dead's distribution); the destruction.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     invasionOfAlaraOracleID + "#1",
		Name:         "Awaken the Maelstrom",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPlayer("target player"),
			TargetPermanent("target permanent an opponent controls", OpponentControls()),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetPlayer {
					return DrawCards{Player: t.ID, N: 2}.Apply(ctx)
				}
				return nil
			}),
			PutFromHandOntoBattlefield{
				Match:    Artifact(),
				Optional: true,
				Label:    "Awaken the Maelstrom — you may put an artifact card from your hand onto the battlefield",
			},
			Instruction(func(ctx *Context) error {
				return ChoosePermanents{
					Question: "Awaken the Maelstrom — choose a permanent you control; create a token that's a copy of it",
					Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
						var out []uuid.UUID
						for _, c := range g.BattlefieldCardsForEffect() {
							if c.Controller == of {
								out = append(out, c.InstanceID)
							}
						}
						return out, 1, 1
					},
					Keyed: awakenTheMaelstromCopy,
				}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error { return matDistributePlusOneCounters(ctx, 3) }),
			Instruction(func(ctx *Context) error {
				if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
					return DestroyTarget{Target: t.ID}.Apply(ctx)
				}
				return nil
			}),
		),
	})
}

// awakenTheMaelstromCopy is "create a token that's a copy of a permanent you
// control", keyed. An on-disk key.
var awakenTheMaelstromCopy = RegisterPermanentsPick("prompt-run/awaken-the-maelstrom-copy", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		ids := picked.Cards()
		if len(ids) == 0 {
			return nil
		}
		return CreateTokenCopy{Controller: ctx.Controller(), Copy: ids[0], N: 1}.Apply(ctx)
	},
})
