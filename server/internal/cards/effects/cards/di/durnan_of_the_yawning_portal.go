package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Durnan of the Yawning Portal — {3}{G} Legendary Creature — Human Warrior 3/3
//
//	"Whenever Durnan attacks, look at the top four cards of your library. You
//	 may exile a creature card from among them. Put the rest on the bottom of
//	 your library in any order. For as long as that card remains exiled, you
//	 may cast it. That spell has undaunted.
//	 Choose a Background"
//
// A look, so only you see the cards. The exile is optional; the rest go to the
// bottom in an order you choose. The cast is a standing permission over that
// one card for as long as it stays in exile (a creature card, so it is cast,
// never played), at any time you could cast it, and carries undaunted: it costs
// {1} less for each opponent as it is cast (CR 702.125a; the engine reads the
// opponents then). "Choose a Background" is the deck-construction rule
// (CR 702.124k) read from the oracle text.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21d3858f-441d-4bb2-888e-8b147ecad7e9",
		Name:         "Durnan of the Yawning Portal",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Durnan of the Yawning Portal — look at the top four cards of your library; you may exile a creature card from among them", durnanLook),
		},
	})
}

func durnanLook(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player := ctx.Controller()
	looked := g.LookAtTopOfLibraryForEffect(player, 4)
	if len(looked) == 0 {
		return nil
	}
	var creatures []uuid.UUID
	for _, id := range looked {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsCreature() {
			creatures = append(creatures, id)
		}
	}
	if len(creatures) == 0 {
		return durnanBottom(ctx, looked)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Durnan of the Yawning Portal — you may exile a creature card from among them (the rest go on the bottom in any order); you may cast it for as long as it remains exiled",
		Cards:    creatures,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneLibrary,
		ThenKey:  durnanPicked,
		Carry:    looked,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// durnanBottom puts the cards on the bottom of the library in the order the
// controller chooses.
func durnanBottom(ctx *effects.Context, rest []uuid.UUID) error {
	return ctx.Game.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   ctx.Controller(),
		Source:    ctx.Source(),
		Cards:     rest,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceBottom,
		Reason:    "Durnan of the Yawning Portal — put the rest on the bottom of your library in any order",
	})
}

// durnanPicked exiles the pick and puts the rest on the bottom. An on-disk key:
// never renamed, never reused.
var durnanPicked = effects.AskedPick("choose-cards/gap-b216-durnan-exile", func(ctx *effects.Context, r game.CardsChosen) error {
	var picked uuid.UUID
	if len(r.Picked) > 0 {
		picked = r.Picked[0]
	}
	var rest []uuid.UUID
	for _, id := range r.Carry {
		if id != picked {
			rest = append(rest, id)
		}
	}
	if len(rest) > 0 {
		if err := durnanBottom(ctx, rest); err != nil {
			return err
		}
	}
	if picked == uuid.Nil {
		return nil
	}
	return effects.ExileTarget{Target: picked, ThenTail: ctx.ItemTail(durnanExiled)}.Apply(ctx)
})

// durnanExiled grants the cast over the exiled card. An on-disk key.
var durnanExiled = effects.CardTail("tail/gap-b216-durnan-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	ctx.Game.GrantCastPermissionOverCardForEffect(o.IDs[0], game.CastPermission{
		Player:     ctx.Controller(),
		Zone:       game.ZoneExile,
		Duration:   game.WhileInZoneDuration(),
		CastOnly:   true,
		Undaunted:  true,
		Source:     ctx.Source(),
		SourceName: "Durnan of the Yawning Portal",
		Label:      "Durnan of the Yawning Portal — you may cast it for as long as it remains exiled; it has undaunted",
	})
	return nil
})
