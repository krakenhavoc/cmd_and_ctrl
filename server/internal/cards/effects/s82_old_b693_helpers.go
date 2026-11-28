package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b693_helpers.go — the Mercadian Masques lands (#5619, S82).

// MmqStorageLand is the whole Spec of a Mercadian Masques storage land:
//
//	"This land enters tapped.
//	 {T}: Put a storage counter on this land.
//	 {T}, Remove any number of storage counters from this land: Add {C}
//	 for each storage counter removed this way."
//
// The counter is a plain activated ability that uses the stack; the
// payout is a mana ability (CR 605.1a) whose cost removes any number of
// counters, zero included, and whose mana counts the counters that came
// off. `color` is the mana symbol's letter.
func MmqStorageLand(oracleID, name, color string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				Tap:            true,
				RemoveCounters: RemoveCountersXFromThis(game.CounterStorage, 0).RemoveCounters,
			},
			ProducedForPaid: ProducedPerCounterRemoved(color),
			Label:           "{T}, Remove any number of storage counters: Add {" + color + "} for each storage counter removed this way",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Put a storage counter on this land.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect:  PutStorageCounterOnThis,
		}},
	}
}

// MmqDepletionCounter is the counter kind the depletion lands print.
const MmqDepletionCounter = "depletion"

// MmqDepletionLand is the whole Spec of a Mercadian Masques depletion
// land:
//
//	"This land enters tapped with two depletion counters on it.
//	 {T}, Remove a depletion counter from this land: Add {C}{C}. If there
//	 are no depletion counters on this land, sacrifice it."
//
// The sacrifice is the mana ability's rider (CR 605.3b), read after the
// counter has come off as the cost. `color` is the mana symbol's letter.
func MmqDepletionLand(oracleID, name, color string) Spec {
	sym := "{" + color + "}"
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedWithCounters(MmqDepletionCounter, 2)},
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				Tap:            true,
				RemoveCounters: RemoveCountersFromThis(MmqDepletionCounter, 1).RemoveCounters,
			},
			Produced: sym + sym,
			Label:    "{T}, Remove a depletion counter: Add " + sym + sym + ". If there are no depletion counters on this land, sacrifice it.",
			Rider: func(g *game.Game, _ uuid.UUID, source uuid.UUID) error {
				c, ok := g.LookupCardForEffect(source)
				if !ok || c.Counters[MmqDepletionCounter] > 0 {
					return nil
				}
				return g.SacrificePermanentForEffect(source)
			},
		}},
	}
}
