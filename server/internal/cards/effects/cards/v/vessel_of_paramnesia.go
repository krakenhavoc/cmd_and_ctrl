package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vessel of Paramnesia — Enchantment {1}{U} (Shadows over Innistrad,
// #5543):
//
//	"{U}, Sacrifice this enchantment: Target player mills three cards.
//	 Draw a card."
//
// The sacrifice is a cost (CR 602.2c), so the ability resolves even though
// the enchantment is gone. The target is any player, yourself included; if
// it is illegal as the ability resolves the whole ability does nothing,
// the draw included (CR 608.2b). The controller draws, not the target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f6894486-0bac-4b72-92b7-5ce99fa426e6",
		Name:         "Vessel of Paramnesia",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, Sacrifice this enchantment: Target player mills three cards. Draw a card.",
			Purpose: game.Purpose{Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.SacrificeThis()),
			Targets: effects.TargetPlayer("target player"),
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				player, ok := effects.FirstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				if err := (effects.MillCards{Player: player, N: 3}).Apply(ctx); err != nil {
					return err
				}
				return effects.DrawCards{N: 1}.Apply(ctx)
			})),
		}},
	})
}
