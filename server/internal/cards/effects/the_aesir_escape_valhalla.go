package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Aesir Escape Valhalla — Enchantment — Saga {2}{G} (Assassin's Creed,
// #5948):
//
//	"(As this Saga enters and after your draw step, add a lore counter.)
//	 I — Exile a permanent card from your graveyard. You gain life equal to its
//	 mana value.
//	 II — Put a number of +1/+1 counters on target creature you control equal to
//	 the mana value of the exiled card.
//	 III — Return this Saga and the exiled card to their owner's hand."
//
// Chapter I is a choice, not a target: the controller picks a permanent card in
// their graveyard (a choose-cards prompt, mandatory while there is one) and it is
// exiled linked to this Saga object (ExileTarget.Linked, CR 607.2a: The Measure of a
// Man's link), then the life is the mana value of the card as it sits in exile (a
// card with no mana cost is 0, CR 202.3). The gain waits for the exile to happen
// (a commander offered the command zone gains nothing). Chapter II reads "the exiled
// card" back through that link as it resolves, and puts that many counters on its
// target (none if nothing was exiled or the Saga is a new object, CR 400.7).
// Chapter III returns the Saga, if it is still this object, and the card exiled with
// it to their owner's hands (Angel of Serenity's bounce); the Saga going to the hand
// is what ends it, as it would be sacrificed after III.
//
// No simplification.
func init() {
	const name = "The Aesir Escape Valhalla"
	Register(Spec{
		OracleID:     "6afe197b-0779-4284-a192-357a3c89b7e3",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "exile a permanent card from your graveyard; you gain life equal to its mana value"),
				aesirChapterOne),
			ChapterTriggerTargeting(2, SagaChapterLabel(name, 2, "put a number of +1/+1 counters on target creature you control equal to the mana value of the exiled card"),
				TargetCreature("target creature you control", YouControl()),
				aesirChapterTwo),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "return this Saga and the exiled card to their owner's hand"),
				aesirChapterThree),
		},
	})
}

// aesirChapterOne asks for the card; with none to exile nothing happens.
func aesirChapterOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	candidates := graveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool { return c.IsPermanent() })
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "The Aesir Escape Valhalla — exile a permanent card from your graveyard",
		Cards:    candidates,
		Min:      1,
		Max:      1,
		ThenKey:  aesirExilePicked,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

// aesirExilePicked exiles the chosen card linked to the Saga. An on-disk key.
var aesirExilePicked = itemChoice("choose-cards/aesir-escape-valhalla-exile", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	id := r.Picked[0]
	if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ExileTarget{Target: id, Linked: true, ThenTail: withCarry(ctx.ItemTail(aesirExiled), id)}.Apply(ctx)
})

// aesirExiled is "you gain life equal to its mana value", once the exile
// happened. An on-disk key.
var aesirExiled = CardTail("aesir-escape-valhalla/after-exile", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(tailCarry0(o))
	if !ok {
		return nil
	}
	return GainLife{Amount: c.ManaValue()}.Apply(ctx)
})

// aesirExiledWithThisSaga is the cards in exile linked to this Saga object.
func aesirExiledWithThisSaga(ctx *Context) []game.Card {
	ref, ok := ctx.Game.SourceObjectForEffect(ctx.Item)
	if !ok {
		return nil
	}
	var out []game.Card
	for _, c := range ctx.Game.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		out = append(out, *c)
	}
	return out
}

// aesirChapterTwo puts counters equal to the exiled card's mana value on the
// target creature.
func aesirChapterTwo(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	n := 0
	for _, c := range aesirExiledWithThisSaga(ctx) {
		n += c.ManaValue()
	}
	if n <= 0 {
		return nil
	}
	return AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
}

// aesirChapterThree returns this Saga and the exiled card to their owner's hand.
func aesirChapterThree(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range aesirExiledWithThisSaga(ctx) {
		ids = append(ids, c.InstanceID)
	}
	if ref, ok := ctx.SourceRef(); ok {
		if c, onbf := battlefieldCardForEffect(g, ref.ID); onbf && c.ObjectEpoch == ref.Epoch {
			if err := (BounceToHand{Target: ref.ID}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	for _, id := range ids {
		if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
