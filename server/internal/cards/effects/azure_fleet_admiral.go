package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Azure Fleet Admiral — Creature — Human Pirate {3}{U}, 3/3 (Lost Caverns
// of Ixalan Commander, #5368):
//
//	"When this creature enters, you become the monarch.
//	 This creature can't be blocked by creatures the monarch controls."
//
// The evasion is a block rule on the Admiral (CantBeBlockedBy over
// creatures controlled by whoever holds the crown), read as blockers are
// declared; with no monarch it refuses nothing. The monarch may be its own
// controller, whose creatures are not blocking it anyway.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b39f228f-d146-49c6-bab4-5304e0f742d8",
		Name:         "Azure Fleet Admiral",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), cmdr2S48ControlledByTheMonarch, "creatures the monarch controls"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Azure Fleet Admiral"),
		},
	})
}

// cmdr2S48ControlledByTheMonarch is "creatures the monarch controls".
func cmdr2S48ControlledByTheMonarch(g *game.Game, _ uuid.UUID, c game.Card) bool {
	return g.Monarch != uuid.Nil && c.Controller == g.Monarch
}
