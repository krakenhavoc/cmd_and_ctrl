package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goldnight Castigator — Creature — Angel {2}{R}{R}, 4/9 (Shadows over
// Innistrad, S82 OLD):
//
//	"Flying, haste
//	 If a source would deal damage to you, it deals double that damage to you
//	 instead.
//	 If a source would deal damage to this creature, it deals double that
//	 damage to this creature instead."
//
// Two replacement effects (CR 614.1) with disjoint recipients, as Gisela,
// Blade of Goldnight's doubling is one: any source counts, combat damage and
// burn alike, yours and theirs.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d46a6e0b-d39e-42ad-b95f-6902e1b72e62",
		Name:            "Goldnight Castigator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Replacements: []game.ReplacementEffect{
			{
				Watches:  []game.EventKind{game.EventDealDamage},
				RowLabel: "If a source would deal damage to you, it deals double that damage to you instead.",
				Label:    "Goldnight Castigator — double that damage to you",
				AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
					return src != nil && ev.Kind == game.RepEventDamage && ev.DamageAmount > 0 && effects.OldB287DamageToYou(ev.DamageTarget, src)
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.DamageAmount *= 2
					return nil
				},
				Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
			},
			{
				Watches:  []game.EventKind{game.EventDealDamage},
				RowLabel: "If a source would deal damage to this creature, it deals double that damage to this creature instead.",
				Label:    "Goldnight Castigator — double that damage to this creature",
				AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
					return src != nil && ev.Kind == game.RepEventDamage && ev.DamageAmount > 0 && ev.DamageTarget == src.InstanceID
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.DamageAmount *= 2
					return nil
				},
				Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
			},
		},
	})
}
