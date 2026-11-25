package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ghastly Conscription — Sorcery {5}{B}{B} (Commander 2019):
//
//	"Exile all creature cards from target player's graveyard in a face-down
//	 pile, shuffle that pile, then manifest those cards. (To manifest a card,
//	 put it onto the battlefield face down as a 2/2 creature. Turn it face up
//	 any time for its mana cost if it's a creature card.)"
//
// Every creature card in the target player's graveyard, read as the spell
// resolves, is exiled together (a target that has left or become illegal does
// nothing, CR 608.2b), then manifested under the caster's control. The cards
// enter together face down (CR 701.40a): the controller becomes the only
// knower of each (CR 708.5), so the printed shuffle of the pile cannot change
// what anyone can tell and the entry order is not one. A card a replacement kept
// out stays in exile. The face-down pile groups cards into a pile (CR 700.3),
// so the spell notes the turn's pile-or-guess tally as it resolves (#6175).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3bf20e68-0e19-4242-bfba-a8a1bb244685",
		Name:         "Ghastly Conscription",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetPlayer {
				return nil
			}
			effects.NotePileOrGuess(ctx)
			ids := effects.GraveyardIDs(ctx.Game, t.ID, func(c game.Card) bool { return c.IsCreature() })
			if len(ids) == 0 {
				return nil
			}
			return ctx.Game.ExileCardsKeyedForEffect(ids, ctx.ItemTail(ghastlyConscriptionExiled))
		},
	})
}

// ghastlyConscriptionExiled manifests the exiled cards, all at once. An
// on-disk key: never renamed, never reused.
var ghastlyConscriptionExiled = effects.CardTail("then/ghastly-conscription-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.PutCardsFromZoneTogetherWith(ctx, game.ZoneExile, append([]uuid.UUID(nil), o.IDs...), game.ZoneEntryOptions{
		Controller: ctx.Controller(),
		FaceDown:   game.FaceDownManifested,
	}, game.TailRef{}, 0)
})
