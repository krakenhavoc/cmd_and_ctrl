package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scrapheap — Artifact {3} (Urza's Legacy, S82 OLD b716, #5623):
//
//	"Whenever an artifact or enchantment is put into your graveyard from the battlefield, you gain 1 life."
//
// Soulcatchers' Aerie's put-into-your-graveyard reading: the card as it last
// existed on the battlefield was an artifact or an enchantment, and its OWNER
// is the Scrapheap's controller, so a permanent of yours that an opponent
// stole counts and one of theirs that you stole does not. A token never
// reaches a graveyard that counts as "put into". One trigger per permanent
// (CR 603.2c); Scrapheap sees itself go with the others (CR 603.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dfed42b8-b35e-4c70-9e75-25a4da158e76",
		Name:         "Scrapheap",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !c.IsToken() && c.Owner == source.Controller && (effects.LeftAsType(ev, c, "artifact") || effects.LeftAsType(ev, c, "enchantment"))
			}, "Scrapheap — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
