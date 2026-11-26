package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rally the Ancestors — {X}{W}{W} Instant (Fate Reforged):
//
//	"Return each creature card with mana value X or less from your graveyard to
//	 the battlefield. Exile those creatures at the beginning of your next
//	 upkeep. Exile Rally the Ancestors."
//
// X is the announced value (CR 107.3). The cards are chosen as the spell
// resolves and enter together (CR 608.2c); the delayed exile (CR 603.7) names
// the creatures that actually entered, as objects, so one that has left and
// returned is not exiled. Rally exiles itself, so it never reaches the
// graveyard.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d612686-2d7f-49a3-b9c4-1e99a30134f6",
		Name:         "Rally the Ancestors",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			ids := effects.GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
				return c.IsCreature() && !c.IsToken() && c.ManaValue() <= x
			})
			if len(ids) > 0 {
				if err := (effects.ReturnFromGraveyardTogether{
					Targets:    ids,
					Controller: ctx.Controller(),
					ThenTail:   s82OldB338RallyExile,
				}).Apply(ctx); err != nil {
					return err
				}
			}
			if effects.SourceIsNewObject(ctx.Game, item) {
				return nil
			}
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}

// s82OldB338RallyExile schedules "exile those creatures at the beginning of
// your next upkeep" for the creatures that entered. An on-disk key: never
// renamed, never reused.
var s82OldB338RallyExile = effects.TogetherTail("then/s82-old-b338-rally-the-ancestors-exile", func(ctx *effects.Context, entered []uuid.UUID, _ int) error {
	if len(entered) == 0 {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		At:                 game.StepUpkeep,
		ControllerTurnOnly: true,
		Label:              "Rally the Ancestors — exile those creatures",
		Cards:              entered,
		Body:               effects.ExileListedCardsBody,
	}.Apply(ctx)
})
