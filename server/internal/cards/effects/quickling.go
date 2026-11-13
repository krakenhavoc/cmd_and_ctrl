package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Quickling — Creature — Faerie Rogue {1}{U}, 2/2 (Wilds of Eldraine
// Commander, #5368):
//
//	"Flash
//	 Flying
//	 When this creature enters, sacrifice it unless you return another
//	 creature you control to its owner's hand."
//
// Asked as the trigger resolves. With another creature under your control
// you may name one to return (a choose_cards prompt answered with none, or
// with a creature that has since left, sacrifices the Quickling); with
// none the Quickling is sacrificed at once. The return is Mistbreath
// Elder's, and only if the creature really reached a hand does the
// Quickling stay. A Quickling that has left or come back as a new object
// is not sacrificed (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "109899e1-abd7-4582-8b67-b0852755bf75",
		Name:            "Quickling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Quickling — sacrifice it unless you return another creature you control to its owner's hand", cmdr2B11QuicklingEnters),
		},
	})
}

func cmdr2B11QuicklingEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var others []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsCreature() && c.InstanceID != item.SourceCardID {
			others = append(others, c.InstanceID)
		}
	}
	return sacrificeUnlessYouReturnOne(ctx, item, others,
		"Quickling — return another creature you control to its owner's hand (or sacrifice Quickling)",
		cmdr2B11QuicklingPicked)
}

// sacrificeUnlessYouReturnOne is "sacrifice it unless you return <one of
// candidates> to its owner's hand" as the enters trigger resolves
// (Quickling, Drake Familiar): with no candidate the source is sacrificed at
// once; otherwise a choose_cards prompt names one or none, and `picked`
// (built on returnPickedOrSacrifice) does the rest.
func sacrificeUnlessYouReturnOne(ctx *Context, item *game.StackItem, candidates []uuid.UUID, question string, picked game.ChooseCardsThen) error {
	if len(candidates) == 0 {
		return SacrificePermanent{Target: item.SourceCardID}.Apply(ctx)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: ctx.Controller(),
		Source:     item.SourceCardID,
		Question:   question,
		Cards:      candidates,
		Min:        0,
		Max:        1,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    picked,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// returnPickedOrSacrifice returns the permanent named, with `bounced` (built
// on sacrificeSourceUnlessBounced) as the return's tail, or sacrifices the
// source when none was named or the one named is gone.
func returnPickedOrSacrifice(ctx *Context, r game.CardsChosen, bounced game.TailRef) error {
	for _, id := range r.Picked {
		if onBattlefield(ctx.Game, id) {
			return BounceToHand{Target: id, ThenTail: ctx.ItemTail(bounced)}.Apply(ctx)
		}
	}
	return SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
}

// sacrificeSourceUnlessBounced sacrifices the source when the returned
// permanent did not reach a hand.
func sacrificeSourceUnlessBounced(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) > 0 {
		return nil
	}
	return SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
}

// cmdr2B11QuicklingPicked returns the chosen creature, or sacrifices the
// Quickling when none was named or the one named is gone. An on-disk key.
var cmdr2B11QuicklingPicked = askedPick("choose-cards/cmdr2-b11-quickling-return", func(ctx *Context, r game.CardsChosen) error {
	return returnPickedOrSacrifice(ctx, r, cmdr2B11QuicklingBounced)
})

// cmdr2B11QuicklingBounced sacrifices the Quickling when the creature did
// not reach a hand. An on-disk key.
var cmdr2B11QuicklingBounced = CardTail("cmdr2-b11/quickling-after-bounce", sacrificeSourceUnlessBounced)
