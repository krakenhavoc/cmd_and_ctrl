package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gemstone Mine — Land (EDHREC rank 3324):
//
//	"This land enters with three mining counters on it.
//	 {T}, Remove a mining counter from this land: Add one mana of any
//	 color. If there are no mining counters on this land, sacrifice
//	 it."
//
// Three untapped any-colour mana, then the land is gone. The
// counters go on through one entry replacement (b10EntersWithCounters
// — on the card before any trigger or state check sees it).
//
// The mana ability's cost has no counter component in the engine's
// mana-cost vocabulary (tap, sacrifice, life, mana — the counter case
// never had a card asking for it). It is modelled the way Runaway
// Steam-Kin is, with two components that DO exist and together
// enforce the cost: a Condition gating activation on a mining
// counter being there (checked before anything is paid, so an empty
// Mine cannot activate — CR 602.5), and a Rider that removes the
// counter as the mana lands and sacrifices the Mine when it was the
// last (b31RemoveMiningCounterOrSacrifice). A mana ability resolves
// as one atomic step without the stack (CR 605.3b), so nothing can
// observe that the counter left after the mana arrived rather than
// before. The auto-tapper never reaches for an ability with a rider,
// so a counter is only ever spent on purpose. "Any color" is the
// five-way pick with no identity narrowing — the card does not say
// "in your commander's color identity".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c828f10-4775-492f-9224-1e2814ad2cad",
		Name:         "Gemstone Mine",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters("mining", 3, "Gemstone Mine: enters with three mining counters"),
		},
		ManaAbilities: []ManaAbility{{
			Cost:      ManaAbilityCost{Tap: true},
			Produced:  "{W|U|B|R|G}",
			Label:     "{T}, Remove a mining counter: Add one mana of any color. With no mining counters left, sacrifice this land.",
			Condition: b31HasMiningCounter,
			Rider:     b31RemoveMiningCounterOrSacrifice,
		}},
	})
}
