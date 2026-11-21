package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Monstrosity of the Lake — Legendary Creature — Kraken {4}{U}, 4/6 (Tales
// of Middle-earth Commander):
//
//	"When Monstrosity of the Lake enters, you may pay {5}. If you do, tap
//	 all creatures your opponents control, then put a stun counter on each
//	 of those creatures. (If a permanent with a stun counter would become
//	 untapped, remove one from it instead.)
//	 Islandcycling {2} ({2}, Discard this card: Search your library for an
//	 Island card, reveal it, put it into your hand, then shuffle.)"
//
// Bitter Chill's "you may pay": the {5} is asked as the ability resolves and
// the paid branch runs only if it was really paid. The creatures are those
// your opponents control as the branch runs, each tapped and then given a
// stun counter, whether or not it was already tapped (a tapped creature with
// a stun counter loses the counter instead of untapping, which is the
// engine's own stun rule). Islandcycling is Jhessian Zombies'.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5f52c7f-5dba-4d4c-8acc-068732964031",
		Name:         "Monstrosity of the Lake",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Islandcycling", "{2}", "an Island card", effects.IsLandWithSubtype("Island")),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Monstrosity of the Lake — you may pay {5}; if you do, tap all creatures your opponents control, then put a stun counter on each of them",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{5}",
						Question: "Monstrosity of the Lake — pay {5} to tap all creatures your opponents control and stun them?",
						Then:     s82GapB118MonstrosityPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var s82GapB118MonstrosityPaid = effects.PayUnlessBranches("pay-unless/monstrosity-of-the-lake-stun",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		var victims []uuid.UUID
		for _, c := range ctx.Game.BattlefieldCardsForEffect() {
			if c.IsCreature() && c.Controller != ctx.Controller() {
				victims = append(victims, c.InstanceID)
			}
		}
		for _, id := range victims {
			if err := (effects.TapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		for _, id := range victims {
			if err := (effects.AddCounter{Target: id, Kind: game.CounterStun, N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}, nil)
