package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Radha, Heir to Keld — Legendary Creature — Elf Warrior {R}{G}, 2/2
// (Dominaria Remastered, #5948):
//
//	"Whenever Radha attacks, you may add {R}{R}.
//	 {T}: Add {G}."
//
// The attack trigger is a "may" asked as it resolves (Soulbright
// Flamekin's shape); the tap ability is a plain mana ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "61d94fe7-b559-4bd4-94a7-f4ed4c853b4b",
		Name:         "Radha, Heir to Keld",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Radha, Heir to Keld — you may add {R}{R}", func(g *game.Game, item *game.StackItem) error {
				return effects.MayChoice{
					Question: "Radha, Heir to Keld — add {R}{R}?",
					Then:     radhaAddMana,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

var radhaAddMana = effects.MayYes("may/radha-heir-to-keld-add-mana", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return effects.AddMana{Produced: "{R}{R}"}.Apply(ctx)
})
