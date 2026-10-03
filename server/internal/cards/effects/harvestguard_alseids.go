package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Harvestguard Alseids — Enchantment Creature — Nymph {2}{W}, 2/3:
//
//	"Constellation — Whenever this creature or another enchantment you
//	 control enters, prevent all damage that would be dealt to target
//	 creature this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): constellation is Grim Guardian's
// trigger (b30SelfOrAnotherEnchantmentYouControlEntered), and its
// effect is the not-one-use shield pinned to the targeted creature.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "ad88bf37-bbb4-46f0-9994-1863c7c31a2a",
		Name:         "Harvestguard Alseids",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b30SelfOrAnotherEnchantmentYouControlEntered(ev, source, g)
			}, "Harvestguard Alseids — prevent all damage to target creature this turn (constellation)", func(g *game.Game, item *game.StackItem) error {
				return PreventDamageFromSource{Protect: ShieldTheTarget}.Apply(NewContext(g, item))
			}), TargetCreature("target creature")),
		},
	})
}
