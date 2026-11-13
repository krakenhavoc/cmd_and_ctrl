package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8aae3a28-dcf3-460e-89f9-4b0c96d6f55e",
		Name:         "Chieftain en-Dal",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Chieftain en-Dal — attacking creatures gain first strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return GrantKeywordUntilEOT{
						Match:    AttackingCreature(),
						Keywords: []string{"first strike"},
						Label:    "Chieftain en-Dal — first strike until end of turn",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
