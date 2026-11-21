package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Forced Worship — Enchantment — Aura {1}{W} (Jumpstart 2022, #5948):
//
//	"Enchant creature
//	 Enchanted creature can't attack.
//	 {2}{W}: Return this Aura to its owner's hand."
//
// A one-sided Pacifism and Flickering Ward's bounce: a mana activation with
// no target, so the Aura can save itself in response to removal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea6b1e8e-b6e7-4fce-bfb7-7adf66b9f240",
		Name:         "Forced Worship",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.RestrictAttached(game.CantAttack)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}: Return this Aura to its owner's hand",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{2}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BounceToHand{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
