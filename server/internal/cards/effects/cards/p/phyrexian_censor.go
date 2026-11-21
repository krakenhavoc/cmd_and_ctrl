package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phyrexian Censor — Creature — Phyrexian Wizard {2}{W}, 3/3 (March of
// the Machine):
//
//	"Each player can't cast more than one non-Phyrexian spell each turn.
//	 Non-Phyrexian creatures enter tapped."
//
// The first sentence is Archon of Emeria's cast restriction with a
// filtered count: the spell being announced must itself be non-Phyrexian,
// and the player must already have cast one this turn
// (CastTally.NonPhyrexian, counted as each spell became cast, CR 601.2i).
// A Phyrexian spell is never refused and never counted. The second is an
// entry replacement (CR 614.1c) for every creature, any controller, that
// is not a Phyrexian as it enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c77de57-29fd-45ba-b2c7-eef231c6b7c7",
		Name:         "Phyrexian Censor",
		Completeness: effects.CompletenessFull,
		CastRestrictions: []game.CastRestriction{{
			Label: "Phyrexian Censor — each player can't cast more than one non-Phyrexian spell each turn.",
			Forbids: func(q game.CastQuery) bool {
				if q.Card.HasSubtype(q.Source.TextSubtype("Phyrexian")) {
					return false
				}
				return q.Game.CastTallyFor(q.Controller).NonPhyrexian >= 1
			},
		}},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield {
					return false
				}
				entering, ok := g.LookupCardForEffect(ev.CardID)
				return ok && entering.IsCreature() && !entering.HasSubtype(src.TextSubtype("Phyrexian"))
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.EntersTapped = true
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Phyrexian Censor: non-Phyrexian creatures enter tapped",
		}},
	})
}
