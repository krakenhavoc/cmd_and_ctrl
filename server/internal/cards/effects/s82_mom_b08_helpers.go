package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b08_helpers.go — keyed continuations for March of the Machine,
// batch 8 (S82, #5477): Moment of Truth's graveyard pick and Seed of Hope's
// permanent pick. Keys are on-disk identities: never renamed, never reused.

// momentOfTruthPickGraveyard is the middle of Moment of Truth: "one of those
// cards into your graveyard". The card not chosen goes on the bottom.
var momentOfTruthPickGraveyard = game.RegisterChooseCardsThen("choose-cards/moment-of-truth-graveyard", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	if err := restIntoGraveyard(g, r.Picked); err != nil {
		return err
	}
	return momentOfTruthBottom(g, r.Chooser, r.Source, subtractIDs(r.Candidates, r.Picked))
})

// momentOfTruthBottom puts the last of the looked-at cards on the bottom of
// the library. A pile of one has no order to choose.
func momentOfTruthBottom(g *game.Game, player, source uuid.UUID, rest []uuid.UUID) error {
	return TakeRestOnBottomInAnyOrder(g, TakeFromLibraryResult{Player: player, Source: source, Rest: rest})
}

func subtractIDs(from, remove []uuid.UUID) []uuid.UUID {
	drop := make(map[uuid.UUID]bool, len(remove))
	for _, id := range remove {
		drop[id] = true
	}
	var out []uuid.UUID
	for _, id := range from {
		if !drop[id] {
			out = append(out, id)
		}
	}
	return out
}

// momentOfTruthAfterTake runs once the card for the hand is chosen: of the
// rest, one goes into the graveyard (the player's choice) and the other to
// the bottom of the library.
func momentOfTruthAfterTake(g *game.Game, res TakeFromLibraryResult) error {
	rest := res.Rest
	switch len(rest) {
	case 0:
		return nil
	case 1:
		// Only two cards were there to look at: the second sentence's card
		// is the one that goes into the graveyard.
		return restIntoGraveyard(g, rest)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  res.Player,
		Source:   res.Source,
		Question: "Moment of Truth — put one of those cards into your graveyard (the other goes on the bottom of your library)",
		Cards:    rest,
		Min:      1,
		Max:      1,
		ThenKey:  momentOfTruthPickGraveyard,
		LiveZone: game.ZoneLibrary,
	})
	return nil
}

var momentOfTruthTake = RegisterTake("choose-cards/moment-of-truth-take", TakeDef{Then: momentOfTruthAfterTake})

// seedOfHopePick is Seed of Hope's "you may put a permanent card from among
// the milled cards into your hand"; the life is gained after the choice.
var seedOfHopePick = askedPick("choose-cards/seed-of-hope-permanent", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) > 0 {
		if err := (ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return GainLife{Amount: 2}.Apply(ctx)
})

var seedOfHopeMilled = CardTail("seed-of-hope/after-mill", func(ctx *Context, o game.TailOutcome) error {
	player := ctx.Controller()
	var perms []uuid.UUID
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.IsToken() || !c.IsPermanent() {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		perms = append(perms, id)
	}
	if len(perms) == 0 {
		return GainLife{Amount: 2}.Apply(ctx)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Seed of Hope — you may put a permanent card from among the milled cards into your hand",
		Cards:    perms,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  seedOfHopePick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})
