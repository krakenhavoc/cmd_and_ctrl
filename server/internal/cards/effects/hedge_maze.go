package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hedge Maze — Land — Forest Island (EDHREC top 1,000, #3654):
//
//	"({T}: Add {G} or {U}.)
//	 This land enters tapped.
//	 When this land enters, surveil 1."
//
// The surveil lands' shape (surveil_lands.go): a nonbasic land with two
// basic land types, so its mana ability is declared, plus the real
// surveil trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "ca4b6689-04ee-4227-9bdc-cb5a9590c745",
		Name:          "Hedge Maze",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("G", "U")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Hedge Maze — surveil 1", Do(Surveil{N: 1})),
		},
	})
}
