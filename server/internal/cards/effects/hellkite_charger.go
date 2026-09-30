package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hellkite Charger — Creature — Dragon {4}{R}{R}, 5/5:
//
//	"Flying, haste
//	 Whenever this creature attacks, you may pay {5}{R}{R}. If you do,
//	 untap all attacking creatures and after this phase, there is an
//	 additional combat phase."
//
// The payment is MayPay held in the step it was asked in (InThisStep),
// so the answer is given before the combat it is about can end: "after
// this phase" is the combat in progress, and an answer that arrived in
// the next main phase would add the combat after the wrong phase. The
// additional combat has no main phase before it (ADR 0059 sub-PR 2b,
// #753). It attacks again in the added combat and asks again, which is
// how the card goes infinite with enough mana.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d4f28a4b-d821-4132-bdec-4c528318f8e2",
		Name:            "Hellkite Charger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Hellkite Charger — pay {5}{R}{R} for an additional combat", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:    item.Controller,
					Cost:       "{5}{R}{R}",
					Question:   "Hellkite Charger — pay {5}{R}{R} to untap all attacking creatures and get an additional combat phase?",
					OnPay:      untapAttackersThenExtraCombat,
					InThisStep: true,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
