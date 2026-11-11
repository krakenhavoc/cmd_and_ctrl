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
	if len(others) == 0 {
		return SacrificePermanent{Target: item.SourceCardID}.Apply(ctx)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: ctx.Controller(),
		Source:     item.SourceCardID,
		Question:   "Quickling — return another creature you control to its owner's hand (or sacrifice Quickling)",
		Cards:      others,
		Min:        0,
		Max:        1,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    cmdr2B11QuicklingPicked,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// cmdr2B11QuicklingPicked returns the chosen creature, or sacrifices the
// Quickling when none was named or the one named is gone. An on-disk key.
var cmdr2B11QuicklingPicked = askedPick("choose-cards/cmdr2-b11-quickling-return", func(ctx *Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if onBattlefield(ctx.Game, id) {
			return BounceToHand{Target: id, ThenTail: ctx.ItemTail(cmdr2B11QuicklingBounced)}.Apply(ctx)
		}
	}
	return SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
})

// cmdr2B11QuicklingBounced sacrifices the Quickling when the creature did
// not reach a hand. An on-disk key.
var cmdr2B11QuicklingBounced = CardTail("cmdr2-b11/quickling-after-bounce", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) > 0 {
		return nil
	}
	return SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
})
