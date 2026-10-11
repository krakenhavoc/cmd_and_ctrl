package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Black Mana Battery — Artifact {4}:
//
//	"{2}, {T}: Put a charge counter on this artifact.
//	 {T}, Remove any number of charge counters from this artifact: Add
//	 {B}, then add an additional {B} for each charge counter removed
//	 this way."
//
// Mage-Ring Network's shape on an artifact. Banking is an activated
// ability, not a mana one: it adds no mana, so it uses the stack (CR
// 605.1a). The payout is a mana ability whose cost removes ANY NUMBER
// of charge counters, zero included; the count is announced and read
// back through the paid-cost record, because by the time the mana is
// minted the counters are gone. It always adds one {B} and one more
// for each counter that really came off, so removing none is {B}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc3b2bab-755d-4b2f-97c0-70a74d721a79",
		Name:         "Black Mana Battery",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				Tap:            true,
				RemoveCounters: RemoveCountersXFromThis(game.CounterCharge, 0).RemoveCounters,
			},
			ProducedForPaid: func(g *game.Game, controller, source uuid.UUID, paid game.PaidCost) string {
				return "{B}" + ProducedPerCounterRemoved("{B}")(g, controller, source, paid)
			},
			Label: "Remove any number of charge counters: Add {B}, then {B} for each",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Put a charge counter on this artifact.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Effect:  putChargeCounterOnThis,
		}},
	})
}
