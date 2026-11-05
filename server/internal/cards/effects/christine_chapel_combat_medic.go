package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Christine Chapel, Combat Medic — Legendary Creature — Human Doctor
// {3}{G}, 2/2 (Star Trek Commander, #4974):
//
//	"Strange new worlds — When Christine Chapel enters, you gain life
//	 equal to the number of differently named lands you control.
//	 Whenever you gain life, put two +1/+1 counters on Christine Chapel."
//
// "Strange new worlds" is an ability word with no rules meaning. The
// lands are counted as the enter trigger resolves, each land you control
// once per name (socB23DifferentlyNamedLands); with none, no life is
// gained. The second trigger is Ajani's Pridemate's: once per life-gain
// event, not once per point, so the enter trigger's own gain puts two
// counters on her.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "330e8a18-d224-4416-b59d-02b749c3a674",
		Name:         "Christine Chapel, Combat Medic",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Christine Chapel, Combat Medic — you gain life equal to the number of differently named lands you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return GainLife{Amount: socB23DifferentlyNamedLands(ctx)}.Apply(ctx)
				}),
			WheneverYouGainLife("Christine Chapel, Combat Medic — put two +1/+1 counters on her",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 2})),
		},
	})
}
