package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ulvenwald Mystics // Ulvenwald Primordials — {2}{G}{G} Creature — Human
// Shaman Werewolf 3/3 // Creature — Werewolf 5/5 (Innistrad, S82 OLD b446,
// #5572):
//
//	Front: "At the beginning of each upkeep, if no spells were cast last turn,
//	        transform this creature."
//	Back:  "{G}: Regenerate this creature.
//	        At the beginning of each upkeep, if a player cast two or more spells
//	        last turn, transform this creature."
//
// The original-Innistrad werewolf pair of triggers (innistrad_werewolves.go);
// the back face's regeneration is the shared "<cost>: Regenerate this
// creature." row.
//
// No simplification.
func init() {
	const oracle = "4418ff27-57c0-4c94-8aee-6151e4f8b0f9"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Ulvenwald Mystics",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfNoSpellsLastTurn("Ulvenwald Mystics"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Ulvenwald Primordials",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.RegenerateThisRow("{G}")},
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Ulvenwald Primordials"),
		},
	})
}
