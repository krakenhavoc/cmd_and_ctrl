package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Orphans of the Wheat — Creature — Human {1}{W}, 1/1 (Duskmourn):
//
//	"Whenever this creature attacks, tap any number of untapped creatures
//	 you control. This creature gets +1/+1 until end of turn for each
//	 creature tapped this way."
//
// The pick is Teferi Akosa's "tap any number" over the controller's
// untapped creatures, floor zero. Each pick is re-read as it is tapped (the
// board can move under the prompt): only a creature that really became
// tapped counts. The boost goes on the Orphans if it is still the permanent
// that attacked (CR 400.7), and nothing happens for zero.
func init() {
	Register(Spec{
		OracleID:     "f758bad0-8f36-4985-ad1d-66b695f3bbb6",
		Name:         "Orphans of the Wheat",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Orphans of the Wheat — tap any number of untapped creatures you control; +1/+1 for each",
				orphansOfTheWheatAsk),
		},
	})
}

// orphansOfTheWheatAsk offers the controller's untapped creatures.
func orphansOfTheWheatAsk(g *game.Game, item *game.StackItem) error {
	candidates := untappedCreaturesControlledBy(g, item.Controller)
	if len(candidates) == 0 || !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Orphans of the Wheat — tap any number of untapped creatures you control",
		Cards:    candidates,
		Min:      0,
		// Max 0 means "all of them", which is what "any number" is.
		Max:      0,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  orphansOfTheWheatTap,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// orphansOfTheWheatTap taps the picks and pumps the Orphans by the number
// that really became tapped. An on-disk key: never renamed, never reused.
var orphansOfTheWheatTap = askedPick("choose-cards/orphans-of-the-wheat-tap", func(ctx *Context, r game.CardsChosen) error {
	n := 0
	for _, id := range r.Picked {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.Tapped || !c.IsCreature() || c.Controller != ctx.Controller() {
			continue
		}
		if err := (TapTarget{Target: id}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
		n++
	}
	if n == 0 || !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return BoostUntilEOT{Target: ctx.Source(), Power: n, Toughness: n,
		Label: "Orphans of the Wheat — +1/+1 until end of turn for each creature tapped this way"}.Apply(ctx)
})
