package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodbraid Marauder — Creature — Human Berserker {1}{R}, 3/1 (Modern
// Horizons 2, #5501):
//
//	"This creature can't block.
//	 Delirium — This spell has cascade as long as there are four or more
//	 card types among cards in your graveyard."
//
// Cascade is the keyword's own trigger (CR 702.85a), which fires as the
// spell is cast; the delirium test (four card types in the caster's
// graveyard, CR 207.2c) is read then, as the spell goes on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25ca4b6c-2c4a-40f1-aee8-3bd373572eb7",
		Name:         "Bloodbraid Marauder",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Triggered:    []game.TriggeredAbility{effects.S82OldB38CascadeWithDelirium()},
	})
}
