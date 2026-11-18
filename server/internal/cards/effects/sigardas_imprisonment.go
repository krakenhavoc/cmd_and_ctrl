package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sigarda's Imprisonment — Enchantment — Aura {2}{W} (Innistrad: Crimson Vow):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block.
//	 {4}{W}: Exile enchanted creature. Create a Blood token."
//
// The Blood token is created even when the host has already gone: the two
// sentences are not an "if you do".
func init() {
	Register(Spec{
		OracleID:     "574230c7-b949-4c90-b107-c26458a6e469",
		Name:         "Sigarda's Imprisonment",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{RestrictAttached(game.CantAttackOrBlock)},
		Activated: []ActivatedAbility{{
			Label:   "{4}{W}: Exile enchanted creature. Create a Blood token.",
			Cost:    ManaCost("{4}{W}"),
			Purpose: game.Purpose{Answers: game.AnswerRemove, Tokens: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if id := coopedUpHost(g, item.SourceCardID); id != uuid.Nil {
					if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
						return err
					}
				}
				return s82Vow03CreateBlood(1).Apply(ctx)
			},
		}},
	})
}
