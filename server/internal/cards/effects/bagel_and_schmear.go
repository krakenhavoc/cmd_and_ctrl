package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bagel and Schmear — Artifact — Food {1} (Marvel's Spider-Man):
//
//	"Share — {W}, {T}, Sacrifice this artifact: Put a +1/+1 counter on up
//	 to one target creature. Draw a card. Activate only as a sorcery.
//	 Nosh — {2}, {T}, Sacrifice this artifact: You gain 3 life and draw a
//	 card."
//
// Two Food abilities (CR 111.10b names Food's sacrifice; the card prints
// both). The target of Share is "up to one": chosen or not, the draw
// happens, and a target that has gone by resolution only loses the counter
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1389c9fa-fec2-4212-a7aa-6dfb318dc202",
		Name:         "Bagel and Schmear",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:        "Share — {W}, {T}, Sacrifice this artifact: Put a +1/+1 counter on up to one target creature. Draw a card. Activate only as a sorcery.",
				Cost:         Plus(ManaCost("{W}"), TapCost(), SacrificeThis()),
				SorcerySpeed: true,
				Targets:      TargetCreature("up to one target creature").WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if err := PlusOneCounterOnTheTarget(g, item); err != nil {
						return err
					}
					return DrawCards{N: 1}.Apply(NewContext(g, item))
				},
			},
			gainThreeLifeAndDrawAbility("Nosh — {2}, {T}, Sacrifice this artifact: You gain 3 life and draw a card."),
		},
	})
}
