package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Root Maze — Enchantment {G} (Special Guests):
//
//	"Artifacts and lands enter tapped."
//
// Kismet's replacement without its "an opponent controls" clause: every
// artifact or land that would enter the battlefield enters tapped,
// whoever controls it (CR 614.1c).
func init() {
	Register(Spec{
		OracleID:     "2f4abc5e-5aad-455e-b1ec-137d7ac4a10c",
		Name:         "Root Maze",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			{
				Watches: []game.EventKind{game.EventZoneMove},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
					if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield {
						return false
					}
					entering, ok := g.LookupCardForEffect(ev.CardID)
					return ok && (entering.IsArtifact() || entering.IsLand())
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.EntersTapped = true
					return nil
				},
				Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: "Root Maze: artifacts and lands enter tapped",
			},
		},
	})
}
