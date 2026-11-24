package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shattered Ego — Enchantment — Aura {U} (Modern Horizons 2, #5501):
//
//	"Enchant creature
//	 Enchanted creature gets -3/-0.
//	 {3}{U}{U}: Put enchanted creature into its owner's library third from
//	 the top."
//
// The tuck is Riptide Gearhulk's third-from-the-top landing (a library of
// fewer than two cards takes it on the bottom).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "829436df-a48b-4535-84ec-64682a16949b",
		Name:         "Shattered Ego",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.PumpAttached(-3, 0)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}{U}: Put enchanted creature into its owner's library third from the top.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.ManaCost("{3}{U}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				host := effects.AttachedHostFor(g, item.SourceCardID)
				if host == nil {
					return nil
				}
				return effects.B22TuckThirdFromTop(g, host.InstanceID)
			},
		}},
	})
}
