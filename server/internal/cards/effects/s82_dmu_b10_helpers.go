package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_dmu_b10_helpers.go — the look-at-the-top-X sentences of the Dominaria
// United domain cards of S82 batch 10 (#5487).

// s82DmuB10IsLandWithBasicType is "a land card that has a basic land type"
// (CR 305.6): any land card whose layered-in-library subtypes name one of
// the five basic land types, so a Temple Garden counts and a Wastes does
// not. A card in a library has only its printed types.
func s82DmuB10IsLandWithBasicType(c game.Card) bool {
	if !c.IsLand() {
		return false
	}
	for _, t := range game.BasicLandTypes {
		if c.HasSubtype(t) {
			return true
		}
	}
	return false
}

// s82DmuB10LookKeepOneOnTop is "Look at the top X cards of your library,
// where X is the number of basic land types among lands you control. [You
// may] put [up to] one of those cards on top of your library. Put the rest
// on the bottom of your library in a random order" (Voda Sea Scavenger,
// Slimefoot's Survey). X is counted as the clause resolves. The kept card
// is the one the player names; once the rest are under the library, it is
// the top card. With no card named, every card goes under.
func s82DmuB10LookKeepOneOnTop(ctx *Context, label string) error {
	player := ctx.Controller()
	n := DomainCount(ctx.Game, player)
	if n <= 0 {
		return nil
	}
	looked := ctx.Game.LookAtTopOfLibraryForEffect(player, n)
	if len(looked) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: label + " — you may put one of those cards on top of your library; the rest go on the bottom in a random order",
		Cards:    looked,
		Min:      0,
		Max:      1,
		ThenKey:  s82DmuB10KeepOneOnTop,
		Carry:    looked,
	})
	return nil
}

// s82DmuB10KeepOneOnTop sends every looked-at card that was not kept to the
// bottom in a random order. An on-disk key: never renamed, never reused.
var s82DmuB10KeepOneOnTop = game.RegisterChooseCardsThen("choose-cards/s82-dmu-b10-keep-one-on-top",
	func(g *game.Game, r game.CardsChosen) error {
		kept := map[uuid.UUID]bool{}
		if r.Answered {
			for _, id := range r.Picked {
				kept[id] = true
			}
		}
		var rest []uuid.UUID
		for _, id := range r.Carry {
			if !kept[id] {
				rest = append(rest, id)
			}
		}
		if len(rest) == 0 {
			return nil
		}
		return g.PutOnBottomInRandomOrderForEffect(r.Chooser, game.ZoneLibrary, rest)
	})
