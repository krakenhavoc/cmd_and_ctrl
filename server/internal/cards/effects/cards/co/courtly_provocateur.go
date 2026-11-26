package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Courtly Provocateur — Creature — Human Wizard {2}{U}, 1/1 (Magic 2013, S82 OLD, #5568):
//
//	"{T}: Target creature attacks this turn if able.
//	 {T}: Target creature blocks this turn if able."
//
// Imps' Taunt's attack requirement (CR 508.1d) and Hustle's block requirement
// (CR 509.1c), each until end of turn, each an ability of its own. The
// requirement binds only if the creature can attack or block this turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3752277a-5b3c-4902-90a1-63acdac66dc9",
		Name:         "Courtly Provocateur",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Target creature attacks this turn if able.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Effect: effects.Do(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.ScopedEffectFor{
						Target:   id,
						Mods:     []game.Mod{game.AddAttackRequirementMod(uuid.Nil)},
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Courtly Provocateur — attacks this turn if able",
					}.Apply(ctx)
				})),
			},
			{
				Label:   "{T}: Target creature blocks this turn if able.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Effect: effects.Do(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.ScopedEffectFor{
						Target:   id,
						Mods:     []game.Mod{game.AddBlockRequirementMod(game.BlockRequirementBlocks)},
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Courtly Provocateur — blocks this turn if able",
					}.Apply(ctx)
				})),
			},
		},
	})
}
