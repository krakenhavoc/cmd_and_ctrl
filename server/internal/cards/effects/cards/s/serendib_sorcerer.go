package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Serendib Sorcerer — Creature — Human Wizard Sorcerer {1}{U}{U}, 1/1 (Planar Chaos, #5595):
//
//	"{T}: Target creature other than this creature has base power and
//	 toughness 0/2 until end of turn."
//
// Sorceress Queen's ability: layer 7b base power and toughness until end
// of turn, so pumps and counters still apply on top of the 0/2. "Other
// than this creature" is object identity (Another).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "73a59b03-168e-472e-bbb1-cabd782ea014",
		Name:         "Serendib Sorcerer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature other than this creature has base power and toughness 0/2 until end of turn.",
			Cost:    effects.TapCost(),
			Targets: effects.Another(effects.TargetCreature("target creature other than this creature")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				target := effects.FirstLegalBattlefieldTarget(ctx)
				if target == uuid.Nil {
					return nil
				}
				return effects.UntilEndOfTurn(ctx, target, nil,
					"Serendib Sorcerer — base power and toughness 0/2 until end of turn",
					game.SetBasePTMods(0, 2)...)
			},
		}},
	})
}
