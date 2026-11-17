package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feast or Famine — Instant {3}{B} (Alliances, #5634):
//
//	"Choose one —
//	 • Create a 2/2 black Zombie creature token.
//	 • Destroy target nonartifact, nonblack creature. It can't be
//	   regenerated."
//
// A modal spell: the mode is chosen as it is cast (CR 700.2a), and only
// the second has a target. That target must still be a nonartifact,
// nonblack creature as Feast or Famine resolves (CR 608.2b); the
// destruction (CR 701.8a) is not replaced by a regeneration shield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "485f710d-5bc2-4c8f-b7e9-eed6caa261a6",
		Name:         "Feast or Famine",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Create a 2/2 black Zombie creature token.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return CreateToken{Template: TokenCard("2/2 black Zombie"), N: 1}.Apply(ctx)
				}), game.Purpose{Tokens: 1, CreatureTokens: 1}),
			ModeWithPurpose(ModeDoing("Destroy target nonartifact, nonblack creature. It can't be regenerated.",
				TargetCreature("target nonartifact, nonblack creature", Not(Artifact()), NonBlack()),
				func(_ *game.StackItem, ctx *Context, occurrence int) error {
					t, ok := ModeTarget(ctx, occurrence)
					if !ok {
						return nil
					}
					return DestroyTarget{Target: t.ID, CantBeRegenerated: true}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerRemove}),
		),
	})
}
