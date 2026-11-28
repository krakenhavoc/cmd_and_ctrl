package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maddening Imp — Creature — Imp {2}{B}, 1/1 (Tempest, S82 OLD, #5628):
//
//	"Flying
//	 {T}: Non-Wall creatures the active player controls attack this turn if
//	 able. At the beginning of the next end step, destroy each of those
//	 creatures that didn't attack this turn. Activate only during an
//	 opponent's turn and only before combat."
//
// Siren's Call's sweep with Nettling Imp's window: each non-Wall creature the
// active player controls as the ability resolves must attack (CR 508.1d) and
// gets its own delayed destroy at the next end step unless it attacked
// (CR 603.7). Creatures that arrive later are not "those creatures". The
// window is an opponent's turn from untap through the precombat main phase.
//
// No simplification.
func init() {
	const name = "Maddening Imp"
	effects.Register(effects.Spec{
		OracleID:        "97dea59e-2090-4653-9ed9-86820a20605f",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: Non-Wall creatures the active player controls attack this turn if able. At the beginning of the next end step, destroy each of those creatures that didn't attack this turn. Activate only during an opponent's turn and only before combat.",
			// A wipe in the audit's eyes: it destroys the active player's non-Wall creatures that
			// did not attack, a class it spares by condition (ADR 0126 §6).
			Purpose:   game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, OpponentsOnly: true, Partial: true}},
			Cost:      effects.TapCost(),
			Condition: effects.AllConditions(effects.DuringAnOpponentsTurn(), effects.S82OldB761BeforeCombat()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				match := effects.And(effects.Creature(), effects.NonWall(), effects.ControlledByTheActivePlayer())
				for _, c := range g.BattlefieldCardsForEffect() {
					if !match(g, ctx.Controller(), c) {
						continue
					}
					if err := effects.MustAttackOrBeDestroyed(ctx, c.InstanceID, name); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
