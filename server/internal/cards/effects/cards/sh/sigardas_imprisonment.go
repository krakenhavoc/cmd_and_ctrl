package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "574230c7-b949-4c90-b107-c26458a6e469",
		Name:         "Sigarda's Imprisonment",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.RestrictAttached(game.CantAttackOrBlock)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{W}: Exile enchanted creature. Create a Blood token.",
			Cost:    effects.ManaCost("{4}{W}"),
			Purpose: game.Purpose{Answers: game.AnswerRemove, Tokens: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if id := effects.CoopedUpHost(g, item.SourceCardID); id != uuid.Nil {
					if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
						return err
					}
				}
				return effects.S82Vow03CreateBlood(1).Apply(ctx)
			},
		}},
	})
}
