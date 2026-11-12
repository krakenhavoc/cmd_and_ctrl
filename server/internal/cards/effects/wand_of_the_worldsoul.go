package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wand of the Worldsoul — Artifact {2}{W} (March of the Machine
// Commander, #5647):
//
//	"This artifact enters tapped.
//	 {T}: Add {W}.
//	 {T}: The next spell you cast this turn has convoke."
//
// The second ability is a one-use NextSpellPromise carrying Convoke
// (NextSpellHasConvoke): while it stands, the next spell its controller
// casts may be paid for by tapping creatures (CR 702.51a). That spell
// spends it whether or not a creature was tapped, it is a spell that has
// convoke for anything that asks, and an unspent promise lapses at end
// of turn. Not a mana ability: it uses the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3d175a59-c1b4-43b8-8a1f-f7cbcefe130e",
		Name:         "Wand of the Worldsoul",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}: The next spell you cast this turn has convoke.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect:  Do(NextSpellHasConvoke("Wand of the Worldsoul")),
		}},
	})
}
