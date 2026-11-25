package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// March from the Tomb — Sorcery {3}{W}{B} (Battle for Zendikar, #5547):
//
//	"Return any number of target Ally creature cards with total mana value 8
//	 or less from your graveyard to the battlefield."
//
// The set rule is on the target clause, so the total is judged as the cards
// are chosen and again as the spell resolves; they enter together
// (PutCardsFromZoneTogether).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d2b9f3b-9ba1-4f28-8cd9-490803190e24",
		Name:         "March from the Tomb",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets: effects.TargetCardInGraveyard("any number of target Ally creature cards with total mana value 8 or less from your graveyard",
			effects.Creature(), effects.HasSubtype("Ally"), effects.YouOwn()).WithCount(0, 0).WithTotal(effects.TotalManaValueAtMost(8)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			return effects.PutCardsFromZoneTogether(ctx, game.ZoneGraveyard, ids, item.Controller, false)
		},
	})
}
