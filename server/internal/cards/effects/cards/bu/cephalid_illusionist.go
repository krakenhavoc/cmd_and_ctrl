package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "3c3067c4-0ae0-4c6a-9174-e3c981a20235",
		Name:         "Cephalid Illusionist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.Self, "Cephalid Illusionist — mill three cards", effects.Do(effects.MillCards{N: 3})),
		},
		Activated: []effects.ActivatedAbility{
			effects.SourceShieldRow("{2}{U}, {T}: Prevent all combat damage that would be dealt to and dealt by target creature you control this turn.",
				effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost()),
				effects.TargetCreature("target creature you control", effects.YouControl()),
				effects.ToAndByShield(effects.ShieldTheTarget, true)),
		},
	})
}
