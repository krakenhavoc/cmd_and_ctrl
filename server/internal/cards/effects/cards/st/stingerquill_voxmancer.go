package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stingerquill Voxmancer // Vicious Verse — Creature — Goblin Sorcerer
// {B/R}, 1/2 // Sorcery {B/R} (preparation card, CR 722):
//
//	"At the beginning of your upkeep, if this creature isn't prepared,
//	 it becomes prepared. (While it's prepared, you may cast a copy of
//	 its spell. Doing so unprepares it.)"
//
//	Vicious Verse — "Vicious Verse deals 1 damage to target opponent."
//
// It does not enter prepared; the upkeep trigger is its only source of
// the designation.
//
// No simplification.
func init() {
	const id = "e8f755f7-ec93-4d2e-bffc-c719984fa13d"
	effects.Register(effects.Spec{
		OracleID:     id,
		Name:         "Stingerquill Voxmancer",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.FraBecomesPreparedAtUpkeep("Stingerquill Voxmancer")},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Vicious Verse",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve:    effects.ViciousVerseResolve,
	})
}
