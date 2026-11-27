package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dong Zhou, the Tyrant — Legendary Creature — Human Soldier {4}{R}, 3/3 (Portal
// Three Kingdoms, S82 #5622):
//
//	"When Dong Zhou enters, target creature an opponent controls deals damage equal
//	 to its power to that player."
//
// Bite Down's one-sided deal with the creature as the source: the target creature
// deals the damage (so its lifelink, deathtouch or infect count), equal to its
// power as the trigger resolves, to its controller as it then is ("that player").
// A target that is illegal by then deals nothing (CR 608.2b), and a creature with
// no power deals 0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1da64bc-8cff-4c89-9b9c-db0ddff5abc4",
		Name:         "Dong Zhou, the Tyrant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Dong Zhou — target creature an opponent controls deals damage equal to its power to that player",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					c, ok := g.LookupCardForEffect(t.ID)
					if !ok || c.CurrentPower() <= 0 {
						return nil
					}
					return effects.DealDamage{Source: t.ID, Target: c.Controller, Amount: c.CurrentPower()}.Apply(ctx)
				}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
