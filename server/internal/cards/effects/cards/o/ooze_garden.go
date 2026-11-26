package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ooze Garden — Enchantment {1}{G} (Shards of Alara, S82 OLD batch 510,
// #5588):
//
//	"{1}{G}, Sacrifice a non-Ooze creature: Create an X/X green Ooze creature
//	 token, where X is the sacrificed creature's power. Activate only as a
//	 sorcery."
//
// X is the sacrificed creature's power as it last existed on the battlefield
// (CR 608.2h; ctx.SacrificedPower), set on the token as it is created.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed1e065d-93b8-4017-9e12-427443548fc5",
		Name:         "Ooze Garden",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}{G}, Sacrifice a non-Ooze creature: Create an X/X green Ooze creature token, where X is the sacrificed creature's power. Activate only as a sorcery.",
			Purpose:      game.Purpose{Tokens: 1},
			Cost:         effects.Plus(effects.ManaCost("{1}{G}"), effects.SacrificeN(1, "a non-Ooze creature", effects.Creature(), effects.Not(effects.HasSubtype("Ooze")))),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := max(ctx.SacrificedPower(), 0)
				tok := ctx.TextToken(effects.TokenCard("0/0 green Ooze"))
				tok.Power, tok.Toughness = x, x
				return g.CreateTokenForEffect(ctx.Controller(), tok, 1)
			},
		}},
	})
}
