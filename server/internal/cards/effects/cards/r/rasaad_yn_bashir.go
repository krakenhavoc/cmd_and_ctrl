package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rasaad yn Bashir — Legendary Creature — Human Monk {2}{W}, 0/3 (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Each creature you control assigns combat damage equal to its toughness
//	 rather than its power.
//	 Whenever Rasaad yn Bashir attacks, if you have the initiative, double
//	 the toughness of each creature you control until end of turn.
//	 Choose a Background (You can have a Background as a second commander.)"
//
// The first line is Baldin, Century Herdmaster's Characteristic.
// DamageByToughnessAlways (CR 510.1c), on the creatures its controller
// controls at all times. The attack trigger is intervening-if (CR 603.4) on
// the initiative: each creature you control as it resolves gets +0/+T, T
// being its toughness as it resolves, which doubles it (CR 701.10a) until
// end of turn. Choose a Background is deck construction (ADR 0144).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3848ad74-8ee7-4546-9394-30db88e1d1e6",
		Name:         "Rasaad yn Bashir",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Each creature you control assigns combat damage equal to its toughness rather than its power.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.DamageByToughnessAlways = true
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.IfYouHaveTheInitiative(effects.WheneverThisAttacks("Rasaad yn Bashir — double the toughness of each creature you control until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					g.RecomputeLayersIfStaleLocked()
					type boost struct {
						id uuid.UUID
						t  int
					}
					var boosts []boost
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() && c.CurrentToughness() > 0 {
							boosts = append(boosts, boost{c.InstanceID, c.CurrentToughness()})
						}
					}
					for _, b := range boosts {
						if err := (effects.BoostUntilEOT{Target: b.id, Toughness: b.t, Label: "Rasaad yn Bashir — double toughness"}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
					return nil
				})),
		},
	})
}
