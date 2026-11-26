package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of Misfortunes — Enchantment — Aura Curse {4}{B} (Dark Ascension, S82
// OLD, #5571):
//
//	"Enchant player
//	 At the beginning of your upkeep, you may search your library for a Curse
//	 card that doesn't have the same name as a Curse attached to enchanted
//	 player, put it onto the battlefield attached to that player, then
//	 shuffle."
//
// Arachnus Spinner's search that puts the card onto the battlefield attached
// (SearchLibrarySpec.EntersAttachedTo), here to a player: the "you may" is the
// trigger's own question, the names to avoid are read as the search begins (the
// Curses then attached to the enchanted player, this one included), and a
// Curse that can't legally enchant that player stays in the library (CR
// 303.4i). The library is shuffled whether or not a card is found. If the
// Curse has left, nothing is searched for.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b03c7c8d-8206-45cd-8ead-8f7b35c03a7a",
		Name:         "Curse of Misfortunes",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Curse of Misfortunes — search your library for a Curse card with a different name than each Curse attached to enchanted player, put it onto the battlefield attached to that player, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ref, ok := ctx.SourceRef()
					if !ok {
						return nil
					}
					aura, ok := g.PermanentForEffect(ref)
					if !ok || aura.AttachedTo.Kind != game.TargetPlayer {
						return nil
					}
					player := aura.AttachedTo.ID
					names := map[string]bool{}
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.AttachedTo.Kind == game.TargetPlayer && c.AttachedTo.ID == player && c.HasSubtype("Curse") {
							names[c.Name] = true
						}
					}
					return g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
						Player:           item.Controller,
						Source:           ctx.Source(),
						Pred:             func(c game.Card) bool { return c.HasSubtype("Curse") && !names[c.Name] },
						Dest:             game.ZoneBattlefield,
						Limit:            1,
						Shuffle:          true,
						EntersAttachedTo: game.TargetRef{Kind: game.TargetPlayer, ID: player},
						Reason:           "Curse of Misfortunes — search your library for a Curse card",
					})
				}),
				"Curse of Misfortunes — search your library for a Curse card?"),
		},
	})
}
