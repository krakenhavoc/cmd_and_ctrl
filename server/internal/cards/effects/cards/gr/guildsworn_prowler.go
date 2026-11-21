package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Guildsworn Prowler — {1}{B} Creature — Tiefling Rogue Assassin 2/1
//
//	"Deathtouch
//	 When this creature dies, if it wasn't blocking, draw a card."
//
// Whether it was blocking is read from the combat state it last had on the
// battlefield (CR 603.10a, Event.BlockingTarget), so a Prowler that dies
// blocking draws nothing, and one that dies anywhere else does.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "72e2aa4f-b4b4-42ab-adc0-d0ac09cd5f92",
		Name:            "Guildsworn Prowler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Purpose:         game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				if !effects.ThisDied(ev, source, lki, g) {
					return false
				}
				_, blocking := effects.DiedWhileBlocking(ev, g)
				return !blocking
			}, "Guildsworn Prowler — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
