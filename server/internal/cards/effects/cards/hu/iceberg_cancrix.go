package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iceberg Cancrix — Snow Creature — Crab {1}{U}, 0/4 (Modern Horizons):
//
//	"Whenever another snow permanent you control enters, you may have target
//	 player mill two cards."
//
// "Another" excludes the Cancrix itself; the entering permanent must be snow
// (CR 205.4a) as it enters. The player is chosen as the ability is put on the
// stack and "you may" is asked as it resolves, so a target that has gone is
// not milled (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e00188a-0743-4a39-9229-977a0a115659",
		Name:         "Iceberg Cancrix",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
					return ok && c.HasSupertype("snow")
				}, "Iceberg Cancrix — target player mills two cards",
					effects.TargetPlayerMills(2)),
				effects.TargetPlayer("target player")),
				"Iceberg Cancrix — have target player mill two cards?"),
		},
	})
}
