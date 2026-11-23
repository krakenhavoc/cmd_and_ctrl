package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Frontline Medic — Creature — Human Cleric {2}{W}, 3/3 (Tales of
// Middle-earth Commander):
//
//	"Battalion — Whenever this creature and at least two other creatures
//	 attack, creatures you control gain indestructible until end of turn.
//	 Sacrifice this creature: Counter target spell with {X} in its mana
//	 cost unless its controller pays {3}."
//
// Boros Elite's battalion trigger (CR 702.101a-style: the Medic and two
// others, any controllers' attackers counted as the printed word says
// "other creatures" you control); the grant is Beregond's mass grant, read
// as the ability resolves. The counter is Miscalculation's pay-unless, aimed
// at a spell whose own printed mana cost has an {X} (an {X} in an additional
// or alternative cost does not count); its controller is asked to pay {3},
// and the Medic was sacrificed as a cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "36a3c910-95b8-45a2-9f84-5681336d868c",
		Name:         "Frontline Medic",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.GtcBattalion("Frontline Medic — creatures you control gain indestructible until end of turn (battalion)",
				effects.Do(effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Keywords: []string{"indestructible"}, Label: "Frontline Medic — indestructible"})),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "Sacrifice this creature: Counter target spell with {X} in its mana cost unless its controller pays {3}.",
			Cost:  effects.SacrificeThis(),
			Targets: effects.TargetSpell("target spell with {X} in its mana cost", func(g *game.Game, _ uuid.UUID, c game.Card) bool {
				return effects.SocB9HasXInManaCost(g, c.InstanceID)
			}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				return effects.CounterUnlessPaid{
					StackID:  t.ID,
					Cost:     "{3}",
					Question: "Frontline Medic — pay {3} or your spell is countered",
				}.Apply(ctx)
			},
		}},
	})
}
