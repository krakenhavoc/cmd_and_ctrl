package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tegwyll, Duke of Splendor — Legendary Creature — Faerie Noble
// {1}{U}{B}, 2/3 (EDHREC rank 3827):
//
//	"Flying, deathtouch
//	 Other Faeries you control get +1/+1.
//	 Whenever another Faerie you control dies, you draw a card and you
//	 lose 1 life."
//
// The Faerie lord with a Grim Haruspex stapled on. Both keywords
// ride PrintedKeywords; the anthem is the shared TribalAnthem
// builder with "other" and "you control"; the dies trigger fires for
// every OTHER Faerie the controller controlled that goes to a
// graveyard — a Faerie as it last existed on the battlefield (CR
// 603.10a, #1679), grants included; Tegwyll's own death does not
// count, as printed — and
// the draw comes before the life loss, in printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "868f0a0a-ca9e-4baf-8295-6b228aa834e5",
		Name:            "Tegwyll, Duke of Splendor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Faerie"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B36AnotherFaerieYouControlDied(ev, source, g)
			}, "Tegwyll, Duke of Splendor — draw a card and lose 1 life", effects.B36DrawAndLoseOne),
		},
	})
}
