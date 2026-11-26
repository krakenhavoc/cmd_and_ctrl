package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lightmine Field — Enchantment {2}{W}{W} (Rise of the Eldrazi, S82 OLD
// batch 483, #5581):
//
//	"Whenever one or more creatures attack, this enchantment deals damage to
//	 each of those creatures equal to the number of attacking creatures."
//
// One trigger per declaration (OncePerBatch, CR 603.2c) whoever attacks. The
// attackers are read as it resolves, in the declare attackers step before
// anything else can join the attack: each takes damage equal to how many
// creatures are attacking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "20490b29-4f08-48a2-bf54-2ba118e746ac",
		Name:         "Lightmine Field",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventAttack
			}, "Lightmine Field — deals damage to each attacking creature equal to the number of attacking creatures",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					var attackers []game.Card
					for _, c := range ctx.Game.BattlefieldCardsForEffect() {
						if c.IsCreature() && c.AttackingTarget != uuid.Nil {
							attackers = append(attackers, c)
						}
					}
					// One printed sentence is one instance of damage (CR 615.8).
					return ctx.Game.DamageInstanceForEffect(func() error {
						for _, c := range attackers {
							if err := (effects.DealDamage{Source: ctx.Source(), Target: c.InstanceID, Amount: len(attackers)}).Apply(ctx.AsGroupMember()); err != nil {
								return err
							}
						}
						return nil
					})
				})))),
		},
	})
}
