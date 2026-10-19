package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dungeon Descent — Land:
//
//	"This land enters tapped.
//	 {T}: Add {C}.
//	 {4}, {T}, Tap an untapped legendary creature you control: Venture
//	 into the dungeon. Activate only as a sorcery."
//
// The enters-tapped is a self-replacement (SelfEntersTapped), and
// tapping the legendary creature is a cost that is not the {T} symbol,
// so one that arrived this turn may pay it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f086a63c-0c62-4674-bd27-82e7aed12b1a",
		Name:         "Dungeon Descent",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{4}, {T}, Tap an untapped legendary creature you control: Venture into the dungeon. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{4}"), TapCost(), TapAnUntapped("an untapped legendary creature you control", Creature(), Legendary())),
			SorcerySpeed: true,
			Effect:       Venture,
		}},
	})
}
