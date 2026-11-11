package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s48_helpers.go: helpers of batch 48 (Monarch courts and clones).

// cmdr2S48CourtOfEmberethUpkeep makes the Knight, then — if its controller
// is the monarch — deals X damage to each opponent, X the creatures that
// controller controls now (the Knight included).
func cmdr2S48CourtOfEmberethUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (CreateToken{Template: TokenCard("3/1 red Knight"), N: 1}).Apply(ctx); err != nil {
		return err
	}
	if !YoureTheMonarch(g, item.Controller) {
		return nil
	}
	x := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() {
			x++
		}
	}
	return DamageEachOpponent{N: x}.Apply(ctx)
}

// cmdr2S48CourtOfArdenvaleUpkeep returns the chosen card to its owner's hand,
// or to the battlefield (under its owner's control) for the monarch.
func cmdr2S48CourtOfArdenvaleUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 || targets[0].Kind != game.TargetCard {
		return nil
	}
	dest := game.ZoneHand
	if YoureTheMonarch(g, item.Controller) {
		dest = game.ZoneBattlefield
	}
	return ReturnFromGraveyard{Target: targets[0].ID, Dest: dest, Controller: ctx.Controller()}.Apply(ctx)
}

// cmdr2S48CourtOfGarenbrigUpkeep distributes the two counters as announced,
// then, for the monarch, doubles the +1/+1 counters on each creature its
// controller controls.
func cmdr2S48CourtOfGarenbrigUpkeep(g *game.Game, item *game.StackItem) error {
	if err := PutDividedCounters(NewContext(g, item), game.CounterPlusOne); err != nil {
		return err
	}
	if !YoureTheMonarch(g, item.Controller) {
		return nil
	}
	return b08DoubleCountersOnEachCreatureYouControl(g, item)
}
