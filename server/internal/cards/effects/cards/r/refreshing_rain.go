package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Refreshing Rain — Instant {3}{G} (Nemesis, #5618):
//
//	"If an opponent controls a Swamp and you control a Forest, you may
//	 cast this spell without paying its mana cost.
//	 Target player gains 6 life."
//
// Sivvi's Ruse's free cast (the Nemesis cycle's condition: an
// OPPONENT's Swamp and your Forest) on Heroes' Reunion's body at 6. Any
// player may be the target, the caster included; one who left the game
// before it resolves is no longer legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6f69168-71c2-41a2-ab95-cfd7baf0ba68",
		Name:         "Refreshing Rain",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, LifeGain: 6}),
		Targets:      effects.TargetPlayer("target player"),
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "free",
			Label:     "Cast without paying its mana cost (an opponent controls a Swamp and you control a Forest)",
			ManaCost:  "",
			Condition: effects.AnOpponentControlsAAndYouControlA("Swamp", "Forest"),
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if p, ok := effects.FirstLegalPlayerTarget(ctx); ok {
				return effects.GainLife{Player: p, Amount: 6}.Apply(ctx)
			}
			return nil
		},
	})
}
