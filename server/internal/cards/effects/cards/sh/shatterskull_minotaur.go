package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shatterskull Minotaur — Creature — Minotaur Warrior {4}{R}{R}, 5/4
// (Zendikar Rising, #5508):
//
//	"This spell costs {1} less to cast for each creature in your party.
//	 (Your party consists of up to one each of Cleric, Rogue, Warrior, and
//	 Wizard.)
//	 Haste"
//
// The party (CR 700.8) is counted as the spell is cast; the discount takes
// generic mana only, so the two red stay.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "2019004c-ebcf-4665-8736-f01ce7d8fcd2",
		Name:              "Shatterskull Minotaur",
		Completeness:      effects.CompletenessFull,
		PrintedKeywords:   []string{"haste"},
		SelfCostModifiers: []game.CostModifier{effects.OldB095PartyCostsLess()},
	})
}
