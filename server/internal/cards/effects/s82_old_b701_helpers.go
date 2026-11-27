package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b701_helpers.go — shared shapes of S82 OLD batch 701 (Starter 1999,
// #5620): Piracy's permission and Eye Spy's look.

// S82OldB701Piracy is "Until end of turn, you may tap lands you don't control
// for mana. Spend this mana only to cast spells." The permission is the
// controller's, read live by the engine (#6204, game/tap_foreign_lands.go).
func S82OldB701Piracy(g *game.Game, item *game.StackItem) error {
	g.AllowTapForeignLandsThisTurnForEffect(item.Controller)
	return nil
}

// S82OldB701EyeSpy is "Look at the top card of target player's library. You
// may put that card into their graveyard." Only the controller looks (CR
// 701.20); the card goes to its owner's graveyard if they say so, and
// otherwise stays on top. The answer is a registered continuation, so a table
// waiting on it is a restore point.
func S82OldB701EyeSpy(g *game.Game, item *game.StackItem) error {
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
		Question:   "Eye Spy — you may put that card into its owner's graveyard",
		Cards:      looked,
		Min:        0,
		Max:        1,
		LiveZone:   game.ZoneLibrary,
		ThenKey:    s82OldB701EyeSpyPicked,
		Carry:      looked,
	})
	return nil
}

// s82OldB701EyeSpyPicked is the card chosen to go to the graveyard. An on-disk
// key: never renamed, never reused.
var s82OldB701EyeSpyPicked = game.RegisterChooseCardsThen("choose-cards/s82-old-b701-eye-spy-graveyard", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return g.PutCardsIntoGraveyardKeyedForEffect(CardsStillInALibrary(g, r.Picked), game.KeyedTail{})
})
