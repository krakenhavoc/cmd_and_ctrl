package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sheoldred's Edict — Instant {1}{B} (EDHREC rank 1151):
//
//	"Choose one —
//	 • Each opponent sacrifices a nontoken creature of their choice.
//	 • Each opponent sacrifices a creature token of their choice.
//	 • Each opponent sacrifices a planeswalker of their choice."
//
// The modal edict. Three modes, none targeted, each the
// EachPlayerSacrifices fan-out with a different predicate — nontoken
// creature (Accursed Marauder's clause), creature token, planeswalker
// — so every opponent picks their own, a hexproof creature is still a
// legal pick, and an opponent with nothing matching sacrifices
// nothing. Instant speed, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "217062f5-96f1-454c-9507-17f34ef37070",
		Name:         "Sheoldred's Edict",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Each opponent sacrifices a nontoken creature of their choice."),
			Mode("Each opponent sacrifices a creature token of their choice."),
			Mode("Each opponent sacrifices a planeswalker of their choice."),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.HasMode(0) {
				return EachPlayerSacrifices{ExceptController: true, Match: b04NontokenCreature, Label: "a nontoken creature"}.Apply(ctx)
			}
			if ctx.HasMode(1) {
				return EachPlayerSacrifices{ExceptController: true, Match: And(Creature(), IsTokenPredicate()), Label: "a creature token"}.Apply(ctx)
			}
			if ctx.HasMode(2) {
				return EachPlayerSacrifices{ExceptController: true, Match: Planeswalker(), Label: "a planeswalker"}.Apply(ctx)
			}
			return nil
		},
	})
}
