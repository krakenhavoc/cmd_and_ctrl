package cardspr

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pyromancer's Gauntlet — Artifact {5} (Magic 2014, S82 old sets):
//
//	"If a red instant or sorcery spell you control or a red planeswalker you
//	 control would deal damage to a permanent or player, it deals that much
//	 damage plus 2 to that permanent or player instead."
//
// A damage replacement (CR 614, CR 616.1) in Jaya's shape: the source is read as
// it deals the damage (CR 608.2h), must be red, controlled by the Gauntlet's
// controller and be an instant, a sorcery or a planeswalker. A red creature or
// an instant you do not control is not boosted, nor is damage of 0. Two
// Gauntlets add 4.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5cbb70e7-a52c-4b35-bc31-d9a0a7df893c",
		Name:         "Pyromancer's Gauntlet",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageTarget == uuid.Nil || src == nil {
					return false
				}
				ch, ok := effects.DamageSourceCharacteristics(ev, g)
				if !ok || ch.Controller != src.Controller || !slices.Contains(ch.Colors, "R") {
					return false
				}
				return effects.HasFold(ch.Types, "Instant") || effects.HasFold(ch.Types, "Sorcery") || effects.HasFold(ch.Types, "Planeswalker")
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount += 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Pyromancer's Gauntlet: +2 damage from a red instant, sorcery or planeswalker you control",
		}},
	})
}
