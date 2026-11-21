package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guttersnipe — Creature — Goblin Shaman {2}{R}, 2/2 (EDHREC rank
// 402):
//
//	"Whenever you cast an instant or sorcery spell, this creature
//	 deals 2 damage to each opponent."
//
// The spellslinger deck's clock: every cantrip is six damage at a
// four-player table. Storm-Kiln Artist's cast condition with damage
// instead of a Treasure; the damage's source is the Goblin, so a
// noncombat-damage doubler (Solphim) sees it. The trigger goes on
// the stack ABOVE the spell that caused it and resolves first, as in
// paper.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6bdaf76-6a03-4695-9c4b-f040e73435af",
		Name:         "Guttersnipe",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.InstantOrSorceryCastByYou(ev, source, g)
			}, "Guttersnipe — 2 damage to each opponent", effects.Do(effects.DamageEachOpponent{N: 2})),
		},
	})
}
