package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Archelos, Lagoon Mystic — Legendary Creature — Turtle Shaman
// {1}{B}{G}{U}, 2/4:
//
//	"As long as Archelos is tapped, other permanents enter tapped.
//	 As long as Archelos is untapped, other permanents enter untapped."
//
// One replacement effect (CR 614.1d / 614.12) on every other permanent
// entering the battlefield, whoever controls it and whatever it is (a
// land, a token, a creature put onto the battlefield): it sets "enters
// tapped" to Archelos's own tapped state as the permanent would enter,
// so the same effect is both printed sentences. Archelos's own entry is
// not affected ("other"). Being a non-self replacement, it applies after
// the entering permanent's own "enters tapped" (CR 616.1a), so with
// Archelos untapped a tapland enters untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fd19a009-d1ab-4148-a829-3d4c3af1b0b2",
		Name:         "Archelos, Lagoon Mystic",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield && ev.CardID != src.InstanceID
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
				ev.EntersTapped = src.Tapped
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Archelos, Lagoon Mystic: other permanents enter tapped while it is tapped and untapped while it is untapped",
		}},
	})
}
