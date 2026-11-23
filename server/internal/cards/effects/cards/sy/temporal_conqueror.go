package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Temporal Conqueror — {2}{G} Creature — Na'kuhl Warrior 2/3 (Star Trek,
// #4973):
//
//	"When this creature enters, mill three cards. Then return a land card
//	 from your graveyard to your hand. If you can't, create a Treasure
//	 token."
//
// The land is chosen after the mill, as the ability resolves (CR 608.2c),
// so a land the trigger itself milled is a legal pick, and so is one that
// was already there. "Return" is not optional: with a land card in the
// graveyard one goes back. With none, you can't, and you create a
// Treasure instead.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "306002e6-d68d-47ad-879b-e6060efebeff",
		Name:         "Temporal Conqueror",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Temporal Conqueror — mill three, then return a land card to your hand or create a Treasure",
				temporalConquerorResolve),
		},
	})
}

func temporalConquerorResolve(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.MillCards{Player: item.Controller, N: 3}).Apply(ctx); err != nil {
		return err
	}
	lands := effects.GraveyardCardIDs(ctx, item.Controller, game.Card.IsLand)
	if len(lands) == 0 {
		return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(ctx)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "Temporal Conqueror — return a land card from your graveyard to your hand",
		Cards:      lands,
		Min:        1,
		Max:        1,
		ThenKey:    temporalConquerorReturned,
		LiveZone:   game.ZoneGraveyard,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

var temporalConquerorReturned = effects.AskedPick("choose-cards/temporal-conqueror-return-land", func(ctx *effects.Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
