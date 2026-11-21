package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Entity Tracker — Creature — Human Scout {2}{U}, 2/3:
//
//	"Flash
//	Eerie — Whenever an enchantment you control enters and whenever you
//	fully unlock a Room, draw a card."
//
// Flash rides PrintedKeywords; the draw is DrawCards for the controller.
//
// Eerie is one ability with two conditions (Eerie, rooms.go): an
// enchantment entering under your control, or you fully unlocking a Room.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c9db5293-4866-46ec-9ae8-059d27bd50fc",
		Name:            "Entity Tracker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Eerie("Entity Tracker — draw a card (eerie)", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
