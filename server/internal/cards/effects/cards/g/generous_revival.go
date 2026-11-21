package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Generous Revival — Sorcery {2}{W}:
//
//	"Return target creature card with mana value 3 or less from your
//	 graveyard to the battlefield with an additional +1/+1 counter on
//	 it.
//	 Flashback {4}{W} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// The counter is a CR 614.1c "enters with", so it rides the entry event:
// an enters trigger finds it already there, and a counter replacement
// (Doubling Season) sees it. A creature that enters with counters of
// its own keeps them and gains this one in addition.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "f8601f5f-0f7c-4689-9d15-137f1e82c49c",
		Name:             "Generous Revival",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{4}{W}")},
		Targets: effects.TargetCardInGraveyard("target creature card with mana value 3 or less from your graveyard",
			effects.YouOwn(), effects.Creature(), effects.ManaValueLE(3)),
		OnResolve: effects.ReturnTargetCreatureWithPlusOneCounter,
	})
}
