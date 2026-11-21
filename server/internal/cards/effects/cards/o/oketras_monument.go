package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oketra's Monument — Legendary Artifact {3}:
//
//	"White creature spells you cast cost {1} less to cast.
//	 Whenever you cast a creature spell, create a 1/1 white Warrior
//	 creature token with vigilance."
//
// Kefnet's Monument's white sibling, and the same two halves: a
// CR 601.2f reduction keyed on colour and card type, and a cast
// trigger that fires on ANY creature spell you cast — the token is
// not restricted to white creatures, only the discount is.
//
// The trigger is a cast trigger (CR 603.2 on EventCast), so the
// Warrior arrives while the creature spell is still on the stack and
// resolves first; countering the creature does not take the token
// back.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0370afa0-07d3-4787-8a5b-10272cb3a486",
		Name:         "Oketra's Monument",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "White creature spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.CreatureSpell(), effects.ColoredSpell("W")),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCast(effects.Creature()), "Oketra's Monument — create a 1/1 white Warrior with vigilance",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Warrior with vigilance"), N: 1})),
		},
	})
}
