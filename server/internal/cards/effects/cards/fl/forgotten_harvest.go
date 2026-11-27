package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forgotten Harvest — Enchantment {1}{G} (Prophecy, S82 OLD batch 675, #5617):
//
//	"At the beginning of your upkeep, you may exile a land card from your
//	 graveyard. If you do, put a +1/+1 counter on target creature."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d), so a
// trigger with no creature to target is removed. As it resolves, a controller
// with a land card in their graveyard is offered one to exile (a choice of none
// or one, Ashiok's keyed pick); declining, or having none, ends the ability.
// Only when the exile happened is the +1/+1 counter put on the target, which is
// checked again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e5de2ed-e8a1-4d52-832f-c4a1bfdf2ae5",
		Name:         "Forgotten Harvest",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtYourUpkeep("Forgotten Harvest — you may exile a land card from your graveyard; if you do, put a +1/+1 counter on target creature",
					forgottenHarvestOffer),
				effects.TargetCreature("target creature")),
		},
	})
}

// forgottenHarvestOffer asks the controller for a land card in their graveyard,
// none allowed.
func forgottenHarvestOffer(g *game.Game, item *game.StackItem) error {
	lands := effects.GraveyardIDs(g, item.Controller, game.Card.IsLand)
	if len(lands) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Forgotten Harvest — you may exile a land card from your graveyard; if you do, put a +1/+1 counter on the target creature",
		Cards:    lands,
		Min:      0,
		Max:      1,
		ThenKey:  forgottenHarvestPick,
		LiveZone: game.ZoneGraveyard,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// forgottenHarvestPick exiles the chosen land card and, once it is in exile,
// puts the counter on the trigger's target. An on-disk key: never renamed,
// never reused.
var forgottenHarvestPick = effects.AskedPick("choose-cards/s82-old-b675-forgotten-harvest-exile-land", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	if err := (effects.ExileTarget{Target: r.Picked[0]}).Apply(ctx); err != nil {
		return err
	}
	if z := ctx.Game.FindCardZoneForEffect(r.Picked[0]); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard && t.ID != uuid.Nil && effects.OnBattlefield(ctx.Game, t.ID) {
			return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
		}
	}
	return nil
})
