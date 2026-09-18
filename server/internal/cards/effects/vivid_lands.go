package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// vivid_lands.go — the Vivid cycle (Lorwyn / Shadowmoor), the first
// cards to need a counter cost on a MANA ability (#789):
//
//	"This land enters tapped with two charge counters on it.
//	 {T}: Add <its colour>.
//	 {T}, Remove a charge counter from this land: Add one mana of any
//	 color."
//
// Three pieces, all live:
//
//   - ONE entry replacement does both halves of the first line —
//     tapped, and the two charge counters — because two
//     self-replacements on the same event would queue a CR 616
//     ordering prompt for a choice that changes nothing.
//   - Ability 0 is the plain printed colour, and it is deliberately
//     FIRST: the auto-tapper takes one ability per permanent
//     (autoTapAbilityFor walks in order), so a Vivid land planned for
//     a cast spends its free colour rather than a charge counter, and
//     the fixer stays for a deliberate click. That is the same
//     ordering the painland cycle relies on.
//   - Ability 1 is the fixer, with ManaAbilityCost.RemoveCounters.
//     The counter comes off as part of the cost, at activation, so a
//     response cannot spend the same counter twice — and once the two
//     charge counters are gone the land is an ordinary mono-coloured
//     source, which is what the card says and what makes the cycle a
//     fair fixer.
//
// The auto-tapper WILL plan the fixer when the plain ability cannot
// cover a requirement and the counters are there: the cost is the self
// form with a printed kind and a printed count, which is the one shape
// the planner can both decide and afford (ADR 0020 §20). With no
// charge counter left it is not a mana source at all.
//
// No simplification.

// vividLand is one member of the cycle: its printed colour, and the
// any-colour fixer that spends a charge counter.
func vividLand(oracleID, name, color string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			SelfEntersTappedWithCounters(game.CounterCharge, 2),
		},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{" + color + "}",
				Label:    "Add {" + color + "}",
			},
			{
				Cost: ManaAbilityCost{
					Tap:            true,
					RemoveCounters: RemoveCountersFromThis(game.CounterCharge, 1).RemoveCounters,
				},
				Produced: "{W|U|B|R|G}",
				Label:    "Remove a charge counter: Add one mana of any color",
			},
		},
	}
}

func init() {
	Register(vividLand("2da7c49f-cc1e-45d9-9cbf-067e92b0daef", "Vivid Creek", "U"))
	Register(vividLand("b7a68899-c0d3-49e0-854b-19268ae9b89d", "Vivid Grove", "G"))
}
