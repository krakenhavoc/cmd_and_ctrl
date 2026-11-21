package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jeweled Amulet — Artifact {0} (Mystery Booster Commander, #5948):
//
//	"{1}, {T}: Put a charge counter on this artifact. Note the type of mana
//	 spent to pay this activation cost. Activate only if there are no charge
//	 counters on this artifact.
//	 {T}, Remove a charge counter from this artifact: Add one mana of this
//	 artifact's last noted type."
//
// The first ability is an ordinary activation, gated on there being no charge
// counter. As it resolves it puts the counter on and notes the type of the one
// mana that paid for {1}, read off the activation's paid-cost record (W, U, B,
// R, G, or C for colorless): the note is game.NoteManaTypeForEffect, kept with
// the permanent where a chosen color is (per instance, cleared when it leaves,
// shown as the same chip), and the last note wins. The second is a mana ability
// (CR 605.1a) whose cost removes the counter, producing the noted type; with
// nothing noted it produces nothing, though the cost is paid, which is the
// weaker reading and unreachable, since a counter is only ever put on together
// with a note. A cost paid on paper (no mana recorded) notes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e2ad4b4d-d5fe-4a64-9b40-296c8fc8bcce",
		Name:         "Jeweled Amulet",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				Tap:            true,
				RemoveCounters: RemoveCountersFromThis(game.CounterCharge, 1).RemoveCounters,
			},
			ProducedFunc: ProducedChosenColor(),
			Label:        "{T}, Remove a charge counter from this artifact: Add one mana of this artifact's last noted type.",
		}},
		Activated: []ActivatedAbility{{
			Label:     "{1}, {T}: Put a charge counter on this artifact. Note the type of mana spent to pay this activation cost. Activate only if there are no charge counters on this artifact.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      Plus(ManaCost("{1}"), TapCost()),
			Condition: jeweledAmuletNoCharge,
			Effect: func(g *game.Game, item *game.StackItem) error {
				if sourceIsNewObject(g, item) {
					return nil
				}
				if err := g.AddCounterForEffect(item.SourceCardID, game.CounterCharge, 1); err != nil {
					return err
				}
				if len(item.Paid.Mana) > 0 {
					g.NoteManaTypeForEffect(item.SourceCardID, item.Paid.Mana[0].Color)
				}
				return nil
			},
		}},
	})
}

// jeweledAmuletNoCharge is "Activate only if there are no charge counters on
// this artifact".
func jeweledAmuletNoCharge(g *game.Game, _, source uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(source)
	return ok && c.Counters[game.CounterCharge] <= 0
}
