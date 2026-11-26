package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// King Macar, the Gold-Cursed — Legendary Creature — Human Noble {2}{B}{B}, 2/3
// (Journey into Nyx, S82 OLD, #5558):
//
//	"Inspired — Whenever King Macar becomes untapped, you may exile target
//	 creature. If you do, create a Gold token. (It's an artifact with
//	 "Sacrifice this token: Add one mana of any color.")"
//
// Vizier of Deferment's "you may exile target creature" on the untap watch; the
// Gold token (Gild's predefined token) is made only for a creature that really
// reached exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ebafd78-269a-4974-88d2-e75b44b744c2",
		Name:         "King Macar, the Gold-Cursed",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.On(game.EventUntapCard, effects.Self,
					"King Macar, the Gold-Cursed — you may exile target creature; if you do, create a Gold token (inspired)",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						t, ok := effects.FirstLegalTarget(ctx)
						if !ok || t.Kind != game.TargetCard {
							return nil
						}
						return effects.ExileTarget{Target: t.ID, ThenTail: ctx.ItemTail(kingMacarGold)}.Apply(ctx)
					}),
				effects.TargetCreature("target creature")),
				"King Macar, the Gold-Cursed — exile target creature?"),
		},
	})
}

// kingMacarGold is the "if you do": a Gold token for the creature that reached
// exile. An on-disk key.
var kingMacarGold = effects.CardTail("then/s82-old-b371-king-macar-gold", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.CreateToken{Template: effects.GoldToken(), N: 1}.Apply(ctx)
})
