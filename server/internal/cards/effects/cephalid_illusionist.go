package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cephalid Illusionist — Creature — Octopus Wizard {1}{U}, 1/1:
//
//	"Whenever this creature becomes the target of a spell or ability,
//	 mill three cards.
//	 {2}{U}, {T}: Prevent all combat damage that would be dealt to and
//	 dealt by target creature you control this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the activated ability is one
// to-and-by record (Mod.AndDealtBy) on the target. Targeting the
// Illusionist itself with it triggers the mill, as printed. The trigger
// reads the engine's EventBecomesTarget, emitted for spells and
// abilities alike, once per target slot (CR 115.3), as every "becomes
// the target" card in the catalog does.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "3c3067c4-0ae0-4c6a-9174-e3c981a20235",
		Name:         "Cephalid Illusionist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, Self, "Cephalid Illusionist — mill three cards", func(g *game.Game, item *game.StackItem) error {
				return MillCards{Player: item.Controller, N: 3}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{
			sourceShieldRow("{2}{U}, {T}: Prevent all combat damage that would be dealt to and dealt by target creature you control this turn.",
				Plus(ManaCost("{2}{U}"), TapCost()),
				TargetCreature("target creature you control", YouControl()),
				toAndByShield(ShieldTheTarget, true)),
		},
	})
}
