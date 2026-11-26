package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flowstone Blade — Enchantment — Aura {R} (Conspiracy, S82 OLD batch 362,
// #5557):
//
//	"Enchant creature
//	 {R}: Enchanted creature gets +1/-1 until end of turn."
//
// The Aura's own activated ability (its controller pays, whoever controls the
// creature): the boost goes to the creature the Aura is attached to as it
// resolves, and nothing happens if it is no longer attached (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77a77622-dc5b-4ff2-a98b-732d44c8856e",
		Name:         "Flowstone Blade",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Enchanted creature gets +1/-1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				host := effects.EnchantedCreatureOf(ctx)
				if host == uuid.Nil {
					return nil
				}
				return effects.BoostUntilEOT{Target: host, Power: 1, Toughness: -1, Label: "Flowstone Blade — +1/-1 until end of turn"}.Apply(ctx)
			},
		}},
	})
}
