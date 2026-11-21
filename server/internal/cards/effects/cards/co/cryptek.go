package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cryptek — Artifact Creature — Necron Wizard {3}{B}, 3/3 (Warhammer
// 40,000 Commander, S82 GAP batch 187, #5948):
//
//	"{1}{B}, {T}: Choose another target artifact creature you control. When
//	 that creature dies this turn, return it to the battlefield tapped
//	 under your control."
//
// The ability only sets up a delayed trigger on the chosen object (CR
// 603.7): it fires if that creature dies this turn, and returns the card as
// it was put into the graveyard, tapped, under Cryptek's controller (not
// necessarily the owner). A creature that has become another object, or a
// token, is not returned.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27a4b633-5a62-4d8c-8cc6-44959c311de9",
		Name:         "Cryptek",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, {T}: Choose another target artifact creature you control. When that creature dies this turn, return it to the battlefield tapped under your control.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost()),
			Targets: effects.Another(effects.TargetCreature("another target artifact creature you control", effects.Artifact(), effects.YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := effects.FirstLegalTarget(ctx)
				if !ok {
					return nil
				}
				return effects.WhenThatCreatureDiesThisTurn(ctx, t.ID, "Cryptek — return it to the battlefield tapped under your control", effects.GapB187ReturnDiedTappedBody)
			},
		}},
	})
}
