package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b574_helpers.go — shared shapes of S82 OLD batch 574 (Ravnica:
// City of Guilds, #5600): the two library looks of Lurking Informant and
// Thoughtpicker Witch.

// S82OldB574LurkingInformantLook is Lurking Informant's ability: "Look at the
// top card of target player's library. You may put that card into that
// player's graveyard." Only the controller looks (CR 701.20); the pick is
// optional, and a card that is not put into the graveyard stays on top. The
// answer is a registered continuation, so a table waiting on it is a restore
// point.
func S82OldB574LurkingInformantLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	owner, ok := FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	looked := g.LookAtTopOfPlayersLibraryForEffect(ctx.Controller(), owner, 1)
	if len(looked) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: owner,
		Source:     ctx.Source(),
		Question:   "Lurking Informant — you may put that card into its owner's graveyard",
		Cards:      looked,
		Min:        0,
		Max:        1,
		LiveZone:   game.ZoneLibrary,
		ThenKey:    s82OldB574InformantPicked,
		Carry:      looked,
	})
	return nil
}

// s82OldB574InformantPicked is the card chosen to go to the graveyard. An
// on-disk key: never renamed, never reused.
var s82OldB574InformantPicked = game.RegisterChooseCardsThen("choose-cards/s82-old-b574-lurking-informant-graveyard", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return g.PutCardsIntoGraveyardKeyedForEffect(CardsStillInALibrary(g, r.Picked), game.KeyedTail{})
})

// S82OldB574ThoughtpickerLook is Thoughtpicker Witch's ability: "Look at the
// top two cards of target opponent's library, then exile one of them." The
// controller looks and picks exactly one (a library of one card gives no
// choice); the other stays where it was, on top.
func S82OldB574ThoughtpickerLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	owner, ok := FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	looked := g.LookAtTopOfPlayersLibraryForEffect(ctx.Controller(), owner, 2)
	switch len(looked) {
	case 0:
		return nil
	case 1:
		return g.ExileCardsKeyedForEffect(looked, game.KeyedTail{})
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: owner,
		Source:     ctx.Source(),
		Question:   "Thoughtpicker Witch — exile one of them",
		Cards:      looked,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneLibrary,
		ThenKey:    s82OldB574ThoughtpickerPicked,
		Carry:      looked,
	})
	return nil
}

// s82OldB574ThoughtpickerPicked is the card chosen to be exiled. An on-disk
// key: never renamed, never reused.
var s82OldB574ThoughtpickerPicked = game.RegisterChooseCardsThen("choose-cards/s82-old-b574-thoughtpicker-witch-exile", func(g *game.Game, r game.CardsChosen) error {
	var ids []uuid.UUID
	ids = append(ids, CardsStillInALibrary(g, r.Picked)...)
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(ids, game.KeyedTail{})
})
