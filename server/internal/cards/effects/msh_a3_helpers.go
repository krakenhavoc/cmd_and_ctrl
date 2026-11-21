package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// msh_a3_helpers.go — shared by the Marvel Super Heroes half A batch 3
// cards (#4455). Every package-level name carries the mshA3 prefix.

// mshA3PlanCounter is the counter kind of the Plan enchantments.
const MshA3PlanCounter = "plan"

// mshA3PlanCounterReaches is "when the Nth plan counter is put on this
// enchantment": a plan-counter placement on the source whose new total
// reaches n from below, so a counter put on above n, or a placement that
// starts above n, fires nothing (Construct a Cosmic Cube's seventh).
func MshA3PlanCounterReaches(n int) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Kind != game.EventCounterPlaced || ev.Target != source.InstanceID || ev.Label != MshA3PlanCounter || ev.Amount < n {
			return false
		}
		return B12CounterTotalBefore(ev, g) < n
	}
}
