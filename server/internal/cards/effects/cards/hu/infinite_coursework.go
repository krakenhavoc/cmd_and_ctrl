package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infinite Coursework — Enchantment — Aura {2}{U}:
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature. It becomes
//	 unprepared.
//	 Enchanted creature loses all abilities and doesn't untap during its
//	 controller's untap step."
//
// Claustrophobia's shape plus the unprepare and the ability loss. The
// entry trigger reads the Aura's attachment as it resolves
// (tapEnchantedCreatureOnEntry's rule: if the Aura has left in response,
// the creature it last enchanted), taps it and removes the prepared
// designation (CR 722.3b), which also removes the copy of its prepare
// spell from exile. Only the "doesn't untap" and "loses all abilities"
// halves end with the Aura; the tap and the unprepare do not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "666b34b5-a0c2-43e3-befd-7c3979313682",
		Name:                  "Infinite Coursework",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Static:                []game.StaticAbility{effects.LoseAllAbilities()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Infinite Coursework — tap enchanted creature; it becomes unprepared",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					aura, ok := ctx.TriggeringPermanent()
					if !ok || aura.AttachedTo.Kind != game.TargetCard {
						return nil
					}
					if err := (effects.TapTarget{Target: aura.AttachedTo.ID}).Apply(ctx); err != nil {
						return err
					}
					return effects.BecomeUnprepared{Target: aura.AttachedTo.ID}.Apply(ctx)
				}),
		},
	})
}
