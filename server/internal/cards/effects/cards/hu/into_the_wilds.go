package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Into the Wilds — Enchantment {3}{G} (Magic 2014, S82 OLD batch 401, #5563):
//
//	"At the beginning of your upkeep, look at the top card of your library. If
//	 it's a land card, you may put it onto the battlefield."
//
// Fecund Greenshell's look: the look is private, a land card may go onto the
// battlefield (untapped), and a card that is not a land, or a land you decline,
// stays on top of the library.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70d6cd9c-6311-4d8e-b74a-8edc11d1c05e",
		Name:         "Into the Wilds",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Into the Wilds — look at the top card of your library; a land card may go onto the battlefield", intoTheWildsLook),
		},
	})
}

func intoTheWildsLook(g *game.Game, item *game.StackItem) error {
	player := item.Controller
	top := g.LookAtTopOfLibraryForEffect(player, 1)
	if len(top) == 0 {
		return nil
	}
	c, ok := g.LookupCardForEffect(top[0])
	if !ok || !c.IsLand() {
		return nil
	}
	return effects.PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    top,
		Match:    effects.Land(),
		Max:      1,
		Optional: true,
		Label:    "Into the Wilds — you may put the land card onto the battlefield",
	}.Apply(effects.NewContext(g, item))
}
