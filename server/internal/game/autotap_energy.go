package game

import "github.com/google/uuid"

// autotap_energy.go — ADR 0129 §5, owner decision 2: the auto-tapper's
// energy tier.
//
// A mana ability that pays energy (Aether Hub's "{T}, Pay {E}: Add one
// mana of any color") is plannable. Its tier sits after every source
// that costs nothing but its tap and before the pain tier (#2392), so a
// plan spends energy only when no plan without it pays the cost, and
// spends energy before it spends life. Within the tier, the source
// spending the least energy comes first.
//
// The planner holds the WHOLE plan to the controller's energy, as it
// holds it to the life the pain tier may spend: two Aether Hubs with one
// energy between them pay for one coloured pip, not two. The budget for
// both rides one spendBudget through the coloured pass, the generic pass
// and #2455's costed planner.

// spendBudget is what an auto-tap plan may still spend on the tiers that
// cost the player something other than a tap: life (painBudgetFor) and
// energy (the controller's energy counters).
type spendBudget struct {
	life   int
	energy int
}

// autoTapSpendBudget is the budget a plan for `controller` starts with.
//
// Caller must hold g.mu.
func autoTapSpendBudget(g *Game, controller uuid.UUID) spendBudget {
	return spendBudget{
		life:   painBudgetFor(g, controller),
		energy: PlayerEnergy(g.playerByIDLocked(controller)),
	}
}

// covers reports whether the budget can pay for source s.
func (b *spendBudget) covers(s tapSource) bool {
	return s.Pain <= b.life && s.Energy <= b.energy
}

// spend books source s against the budget.
func (b *spendBudget) spend(s tapSource) {
	b.life -= s.Pain
	b.energy -= s.Energy
}

// refund returns source s to the budget when the search backs out.
func (b *spendBudget) refund(s tapSource) {
	b.life += s.Pain
	b.energy += s.Energy
}

// overdrawn reports whether the bookings so far spend more than the
// budget held.
func (b spendBudget) overdrawn() bool {
	return b.life < 0 || b.energy < 0
}

// costTier is the tier a source's non-tap cost puts it in, for both
// comparators: 0 for a source that spends neither life nor energy, 1
// for the energy tier, 2 for the pain tier (owner decision 2: energy
// before life).
func (s tapSource) costTier() int {
	switch {
	case s.Pain > 0:
		return 2
	case s.Energy > 0:
		return 1
	}
	return 0
}
