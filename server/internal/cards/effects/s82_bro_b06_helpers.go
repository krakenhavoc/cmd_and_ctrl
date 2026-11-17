package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bro_b06_helpers.go — the shared bodies of The Brothers' War batch 6:
// the "mill three cards, you may put a <kind> card from among the cards
// milled this way into your hand, if you don't, put a +1/+1 counter on
// this creature" cycle, and the graveyard-count condition of the Gixian
// Skullflayer and Gaea's Courser triggers.

// bro6MillThreeTakeOne is the cycle's enters trigger. `match` says which
// of the milled cards may be taken; `what` is the printed phrase for the
// prompt. The pick is over the cards this mill put into the graveyard
// (CR 400.7), at most one, and a card not taken, whether none matched or
// the player declined, is the counter. The two keys are on-disk
// identities derived from slug: never renamed, never reused.
func bro6MillThreeTakeOne(card, slug, what string, match CardPredicate) game.TriggeredAbility {
	pick := askedPick("choose-cards/bro-b06-"+slug+"-take-or-counter", func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) == 0 {
			return CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
		}
		return ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneHand}.Apply(ctx)
	})
	tail := CardTail("bro-b06/"+slug+"/after-mill", func(ctx *Context, o game.TailOutcome) error {
		player := ctx.Controller()
		var eligible []uuid.UUID
		for _, id := range o.IDs {
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || !match(ctx.Game, player, c) {
				continue
			}
			if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
				continue
			}
			eligible = append(eligible, id)
		}
		if len(eligible) == 0 {
			return CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  player,
			Source:   ctx.Source(),
			Question: card + " — you may put " + what + " from among the cards milled this way into your hand (if you don't, put a +1/+1 counter on it)",
			Cards:    eligible,
			Min:      0,
			Max:      1,
			LiveZone: game.ZoneGraveyard,
			ThenKey:  pick,
			Asked:    game.NewAskedItem(ctx.Item),
		})
		return nil
	})
	return WhenThisEnters(card+" — mill three cards; you may put "+what+" from among them into your hand, or put a +1/+1 counter on it",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			return MillToZone{N: 3, ThenTail: ctx.ItemTail(tail)}.Apply(ctx)
		})
}

// CondThreeCreatureCardsInYourGraveyard is "if there are three or more
// creature cards in your graveyard" declared as an intervening if
// (InterveningIf): asked as the ability triggers and again as it resolves
// (CR 603.4). A card in a graveyard has no layer cache, so this is the
// printed type line.
type CondThreeCreatureCardsInYourGraveyard struct{}

func (CondThreeCreatureCardsInYourGraveyard) Holds(ctx *Context) bool {
	return b11CreatureCardsInGraveyard(ctx.Game, ctx.Controller()) >= 3
}

func (CondThreeCreatureCardsInYourGraveyard) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondGraveyard, 3), true
}
