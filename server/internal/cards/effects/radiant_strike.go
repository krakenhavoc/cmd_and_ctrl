package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Radiant Strike — Instant {3}{W} (Edge of Eternities, #4461):
//
//	"Destroy target artifact or tapped creature. You gain 3 life."
//
// Aerial Assault's destroy-then-gain-life with Royal Assassin's tapped
// creature (tappedPermanent) or any artifact as the target. The life is
// gained even if the target survives (an indestructible one); a creature
// untapped in response is no longer a legal target, so the spell doesn't
// resolve and no life is gained (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "186efd85-83d4-40cc-b617-76f32dbe99ce",
		Name:         "Radiant Strike",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or tapped creature", Or(Artifact(), And(Creature(), tappedPermanent()))),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Amount: 3}.Apply(ctx)
		},
	})
}
