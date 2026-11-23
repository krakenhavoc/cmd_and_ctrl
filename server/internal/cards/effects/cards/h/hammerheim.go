package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hammerheim — Legendary Land (Tales of Middle-earth Commander):
//
//	"{T}: Add {R}.
//	 {T}: Target creature loses all landwalk abilities until end of turn."
//
// A layer-6 removal (CR 613.1f) of every landwalk keyword the engine knows
// (plainswalk, islandwalk, swampwalk, mountainwalk, forestwalk, desertwalk
// and nonbasic landwalk), pinned to the target with its own timestamp, so a
// later grant puts one back (Thieves' Auction-style). A target that is gone
// does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7476beb-7923-4994-8476-bc69187ecb72",
		Name:         "Hammerheim",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature loses all landwalk abilities until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.UntilEndOfTurn(ctx, id, nil, "Hammerheim — loses all landwalk abilities until end of turn",
					game.RemoveKeywordsMod("plainswalk", "islandwalk", "swampwalk", "mountainwalk", "forestwalk", "desertwalk", "nonbasic landwalk"))
			},
		}},
	})
}
