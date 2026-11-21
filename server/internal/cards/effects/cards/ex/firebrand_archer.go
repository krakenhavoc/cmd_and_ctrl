package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "2e9289d6-dbc6-456d-88cf-d1f534e731d6",
		Name:         "Firebrand Archer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Firebrand Archer — 1 damage to each opponent", effects.Do(effects.DamageEachOpponent{N: 1})),
		},
	})
}
