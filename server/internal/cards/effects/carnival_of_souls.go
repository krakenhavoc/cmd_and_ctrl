package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Carnival of Souls — Enchantment {1}{B}:
//
//	"Whenever a creature enters, you lose 1 life and add {B}."
//
// An ORDINARY triggered ability that goes on the stack and can be
// answered. It is not a triggered mana ability (CR 605.1b): that needs
// a trigger that resolves off a mana ability's activation, and a
// creature entering is not one. The mana lands in the pool when the
// trigger resolves, so it empties at the end of that step (CR 106.4).
//
// Any player's creature, card or token, counts. The two halves run in
// printed order: the life loss first, and the mana in its continuation,
// so a life-loss replacement that pauses on an ordering prompt (CR 616)
// holds the mana back until it has settled. The mana is added even
// when the loss was replaced away: the sentence says "and", not "if you
// do".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "95b10ca7-7360-4da5-bd93-686ae3051833",
		Name:         "Carnival of Souls",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b33AnyCreatureEntered(ev, g)
			}, "Carnival of Souls — you lose 1 life and add {B}", carnivalOfSoulsLoseLifeAddMana),
		},
	})
}

// carnivalOfSoulsLoseLifeAddMana loses the life, then adds the mana
// once the change has settled.
func carnivalOfSoulsLoseLifeAddMana(g *game.Game, item *game.StackItem) error {
	return g.ChangePlayerLifeKeyedForEffect(item.SourceCardID, item.Controller, -1, NewContext(g, item).ItemTail(carnivalOfSoulsAddMana))
}

// carnivalOfSoulsAddMana is "and add {B}", once the life loss has settled. An
// on-disk key: never renamed, never reused.
var carnivalOfSoulsAddMana = CardTail("carnival-of-souls/add-mana", func(ctx *Context, _ game.TailOutcome) error {
	return AddMana{Player: ctx.Controller(), Produced: "{B}"}.Apply(ctx)
})
