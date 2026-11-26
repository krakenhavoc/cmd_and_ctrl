package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hold the Perimeter — Conspiracy (Conspiracy: Take the Crown, S82 OLD
// batch 267, #5540):
//
//	"(Start the game with this conspiracy face up in the command zone.)
//	 At the beginning of your first upkeep, create a 1/2 white Soldier
//	 creature token with defender.
//	 At the beginning of each other player's first upkeep, that player
//	 creates a 1/1 red Goblin creature token with "This token can't block.""
//
// A face-up conspiracy (CR 315.5, ADR 0227): two triggers, each read from
// the upkeep that is beginning being its player's first (Player.UpkeepsBegun,
// the Devourer of Destiny count). The Goblin is made under that player's
// control, not the conspiracy's owner's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "20f7f33d-0c49-49f7-80fa-aed86e33a528",
		Name:         "Hold the Perimeter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.S82OldB267FirstUpkeep(true),
				"Hold the Perimeter — create a 1/2 white Soldier creature token with defender",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/2 white Soldier with defender"), N: 1})),
			effects.On(game.EventBeginUpkeep, effects.S82OldB267FirstUpkeep(false),
				"Hold the Perimeter — that player creates a 1/1 red Goblin creature token that can't block",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Trigger.Event.Actor, Template: effects.S82OldB267GoblinToken(), N: 1}.
						Apply(effects.NewContext(g, item))
				}),
		},
	})
}
