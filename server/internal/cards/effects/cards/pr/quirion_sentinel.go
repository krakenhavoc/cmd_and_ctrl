package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quirion Sentinel — Creature — Elf Druid {1}{G}, 2/1 (Invasion, #5616):
//
//	"When this creature enters, add one mana of any color."
//
// Not a mana ability (CR 605.1b): it uses the stack and adds the mana as it
// resolves, Outcaster Trailblazer's pick of all five colours.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70bb2690-653b-4787-bdd1-d1739a8ae545",
		Name:         "Quirion Sentinel",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Quirion Sentinel — add one mana of any color", effects.Do(effects.AddMana{Produced: "{W|U|B|R|G}"})),
		},
	})
}
