package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deep-Sea Terror — Creature — Serpent {4}{U}{U}, 6/6 (Magic Origins, S82 older
// sets):
//
//	"This creature can't attack unless there are seven or more cards in your
//	 graveyard."
//
// River Serpent's restriction with a floor of seven, read live at declaration
// (CR 508.1c) against the controller's own graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "36775412-0280-4acc-bae8-3ea5d4e341be",
		Name:         "Deep-Sea Terror",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.CantAttackUnlessYouHaveCardsInGraveyard(7)},
	})
}
