package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Charge of the Mites — Instant {2}{W}:
//
//	"Choose one —
//	 • Charge of the Mites deals damage equal to the number of creatures you
//	   control to target creature or planeswalker.
//	 • Create two 1/1 colorless Phyrexian Mite artifact creature tokens with
//	   toxic 1 and 'This token can't block.'"
//
// The damage reads the creature count as the spell resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa47c190-d66a-41b0-aa08-21036cda4223",
		Name:         "Charge of the Mites",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 2, CreatureTokens: 2},
		Modes: ChooseOne(
			Mode("Charge of the Mites deals damage equal to the number of creatures you control to target creature or planeswalker.",
				TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker()))),
			Mode("Create two 1/1 colorless Phyrexian Mite artifact creature tokens with toxic 1 and \"This token can't block.\""),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.HasMode(0) {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				n := 0
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.IsCreature() && c.Controller == ctx.Controller() {
						n++
					}
				}
				return DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: n}.Apply(ctx)
			}
			if ctx.HasMode(1) {
				return CreateToken{Template: PhyrexianMiteToken(), N: 2}.Apply(ctx)
			}
			return nil
		},
	})
}
