package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Storm Entity — Creature — Elemental {1}{R}, 1/1 (Dominaria Remastered,
// #5948):
//
//	"Haste
//	 This creature enters with a +1/+1 counter on it for each other spell
//	 cast this turn."
//
// A replacement effect on its own entry (CR 614.1c), counted as it enters:
// every spell any player has cast this turn, not counting the Entity's own
// cast. The per-turn tally already holds the Entity's own spell once it is
// cast, so one is taken off when it entered as a cast spell; put onto the
// battlefield any other way there is nothing of its own to leave out.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6c3b9136-5e4b-4bb5-bef5-17887d7211dd",
		Name:            "Storm Entity",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
					src != nil && ev.CardID == src.InstanceID
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
				n := effects.GapB138SpellsCastThisTurnByAnyone(g)
				if ev.OldZone == game.ZoneStack && effects.GapB138WasCast(g, src.InstanceID) && n > 0 {
					n--
				}
				ev.AddCounterAtETB(game.CounterPlusOne, n)
				return nil
			},
			Label: "Storm Entity: enters with a +1/+1 counter for each other spell cast this turn",
		}},
	})
}
