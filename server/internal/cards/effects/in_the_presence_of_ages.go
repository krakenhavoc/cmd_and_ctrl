package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// In the Presence of Ages — Instant {2}{G} (The Lost Caverns of Ixalan,
// #4470):
//
//	"Reveal the top four cards of your library. You may put a creature
//	 card and/or a land card from among them into your hand. Put the rest
//	 into your graveyard."
//
// Satyr Wayfinder's reveal-and-take with Explore the Vastlands' two
// slots: the four are revealed to every player, then one optional pick
// of up to one creature card and up to one land card (a card that is both
// fills only one slot), and everything not picked is put into the
// graveyard (TakeRestIntoGraveyard: put, not milled). A shorter library
// reveals what it has.
//
// No simplification.
func init() {
	const label = "In the Presence of Ages — you may put a creature card and/or a land card into your hand and the rest into your graveyard"
	Register(Spec{
		OracleID:     "142a960c-4d8e-4b3b-a0db-435c22df8f92",
		Name:         "In the Presence of Ages",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			return TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.RevealTopOfLibraryForEffect(player, ctx.Source(), 4, "In the Presence of Ages — reveal the top four cards"),
				Optional: true,
				Label:    label,
				Key:      inThePresenceOfAgesTake,
			}.Apply(ctx)
		},
	})
}

// inThePresenceOfAgesTake is the look's registered rule and rest: a
// creature slot and a land slot, then the rest into the graveyard. An
// on-disk key: never renamed, never reused.
var inThePresenceOfAgesTake = RegisterTake("choose-cards/in-the-presence-of-ages-take",
	TakeDef{
		Slots: []TakeSlot{
			{Label: "a creature card", Match: Creature(), Max: 1},
			{Label: "a land card", Match: Land(), Max: 1},
		},
		Then: TakeRestIntoGraveyard,
	})
