package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rabaroo Troop — Creature — Rabbit Kangaroo {3}{W}{W}, 3/5:
//
//	"Landfall — Whenever a land you control enters, this creature
//	 gains flying until end of turn and you gain 1 life.
//	 Plainscycling {2}"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "48252207-d345-455c-b655-c2027664e6a6",
		Name:         "Rabaroo Troop",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Rabaroo Troop — gains flying until end of turn and you gain 1 life", Do(
				Instruction(func(ctx *Context) error {
					return GrantKeywordUntilEOT{Target: ctx.Item.SourceCardID, Keywords: []string{"flying"},
						Label: "Rabaroo Troop — flying until end of turn"}.Apply(ctx)
				}),
				GainLife{Amount: 1},
			)),
		},
		Activated: []ActivatedAbility{
			Typecycling("Plainscycling", "{2}", "a Plains card", IsLandWithSubtype("Plains")),
		},
	})
}
