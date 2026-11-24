package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// So Shiny — Enchantment — Aura {2}{U} (Modern Horizons 2, #5501):
//
//	"Enchant creature
//	 When this Aura enters, if you control a token, tap enchanted creature,
//	 then scry 2.
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// An intervening "if" (CR 603.4) on the enters trigger; the tap and the
// scry are Bitter Chill's pieces. The untap lock is the shared Aura rule.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "4bc34109-2050-4736-8247-ba06f0a9114a",
		Name:                  "So Shiny",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm, Scry: 2},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.S82OldB38ControlsAToken(),
				effects.WhenThisEnters("So Shiny — tap enchanted creature, then scry 2",
					func(g *game.Game, item *game.StackItem) error {
						if err := effects.TapEnchantedCreatureOnEntry(g, item); err != nil {
							return err
						}
						return effects.Scry{Player: item.Controller, N: 2}.Apply(effects.NewContext(g, item))
					})),
		},
	})
}
