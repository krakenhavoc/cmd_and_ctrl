package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Argentum Masticore — Artifact Creature — Phyrexian Masticore {5}, 5/5:
//
//	"First strike, protection from multicolored
//	 At the beginning of your upkeep, sacrifice this creature unless you
//	 discard a card. When you discard a card this way, destroy target
//	 nonland permanent an opponent controls with mana value less than or
//	 equal to the mana value of the discarded card."
//
// The "unless" is a discard of up to one card: choosing none (or having an
// empty hand, CR 701.9a) sacrifices the Masticore, but only if it is still
// the same object that triggered (CR 400.7). A card really discarded this
// way starts the reflexive trigger (CR 603.12), which chooses its target as
// it goes on the stack and reads the mana value the discarded card had in
// the hand (a land is 0, so the ceiling is 0; X in a cost is 0 there too,
// CR 202.3e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3ceee5c2-321c-49ec-ba49-1fac7f6ee2cc",
		Name:            "Argentum Masticore",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "protection from multicolored"},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Argentum Masticore — sacrifice this creature unless you discard a card; when you do, destroy a nonland permanent an opponent controls",
				argentumMasticoreUpkeep),
		},
	})
}

// argentumMasticoreUpkeep asks for the discard, or sacrifices at once when
// there is no card to discard.
func argentumMasticoreUpkeep(g *game.Game, item *game.StackItem) error {
	if !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Eliminated {
		return nil
	}
	if p.Hand == nil || p.Hand.Size() == 0 {
		return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
	}
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   item.Controller,
		Source:   item.SourceCardID,
		N:        1,
		UpTo:     true,
		Question: "Argentum Masticore — discard a card, or sacrifice it?",
		Asked:    game.NewAskedItem(item),
	}, argentumMasticoreDiscarded, nil)
}

// argentumMasticoreDiscarded is the discard answered. Keyed (#4253); an
// on-disk key: never renamed, never reused.
var argentumMasticoreDiscarded = game.RegisterDiscardThen("choose-cards/argentum-masticore-discard", func(g *game.Game, r game.DiscardsDone) error {
	if r.Item == nil {
		return nil
	}
	ctx := NewContext(g, r.Item)
	discarded := r.Discarded.Cards()
	if len(discarded) == 0 {
		if !sourceIsStillThisPermanent(g, r.Item) {
			return nil
		}
		return SacrificePermanent{Target: r.Item.SourceCardID}.Apply(ctx)
	}
	card, ok := g.LookupCardForEffect(discarded[0])
	if !ok {
		return nil
	}
	return ReflexiveTrigger{
		Label:  "Argentum Masticore — destroy target nonland permanent an opponent controls with mana value no greater than the discarded card's",
		Cards:  discarded,
		Body:   argentumMasticoreDestroyBody,
		Params: game.EffectParams{Amount: card.ManaValue()},
	}.Apply(ctx)
})

// argentumMasticoreDestroy is the reflexive half.
func argentumMasticoreDestroy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if len(item.Targets) == 0 {
		return nil
	}
	t := item.Targets[0]
	if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
		return nil
	}
	return DestroyTarget{Target: t.ID}.Apply(ctx)
}
