package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gorilla War Cry — Instant {1}{R} (Alliances, #5634):
//
//	"Cast this spell only during combat before blockers are declared.
//	 All creatures gain menace until end of turn. (They can't be blocked
//	 except by two or more creatures.)
//	 Draw a card at the beginning of the next turn's upkeep."
//
// The cast window is the combat phase before its declare blockers step (the
// beginning of combat and declare attackers steps, CR 506.8). The creatures
// are fixed as the spell resolves (CR 611.2c): every creature on the
// battlefield, everyone's, gains menace until end of turn, and the delayed draw
// at the next upkeep at the table is scheduled either way (CR 603.7a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "60a7f61c-41ad-4dce-b556-a99bf38c3d73",
		Name:               "Gorilla War Cry",
		Completeness:       effects.CompletenessFull,
		Purpose:            game.Purpose{Draws: 1},
		CastCondition:      effects.OldB814DuringCombatBeforeBlockers,
		CastConditionLabel: "Cast this spell only during combat before blockers are declared.",
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.GrantKeywordUntilEOT{
				Match:    effects.Creature(),
				Keywords: []string{"menace"},
				Label:    "Gorilla War Cry — menace until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawACardAtTheNextUpkeep("Gorilla War Cry — draw a card").Apply(ctx)
		},
	})
}
