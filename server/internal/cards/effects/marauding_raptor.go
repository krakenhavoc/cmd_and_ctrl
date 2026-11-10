package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Marauding Raptor — Creature — Dinosaur {1}{R}, 2/3 (Lost Caverns of
// Ixalan Commander):
//
//	"Creature spells you cast cost {1} less to cast.
//	 Whenever another creature you control enters, this creature deals 2
//	 damage to it. If a Dinosaur is dealt damage this way, this creature
//	 gets +2/+0 until end of turn."
//
// The discount is a cost modifier (CR 601.2f). The trigger fires once for
// each creature that enters under your control, including a token and
// including a Dinosaur; the damage is dealt to the creature that entered,
// and only a Dinosaur that actually took damage (a prevented hit does not
// count, CR 120.3) pumps the Raptor. The pump is a continuation told what
// the creature took, so a prompt in the damage window does not lose it.
func init() {
	const label = "Marauding Raptor — 2 damage to the creature that entered; +2/+0 if it is a Dinosaur"
	Register(Spec{
		OracleID:     "b3fe2f61-c58e-4ebc-a5e3-99d6b6f62caf",
		Name:         "Marauding Raptor",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Creature spells you cast cost {1} less to cast.", YourSpell(), CreatureSpell()),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature()
			}, label, maraudingRaptorDamage),
		},
	})
}

func maraudingRaptorDamage(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	return DealDamageThen(ctx, item.Trigger.Event.CardID, 2, itemRecipient(item, maraudingRaptorDealt))
}

var maraudingRaptorDealt = recipientTail("recipient/marauding-raptor-dinosaur-pump", func(ctx *Context, target uuid.UUID, dealt int, _ game.TailOutcome) error {
	if dealt <= 0 || !onBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(target)
	if !ok || !c.HasSubtype("Dinosaur") {
		return nil
	}
	return BoostUntilEOT{Target: ctx.Source(), Power: 2, Label: "Marauding Raptor — +2/+0"}.Apply(ctx)
})
