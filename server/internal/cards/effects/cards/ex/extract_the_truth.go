package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Extract the Truth — Sorcery {1}{B}:
//
//	"Choose one —
//	 • Target opponent reveals their hand. You may choose a creature,
//	   enchantment, or planeswalker card from it. That player discards
//	   that card.
//	 • Target opponent sacrifices an enchantment of their choice."
//
// The first bullet is the optional revealed-hand pick (#2115, ADR
// 0116's 2026-10-05 amendment): the whole table sees the hand (CR
// 701.20a), and you may choose a creature, enchantment or planeswalker
// card — or nothing, even when there is one to take (its 2022-04-29
// ruling). A chosen card is discarded.
//
// The second bullet is an ordinary edict: the opponent chooses one of
// their enchantments as it resolves, and nothing is targeted but the
// player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "18207fe8-41e4-418d-b5f9-e1239c527e67",
		Name:         "Extract the Truth",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Target opponent reveals their hand. You may choose a creature, enchantment, or planeswalker card from it. That player discards that card.",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				extractTheTruthPick),
			effects.ModeDoing("Target opponent sacrifices an enchantment of their choice.",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				extractTheTruthEdict),
		),
	})
}

// extractTheTruthPick is the first bullet.
func extractTheTruthPick(_ *game.StackItem, ctx *effects.Context, occ int) error {
	t, ok := effects.ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	return effects.ChooseFromRevealedHand{
		Player:   t.ID,
		Filter:   effects.Or(effects.Creature(), effects.Enchantment(), effects.Planeswalker()),
		Label:    "creature, enchantment, or planeswalker card",
		Optional: true,
	}.Apply(ctx)
}

// extractTheTruthEdict is the second bullet.
func extractTheTruthEdict(_ *game.StackItem, ctx *effects.Context, occ int) error {
	t, ok := effects.ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	ctx.Game.PlayerSacrificesForEffect(ctx.Source(), t.ID,
		effects.SacrificeSpec("an enchantment", effects.Enchantment()),
		"Extract the Truth — sacrifice an enchantment")
	return nil
}
