package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Supernatural Rescue — Enchantment — Aura {3}{W} (Innistrad: Crimson Vow,
// S82):
//
//	"This spell has flash as long as you control a Spirit.
//	 When you cast this spell, tap up to two target creatures you don't
//	 control.
//	 Enchant creature you control
//	 Enchanted creature gets +1/+2."
//
// The flash is Bard's Company's FlashIf, read as the cast begins (CR
// 601.3). The cast trigger (Abundant Maw's WhenYouCastThisSpell) chooses
// its targets as it goes on the stack and resolves before the Aura does
// (CR 603.3a); with no creature to tap it is simply removed (CR 603.3d),
// and the Aura's own target is the "enchant" restriction.
//
// No simplification.
func init() {
	cast := effects.WhenYouCastThisSpell("Supernatural Rescue — tap up to two target creatures you don't control", effects.Msc13TapTheLegalTargets)
	cast.Targets = effects.TargetCreature("up to two target creatures you don't control", effects.Not(effects.YouControl())).WithCount(0, 2)
	effects.Register(effects.Spec{
		OracleID:     "a77510ef-baf0-4cc5-b659-6aaaed0b02f4",
		Name:         "Supernatural Rescue",
		Completeness: effects.CompletenessFull,
		FlashIf:      effects.FlashIfYou("as long as you control a Spirit", effects.VowB13YouControlASpirit),
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Static:       []game.StaticAbility{effects.PumpAttached(1, 2)},
		Triggered:    []game.TriggeredAbility{cast},
	})
}
