package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thalia, Heretic Cathar — Legendary Creature — Human Soldier {2}{W},
// 3/2 (EDHREC rank 1473):
//
//	"First strike
//	 Creatures and nonbasic lands your opponents control enter
//	 tapped."
//
// The tempo hatebear. Kismet's replacement narrowed to creatures and
// NONBASIC lands: the entering permanent's controller must not be
// Thalia's, and "nonbasic" is the Basic supertype read off the
// effective characteristic — a land Urborg made a Swamp is still
// nonbasic, a Snow-Covered Plains is still basic. Every entry path
// runs the pipeline (cast, played, fetched, reanimated, a token), so
// an opponent's fetched shockland and their Zombie tokens all arrive
// tapped, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "13be47e4-9da2-4f9e-baf0-db96f7777ccb",
		Name:            "Thalia, Heretic Cathar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield {
					return false
				}
				entering, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || entering.Controller == src.Controller {
					return false
				}
				return entering.IsCreature() || b10NonbasicLand()(g, uuid.Nil, entering)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.EntersTapped = true
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Thalia, Heretic Cathar: enters tapped",
		}},
	})
}
