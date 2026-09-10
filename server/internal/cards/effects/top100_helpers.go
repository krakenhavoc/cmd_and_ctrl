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

// manaValueOfCard is CR 202.3, a card's mana value: the generic
// component plus one for each coloured symbol. Feed the Swarm and
// Reanimate both charge life equal to it.
//
// It reads the PRINTED mana cost. Cost reductions change what you
// pay, not the card's mana value (CR 202.3b), so the printed string
// is the right source — and it is the same derivation ManaValueLE /
// ManaValueGE already use for targeting predicates.
//
// A cost that fails to parse — a token, a land, anything with no
// mana cost — is mana value 0, which is what CR 202.3a says for a
// card with no mana cost.
func manaValueOfCard(c game.Card) int {
	cost, err := game.ParseCost(c.ManaCost)
	if err != nil {
		return 0
	}
	return cost.Generic + len(cost.Required)
}

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
