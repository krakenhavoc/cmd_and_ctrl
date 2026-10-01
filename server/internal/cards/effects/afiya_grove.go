package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Afiya Grove — Enchantment {1}{G}:
//
//	"This enchantment enters with three +1/+1 counters on it.
//	 At the beginning of your upkeep, move a +1/+1 counter from this
//	 enchantment onto target creature.
//	 When this enchantment has no +1/+1 counters on it, sacrifice it."
//
// ADR 0107 §1 (#1858).
//
//   - The three counters are an entry replacement, there before anything
//     looks at the Grove.
//   - The upkeep move is Simic Fluxmage's (CR 122.5): the counter is put
//     on the target and, only if it landed, taken off the Grove. With no
//     creature to target the trigger is removed (CR 603.3d).
//   - The sacrifice is a CR 603.8 state trigger: the upkeep that moves the
//     last counter triggers it, and so does anything else that strips them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "056a98aa-f3f7-4ac9-9755-3e8a60a59abb",
		Name:         "Afiya Grove",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 3, "Afiya Grove: enters with three +1/+1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(AtYourUpkeep("Afiya Grove — move a +1/+1 counter onto target creature", simicFluxmageMove),
				TargetCreature("target creature")),
			WhenThisHasNo(game.CounterPlusOne, "Afiya Grove — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
	})
}
