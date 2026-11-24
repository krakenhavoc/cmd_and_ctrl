package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Meep — Legendary Creature — Alien {2}{B}, 0/4:
//
//	"Ward—Pay 3 life.
//	 Whenever The Meep attacks, you may sacrifice another creature. If you do,
//	 creatures you control have base power and toughness X/X until end of turn,
//	 where X is the sacrificed creature's mana value."
//
// The ward is the life payment (WardLife). The attack trigger asks for another
// creature to sacrifice (the Meep itself is not offered) and, when one really
// left, sets the base power and toughness of the creatures you control as the
// ability resolves (CR 611.2c: the set is locked then, so a creature that enters
// later this turn keeps its own) to X/X, layer 7b, until end of turn, where X is
// the sacrificed creature's mana value as it last existed on the battlefield
// (CR 608.2h): a token that is no copy is 0, so everything you control becomes
// 0/0 and dies. Counters, pumps and anthems still land on top of the base.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff2448d0-7a45-45be-bf02-19650f490b54",
		Name:         "The Meep",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardLife(3), "The Meep — ward—pay 3 life"),
			effects.WheneverThisAttacks("The Meep — you may sacrifice another creature; if you do, creatures you control have base power and toughness X/X until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "The Meep — sacrifice another creature?", effects.Creature(), true, theMeepSacrificed)
				}),
		},
	})
}

// theMeepSacrificed is "if you do": the creatures you control become X/X, X
// the sacrificed creature's mana value. An on-disk key.
var theMeepSacrificed = effects.RegisterSacrificedOne("prompt-run/the-meep-base-power-and-toughness", func(ctx *effects.Context, sacrificed uuid.UUID) error {
	x := effects.DepartedManaValue(ctx.Game, sacrificed)
	return effects.ScopedEffectFor{
		Match: effects.And(effects.Creature(), effects.YouControl()),
		Mods: []game.Mod{
			game.SetBasePowerMod(x),
			game.SetBaseToughnessMod(x),
		},
		Duration: effects.DurationUntilEndOfTurn(ctx),
		Label:    "The Meep — creatures you control have base power and toughness X/X",
	}.Apply(ctx)
})
