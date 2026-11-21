package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Young Pyromancer — Creature — Human Shaman {1}{R}, 2/1 (EDHREC rank
// 826):
//
//	"Whenever you cast an instant or sorcery spell, create a 1/1 red
//	 Elemental creature token."
//
// The spellslinger's go-wide engine — Storm-Kiln Artist's trigger
// paying out in bodies instead of Treasures. The token is created
// when the trigger resolves, before the spell that made it (LIFO).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5fac139a-07d3-4e6c-98e3-d98b199f7a6f",
		Name:         "Young Pyromancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && (spell.IsInstant() || spell.IsSorcery())
			}, "Young Pyromancer — create an Elemental", effects.Do(effects.CreateToken{
				Template: effects.TokenCard("1/1 red Elemental"),
				N:        1,
			})),
		},
	})
}
