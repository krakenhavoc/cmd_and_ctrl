package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "217062f5-96f1-454c-9507-17f34ef37070",
		Name:         "Sheoldred's Edict",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Each opponent sacrifices a nontoken creature of their choice."),
			effects.Mode("Each opponent sacrifices a creature token of their choice."),
			effects.Mode("Each opponent sacrifices a planeswalker of their choice."),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.HasMode(0) {
				return effects.EachPlayerSacrifices{ExceptController: true, Match: effects.B04NontokenCreature, Label: "a nontoken creature"}.Apply(ctx)
			}
			if ctx.HasMode(1) {
				return effects.EachPlayerSacrifices{ExceptController: true, Match: effects.And(effects.Creature(), effects.IsTokenPredicate()), Label: "a creature token"}.Apply(ctx)
			}
			if ctx.HasMode(2) {
				return effects.EachPlayerSacrifices{ExceptController: true, Match: effects.Planeswalker(), Label: "a planeswalker"}.Apply(ctx)
			}
			return nil
		},
	})
}
