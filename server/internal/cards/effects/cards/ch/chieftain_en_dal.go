package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chieftain en-Dal — Creature — Human Knight {1}{W}{W}, 2/2
// (Nemesis, #5618):
//
//	"Whenever this creature attacks, attacking creatures gain first strike until end of turn."
//
// The attack trigger reads "attacking creatures" as it resolves (CR
// 611.2c): every attacking creature then, whoever controls it, gains
// first strike until the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8aae3a28-dcf3-460e-89f9-4b0c96d6f55e",
		Name:         "Chieftain en-Dal",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Chieftain en-Dal — attacking creatures gain first strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{
						Match:    effects.AttackingCreature(),
						Keywords: []string{"first strike"},
						Label:    "Chieftain en-Dal — first strike until end of turn",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
