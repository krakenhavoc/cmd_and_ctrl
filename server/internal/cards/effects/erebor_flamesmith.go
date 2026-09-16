package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Erebor Flamesmith — Creature — Dwarf Artificer {1}{R}, 2/1 (EDHREC
// rank 2427):
//
//	"Whenever you cast an instant or sorcery spell, this creature
//	 deals 1 damage to each opponent."
//
// The spellslinger pinger. Thousand-Year Storm's condition (the
// spell's type read off the stack, where its type line is intact)
// with Impact Tremors' payoff; the Flamesmith itself is the damage
// source, so a doubler sees it. Cast, not resolve, so a countered
// spell still pings.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6eb93545-a1a6-4447-82c5-421d8e9f023d",
		Name:         "Erebor Flamesmith",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return instantOrSorceryCastByYou(ev, source, g)
			}, "Erebor Flamesmith — 1 damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 1)
			}),
		},
	})
}
