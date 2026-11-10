package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hakbal of the Surging Soul — Legendary Creature — Merfolk Scout
// {2}{G}{U}, 3/3 (The Lost Caverns of Ixalan Commander, #5425):
//
//	"At the beginning of combat on your turn, each Merfolk creature you
//	 control explores. (Reveal the top card of your library. Put that card
//	 into your hand if it's a land. Otherwise, put a +1/+1 counter on the
//	 exploring creature, then put the card back or put it into your
//	 graveyard.)
//	 Whenever Hakbal attacks, you may put a land card from your hand onto
//	 the battlefield. If you don't, draw a card."
//
// The Merfolk are read as the first ability resolves, through the layered
// view, and they explore one at a time in an order you choose as you go
// (CR 701.44d, game.ExploreEachForEffect): each explore finishes, its
// graveyard question answered, before you choose the next, so you know
// whether a nonland card is still on top when you pick who gets it.
//
// The attack trigger is PutFromHandOntoBattlefield, optional, and the
// shared "If you don't, draw a card" tail: a declined offer, a hand with
// no land and an entry a replacement cancelled all draw. Putting a land
// onto the battlefield is not playing one (CR 305.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "687e772a-cdfe-41fd-a066-6cb3aee04193",
		Name:         "Hakbal of the Surging Soul",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Hakbal of the Surging Soul — each Merfolk creature you control explores",
				hakbalEachMerfolkExplores),
			WheneverThisAttacks("Hakbal of the Surging Soul — you may put a land card from your hand onto the battlefield. If you don't, draw a card",
				hakbalLandOrDraw),
		},
	})
}

// hakbalEachMerfolkExplores is the combat trigger.
func hakbalEachMerfolkExplores(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var merfolk []game.ObjectRef
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == ctx.Controller() && c.IsCreature() && c.HasSubtype("Merfolk") {
			merfolk = append(merfolk, game.ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch})
		}
	}
	return g.ExploreEachForEffect(ctx.Source(), merfolk, ctx.Controller())
}

// hakbalLandOrDraw is the attack trigger: the optional put, then the draw
// if no card arrived.
func hakbalLandOrDraw(g *game.Game, item *game.StackItem) error {
	return PutFromHandOntoBattlefield{
		Match:    Land(),
		Optional: true,
		Label:    "Hakbal of the Surging Soul — you may put a land card from your hand onto the battlefield",
		ThenTail: mshA6OriginDrawIfNone,
	}.Apply(NewContext(g, item))
}
