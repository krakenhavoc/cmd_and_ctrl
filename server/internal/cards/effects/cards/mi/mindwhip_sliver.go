package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const mindwhipSliverGrant = "mindwhip-sliver/discard-at-random"

// Mindwhip Sliver — Creature — Sliver {2}{B}, 2/2 (Tempest, S82 OLD, #5628):
//
//	"All Slivers have "{2}, Sacrifice this permanent: Target player discards a
//	 card at random. Activate only as a sorcery.""
//
// Crypt Sliver's shape: one activated ability granted to every Sliver, any
// player's, each paying for itself. The sacrifice is a cost, so the discard
// still happens when the Sliver is gone; the target is re-checked at
// resolution. The discard is random (CR 701.9b), not a choice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9e43839-7cc1-40ff-a8ec-a47316de6440",
		Name:         "Mindwhip Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: mindwhipSliverGrant,
			Activated: []effects.ActivatedAbility{{
				Label:        "{2}, Sacrifice this permanent: Target player discards a card at random. Activate only as a sorcery.",
				Cost:         effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
				SorcerySpeed: true,
				Targets:      effects.TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return effects.DiscardCards{Player: t.ID, N: 1}.Apply(ctx)
				},
			}},
			Text: "{2}, Sacrifice this permanent: Target player discards a card at random. Activate only as a sorcery.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, mindwhipSliverGrant),
		},
	})
}
