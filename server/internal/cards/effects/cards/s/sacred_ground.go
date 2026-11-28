package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sacred Ground — Enchantment {1}{W} (Eighth Edition, #5607):
//
//	"Whenever a spell or ability an opponent controls causes a land to be
//	 put into your graveyard from the battlefield, return that card to the
//	 battlefield."
//
// Reads the controller of the resolving spell or ability behind the exit
// (Event.LeaveController): destruction or a sacrifice an opponent's effect
// causes counts, a cost, a state-based action or one of your own effects does
// not. The land must have been a land as it left, and go to your graveyard
// (you own it). It returns under its owner's control, as the same card that
// is now in the graveyard (CR 400.7).
//
// No simplification.
func init() {
	const label = "Sacred Ground — return that land to the battlefield"
	effects.Register(effects.Spec{
		OracleID:     "b18e773c-611b-4c18-8b2d-8d3a7e5ddc93",
		Name:         "Sacred Ground",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventZoneMove},
			AppliesTo: effects.S82OldB615OpponentPutYourLandInYourGraveyard,
			Key:       label,
			Build:     effects.EventCardTriggerItem(label),
			Effect:    effects.S82OldB615ReturnTheNamedCardToTheBattlefield,
		}},
	})
}
