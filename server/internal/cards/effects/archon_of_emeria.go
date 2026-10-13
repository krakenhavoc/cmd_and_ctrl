package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Archon of Emeria — Creature — Archon {2}{W}, 2/3:
//
//	"Flying
//	 Each player can't cast more than one spell each turn.
//	 Nonbasic lands your opponents control enter tapped."
//
// The second line is Rule of Law's restriction, derived from the Archon
// being on the battlefield, so it lifts the moment the Archon leaves.
// The third is Thalia, Heretic Cathar's replacement narrowed to
// nonbasic lands: a land played, fetched or put onto the battlefield
// under an opponent's control arrives tapped; the Archon's own
// controller's lands are untouched.
//
// No simplification. Mass Casualty deck request (#3245).
func init() {
	Register(Spec{
		OracleID:        "ceef2d5a-77ea-4e56-9806-fd1a2d5be400",
		Name:            "Archon of Emeria",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CastRestrictions: []game.CastRestriction{
			EachPlayerMaxSpellsPerTurn(1,
				"Archon of Emeria — each player can't cast more than one spell each turn."),
		},
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
				return b10NonbasicLand()(g, uuid.Nil, entering)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.EntersTapped = true
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Archon of Emeria: enters tapped",
		}},
	})
}
