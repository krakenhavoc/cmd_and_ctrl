package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vizier of the Anointed — Creature — Human Cleric {3}{U}, 2/4 (Hour of
// Devastation, #5534):
//
//	"When this creature enters, you may search your library for a creature
//	 card with eternalize or embalm, put that card into your graveyard, then
//	 shuffle.
//	 Whenever you activate an eternalize or embalm ability, draw a card."
//
// The search is Entomb's, optional, over the library's creature cards that
// offer eternalize or embalm. The draw is a trigger on the activation event
// (CR 603.2): it fires for the controller's own activation of either keyword
// from the graveyard, including one of the Vizier's own kind; each activation
// draws once. The activation is told apart by its label, which the keyword's
// shared constructor begins with the keyword, because the exiled card is
// already gone when the watcher runs.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb5aa38e-6a76-4c0c-be14-3c546acd353d",
		Name:         "Vizier of the Anointed",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{SelfMillTutor: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Vizier of the Anointed — you may search your library for a creature card with eternalize or embalm, put that card into your graveyard, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player: item.Controller,
						Predicate: func(c game.Card) bool {
							return c.IsCreature() && effects.S82OldB235HasEternalizeOrEmbalm(c)
						},
						Dest:     game.ZoneGraveyard,
						Limit:    1,
						Optional: true,
						Shuffle:  true,
						Reason:   "Vizier of the Anointed — a creature card with eternalize or embalm",
					}.Apply(effects.NewContext(g, item))
				}),
			effects.On(game.EventActivateAbility, effects.AllOf(effects.ByYou, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.S82OldB235IsEternalizeOrEmbalm(ev)
			}), "Vizier of the Anointed — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
