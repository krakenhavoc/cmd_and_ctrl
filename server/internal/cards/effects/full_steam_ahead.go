package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Full Steam Ahead — Sorcery {3}{G}{G} (Outlaws of Thunder Junction):
//
//	"Until end of turn, each creature you control gets +2/+2 and gains
//	 trample and "This creature can't be blocked by more than one
//	 creature.""
//
// One record over the creatures you control as it resolves (CR 611.2c):
// the pump, the keyword and the block-count rule (game.ModCantBeBlockedByMoreThan,
// the seam this card needed: MaxBlockers as a duration effect rather than
// a printed rule) share one timestamp and one set, so a creature that
// arrives later gets none of it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "65e64c50-4757-4e8e-8043-bc61a5f21750",
		Name:         "Full Steam Ahead",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ScopedEffectFor{
				Match: And(Creature(), YouControl()),
				Mods: []game.Mod{
					game.ModifyPTMod(2, 2),
					game.AddKeywordsMod("trample"),
					game.CantBeBlockedByMoreThanMod(1),
				},
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Full Steam Ahead — +2/+2, trample, can't be blocked by more than one creature",
			}.Apply(ctx)
		},
	})
}
