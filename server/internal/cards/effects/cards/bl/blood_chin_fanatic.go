package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood-Chin Fanatic — Creature — Orc Warrior {1}{B}{B}, 3/3 (Dragons of Tarkir, #5551):
//
//	"{1}{B}, Sacrifice another Warrior creature: Target player loses X life and you gain X life, where X is the sacrificed creature's power."
//
// Brion Stoutarm's sacrifice-for-its-power shape: X is the sacrificed
// creature's last-known power (CR 608.2h), floored at zero (CR 107.1b).
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "b704f231-f64c-49fc-84f6-0fc53081bcd0",
		Name:         "Blood-Chin Fanatic",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice another Warrior creature: Target player loses X life and you gain X life, where X is the sacrificed creature's power.",
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.SacrificeAnotherN(1, "another Warrior creature", effects.And(effects.Creature(), effects.HasSubtype("Warrior")))),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := ctx.SacrificedPower()
				player, ok := effects.FirstLegalPlayerTarget(ctx)
				if !ok || x <= 0 {
					return nil
				}
				if err := g.ChangePlayerLifeForEffect(ctx.Source(), player, -x); err != nil {
					return err
				}
				return effects.GainLife{Amount: x}.Apply(ctx)
			},
		}},
	})
}
