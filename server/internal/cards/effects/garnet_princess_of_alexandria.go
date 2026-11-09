package effects

import (
	"fmt"
	"os"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Garnet, Princess of Alexandria — Legendary Creature — Human Noble Cleric
// {G}{W}, 2/2 (Final Fantasy, #4462):
//
//	"Lifelink
//	 Whenever Garnet attacks, you may remove a lore counter from each of any
//	 number of Sagas you control. Put a +1/+1 counter on Garnet for each lore
//	 counter removed this way."
//
// The attack trigger offers the Sagas the controller controls that carry a
// lore counter, any number of them, zero included (the "may"). Each one
// chosen loses one lore counter as the pick is answered (removing counters
// triggers no earlier chapter, per the rulings), and Garnet gets a +1/+1
// counter for each one really removed, if she is still on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8c709873-84b5-46c2-bb29-85b041a6df88",
		Name:            "Garnet, Princess of Alexandria",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Garnet, Princess of Alexandria — you may remove a lore counter from each of any number of Sagas you control; put a +1/+1 counter on Garnet for each removed",
				garnetAttacks),
		},
	})
}

func garnetAttacks(g *game.Game, item *game.StackItem) error {
	var sagas []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.HasSubtype("Saga") && c.Counters[game.CounterLore] > 0 {
			sagas = append(sagas, c.InstanceID)
		}
	}
	if len(sagas) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Garnet, Princess of Alexandria — remove a lore counter from each of any number of Sagas you control",
		Cards:    sagas,
		Min:      0,
		Max:      0, // "any number"
		ThenKey:  garnetRemovePick,
		LiveZone: game.ZoneBattlefield,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// garnetRemovePick removes one lore counter from each Saga chosen and puts a
// +1/+1 counter on Garnet for each one that really lost it. An on-disk key:
// never renamed, never reused.
var garnetRemovePick = askedPick("choose-cards/garnet-princess-of-alexandria-remove", func(ctx *Context, r game.CardsChosen) error {
	removed := 0
	for _, id := range r.Picked {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !onBattlefield(ctx.Game, id) || c.Counters[game.CounterLore] <= 0 {
			continue
		}
		before := c.Counters[game.CounterLore]
		fmt.Fprintf(os.Stderr, "GARNETDEBUG before=%d id=%s\n", before, id)
		if err := ctx.Game.AddCounterForEffect(id, game.CounterLore, -1); err != nil {
			return err
		}
		if after, ok := ctx.Game.LookupCardForEffect(id); ok {
			fmt.Fprintf(os.Stderr, "GARNETDEBUG after=%d\n", after.Counters[game.CounterLore])
		}
		if after, ok := ctx.Game.LookupCardForEffect(id); ok && after.Counters[game.CounterLore] < before {
			removed++
		}
	}
	if removed == 0 || !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return AddCounter{Target: ctx.Item.SourceCardID, Kind: game.CounterPlusOne, N: removed}.Apply(ctx)
})
