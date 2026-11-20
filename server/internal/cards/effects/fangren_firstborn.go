package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fangren Firstborn — Creature — Beast {1}{G}{G}{G} (Darksteel, #5605):
//
//	"Whenever this creature attacks, put a +1/+1 counter on each
//	 attacking creature."
//
// An attack trigger (CR 508.3a). "Each attacking creature" is read as it
// resolves (AttackingCreatures), so a creature removed from combat in
// response gets nothing, and the Firstborn itself gets one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2affbe3f-5292-43d9-a08c-23bf44629cc9",
		Name:         "Fangren Firstborn",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Fangren Firstborn — put a +1/+1 counter on each attacking creature",
				func(g *game.Game, item *game.StackItem) error {
					return b13PutCounterOnEach(NewContext(g, item), AttackingCreatures(g))
				}),
		},
	})
}
