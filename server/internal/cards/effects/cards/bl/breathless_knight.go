package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Breathless Knight — Creature — Spirit Knight {1}{W}{B}, 2/2 (Modern
// Horizons 2, #5501):
//
//	"Flying, lifelink
//	 Whenever this creature or another creature you control enters, if that
//	 creature entered from a graveyard or you cast it from a graveyard, put a
//	 +1/+1 counter on this creature."
//
// One trigger, for the Knight and each other creature you control. "Entered
// from a graveyard" is the entry's own record (any graveyard); "cast it from
// a graveyard" is the permanent's cast provenance (CR 400.7d), read as the
// creature enters, and only for a spell its controller cast. Both are facts
// of the entry that cannot change afterwards, so the "if" is judged as the
// trigger is put on the stack and needs no second look at resolution
// (CR 603.4); the counter goes on the Knight only if it is still there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e15fb040-2e11-478e-87f1-55675d614f8e",
		Name:            "Breathless Knight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.OldB031EnteredFromOrCastFromAGraveyard,
				"Breathless Knight — put a +1/+1 counter on this creature", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
