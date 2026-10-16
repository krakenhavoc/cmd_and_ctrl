package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dwarven Thaumaturgist — Creature — Dwarf Shaman {2}{R}, 1/2:
//
//	"{T}: Switch target creature's power and toughness until end of turn."
//
// The switch is layer 7d (CR 613.4d, #3996): it applies after every
// set, pump and counter, whatever their timestamps.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "019d0013-52fd-4261-9cc8-7efcee5a4616",
		Name:         "Dwarven Thaumaturgist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Switch target creature's power and toughness until end of turn.",
				Cost:    TapCost(),
				Targets: TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return switchEachTargetPT(NewContext(g, item), "Dwarven Thaumaturgist — switch power and toughness")
				},
			},
		},
	})
}
