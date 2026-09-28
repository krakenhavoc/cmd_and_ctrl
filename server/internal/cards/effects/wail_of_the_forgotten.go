package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wail of the Forgotten — {U}{B} Sorcery:
//
//	"Descend 8 — Choose one. If there are eight or more permanent cards
//	 in your graveyard as you cast this spell, choose one or more
//	 instead.
//	 • Return target nonland permanent to its owner's hand.
//	 • Target opponent discards a card.
//	 • Look at the top three cards of your library. Put one of them
//	   into your hand and the rest into your graveyard."
//
// Descend 8 is a board condition read "as you cast this spell"
// (#1655): the maximum becomes every bullet and the minimum stays one,
// so AnyNumberIf(Descended8). A permanent card is one whose types
// include artifact, creature, enchantment, land, planeswalker or
// battle (CR 110.4), counted off the printed type line in the
// graveyard.
//
// "The rest into your graveyard" is a move, not a mill, and routes
// through the graveyard replacements (TakeRestIntoGraveyard). No
// simplification.
func init() {
	Register(Spec{
		OracleID:     "030b5408-f216-43e4-8593-f78d22821876",
		Name:         "Wail of the Forgotten",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Return target nonland permanent to its owner's hand.",
				TargetPermanent("target nonland permanent", Nonland()),
				BounceTheModesTarget),
			ModeDoing("Target opponent discards a card.",
				TargetPlayer("target opponent", Opponent()),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
						Player: t.ID,
						Source: item.SourceCardID,
						N:      1,
					})
					return nil
				}),
			ModeDoing("Look at the top three cards of your library. Put one of them into your hand and the rest into your graveyard.",
				nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					player := item.Controller
					return TakeFromLibraryToHand{
						Player: player,
						Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 3),
						Max:    1,
						Label:  "Wail of the Forgotten — put one into your hand",
						Then:   TakeRestIntoGraveyard,
					}.Apply(ctx)
				}),
		).AnyNumberIf(Descended8),
	})
}

// Descended8 is descend 8's mode count — "If there are eight or more
// permanent cards in your graveyard as you cast this spell" (#1655).
var Descended8 = game.ModeCondition("descend-8", func(g *game.Game, chooser uuid.UUID) bool {
	return permanentCardsInGraveyard(g, chooser) >= 8
})

// permanentCardsInGraveyard counts the permanent cards (CR 110.4) in
// `player`'s graveyard — descend's count. Printed characteristics: a
// card in a graveyard has no layer cache.
func permanentCardsInGraveyard(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.IsPermanent() {
			n++
		}
	}
	return n
}
