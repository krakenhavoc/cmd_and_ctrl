package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Revivify — Instant {2}{W} (Forgotten Realms Commander, #5681):
//
//	"Roll a d20 and add the number of creature cards in your graveyard
//	 that were put there from the battlefield this turn.
//	 1—14 | Return all creature cards in your graveyard that were put
//	 there from the battlefield this turn to your hand.
//	 15+ | Return those cards from your graveyard to the battlefield."
//
// The count is the roll's modifier (CR 706.2, RollDice.Modifier), read as
// the spell resolves from this turn's record of what went to the
// graveyard from the battlefield (Brought Back's clause): a creature card
// that was discarded or milled does not count, and neither does one that
// left the graveyard and came back another way. The result, natural roll
// plus count, picks the row (CR 706.3a). The 15+ row puts every such card
// onto the battlefield at once (ReturnFromGraveyardTogether, CR 603.6a),
// under its owner's control, which is you. With no such card the roll
// still happens and nothing returns.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "55eb3c21-6834-4c62-8ab9-54b901531fed",
		Name:         "Revivify",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return RollDice{Sides: 20, N: 1, Modifier: len(revivifyCards(ctx)), Then: revivifyResult}.Apply(ctx)
		},
	})
}

// revivifyCards are the creature cards in the caster's graveyard that
// were put there from the battlefield this turn.
func revivifyCards(ctx *Context) []uuid.UUID {
	return graveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
		return c.IsCreature() && putIntoGraveyardFromBattlefieldThisTurn(ctx.Game, c.InstanceID)
	})
}

// revivifyResult is the results table, after the d20. An on-disk key:
// never renamed, never reused.
var revivifyResult = CardTail("dice/revivify-result", func(ctx *Context, o game.TailOutcome) error {
	result := DiceResult(o)
	cards := revivifyCards(ctx)
	if len(o.Rolls) == 0 || result < 1 || len(cards) == 0 {
		return nil
	}
	if result >= 15 {
		return ReturnFromGraveyardTogether{Targets: cards}.Apply(ctx)
	}
	for _, id := range cards {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
