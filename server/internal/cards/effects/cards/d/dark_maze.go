package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dark Maze — Creature — Wall {4}{U}, 4/5 (Homelands, S82 OLD b821, #5635):
//
//	"Defender
//	{0}: This creature can attack this turn as though it didn't have defender.
//	Exile it at the beginning of the next end step."
//
// Mirror Wall's waiver of defender (CR 702.3b) with a cost of {0}, and the
// exile that comes with it is a delayed trigger (CR 603.7) on the Maze as it
// is now: Molten Echoes's ExileListedCardsBody. Each activation schedules its
// own exile; a Maze that has left by then is not exiled again (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "44c65424-9c44-4bd8-9022-7ef5bb5b898e",
		Name:            "Dark Maze",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{0}: This creature can attack this turn as though it didn't have defender. Exile it at the beginning of the next end step.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{0}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				if err := effects.WaiveDefenderUntilEOT(ctx, item.SourceCardID, "Dark Maze — can attack this turn as though it didn't have defender"); err != nil {
					return err
				}
				return effects.ScheduleDelayedTrigger{
					Label: "Dark Maze — exile it",
					Cards: []uuid.UUID{item.SourceCardID},
					Body:  effects.ExileListedCardsBody,
				}.Apply(ctx)
			},
		}},
	})
}
