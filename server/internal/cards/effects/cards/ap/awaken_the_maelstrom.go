package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     effects.InvasionOfAlaraOracleID + "#1",
		Name:         "Awaken the Maelstrom",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetPlayer("target player"),
			effects.TargetPermanent("target permanent an opponent controls", effects.OpponentControls()),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetPlayer {
					return effects.DrawCards{Player: t.ID, N: 2}.Apply(ctx)
				}
				return nil
			}),
			effects.PutFromHandOntoBattlefield{
				Match:    effects.Artifact(),
				Optional: true,
				Label:    "Awaken the Maelstrom — you may put an artifact card from your hand onto the battlefield",
			},
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.ChoosePermanents{
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
			effects.Instruction(func(ctx *effects.Context) error { return effects.MatDistributePlusOneCounters(ctx, 3) }),
			effects.Instruction(func(ctx *effects.Context) error {
				if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
					return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
				}
				return nil
			}),
		),
	})
}

// awakenTheMaelstromCopy is "create a token that's a copy of a permanent you
// control", keyed. An on-disk key.
var awakenTheMaelstromCopy = effects.RegisterPermanentsPick("prompt-run/awaken-the-maelstrom-copy", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		ids := picked.Cards()
		if len(ids) == 0 {
			return nil
		}
		return effects.CreateTokenCopy{Controller: ctx.Controller(), Copy: ids[0], N: 1}.Apply(ctx)
	},
})
