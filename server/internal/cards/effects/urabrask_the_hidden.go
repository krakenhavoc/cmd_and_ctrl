package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Urabrask the Hidden — Legendary Creature — Phyrexian Praetor
// {3}{R}{R}, 4/4 (EDHREC rank 1774):
//
//	"Creatures you control have haste.
//	 Creatures your opponents control enter tapped."
//
// The red Praetor. Haste is a Layer 6 grant over "creatures you
// control" (Concordant Crossroads without the world rule); the
// second line is Thalia, Heretic Cathar's replacement narrowed to
// creatures — the entering permanent's controller must not be
// Urabrask's. A cast, reanimated, fetched or flickered creature runs
// the entry pipeline and arrives tapped, as printed.
//
// An opponent's creature TOKEN enters tapped too, since #762: a
// created token now runs the same battlefield-entry pipeline every
// other permanent runs, so this replacement sees it exactly as it
// sees a cast creature. Kismet and Thalia gained the same reach in
// the same change.
func init() {
	Register(Spec{
		OracleID:     "5b2ffb53-86b7-4665-a5c7-b85b035b6c81",
		Name:         "Urabrask the Hidden",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16GrantKeywords(b16CreaturesYouControl, "haste"),
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
				return entering.IsCreature()
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.EntersTapped = true
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Urabrask the Hidden: enters tapped",
		}},
	})
}
