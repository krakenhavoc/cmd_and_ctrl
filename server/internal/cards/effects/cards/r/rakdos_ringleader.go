package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rakdos Ringleader — Creature — Skeleton Warrior {4}{B}{R}, 3/1 (Return to
// Ravnica, S82 OLD):
//
//	"First strike
//	 Whenever this creature deals combat damage to a player, that player
//	 discards a card at random.
//	 {B}: Regenerate this creature."
//
// The damaged player is read as the trigger fires, so the discard still
// happens if the Ringleader has left by resolution; an empty hand discards
// nothing (CR 701.9b).
//
// No simplification.
func init() {
	const key = "Rakdos Ringleader — that player discards a card at random"
	effects.Register(effects.Spec{
		OracleID:        "f8ce8573-7c87-4f47-a0d4-14bd335f2c1d",
		Name:            "Rakdos Ringleader",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated:       []effects.ActivatedAbility{effects.LrwB1RegenerateThisRow("{B}")},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: effects.ThisDealtCombatDamageToAPlayer,
			Key:       key,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, key)
				item.Params.Player = ev.Target
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DiscardCards{Player: item.Params.Player, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
