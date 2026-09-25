package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seizan, Perverter of Truth — Legendary Creature — Demon Spirit
// {3}{B}{B}, 6/5 (EDHREC rank 3627):
//
//	"At the beginning of each player's upkeep, that player loses 2
//	 life and draws two cards."
//
// The symmetrical Phyrexian Arena. Fires on every seat's upkeep,
// the controller's included; "that player" is the active player,
// captured when the trigger fires. Life loss, not damage, so no
// prevention shield sees it; then two ordinary draws, so a "whenever
// an opponent draws" payoff sees both.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "04d0d20f-720e-4cb6-a3ad-ea2b57bb7efa",
		Name:         "Seizan, Perverter of Truth",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil
			},
			Key: "Seizan, Perverter of Truth — that player loses 2 life and draws two cards",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Seizan, Perverter of Truth — that player loses 2 life and draws two cards", nil)
				item.Params.Player = ev.Actor
				return item
			},
			Effect: b34ThatPlayerLosesLifeAndDraws(2, 2),
		}},
	})
}
