package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Maarika, Brutal Gladiator — Legendary Creature — Human Warrior
// {2}{B}{R}{G}, 7/4 (Secret Lair Drop, #5462):
//
//	"Maarika, Brutal Gladiator must be blocked if able.
//	 As long as it's your turn, Maarika has indestructible.
//	 Whenever Maarika deals damage to a creature, if that creature was dealt
//	 excess damage this turn, that creature's controller sacrifices a
//	 noncreature, nonland permanent."
//
// The block rule is the shared MustBeBlocked static. The indestructible is
// Faithful Pikemaster's turn-gated layer 6 grant. The trigger asks the turn
// tally which creature OBJECTS were dealt excess damage (CR 120.10,
// ExcessDamageRecord.Object, by any source this turn, this damage included),
// when the damage lands. The tally can only grow within a turn, so the
// intervening "if" (CR 603.4) still holds as the ability resolves, and it
// names who controlled the creature as it was dealt the damage: that player
// chooses a noncreature, nonland permanent they control to sacrifice, and
// does nothing if they have none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a46f9f5c-775c-470d-8935-9556e1bbb23d",
		Name:         "Maarika, Brutal Gladiator",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.MustBeBlocked(),
			{
				Label: "As long as it's your turn, Maarika has indestructible.",
				Layer: game.Layer6Ability,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID && effects.IsActivePlayer(g, source.Controller)
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Abilities = game.AppendKeywordAbility(c.Abilities, "indestructible")
				},
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, maarikaDamagedAnExcessCreature,
				"Maarika, Brutal Gladiator — that creature's controller sacrifices a noncreature, nonland permanent",
				maarikaSacrifice),
		},
	})
}

// maarikaDamagedAnExcessCreature is "Whenever Maarika deals damage to a
// creature, if that creature was dealt excess damage this turn".
func maarikaDamagedAnExcessCreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !effects.ThisDealtDamageToACreature(ev, source, game.Characteristic{}, g) {
		return false
	}
	target, ok := g.LookupCardForEffect(ev.Target)
	if !ok {
		return false
	}
	_, excess := g.ObjectDealtExcessDamageThisTurn(game.ObjectRef{ID: target.InstanceID, Epoch: target.ObjectEpoch})
	return excess
}

func maarikaSacrifice(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	controller, excess := g.ObjectIDDealtExcessDamageThisTurn(item.Trigger.Event.Target)
	if !excess || controller == uuid.Nil {
		return nil
	}
	g.PlayerSacrificesForEffect(item.SourceCardID, controller,
		effects.SacrificeSpec("a noncreature, nonland permanent", effects.Not(effects.Creature()), effects.Nonland()),
		"Maarika, Brutal Gladiator — sacrifice a noncreature, nonland permanent")
	return nil
}
