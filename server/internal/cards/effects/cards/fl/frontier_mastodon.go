package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frontier Mastodon — {2}{G} Creature — Elephant 3/2 (Jumpstart 2022, #5948):
//
//	"Ferocious — This creature enters with a +1/+1 counter on it if you
//	 control a creature with power 4 or greater."
//
// An enters-with-counters replacement (CR 614.1c) whose condition is read
// as it enters: the Mastodon is not on the battlefield yet, so it never
// counts itself (Revolt's entry shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae81040b-572d-4906-bacc-e7bae16e1130",
		Name:         "Frontier Mastodon",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			RowLabel:  "Ferocious — enters with a +1/+1 counter on it if you control a creature with power 4 or greater.",
			EntryOnly: true,
			Watches:   []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventMove &&
					ev.NewZone == game.ZoneBattlefield &&
					src != nil && ev.CardID == src.InstanceID &&
					effects.B24ControlsCreatureWithPowerAtLeast(g, src.Controller, 4)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.AddCounterAtETB(game.CounterPlusOne, 1)
				return nil
			},
		}},
	})
}
