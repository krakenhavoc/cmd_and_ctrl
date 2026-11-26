package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Angrath's Rampage — Sorcery {B}{R} (War of the Spark, #5520):
//
//	"Choose one —
//	 • Target player sacrifices an artifact of their choice.
//	 • Target player sacrifices a creature of their choice.
//	 • Target player sacrifices a planeswalker of their choice."
//
// Rush of Dread's per-bullet target player and its sacrifice seam
// (PlayersSacrificeForEffect): the player picks the permanent as the spell
// resolves; a player with none sacrifices nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "850a8369-94ab-43b6-8cfe-286e3623b19d",
		Name:         "Angrath's Rampage",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			angrathsRampageMode("artifact", effects.Artifact()),
			angrathsRampageMode("creature", effects.Creature()),
			angrathsRampageMode("planeswalker", effects.Planeswalker()),
		),
	})
}

func angrathsRampageMode(kind string, match effects.CardPredicate) game.ModeOption {
	return effects.ModeDoing("Target player sacrifices "+aOrAn(kind)+" "+kind+" of their choice.",
		effects.TargetPlayer("target player"),
		func(_ *game.StackItem, ctx *effects.Context, occ int) error {
			t, ok := effects.ModeTarget(ctx, occ)
			if !ok || t.Kind != game.TargetPlayer || !ctx.IsTargetLegal(t) {
				return nil
			}
			ctx.Game.PlayersSacrificeForEffect(ctx.Source(), []uuid.UUID{t.ID}, effects.SacrificeSpec(aOrAn(kind)+" "+kind, match),
				"Angrath's Rampage — sacrifice "+aOrAn(kind)+" "+kind)
			return nil
		})
}

func aOrAn(kind string) string {
	if kind == "artifact" {
		return "an"
	}
	return "a"
}
