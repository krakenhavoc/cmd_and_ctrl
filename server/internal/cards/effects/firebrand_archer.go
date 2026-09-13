package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Firebrand Archer — Creature — Human Archer {1}{R}, 2/1 (EDHREC rank
// 1125):
//
//	"Whenever you cast a noncreature spell, this creature deals 1
//	 damage to each opponent."
//
// The spellslinger deck's cheapest pinger. Flux Channeler's trigger
// (noncreature is every instant, sorcery, artifact, enchantment and
// planeswalker) with Guttersnipe's payoff, from the Archer itself as
// the source, so a damage doubler or a prevention shield sees a red
// creature source. Fires on CAST, so it resolves before the spell
// that caused it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2e9289d6-dbc6-456d-88cf-d1f534e731d6",
		Name:         "Firebrand Archer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: b10NoncreatureSpellCastByYou,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Firebrand Archer — 1 damage to each opponent",
					func(g *game.Game, item *game.StackItem) error {
						return damageToEachOpponent(g, item, 1)
					})
			},
		}},
	})
}
