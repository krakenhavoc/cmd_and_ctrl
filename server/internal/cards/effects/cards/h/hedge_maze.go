package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "ca4b6689-04ee-4227-9bdc-cb5a9590c745",
		Name:          "Hedge Maze",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("G", "U")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Hedge Maze — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
	})
}
