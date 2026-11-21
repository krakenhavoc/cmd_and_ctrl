package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rabaroo Troop — Creature — Rabbit Kangaroo {3}{W}{W}, 3/5:
//
//	"Landfall — Whenever a land you control enters, this creature
//	 gains flying until end of turn and you gain 1 life.
//	 Plainscycling {2}"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "48252207-d345-455c-b655-c2027664e6a6",
		Name:         "Rabaroo Troop",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Rabaroo Troop — gains flying until end of turn and you gain 1 life", effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.GrantKeywordUntilEOT{Target: ctx.Item.SourceCardID, Keywords: []string{"flying"},
						Label: "Rabaroo Troop — flying until end of turn"}.Apply(ctx)
				}),
				effects.GainLife{Amount: 1},
			)),
		},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Plainscycling", "{2}", "a Plains card", effects.IsLandWithSubtype("Plains")),
		},
	})
}
