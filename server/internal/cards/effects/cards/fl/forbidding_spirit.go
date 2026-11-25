package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forbidding Spirit — Creature — Spirit Cleric {1}{W}{W}, 3/3 (Ravnica
// Allegiance):
//
//	"When this creature enters, until your next turn, creatures can't attack you
//	 or planeswalkers you control unless their controller pays {2} for each of
//	 those creatures."
//
// Summon: Yojimbo's timed attack tax, widened to the controller's planeswalkers
// (Game.AttackTaxPlaneswalkersForEffect, a ModAttackTax record whose Types name
// planeswalker). It lasts until the start of the controller's next turn.
//
// No simplification.
func init() {
	const clause = "Forbidding Spirit — until your next turn, creatures can't attack you or planeswalkers you control unless their controller pays {2} for each of those creatures."
	effects.Register(effects.Spec{
		OracleID:     "efd8dc6c-8233-45b2-99bd-e38ca2e0b35f",
		Name:         "Forbidding Spirit",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(clause, func(g *game.Game, item *game.StackItem) error {
				g.AttackTaxPlaneswalkersForEffect(item.SourceCardID, item.Controller, "{2}",
					g.UntilYourNextTurnDuration(item.Controller), clause)
				return nil
			}),
		},
	})
}
