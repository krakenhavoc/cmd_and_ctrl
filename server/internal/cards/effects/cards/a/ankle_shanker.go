package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ankle Shanker — Creature — Goblin Berserker {2}{R}{W}{B}, 2/2 (Commander
// 2016, S82 older sets, #5537):
//
//	"Haste
//	 Whenever this creature attacks, creatures you control gain first strike
//	 and deathtouch until end of turn."
//
// "Creatures you control" is fixed as the trigger resolves (CR 611.2c): a
// creature that enters later this turn does not gain the keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "22052572-0b7d-4dd8-84a8-e2658f77b85e",
		Name:            "Ankle Shanker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Ankle Shanker — creatures you control gain first strike and deathtouch until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{
						Match:    effects.And(effects.Creature(), effects.YouControl()),
						Keywords: []string{"first strike", "deathtouch"},
						Label:    "Ankle Shanker — first strike and deathtouch",
					}.Apply(effects.NewContext(g, item).AsGroupMember())
				}),
		},
	})
}
