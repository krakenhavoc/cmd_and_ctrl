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
// upkeep trigger, Beastbond Outcaster's enters trigger and Garruk's
// Uprising's enters trigger share it. A Do with an If, so the rows
// declare a footprint (#3033).
var drawIfYouControlPowerFourOrGreater = Do(If{
	Cond: CondPowerAtLeast{N: 4},
	Then: []Applier{DrawCards{N: 1}},
})
