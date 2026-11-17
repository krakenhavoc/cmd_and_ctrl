package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Queen Kayla bin-Kroog — Legendary Creature — Human Noble {1}{R}{W}, 2/3
// (The Brothers' War promo, S82):
//
//	"{4}, {T}: Discard all the cards in your hand, then draw that many
//	 cards. You may choose an artifact or creature card with mana value 1
//	 you discarded this way, then do the same for artifact or creature cards
//	 with mana values 2 and 3. Return those cards to the battlefield.
//	 Activate only as a sorcery."
//
// Forgotten Creation's wheel: the whole hand is discarded and the draw is as
// many as REALLY reached a graveyard (read off the discard's continuation).
// Then, for mana value 1, 2 and 3 in that order, the controller is asked for
// an artifact or creature card of that value among the cards just discarded
// and still in a graveyard (Atraxa's one-prompt-per-step pattern; a step with
// no candidate is skipped, each answer is optional), and the cards chosen are
// returned to the battlefield together, under their owners' control (CR
// 110.2). A card with X in its cost has mana value 0 off the stack (CR
// 202.3e), so it is never a 1, 2 or 3.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "10c2428f-6e4a-4b92-9087-720242a5c25c",
		Name:         "Queen Kayla bin-Kroog",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{4}, {T}: Discard all the cards in your hand, then draw that many cards. You may choose an artifact or creature card with mana value 1 you discarded this way, then do the same for artifact or creature cards with mana values 2 and 3. Return those cards to the battlefield. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{4}"), TapCost()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return socB21DiscardHandThen(NewContext(g, item), item, queenKaylaDiscarded)
			},
		}},
	})
}

// queenKaylaDiscarded draws as many cards as were discarded, then starts the
// choices over the discarded cards. An on-disk key.
var queenKaylaDiscarded = CardTail("queen-kayla-bin-kroog/after-discard", func(ctx *Context, o game.TailOutcome) error {
	discarded := o.IDs
	if err := (DrawCards{N: len(discarded)}).Apply(ctx); err != nil {
		return err
	}
	return queenKaylaAsk(ctx, discarded, nil, 1)
})

// queenKaylaAsk asks for the card of mana value `mv` (1..3) among the
// discarded cards still in a graveyard, `chosen` being what the earlier steps
// picked, and returns the chosen cards together once mana value 3 is done.
func queenKaylaAsk(ctx *Context, discarded, chosen []uuid.UUID, mv int) error {
	g, you := ctx.Game, ctx.Controller()
	for ; mv <= 3; mv++ {
		var candidates []uuid.UUID
		for _, id := range discarded {
			z := g.FindCardZoneForEffect(id)
			c, ok := g.LookupCardForEffect(id)
			if !ok || z == nil || z.Kind != game.ZoneGraveyard {
				continue
			}
			if (c.IsArtifact() || c.IsCreature()) && c.ManaValue() == mv {
				candidates = append(candidates, id)
			}
		}
		if len(candidates) == 0 {
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    you,
			FromPlayer: you,
			Source:     ctx.Source(),
			Question:   "Queen Kayla bin-Kroog — you may choose an artifact or creature card with mana value " + string(rune('0'+mv)) + " you discarded this way",
			Cards:      candidates,
			Min:        0,
			Max:        1,
			LiveZone:   game.ZoneGraveyard,
			ThenKey:    queenKaylaPicked,
			Carry:      packIDLists(discarded, chosen),
			CarryNums:  []int{mv},
			Asked:      game.NewAskedItem(ctx.Item),
		})
		return nil
	}
	var back []uuid.UUID
	for _, id := range chosen {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneGraveyard {
			back = append(back, id)
		}
	}
	if len(back) == 0 {
		return nil
	}
	return ReturnFromGraveyardTogether{Targets: back}.Apply(ctx)
}

// queenKaylaPicked is one step's answer: the pick joins the chosen cards, then
// the next mana value is asked. An on-disk key.
var queenKaylaPicked game.ChooseCardsThen

func init() {
	queenKaylaPicked = askedPick("choose-cards/queen-kayla-bin-kroog-mana-value", func(ctx *Context, r game.CardsChosen) error {
		discarded, chosen := unpackIDPair(r.Carry)
		chosen = append(chosen, r.Picked...)
		mv := 0
		if len(r.Nums) > 0 {
			mv = r.Nums[0]
		}
		return queenKaylaAsk(ctx, discarded, chosen, mv+1)
	})
}
