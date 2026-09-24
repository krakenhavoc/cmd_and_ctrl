package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// top100_helpers.go — the shared bodies behind the play-rate batch
// catalogued in docs/decklists/top-100-commander-staples.md.
//
// Its own file rather than helpers.go, per the convention #231 set:
// concurrent card batches collide on shared helper files.

// youControlPowerFourOrGreater reports whether `controller` has a
// creature with power 4 or greater on the battlefield — the
// intervening-if clause on Garruk's Uprising.
//
// CurrentPower folds counters into the printed value, so a 2/2 with
// two +1/+1 counters counts. It does not consult the layer engine's
// effective characteristics, matching the posture every other
// power-reading predicate in the catalog takes (PowerGE, PowerLE).
func youControlPowerFourOrGreater(g *game.Game, controller uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.CurrentPower() >= 4 {
			return true
		}
	}
	return false
}

// drawIfYouControlPowerFourOrGreater is the resolution half of "…, if
// you control a creature with power 4 or greater, draw a card" — the
// intervening-if (CR 603.4) checked again as the trigger resolves, so
// a big creature killed in response leaves no card. Colossal Majesty's
// upkeep trigger and Beastbond Outcaster's enters trigger share it.
func drawIfYouControlPowerFourOrGreater(g *game.Game, item *game.StackItem) error {
	if !youControlPowerFourOrGreater(g, item.Controller) {
		return nil
	}
	return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
}
