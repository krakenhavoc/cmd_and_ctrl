package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Halo-Charged Skaab — Creature — Zombie {4}{U}, 4/4 (March of the
// Machine):
//
//	"When this creature enters, each player mills two cards. Then you may
//	 put an instant, sorcery, or battle card from your graveyard on top of
//	 your library."
//
// The mills run one library at a time in APNAP order (millEachPlayerInTurn);
// the choice follows once every mill has landed, so a card the mill just put
// into your graveyard is a legal pick. The pick is optional and chosen, not
// targeted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7dcd1a3-3163-4e26-94e8-f4072e477e6a",
		Name:         "Halo-Charged Skaab",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Halo-Charged Skaab — each player mills two cards, then you may put an instant, sorcery, or battle card from your graveyard on top of your library",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillEachPlayerInTurn(ctx, effects.ApnapPlayers(g), effects.ConstantMill(2), ctx.ItemTail(s82MomB12SkaabAfterMill))
				}),
		},
	})
}

// s82MomB12SkaabAfterMill is "then you may put an instant, sorcery, or
// battle card from your graveyard on top of your library", run once every
// mill is done. An on-disk key: never renamed, never reused.
var s82MomB12SkaabAfterMill = effects.CardTail("tail/s82-mom-b12-skaab-after-mill", func(ctx *effects.Context, _ game.TailOutcome) error {
	cands := effects.GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
		return c.IsInstant() || c.IsSorcery() || c.IsBattle()
	})
	if len(cands) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Halo-Charged Skaab — you may put an instant, sorcery, or battle card from your graveyard on top of your library",
		Cards:    cands,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  s82MomB12SkaabTop,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

// s82MomB12SkaabTop is the pick answered: the card, if one was chosen, goes
// on top of its owner's library. An on-disk key.
var s82MomB12SkaabTop = effects.ItemChoice("choose-cards/s82-mom-b12-skaab-top", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ctx.Game.TuckToLibraryForEffect(r.Picked[0], false)
})
