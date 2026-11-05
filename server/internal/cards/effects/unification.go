package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unification — Enchantment — Saga {3}{W}{U} (Star Trek Commander, #4974):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — The owner of up to one target creature draws two cards, then
//	     shuffles that creature into their library.
//	 II — You may put a Spacecraft card from your hand or graveyard onto
//	      the battlefield.
//	 III — You and target opponent each draw two cards."
//
// Chapter I reads the owner off the creature as it resolves, has them
// draw, then tucks the creature and shuffles that owner's library (Chaos
// Warp's tuck, so a commander's owner is offered the command zone, CR
// 903.9). With no creature chosen, or the one chosen no longer legal,
// nobody draws (CR 608.2b).
//
// Chapter II is one optional pick over the Spacecraft cards in your hand
// and your graveyard together, then the put through whichever zone's
// entry door the picked card is still in: the hand door or the graveyard
// door. A card that has moved since the prompt opened is not put.
//
// Chapter III is Love Song of Night and Day's chapter I.
//
// No simplification.
func init() {
	const name = "Unification"
	Register(Spec{
		OracleID:     "b0efb6fb-f2d1-4027-90b5-44fb3f5dd538",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, "the owner of up to one target creature draws two cards, then shuffles it into their library"),
				TargetCreature("up to one target creature").WithCount(0, 1),
				unificationDrawThenShuffle),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "you may put a Spacecraft card from your hand or graveyard onto the battlefield"),
				unificationPutSpacecraft),
			ChapterTriggerTargeting(3, SagaChapterLabel(name, 3, "you and target opponent each draw two cards"),
				TargetPlayer("target opponent", Opponent()),
				loveSongEachDrawTwo),
		},
	})
}

// unificationDrawThenShuffle is chapter I.
func unificationDrawThenShuffle(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := battlefieldCardForEffect(g, id)
	if !ok {
		return nil
	}
	owner := c.Owner
	if err := (DrawCards{Player: owner, N: 2}).Apply(ctx); err != nil {
		return err
	}
	if !onBattlefield(g, id) {
		return nil
	}
	return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{id}, game.TuckOptions{},
		game.KeyedTail{Ref: thenShuffle, Player: owner})
}

// unificationPutSpacecraft is chapter II: the pick over hand and
// graveyard. No Spacecraft card in either asks nothing.
func unificationPutSpacecraft(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	you := ctx.Controller()
	isSpacecraft := func(c game.Card) bool { return c.HasSubtypeWord("Spacecraft") }
	candidates := handCardsMatching(g, you, func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return isSpacecraft(c) })
	candidates = append(candidates, graveyardCardIDs(ctx, you, isSpacecraft)...)
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Destination: game.ChooseOntoBattlefield,
		Chooser:     you,
		Source:      ctx.Source(),
		Question:    "Unification — you may put a Spacecraft card from your hand or graveyard onto the battlefield",
		Cards:       candidates,
		Min:         0,
		Max:         1,
		ThenKey:     unificationSpacecraftPicked,
	})
	return nil
}

// unificationSpacecraftPicked puts the picked Spacecraft card onto the
// battlefield from the zone it is in now. An on-disk key: never renamed,
// never reused.
var unificationSpacecraftPicked = game.RegisterChooseCardsThen("choose-cards/unification-put-spacecraft", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	id := r.Picked[0]
	p := g.PlayerByIDForEffect(r.Chooser)
	if p == nil {
		return nil
	}
	z := g.FindCardZoneForEffect(id)
	if z == nil {
		return nil
	}
	switch z {
	case p.Hand:
		return g.PutFromHandOntoBattlefieldKeyedForEffect(id, game.HandEntryOptions{Controller: r.Chooser}, game.KeyedTail{})
	case p.Graveyard:
		return g.ReturnFromGraveyardToBattlefieldKeyedForEffect(id, r.Chooser, false, game.KeyedTail{})
	}
	return nil
})
