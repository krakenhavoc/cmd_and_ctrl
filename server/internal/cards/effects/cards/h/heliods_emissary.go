package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heliod's Emissary — Enchantment Creature — Elk {3}{W}, 3/3 (Theros):
//
//	"Bestow {6}{W} (If you cast this card for its bestow cost, it's an Aura
//	 spell with enchant creature. It becomes a creature again if it's not
//	 attached.)
//	 Whenever this creature or enchanted creature attacks, tap target creature
//	 an opponent controls.
//	 Enchanted creature gets +3/+3."
//
// One printed trigger, one registered trigger: the attack of the source or of
// its host (Spiteful Returned's condition), with a target chosen as it goes on
// the stack. A target gone by resolution is not tapped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "182c9af9-d867-4848-ab36-ac7b1914f41a",
		Name:             "Heliod's Emissary",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Bestow("{6}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventAttack, effects.S82OldB377ThisOrEnchantedAttacked,
					"Heliod's Emissary — tap target creature an opponent controls", effects.TapTheFirstLegalTarget),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
		Static: []game.StaticAbility{effects.PumpAttached(3, 3)},
	})
}
