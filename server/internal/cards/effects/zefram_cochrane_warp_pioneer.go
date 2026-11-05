package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zefram Cochrane, Warp Pioneer — Legendary Creature — Human Pilot {2}{U},
// 2/3 (Star Trek):
//
//	You may look at the top card of your library any time.
//	Once during each of your turns, you may cast an artifact spell from
//	the top of your library.
//	Spacecraft you control enter with two additional charge counters on
//	them.
//
// The peek is the library's top-visible-to-owner flag (CR 401.5). The cast
// is a standing library permission (ADR 0066) on the top card only, cast
// only (an artifact land is not played this way), with Gisa and Geralf's
// once-a-turn budget: PerType names the one type "artifact", so one artifact
// spell per turn per Zefram object, and TimingYourTurnOnly closes it on
// other turns while the spell keeps its own timing. The counters are a CR
// 614.12 entry replacement, Curator Beastie's shape: a Spacecraft entering
// under Zefram's controller gets two extra charge counters as it enters, so
// a Spacecraft with a station threshold of 2 or less arrives already a
// creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "fea67446-f413-4a8c-b79a-49172e798fff",
		Name:              "Zefram Cochrane, Warp Pioneer",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{{
			Zone:             game.ZoneLibrary,
			TopOfLibraryOnly: true,
			CastOnly:         true,
			Filter:           game.PermissionFilter{CardType: game.PermissionTypeArtifact},
			PerType:          []string{game.PermissionTypeArtifact},
			Timing:           game.TimingYourTurnOnly,
			Label:            "Once during each of your turns, cast an artifact spell from the top of your library (Zefram Cochrane)",
		}},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil {
					return false
				}
				entering, ok := g.LookupCardForEffect(ev.CardID)
				return ok && entering.Controller == src.Controller && entering.HasSubtype("Spacecraft")
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.AddCounterAtETB(game.CounterCharge, 2)
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Zefram Cochrane: Spacecraft you control enter with two additional charge counters",
		}},
	})
}
