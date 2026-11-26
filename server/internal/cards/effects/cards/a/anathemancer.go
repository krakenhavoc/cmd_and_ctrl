package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Anathemancer — Creature — Zombie Wizard {1}{B}{R}, 2/2 (Alara Reborn, #5586):
//
//	"When this creature enters, it deals damage to target player equal to the
//	 number of nonbasic lands that player controls.
//	 Unearth {5}{B}{R}"
//
// The lands are counted as the trigger resolves (CR 608.2h), through the layers
// (a land that is basic only by an effect is read as it is now); a basic land is
// one with the basic supertype. The Anathemancer itself deals the damage, from
// the battlefield or, when it has left, as it last was. Unearth is the keyword's
// own sorcery-speed graveyard ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "137b5708-50a9-4f4c-84e6-9043be9f1989",
		Name:         "Anathemancer",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Unearth("{5}{B}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Anathemancer — it deals damage to target player equal to the number of nonbasic lands that player controls",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						victim := effects.TargetedPlayer(ctx)
						if victim == uuid.Nil {
							return nil
						}
						n := 0
						for _, c := range g.BattlefieldCardsForEffect() {
							if c.Controller == victim && c.IsLand() && !effects.IsBasicLand(c) {
								n++
							}
						}
						if n <= 0 {
							return nil
						}
						return effects.DealDamage{Source: ctx.Source(), Target: victim, Amount: n}.Apply(ctx)
					}),
				effects.TargetPlayer("target player")),
		},
	})
}
