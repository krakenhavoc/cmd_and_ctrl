package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Power Depot — Artifact Land:
//
//	"This land enters tapped.
//	 {T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast artifact
//	 spells or activate abilities of artifacts.
//	 Modular 1"
//
// Modular is the engine's keyword (Arcbound Bruiser): it enters with a
// +1/+1 counter (inert on a noncreature) and its dies half never fires for
// a land that is not a creature. The restricted mana is Cargo Ship's
// either-or restriction.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "64687880-03f9-4f38-985b-1027c797e33f",
		Name:            "Power Depot",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"modular 1"},
		Replacements:    []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "{T}: Add one mana of any color. Spend this mana only to cast artifact spells or activate abilities of artifacts.",
				Restrictions: []string{game.ManaRestrictAnyOf(
					[]string{game.ManaRestrictCast, game.ManaRestrictType("Artifact")},
					[]string{game.ManaRestrictActivate, game.ManaRestrictType("Artifact")},
				)},
			},
		},
	})
}
