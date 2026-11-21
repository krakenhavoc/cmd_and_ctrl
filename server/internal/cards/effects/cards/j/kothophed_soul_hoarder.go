package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kothophed, Soul Hoarder — Legendary Creature — Demon {4}{B}{B}, 6/6:
//
//	"Flying
//	 Whenever a permanent owned by another player is put into a graveyard
//	 from the battlefield, you draw a card and you lose 1 life."
//
// One trigger per permanent put into a graveyard (CR 603.10a: it looks back
// at the event, so it sees the permanent that left, Kothophed's own leaving
// included). "Another player" is a player other than Kothophed's controller,
// read from the permanent's owner, not whoever controlled it: a permanent you
// own that an opponent controlled does not trigger, and one an opponent owns
// that you controlled does. A token is owned by the player who created it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "85e1791f-f9a0-4e82-baf6-33cff2dcf60b",
		Name:            "Kothophed, Soul Hoarder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.NewZone != game.ZoneGraveyard {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.Owner != source.Controller
			}, "Kothophed, Soul Hoarder — you draw a card and you lose 1 life",
				effects.Do(effects.DrawCards{N: 1}, effects.YouLoseLife(1))),
		},
	})
}
