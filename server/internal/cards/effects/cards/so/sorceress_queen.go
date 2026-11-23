package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sorceress Queen — Creature — Human Wizard Sorcerer {1}{B}{B}, 1/1
// (Arabian Nights, #5641):
//
//	"{T}: Target creature other than this creature has base power and
//	 toughness 0/2 until end of turn."
//
// Genemorph Imago's layer 7b base power and toughness until end of
// turn, so pumps and counters still apply on top of the 0/2. "Other
// than this creature" is object identity (Another).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e4cb1b2-e2cc-4925-a226-6c6f1501d9c1",
		Name:         "Sorceress Queen",
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
					"Sorceress Queen — base power and toughness 0/2 until end of turn",
					game.SetBasePTMods(0, 2)...)
			},
		}},
	})
}
