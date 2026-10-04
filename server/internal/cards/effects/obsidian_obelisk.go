package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Obsidian Obelisk — Artifact {2}:
//
//	"This artifact enters tapped.
//	 {T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast a
//	 multicolored spell."
//
// Pillar of the Paruns' mana on a rock that can also make plain {C}.
// The coloured half carries "cast" and "multicolored" (#1600, CR
// 105.2b); the colourless half is ordinary mana. "Enters tapped" is the
// self-replacement (SelfEntersTapped), not an ETB tap.
//
// The auto-tapper never plans a restricted mana ability (ADR 0040 §7),
// so it funds a generic pip with the Obelisk's {C} and leaves the
// coloured mana to a deliberate click.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "92a5e50d-8477-4654-9c1b-219e4fe3c5ab",
		Name:         "Obsidian Obelisk",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:         ManaAbilityCost{Tap: true},
				Produced:     "{W|U|B|R|G}",
				Restrictions: []string{ManaRestrictCast, ManaRestrictMulticolored},
				Label:        "Add one mana of any color. Spend this mana only to cast a multicolored spell",
			},
		},
	})
}
