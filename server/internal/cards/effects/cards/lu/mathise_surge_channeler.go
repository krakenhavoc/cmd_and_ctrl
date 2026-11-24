package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mathise, Surge Channeler — Legendary Creature — Human Elf Shaman Sorcerer
// {2}{U}, 1/1 (Secret Lair Drop):
//
//	"Whenever you cast an instant or sorcery spell with mana value 3 or
//	 greater, roll a d20.
//	 1-9 | Each player draws a card.
//	 10-19 | You draw a card.
//	 20 | Copy that spell. You may choose new targets for the copy."
//
// The spell's mana value is read as it is cast (CR 202.3's value on the stack,
// so X counts). The spell rides the roll's tail as data, and the copy of "that
// spell" is made from last-known information if the spell has left the stack in
// the meantime (it names the spell and does not target it); the copy is not
// cast, so it does not trigger Mathise again. A roll of 1-9 has each player
// draw in turn order from the active player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "61be7df8-cc78-4822-9e70-f226a8a96c9d",
		Name:         "Mathise, Surge Channeler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(mathiseBigInstantOrSorcery, "Mathise, Surge Channeler — roll a d20",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return effects.RollDice{Sides: 20, N: 1, Then: mathiseRolled,
						TailCarry: []uuid.UUID{item.Trigger.Event.CardID}}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// mathiseBigInstantOrSorcery is "an instant or sorcery spell with mana value 3
// or greater".
func mathiseBigInstantOrSorcery(g *game.Game, _ uuid.UUID, c game.Card) bool {
	if !c.IsInstant() && !c.IsSorcery() {
		return false
	}
	mv, ok := g.ManaValueForEffect(c)
	return ok && mv >= 3
}

// mathiseRolled is the results table. An on-disk key: never renamed, never
// reused.
var mathiseRolled = effects.CardTail("dice/mathise-surge-channeler-rolled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Rolls) == 0 {
		return nil
	}
	switch roll := o.Rolls[0]; {
	case roll >= 20:
		if len(o.Tail.Carry) == 0 {
			return nil
		}
		return effects.CopySpell{
			StackID:          o.Tail.Carry[0],
			Controller:       ctx.Controller(),
			FromLastKnown:    true,
			ChooseNewTargets: true,
		}.Apply(ctx)
	case roll >= 10:
		return effects.DrawCards{N: 1}.Apply(ctx)
	default:
		for _, p := range effects.SeatsFromActive(ctx.Game) {
			if err := (effects.DrawCards{Player: p, N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
})
