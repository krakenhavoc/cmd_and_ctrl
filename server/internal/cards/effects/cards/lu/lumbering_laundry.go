package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lumbering Laundry — Artifact Creature — Golem {5}, 4/5 (Murders at
// Karlov Manor, #4469):
//
//	"{2}: Until end of turn, you may look at face-down creatures you
//	 don't control any time.
//	 Disguise {5} (You may cast this card face down for {3} as a 2/2
//	 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)"
//
// The activated ability writes a turn-scoped record
// (game.ModLookAtFaceDownCreatures, #5251) that lasts the turn whether
// or not the Laundry stays (CR 611.2a) and ends at cleanup (CR 514.2).
// While it lasts the activator's view shows the card under every
// face-down creature another player controls; nobody else's view
// changes, and it is never stored on the creatures, so the next frame
// after cleanup they are face-down 2/2s again. A restore point taken
// during the turn keeps the record.
//
// A face-down Laundry has no abilities (CR 708.2a), so it cannot
// activate the look until it is turned face up. Disguise is the shared
// constructor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "6afe9cc7-80e4-4607-8633-6013d96bd19c",
		Name:             "Lumbering Laundry",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Disguise("{5}")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Until end of turn, you may look at face-down creatures you don't control any time.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}"),
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				ctx.Game.LookAtFaceDownCreaturesThisTurnForEffect(ctx.Item.SourceCardID, ctx.Controller(),
					"Lumbering Laundry — look at face-down creatures")
				return nil
			})),
		}},
	})
}
