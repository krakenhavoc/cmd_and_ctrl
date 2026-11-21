package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b03_helpers.go — shared pieces of S82 GAP batch 3 (#5948).

// enchantedPlayerLosesTwiceThatMuchLife is "If enchanted player would lose
// life, they lose twice that much life instead" (Selenia's Curse): a
// CR 614 replacement on a life LOSS (a negative delta) by the player the
// source Aura is attached to. Damage is a loss too (CR 120.3a), so the
// doubling reaches it; the damage dealt is unchanged. The enchanted
// player is the affected player and orders competing replacements.
func enchantedPlayerLosesTwiceThatMuchLife(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventChangeLife},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventLife || ev.LifeDelta >= 0 || src == nil {
				return false
			}
			return src.AttachedTo.Kind == game.TargetPlayer && src.AttachedTo.ID == ev.LifePlayer
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.LifeDelta = timesTwo(ev.LifeDelta)
			return nil
		},
		Controller: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) uuid.UUID {
			return ev.LifePlayer
		},
		Label: label,
	}
}
