package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pack Guardian — Creature — Wolf Spirit {2}{G}{G}, 4/3 (Innistrad
// Remastered, S82 GAP, #5948):
//
//	"Flash
//	 When this creature enters, you may discard a land card. If you do,
//	 create a 2/2 green Wolf creature token."
//
// The "you may" is the pick itself (Ainok Wayfarer's prompt shape): the
// controller chooses a land card from their hand or none, asked as the
// trigger resolves (CR 603.5), and only a card really discarded makes the
// Wolf. With no land card in hand nothing is asked and nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "61ad6bd2-c079-44e7-a1d9-8089f086ec69",
		Name:            "Pack Guardian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Pack Guardian — you may discard a land card; if you do, create a 2/2 green Wolf creature token",
				packGuardianAsk),
		},
	})
}

func packGuardianAsk(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player := ctx.Controller()
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Hand == nil {
		return nil
	}
	var lands []uuid.UUID
	for _, c := range p.Hand.Cards {
		if c.IsLand() {
			lands = append(lands, c.InstanceID)
		}
	}
	if len(lands) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Pack Guardian — you may discard a land card; if you do, create a 2/2 green Wolf creature token",
		Cards:    lands,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneHand,
		ThenKey:  packGuardianPick,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// packGuardianPick is the land discarded, then the Wolf. A card that left
// the hand in the meantime is not discarded and makes no Wolf. An on-disk
// key: never renamed, never reused.
var packGuardianPick = effects.AskedPick("choose-cards/pack-guardian-discard-land", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	if z := ctx.Game.FindCardZoneForEffect(r.Picked[0]); z == nil || z.Kind != game.ZoneHand {
		return nil
	}
	if err := ctx.Game.DiscardCardsKeyedForEffect(ctx.Controller(), r.Picked[:1], ctx.Source(), game.KeyedTail{}); err != nil {
		return err
	}
	return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("2/2 green Wolf"), N: 1}.Apply(ctx)
})
