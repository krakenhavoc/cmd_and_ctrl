package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Charisma — Enchantment — Aura {U}{U}{U} (Mercadian Masques, S82 OLD batch 686, #5619):
//
//	"Enchant creature
//	Whenever enchanted creature deals damage to a creature, gain control of the other creature for as long as this Aura remains on the battlefield."
//
// Giant's Grasp's control effect keyed on the Aura as the object it is now
// (DurationWhileSourceRemains, CR 611.2b), triggered by each damage event the
// enchanted creature deals to a creature, combat or not. A creature that has left
// the battlefield by resolution is not gained, and an Aura that has left (or left
// and returned) takes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2abfcd1-c00c-43fa-a9ca-a9577b3c8a60",
		Name:         "Charisma",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, charismaDamagedACreature,
				"Charisma — gain control of the other creature for as long as this Aura remains on the battlefield",
				charismaGainControl),
		},
	})
}

// charismaDamagedACreature is "whenever enchanted creature deals damage to a
// creature": any damage from the creature the Aura is attached to, to a creature
// still on the battlefield.
func charismaDamagedACreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || !source.IsAttachedTo(ev.Source) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	return ok && c.IsCreature() && effects.OnBattlefield(g, ev.Target)
}

// charismaGainControl takes the damaged creature for as long as this Aura stays.
func charismaGainControl(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || !effects.OnBattlefield(g, item.Trigger.Event.Target) {
		return nil
	}
	ctx := effects.NewContext(g, item)
	d, ok := effects.DurationWhileSourceRemains(ctx, item.SourceCardID)
	if !ok {
		return nil
	}
	return effects.GainControl{
		Target:     item.Trigger.Event.Target,
		Controller: item.Controller,
		Duration:   d,
		Label:      "Charisma — control for as long as this Aura remains",
	}.Apply(ctx)
}
