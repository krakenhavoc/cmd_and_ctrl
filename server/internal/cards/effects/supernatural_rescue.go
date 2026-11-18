package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	cast := WhenYouCastThisSpell("Supernatural Rescue — tap up to two target creatures you don't control",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		})
	cast.Targets = TargetCreature("up to two target creatures you don't control", Not(YouControl())).WithCount(0, 2)
	Register(Spec{
		OracleID:     "a77510ef-baf0-4cc5-b659-6aaaed0b02f4",
		Name:         "Supernatural Rescue",
		Completeness: CompletenessFull,
		FlashIf:      FlashIfYou("as long as you control a Spirit", vowB13YouControlASpirit),
		Targets:      EnchantCreature(YouControl()),
		Static:       []game.StaticAbility{PumpAttached(1, 2)},
		Triggered:    []game.TriggeredAbility{cast},
	})
}
