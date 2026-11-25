package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Devil's Play — Sorcery {X}{R} (Commander 2019):
//
//	"Devil's Play deals X damage to any target.
//	 Flashback {X}{R}{R}{R} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Banefire's X damage to any target (X is the announced value, read as the
// spell resolves, CR 107.3), cast from the graveyard under the shared flashback
// constructor with its own {X} in the flashback cost (CR 107.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "ee9f0b29-8a54-4cb8-8e2c-7bd67c2184ba",
		Name:             "Devil's Play",
		Completeness:     effects.CompletenessFull,
		XMatters:         true,
		Targets:          effects.TargetAny(),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{X}{R}{R}{R}")},
		OnResolve:        effects.DealXToTarget,
	})
}
