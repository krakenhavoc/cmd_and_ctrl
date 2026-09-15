package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angelic Accord — Enchantment {3}{W} (EDHREC rank 3394):
//
//	"At the beginning of each end step, if you gained 4 or more life
//	 this turn, create a 4/4 white Angel creature token with flying."
//
// The lifegain deck's Angel factory. "Each end step" is every
// player's, so the trigger does not read the event's actor; the
// intervening if (CR 603.4) is checked when the trigger would fire
// and again as it resolves — b15LifeGainedThisTurn, the sum of the
// controller's positive life changes since this turn's upkeep began,
// so a lifelink hit on an opponent's turn counts on that turn's end
// step, as printed. The Angel is Sigil of the Empty Throne's 4/4
// flier.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aa95501f-bb59-494a-bcae-b74ca10ad57e",
		Name:         "Angelic Accord",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b32EndStepAndYouGainedLifeThisTurnAtLeast(ev, source, g, 4)
			}, "Angelic Accord — create a 4/4 Angel with flying", b32AngelIfYouGainedLifeThisTurnAtLeast(4)),
		},
	})
}
