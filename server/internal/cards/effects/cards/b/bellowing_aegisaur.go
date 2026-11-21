package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bellowing Aegisaur — Creature — Dinosaur {5}{W}, 3/5 (Lost Caverns of
// Ixalan Commander):
//
//	"Enrage — Whenever this creature is dealt damage, put a +1/+1 counter on
//	 each other creature you control."
//
// The set is read once as the ability resolves; the Aegisaur gets none.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7063abe1-87dd-4526-9867-f7ce255b2cf0",
		Name:         "Bellowing Aegisaur",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Cmdr2B08EnrageTrigger("Bellowing Aegisaur — Enrage: a +1/+1 counter on each other creature you control", effects.PutCounterOnEachOtherCreatureYouControl),
		},
	})
}
