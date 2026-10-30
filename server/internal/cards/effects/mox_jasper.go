package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mox Jasper — Legendary Artifact {0}:
//
//	"{T}: Add one mana of any color. Activate only if you control a Dragon."
//
// Mox Opal's and Tainted Peak's shape: one mana ability, {T} for a
// five-colour pick, gated by ManaAbility.Condition. The gate is checked
// when the ability is activated, before the artifact taps, and never again
// (CR 602.1b); a mana ability does not use the stack (CR 605.3b), so losing
// the Dragon afterwards does not take the mana back. Without a Dragon the
// activation is refused with nothing paid, and the auto-tapper, the bot's
// enumerator and the client's menu all read the same closure, so the Mox is
// not offered as a source either.
//
// "A Dragon" is a permanent you control with the Dragon subtype, read
// through the layers: a Changeling counts (CR 702.73a), a creature that has
// been made a Dragon counts, a face-down permanent (no subtypes, CR 708.2)
// and a Dragon an opponent controls do not. It needs no commander-identity
// narrowing, which the text never mentions, so all five colours are offered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d5fca77c-377f-41b4-a9e9-e91a89d351cc",
		Name:         "Mox Jasper",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:      ManaAbilityCost{Tap: true},
			Produced:  "{W|U|B|R|G}",
			Label:     "Add one mana of any color. Activate only if you control a Dragon",
			Condition: ControlsAtLeast(1, func(c game.Card) bool { return c.HasSubtypeWord("Dragon") }),
		}},
	})
}
