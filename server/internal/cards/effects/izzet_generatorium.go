package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Izzet Generatorium — Artifact {U}{R}:
//
//	"If you would get one or more {E} (energy counters), you get that
//	 many plus one {E} instead.
//	 {T}: Draw a card. Activate only if you've paid or lost four or more
//	 {E} this turn."
//
// ADR 0129 §6 (#1995). The replacement is Aether Refinery's with an
// addition in place of a doubling: ADR 0056's counter window with the
// player set (RepEventCounter, CounterPlayer). With a Refinery out too,
// the controller orders the two (CR 616.1). The activation condition
// reads the turn tally of energy paid or lost (PlayerTurnTally.
// EnergyPaidOrLost), which counts every energy payment and every energy
// counter an effect removes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6a618c4a-e604-4dce-a078-755fd35ac5d9",
		Name:         "Izzet Generatorium",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventCounter && ev.CounterPlayer == src.Controller &&
					ev.CounterName == game.CounterEnergy && ev.CounterDelta > 0
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Izzet Generatorium: that many plus one {E}",
		}},
		Activated: []ActivatedAbility{{
			Label:     "{T}: Draw a card. Activate only if you've paid or lost four or more {E} this turn.",
			Cost:      TapCost(),
			Condition: PaidOrLostEnergyThisTurn(4),
			Purpose:   game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Effect:    Do(DrawCards{N: 1}),
		}},
	})
}
