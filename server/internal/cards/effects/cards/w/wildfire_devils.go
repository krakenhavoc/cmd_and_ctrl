package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wildfire Devils — Creature — Devil {3}{R}, 4/2 (Commander Starter Decks):
//
//	"When this creature enters and at the beginning of your upkeep, choose a
//	 player at random. That player exiles an instant or sorcery card from
//	 their graveyard. Copy that card. You may cast the copy without paying its
//	 mana cost."
//
// One ability on two events (Old Rutstein's ThisEntersOrYourUpkeep). The
// player is drawn from the engine's keyed random stream among every player still in
// the game, the Devils' controller included. That player, not the Devils'
// controller, picks the card (nothing is asked when only one qualifies, and
// nothing happens when none does), and it is exiled through the CR 614 window;
// only a card that really reached exile is copied. The Devils' controller then
// casts the copy for free, or not, while the ability is still resolving
// (CopyAndCast, as Roving Actuator does).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b446b485-c10b-4175-bb49-16ac2f9e7718",
		Name:         "Wildfire Devils",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventBeginUpkeep}, effects.ThisEntersOrYourUpkeep,
				"Wildfire Devils — a random player exiles an instant or sorcery card from their graveyard; copy it and you may cast the copy",
				wildfireDevilsResolve),
		},
	})
}

func wildfireDevilsResolve(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	chosen := g.ChooseAtRandomForEffect(effects.RandomDraw(ctx), effects.ApnapLivingSeats(g), 1)
	if len(chosen) == 0 {
		return nil
	}
	player := g.PlayerByIDForEffect(chosen[0])
	if player == nil {
		return nil
	}
	var cards []uuid.UUID
	for _, c := range player.Graveyard.Cards {
		if c.IsInstant() || c.IsSorcery() {
			cards = append(cards, c.InstanceID)
		}
	}
	switch len(cards) {
	case 0:
		return nil
	case 1:
		return g.ExileCardsKeyedForEffect(cards, ctx.ItemTail(wildfireDevilsExiled))
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    player.ID,
		FromPlayer: player.ID,
		Source:     ctx.Source(),
		Question:   "Wildfire Devils — exile an instant or sorcery card from your graveyard",
		Cards:      cards,
		Min:        1,
		Max:        1,
		Zone:       game.ZoneGraveyard,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    wildfireDevilsChosen,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
}

// wildfireDevilsChosen exiles the card its owner picked. An on-disk key:
// never renamed, never reused.
var wildfireDevilsChosen = effects.AskedPick("choose-cards/wildfire-devils-exile", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(r.Picked[:1], ctx.ItemTail(wildfireDevilsExiled))
})

// wildfireDevilsExiled copies what reached exile and offers the copy to the
// Devils' controller for free. An on-disk key.
var wildfireDevilsExiled = effects.CardTail("wildfire-devils/copy-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.CopyAndCast{Cards: o.IDs, Free: true}.Apply(ctx)
})
