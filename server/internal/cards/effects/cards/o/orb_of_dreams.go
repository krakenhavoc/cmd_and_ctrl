package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Orb of Dreams — Artifact {3} (Betrayers of Kamigawa, S82 OLD, #5602):
//
//	"Permanents enter tapped."
//
// Archelos's replacement effect with the tapped state fixed: every
// permanent entering the battlefield, whoever controls it and whatever it
// is (a land, a token, a creature put onto the battlefield) enters tapped
// (CR 614.1d / 614.12). The Orb applies to itself as it enters (CR 614.12),
// so it enters tapped, unlike Archelos, whose text says "other".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c24b7c3-7c90-4109-a4aa-d9cae6febf34",
		Name:         "Orb of Dreams",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.EntersTapped = true
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Orb of Dreams: permanents enter tapped",
		}},
	})
}
