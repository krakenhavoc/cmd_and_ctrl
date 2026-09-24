package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kari Zev's Expertise — Sorcery {1}{R}{R}:
//
//	"Gain control of target creature or Vehicle until end of turn.
//	 Untap it. It gains haste until end of turn.
//	 You may cast a spell with mana value 2 or less from your hand
//	 without paying its mana cost."
//
// The theft is Act of Treason's three primitives (#756) over a wider
// target — "creature or Vehicle" is Agonasaur Rex's clause,
// Or(Creature(), Subtype("Vehicle")) — and the free cast is Rishkar's
// Expertise's second sentence with the mana-value ceiling lowered from
// five to two: a choose_cards pick over the hand, floor zero ("you
// MAY"), filtered to nonland cards with mana value 2 or less, then a
// one-shot {0} cast permission over whichever card was picked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e0de0121-6185-45c5-9d88-a415b07347e0",
		Name:         "Kari Zev's Expertise",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or Vehicle", Or(Creature(), Subtype("Vehicle"))),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (GainControl{
				Target:   target,
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Kari Zev's Expertise — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"haste"},
				Label:    "Kari Zev's Expertise — haste until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return kariZevsExpertiseOfferFreeCast(ctx, item.Controller)
		},
	})
}

// kariZevsExpertiseOfferFreeCast is "you may cast a spell with mana
// value 2 or less from your hand without paying its mana cost" —
// Rishkar's Expertise's rishkarsExpertiseOfferFreeCast with the
// ceiling lowered to two.
func kariZevsExpertiseOfferFreeCast(ctx *Context, controller uuid.UUID) error {
	pred := And(Not(Land()), ManaValueLE(2))
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
		Question: "Kari Zev's Expertise — cast a spell with mana value 2 or less from your hand without paying its mana cost?",
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
				Label:    "Kari Zev's Expertise — cast it without paying its mana cost",
			})
			return nil
		},
	})
	return nil
}
