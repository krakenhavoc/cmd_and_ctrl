package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flying Drone — Artifact Creature — Robot Scout {1}{U}, 1/2 (Marvel Super
// Heroes Commander, #4707):
//
//	"Flying, vigilance
//	 {1}{U}, {T}: Draw a card, then discard a card. This ability costs {1}{U}
//	 less to activate if another creature with flying entered the battlefield
//	 under your control this turn."
//
// The discount is the ability's own cost clause (ActivatedAbility.CostModifiers,
// CR 601.2f via 602.2b) and a COLOURED reduction: ActivationCostsLessMana takes
// {1}{U} off the mana cost, leaving {T} alone to pay (CR 118.7, ADR 0048's
// 2026-10-30 amendment). The condition is read from the turn tally's record of
// creatures that entered with flying (game.EnteredFlierUnderThisTurn), written
// as each one enters, so a flier that has since died or lost flying still
// counts, one an opponent controls does not, and the Drone's own entry is
// skipped by instance ("another").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "541f7525-a9a4-493d-b724-354ae6d37de5",
		Name:            "Flying Drone",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}, {T}: Draw a card, then discard a card. This ability costs {1}{U} less to activate if another creature with flying entered the battlefield under your control this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}{U}"), TapCost()),
			CostModifiers: []game.CostModifier{
				ActivationCostsLessMana("{1}{U}", "This ability costs {1}{U} less to activate if another creature with flying entered the battlefield under your control this turn.",
					AnotherFlierEnteredUnderYourControlThisTurn()),
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b16DrawThenDiscard(g, item, 1)
			},
		}},
	})
}
