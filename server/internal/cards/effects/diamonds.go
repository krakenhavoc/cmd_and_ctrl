package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// diamonds.go — the Mirage "Diamond" mana rocks:
//
//	"This artifact enters tapped.
//	 {T}: Add {R}."
//
// Two-mana rocks for one colour — the Worn Powerstone shape, one mana
// smaller. Two of the five are in the roadmap's batch 06 (#299); the
// rest belong in this table when their batches reach them.
//
// No simplification.
func init() {
	for _, rock := range []struct{ oracleID, name, color string }{
		{"97b477d8-2e05-475e-8ed6-7d680cb21cd9", "Fire Diamond", "R"},
		{"2224b6e0-c5ff-45d0-84e3-83758c5fc99f", "Sky Diamond", "U"},
	} {
		Register(Spec{
			OracleID:     rock.oracleID,
			Name:         rock.name,
			Completeness: CompletenessFull,
			Replacements: []game.ReplacementEffect{SelfEntersTapped()},
			ManaAbilities: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{" + rock.color + "}",
				Label:    "Add {" + rock.color + "}",
			}},
		})
	}
}
