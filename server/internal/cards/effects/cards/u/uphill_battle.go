package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Uphill Battle — Enchantment {2}{R} (Mercadian Masques, #5619):
//
//	"Creatures played by your opponents enter tapped."
//
// Thalia, Heretic Cathar's entry replacement, narrowed to creatures that
// arrive from the stack: a creature is "played" by being cast, so a
// reanimated creature or a token an opponent makes enters untapped. The
// entering creature's controller must be an opponent of Uphill Battle's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "349a18bf-5c61-4fe6-a5e5-5e7cad6b4ba6",
		Name:         "Uphill Battle",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || ev.OldZone != game.ZoneStack {
					return false
				}
				entering, ok := g.LookupCardForEffect(ev.CardID)
				return ok && entering.IsCreature() && entering.Controller != src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.EntersTapped = true
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Uphill Battle: creatures played by your opponents enter tapped",
		}},
	})
}
