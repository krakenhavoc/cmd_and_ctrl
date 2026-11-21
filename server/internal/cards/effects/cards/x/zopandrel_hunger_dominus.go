package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zopandrel, Hunger Dominus — Legendary Creature — Phyrexian Horror
// {5}{G}{G}, 4/6:
//
//	"Reach
//	 At the beginning of each combat, double the power and toughness
//	 of each creature you control until end of turn.
//	 {G/P}{G/P}, Sacrifice two other creatures: Put an indestructible
//	 counter on Zopandrel. ({G/P} can be paid with either {G} or 2
//	 life.)"
//
// "Each combat" is every player's (AtBeginningOfEachCombat). Doubling
// is CR 701.10b per creature: each creature you control as the
// trigger resolves gets +X/+Y until end of turn, X and Y its own power
// and toughness then, a negative power doubling downward (CR 701.10c).
// The set is locked as it resolves (CR 611.2c), so a creature that
// arrives later in the turn is not doubled. Two Zopandrels double
// twice (the 2023-02-04 ruling), because each trigger reads the board
// the one before it left.
//
// The activated ability is Drivnod's: Phyrexian mana (each {G/P} paid
// with {G} or 2 life, announced with the activation) plus a sacrifice
// of exactly two creatures other than Zopandrel itself, paid at
// announce. The indestructible counter is a keyword counter (CR
// 122.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b168e4e6-f572-4d9f-b98f-95b2611354cb",
		Name:            "Zopandrel, Hunger Dominus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfEachCombat("Zopandrel, Hunger Dominus — double the power and toughness of each creature you control",
				func(g *game.Game, item *game.StackItem) error {
					var mine []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() {
							mine = append(mine, c.InstanceID)
						}
					}
					return effects.DoublePowerAndToughnessOfEachUntilEOT(effects.NewContext(g, item), mine,
						"Zopandrel, Hunger Dominus — double power and toughness")
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G/P}{G/P}, Sacrifice two other creatures: Put an indestructible counter on Zopandrel.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{G/P}{G/P}"), effects.SacrificeAnotherN(2, "two other creatures", effects.Creature())),
			Effect:  effects.PutIndestructibleCounterOnSource(),
		}},
	})
}
