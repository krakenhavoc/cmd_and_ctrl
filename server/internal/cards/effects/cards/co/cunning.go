package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cunning — Enchantment — Aura {1}{U} (Exodus, S82 OLD b739, #5626):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3.
//	 When enchanted creature attacks or blocks, sacrifice this Aura at the
//	 beginning of the next cleanup step."
//
// The pump is the shared attachment pump. The trigger fires once per attack or
// block declaration of the enchanted creature (Primal Cocoon's watch) and
// schedules a delayed trigger for the next cleanup step (CR 603.7), whose
// source is this Aura: it is sacrificed if it is still the same object on the
// battlefield then, and the +3/+3 lasts through the turn's combat and end step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5e48041-5f21-48e6-bd0c-000f02dfbe8a",
		Name:         "Cunning",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(3, 3)},
		Triggered: []game.TriggeredAbility{
			effects.S82OldB476EnchantedAttacksOrBlocks("Cunning — sacrifice this Aura at the beginning of the next cleanup step",
				cunningSacrificeAtNextCleanup("Cunning — sacrifice this Aura")),
		},
	})
}

// cunningSacrificeBody is the delayed trigger "sacrifice this Aura at the
// beginning of the next cleanup step" schedules: the trigger's source, if it
// is still the same object on the battlefield. A registered key (ADR 0041
// phase 3), so a table with it waiting is a restore point.
var cunningSacrificeBody = game.SimpleDelayedBody("cunning/sacrifice-at-next-cleanup", effects.SacrificeDelayedTriggerSource)

// cunningSacrificeAtNextCleanup schedules the delayed trigger (CR 603.7) at
// the next cleanup step, its source being this Aura.
func cunningSacrificeAtNextCleanup(label string) effects.Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return effects.ScheduleDelayedTrigger{
			At:    game.StepCleanup,
			Label: label,
			Body:  cunningSacrificeBody,
		}.Apply(effects.NewContext(g, item))
	}
}
