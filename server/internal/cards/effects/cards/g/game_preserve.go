package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Game Preserve — Enchantment {2}{G} (Mercadian Masques, #5619):
//
//	"At the beginning of your upkeep, each player reveals the top card of
//	 their library. If all cards revealed this way are creature cards, put
//	 those cards onto the battlefield under their owners' control."
//
// Every player still in the game reveals in APNAP order (a player with an
// empty library reveals nothing, and the condition reads only the cards that
// were revealed). Only if every revealed card is a creature card do they all
// go onto the battlefield together (CR 603.6a), each under its owner's
// control.
//
// No simplification.
func init() {
	const name = "Game Preserve"
	effects.Register(effects.Spec{
		OracleID:     "409e0a43-0e85-47ea-9d40-de0860b12952",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep(name+" — each player reveals the top card of their library; if all are creature cards, put them onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					var revealed []uuid.UUID
					for _, p := range effects.ApnapPlayers(g) {
						revealed = append(revealed, g.RevealTopOfLibraryForEffect(p, item.SourceCardID, 1, name+" — revealed from the top of the library")...)
					}
					if len(revealed) == 0 {
						return nil
					}
					entries := make([]game.BatchEntry, 0, len(revealed))
					for _, id := range revealed {
						c, ok := g.LookupCardForEffect(id)
						if !ok || !c.IsCreature() {
							return nil
						}
						entries = append(entries, game.BatchEntry{CardID: id, From: game.ZoneLibrary})
					}
					return g.PutOntoBattlefieldTogetherKeyedForEffect(entries, game.ZoneEntryOptions{}, game.KeyedTail{})
				}),
		},
	})
}
