package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b24_helpers.go — helpers for the Commander deck cards, batch 24
// (#5368). Every name carries the cmdrB24 prefix.

// cmdrB24ZombieDealtCombatDamageToAPlayerThisTurn is Lost Monarch of Ifnir's
// condition: "a player was dealt combat damage by a Zombie this turn", read
// off this turn's damage events. The source is the card as it sits now, so a
// Zombie that died after dealing the damage counts by its printed subtypes.
func cmdrB24ZombieDealtCombatDamageToAPlayerThisTurn(g *game.Game, source *game.Card) bool {
	zombie := source.TextSubtype("Zombie")
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || g.PlayerByIDForEffect(ev.Target) == nil {
			continue
		}
		if src, ok := g.LookupCardForEffect(ev.Source); ok && src.HasSubtype(zombie) {
			return true
		}
	}
	return false
}

// cmdrB24MillThenMayReturnCreature is "mill three cards, then you may return
// a creature card from your graveyard to your hand": the pick is asked after
// the mill, over the creature cards in the graveyard then, with a floor of
// zero.
func cmdrB24MillThenMayReturnCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (MillCards{Player: item.Controller, N: 3}).Apply(ctx); err != nil {
		return err
	}
	controller := item.Controller
	candidates := graveyardCardIDs(ctx, controller, game.Card.IsCreature)
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    controller,
		FromPlayer: controller,
		Source:     item.SourceCardID,
		Question:   "Lost Monarch of Ifnir — you may return a creature card from your graveyard to your hand",
		Cards:      candidates,
		Min:        0,
		Max:        1,
		ThenKey:    cmdrB24ReturnCreatureToHand,
		LiveZone:   game.ZoneGraveyard,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// cmdrB24ReturnCreatureToHand is the chosen creature card going to hand. An
// on-disk key: never renamed, never reused.
var cmdrB24ReturnCreatureToHand = askedPick("choose-cards/lost-monarch-of-ifnir-return-creature", func(ctx *Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
