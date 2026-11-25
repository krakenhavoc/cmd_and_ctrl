package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pharika's Libation — Instant {2}{B} (Theros Beyond Death, #5514):
//
//	"Choose one —
//	 • Target opponent sacrifices a creature of their choice.
//	 • Target opponent sacrifices an enchantment of their choice."
//
// Two edict modes, each with its own target opponent chosen as the spell is cast
// (CR 700.2). The sacrifice is that player's choice among their own permanents
// as the spell resolves, and nothing happens if the target is no longer legal
// (CR 608.2b).
//
// No simplification.
func init() {
	edict := func(what string, match effects.CardPredicate) game.ModeOption {
		return effects.ModeWithPurpose(effects.ModeDoing("Target opponent sacrifices "+what+" of their choice.",
			effects.TargetPlayer("target opponent", effects.Opponent()),
			func(item *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				ctx.Game.PlayerSacrificesForEffect(item.SourceCardID, t.ID, effects.SacrificeSpec(what, match), "Pharika's Libation — sacrifice "+what)
				return nil
			}), game.Purpose{Answers: game.AnswerRemove})
	}
	effects.Register(effects.Spec{
		OracleID:     "32c84176-cf11-4010-b1cc-ed570a8c3e54",
		Name:         "Pharika's Libation",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			edict("a creature", effects.Creature()),
			edict("an enchantment", effects.Enchantment()),
		),
	})
}
