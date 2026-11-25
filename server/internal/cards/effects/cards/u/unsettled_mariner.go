package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unsettled Mariner — Creature — Shapeshifter {W}{U}, 2/2 (Modern Horizons):
//
//	"Changeling (This card is every creature type.)
//	 Whenever you or a permanent you control becomes the target of a spell or
//	 ability an opponent controls, counter that spell or ability unless its
//	 controller pays {1}."
//
// Ward's pay-or-counter body (CR 702.21a) over its controller and every
// permanent that controller has, the Mariner included, watching a spell or
// ability an opponent controls (CR 115.9: the trigger comes from the target
// being chosen, per target). It is a triggered ability, not the ward keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7e4d1ce3-92f0-4acc-962b-8073d41d3673",
		Name:            "Unsettled Mariner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"changeling"},
		Triggered: []game.TriggeredAbility{
			effects.OldB166WardForYouAndYourPermanents(effects.WardMana("{1}"),
				"Unsettled Mariner — whenever you or a permanent you control becomes the target of a spell or ability an opponent controls, counter that spell or ability unless its controller pays {1}"),
		},
	})
}
