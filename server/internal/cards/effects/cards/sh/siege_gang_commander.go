package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Siege-Gang Commander — Creature — Goblin {3}{R}{R}, 2/2 (EDHREC
// rank 1590):
//
//	"When this creature enters, create three 1/1 red Goblin creature
//	 tokens.
//	 {1}{R}, Sacrifice a Goblin: This creature deals 2 damage to any
//	 target."
//
// Four bodies and a sac outlet. The ETB is a real trigger with a
// response window; the Goblins are Krenko's template. The sacrifice
// is Skirk Prospector's "a Goblin" clause — effective subtypes, so a
// changeling counts and the Commander can eat itself, as printed —
// paid at announce, so a Goblin's dies-triggers resolve above the
// damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ddc7f59a-bbb1-4ba1-82c8-6813fd191940",
		Name:         "Siege-Gang Commander",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Siege-Gang Commander — create three 1/1 red Goblins", effects.Do(effects.CreateToken{Template: effects.RedGoblinToken(), N: 3})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, Sacrifice a Goblin: This creature deals 2 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), game.AbilityCost{SacrificeOther: effects.B12SacrificeAGoblin()}),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return ctx.Game.DamageInstanceForEffect(func() error {
					for _, t := range ctx.LegalTargets() {
						if err := (effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 2}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				})
			},
		}},
	})
}
