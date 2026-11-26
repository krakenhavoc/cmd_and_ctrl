package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Illusory Gains — Enchantment — Aura {3}{U}{U} (Dragons of Tarkir, S82 old batch
// 332):
//
//	"Enchant creature
//	 You control enchanted creature.
//	 Whenever a creature an opponent controls enters, attach this Aura to that
//	 creature."
//
// Mind Control's layer-2 control effect (ControlAttachedBySource), so control
// follows the Aura and reverts when it moves or leaves. The trigger is on the Aura:
// a creature entering under an opponent's control (the Aura's controller's
// opponent, tokens included) moves it, as its own ability resolves, onto that
// creature if it is still on the battlefield. An Aura that can't legally enchant
// the creature stays where it is (CR 303.4d, CR 701.3b), and the Aura's controller
// stays the same: moving it does not change who controls the Aura.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ba45691-417b-4e6a-be45-6027e4f04599",
		Name:         "Illusory Gains",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.ControlAttachedBySource()},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.OpponentsCreatureEntered,
				"Illusory Gains — attach this Aura to that creature", effects.AttachThisAuraToTheEnteredCreature),
		},
	})
}
