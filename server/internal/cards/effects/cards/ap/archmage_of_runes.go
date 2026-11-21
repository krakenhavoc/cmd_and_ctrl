package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Archmage of Runes — Creature — Giant Wizard {3}{U}{U}, 3/6:
//
//	"Instant and sorcery spells you cast cost {1} less to cast.
//	 Whenever you cast an instant or sorcery spell, draw a card."
//
// The discount is InstantOrSorcerySpell(), Goblin Electromancer's
// predicate. The draw trigger is Or(Instant(), Sorcery()) on the
// ordinary WheneverYouCast constructor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "28045b32-4c1a-40e5-a15d-524d0f8fe6ec",
		Name:         "Archmage of Runes",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Instant and sorcery spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.InstantOrSorcerySpell()),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Archmage of Runes — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
