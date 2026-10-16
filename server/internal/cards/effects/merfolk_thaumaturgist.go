package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merfolk Thaumaturgist — Creature — Merfolk Wizard {2}{U}, 1/2:
//
//	"{T}: Switch target creature's power and toughness until end of turn."
//
// The switch is layer 7d (CR 613.4d, #3996): it applies after every
// set, pump and counter, whatever their timestamps.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1220e11f-aab0-459b-ae3f-b095aad37294",
		Name:         "Merfolk Thaumaturgist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Switch target creature's power and toughness until end of turn.",
				Cost:    TapCost(),
				Targets: TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return switchEachTargetPT(NewContext(g, item), "Merfolk Thaumaturgist — switch power and toughness")
				},
			},
		},
	})
}
