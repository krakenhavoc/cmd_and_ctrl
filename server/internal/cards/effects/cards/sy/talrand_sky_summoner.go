package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Talrand, Sky Summoner — Legendary Creature — Merfolk Wizard {2}{U}{U},
// 2/2 (EDHREC rank 785):
//
//	"Whenever you cast an instant or sorcery spell, create a 2/2 blue
//	 Drake creature token with flying."
//
// The spellslinger commander: every cantrip is a flier. Storm-Kiln
// Artist's cast trigger with a different payout — the spell's type is
// read off the stack when the cast event fires, and the Drake arrives
// when the trigger resolves, before the spell that made it (LIFO), as
// in paper.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea1eb902-a23c-44ff-9169-19baf71de238",
		Name:         "Talrand, Sky Summoner",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && (spell.IsInstant() || spell.IsSorcery())
			}, "Talrand, Sky Summoner — create a 2/2 Drake with flying", effects.Do(effects.CreateToken{
				Template: effects.TokenCard("2/2 blue Drake with flying"),
				N:        1,
			})),
		},
	})
}
