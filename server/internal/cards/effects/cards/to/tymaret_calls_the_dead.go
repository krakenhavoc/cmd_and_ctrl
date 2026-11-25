package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tymaret Calls the Dead — Enchantment — Saga {2}{B} (Theros Beyond Death, #5514):
//
//	"I, II — Mill three cards. Then you may exile a creature or enchantment card
//	 from your graveyard. If you do, create a 2/2 black Zombie creature token.
//	 III — You gain X life and scry X, where X is the number of Zombies you
//	 control."
//
// Chapters I and II are one body: mill three (CR 701.17a), then one keyed
// choose_cards over the controller's graveyard as it stands AFTER the mill
// ("then you may exile"), at most one card; the Zombie is made from the exile's
// tail only if a card was exiled (Masked Vandal's shape). Chapter III counts the
// Zombies as it resolves, through the layered subtype (a changeling counts), and
// gains that much life, then scries that many.
//
// No simplification.
func init() {
	const name = "Tymaret Calls the Dead"
	effects.Register(effects.Spec{
		OracleID:     "1d09ebb7-c4c9-4bb7-b6bf-3c8330b954e8",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "mill three cards, then you may exile a creature or enchantment card from your graveyard; if you do, create a 2/2 black Zombie"), tymaretMillAndExile),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "mill three cards, then you may exile a creature or enchantment card from your graveyard; if you do, create a 2/2 black Zombie"), tymaretMillAndExile),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "you gain X life and scry X, where X is the number of Zombies you control"), tymaretCount),
		},
	})
}

func tymaretMillAndExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.MillCards{N: 3}).Apply(ctx); err != nil {
		return err
	}
	ids := effects.B42CardsInGraveyardOf(g, item.Controller, func(c game.Card) bool {
		return c.IsCreature() || c.HasCardType("enchantment")
	})
	if len(ids) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Tymaret Calls the Dead — you may exile a creature or enchantment card from your graveyard (if you do, create a 2/2 black Zombie)",
		Cards:    ids,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  tymaretExile,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

var tymaretExile = effects.AskedPick("choose-cards/tymaret-calls-the-dead-exile", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(r.Picked, ctx.ItemTail(tymaretZombie))
})

var tymaretZombie = effects.CardTail("exile-then/tymaret-calls-the-dead-zombie", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	return effects.CreateToken{Controller: ctx.Controller(), Template: effects.BlackZombieToken(), N: 1}.Apply(ctx)
})

func tymaretCount(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	zombie := ctx.TextSubtype("Zombie")
	x := effects.CountControlled(g, item.Controller, func(c game.Card) bool { return c.HasSubtype(zombie) })
	if x <= 0 {
		return nil
	}
	if err := (effects.GainLife{Player: item.Controller, Amount: x}).Apply(ctx); err != nil {
		return err
	}
	return effects.Scry{N: x}.Apply(ctx)
}
