package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcbound Reclaimer — Artifact Creature — Golem {4}, 0/0 (Darksteel,
// #5605):
//
//	"Remove a +1/+1 counter from this creature: Put target artifact card
//	 from your graveyard on top of your library.
//	 Modular 2 (This creature enters with two +1/+1 counters on it. When
//	 it dies, you may put its +1/+1 counters on target artifact
//	 creature.)"
//
// The ability is Academy Ruins' body (PutChosenTargetOnTopOfLibrary) with
// a counter removed as its cost; it has no tap symbol, so it may be
// activated as long as counters remain. Removing the last counter leaves
// a 0/0 that dies to the state-based check, and then its modular trigger
// has nothing to move. Modular is the engine's keyword (#2012,
// CR 702.43a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c6424a3e-dbc7-403f-867f-a81cccd17ab8",
		Name:            "Arcbound Reclaimer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"modular 2"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove a +1/+1 counter from this creature: Put target artifact card from your graveyard on top of your library.",
			Cost:    effects.RemoveCountersFromThis(game.CounterPlusOne, 1),
			Targets: effects.TargetCardInGraveyard("target artifact card from your graveyard", effects.Artifact(), effects.YouOwn()),
			Effect:  effects.PutChosenTargetOnTopOfLibrary,
		}},
	})
}
