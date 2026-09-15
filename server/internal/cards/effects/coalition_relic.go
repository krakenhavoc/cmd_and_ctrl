package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coalition Relic — Artifact {3} (EDHREC rank 3469):
//
//	"{T}: Add one mana of any color.
//	 {T}: Put a charge counter on this artifact.
//	 At the beginning of your first main phase, remove all charge
//	 counters from this artifact. Add one mana of any color for each
//	 charge counter removed this way."
//
// The five-colour rock that banks mana for your own turn. Three
// abilities in their three slots: the mana ability is Birds of
// Paradise's pipe pick; the charge counter is a tap-cost activated
// ability that uses the stack, as printed (it is not a mana ability
// — it adds no mana); and the main-phase trigger is Hulking Raptor's
// EventBeginPrecombatMain, whose body takes every charge counter off
// the Relic and adds one any-colour pick per counter. The mana
// arrives when the trigger resolves, at the start of the first main
// phase, and empties with the pool at the end of that phase — the
// printed behaviour. A Relic that left the battlefield in response
// adds nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "008cb342-79f5-4df6-a6b7-0e9e22ed693f",
		Name:         "Coalition Relic",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:  "{T}: Put a charge counter on Coalition Relic",
			Cost:   TapCost(),
			Effect: b33PutChargeCounterOnSelf,
		}},
		Triggered: []game.TriggeredAbility{
			AtYourPrecombatMain(b33CoalitionRelicLabel, b33RemoveChargeCountersForMana),
		},
	})
}
