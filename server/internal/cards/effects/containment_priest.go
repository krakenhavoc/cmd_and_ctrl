package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Containment Priest — Creature — Human Cleric {1}{W}, 2/2:
//
//	"Flash
//	 If a nontoken creature would enter and it wasn't cast, exile it
//	 instead."
//
// A replacement on the enter-the-battlefield move (CR 614.1, 614.6). "It
// was cast" is the move coming off the stack (a cast creature spell
// resolves from OldZone stack); a creature put onto the battlefield from
// a graveyard, library, hand or exile, or blinked back, has some other
// OldZone and is exiled instead. A token is created from no zone and is
// excluded by name ("nontoken"). Containment Priest itself, cast, is not
// replaced. It applies to every player's creatures.
func init() {
	Register(Spec{
		OracleID:        "8a76a9b9-3127-45fe-b20f-a8f643276281",
		Name:            "Containment Priest",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || ev.OldZone == game.ZoneStack {
					return false
				}
				entering, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !entering.IsToken() && entering.IsCreature()
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.NewZone = game.ZoneExile
				ev.NewZoneOwner = uuid.Nil
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Containment Priest: exile instead of entering",
		}},
	})
}
