package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thrumming Stone — Legendary Artifact {5} (Double Masters 2022):
//
//	"Spells you cast have ripple 4. (Whenever you cast a spell, you may
//	 reveal the top four cards of your library. You may cast spells with the
//	 same name as that spell from among the revealed cards without paying
//	 their mana costs. Put the rest on the bottom of your library.)"
//
// A permanent that gives ripple to every spell its controller casts, so each
// cast is one trigger of the Stone (CR 702.60a, CR 113.6). See
// GapB200RippleTrigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b2bd40a1-1857-4f4c-aed6-d4e224ac0ef7",
		Name:         "Thrumming Stone",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.GapB200RippleTrigger("Thrumming Stone", 4)},
	})
}
