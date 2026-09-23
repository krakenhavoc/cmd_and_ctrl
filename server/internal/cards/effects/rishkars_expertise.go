package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rishkar's Expertise — Sorcery {4}{G}{G} (EDHREC rank 243):
//
//	"Draw cards equal to the greatest power among creatures you
//	 control.
//	 You may cast a spell with mana value 5 or less from your hand
//	 without paying its mana cost."
//
// The draw count is `b42GreatestPowerControlledBy`, the same scalar
// Avatar Kyoshi's mana ability and Orcish Siegemaster's trigger read.
// The free cast is a `choose_cards` pick over the hand — the same
// door Brainstorm's tuck-back and Malcolm's discard offer use —
// filtered to nonland cards with mana value 5 or less, floor zero
// ("you MAY cast"), then a one-shot `{0}` cast permission over
// whichever card was picked (`GrantCastPermissionOverCardForEffect`
// with `Zone` left empty auto-fills to wherever the card sits, which
// for a card that never left the hand is the hand itself).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "97407cd0-2bd2-4074-94d3-4ec3d243fa78",
		Name:         "Rishkar's Expertise",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			g := ctx.Game
			n := b42GreatestPowerControlledBy(g, item.Controller)
			if err := (DrawCards{Player: item.Controller, N: n}).Apply(ctx); err != nil {
				return err
			}
			return rishkarsExpertiseOfferFreeCast(ctx, item.Controller)
		},
	})
}

// rishkarsExpertiseOfferFreeCast is "you may cast a spell with mana
// value 5 or less from your hand without paying its mana cost".
//
// Not handCardsMatching: that helper is put_from_hand.go's "permanent
// card" filter, and "a spell" reaches instants and sorceries too —
// allHandCardIDs (brainstorm.go) is the type-agnostic list, filtered
// here against "nonland, mana value 5 or less".
func rishkarsExpertiseOfferFreeCast(ctx *Context, controller uuid.UUID) error {
	pred := And(Not(Land()), ManaValueLE(5))
	var candidates []uuid.UUID
	for _, id := range allHandCardIDs(ctx.Game, controller) {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !pred(ctx.Game, controller, c) {
			continue
		}
		candidates = append(candidates, id)
	}
	if len(candidates) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   ctx.Source(),
		Question: "Rishkar's Expertise — cast a spell with mana value 5 or less from your hand without paying its mana cost?",
		Cards:    candidates,
		Min:      0,
		Max:      1,
		Zone:     game.ZoneHand,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			g.GrantCastPermissionOverCardForEffect(picked[0], game.CastPermission{
				Player:   controller,
				Cost:     "{0}",
				CastOnly: true,
				Label:    "Rishkar's Expertise — cast it without paying its mana cost",
			})
			return nil
		},
	})
	return nil
}
