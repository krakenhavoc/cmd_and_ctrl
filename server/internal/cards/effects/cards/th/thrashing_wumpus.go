package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thrashing Wumpus — Creature — Beast {3}{B}{B}, 3/3 (Mercadian Masques, S82
// OLD, #5619):
//
//	"{B}: This creature deals 1 damage to each creature and each player."
//
// Pyrohemia's sweep for {B}: one damage instance to every creature (the Wumpus
// included) and every player, activated as often as there is black mana.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef220824-fab4-4d8e-9ba4-65ff5ba1db66",
		Name:         "Thrashing Wumpus",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: This creature deals 1 damage to each creature and each player.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1}},
			Cost:    effects.ManaCost("{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B23DamageEachCreatureAndEachPlayer(effects.NewContext(g, item), 1)
			},
		}},
	})
}
