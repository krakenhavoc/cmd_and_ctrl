package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Papalymo Totolymo — Legendary Creature — Dwarf Wizard {W}{B}, 1/2:
//
//	"Whenever you cast a noncreature spell, Papalymo Totolymo deals 1
//	 damage to each opponent and you gain 1 life.
//	 {4}, {T}, Sacrifice Papalymo Totolymo: Each opponent who lost life
//	 this turn sacrifices a creature with the greatest power among
//	 creatures they control."
//
// The drain is Y'shtola's shape at 1. The sacrifice is paid for as a
// cost, then the opponents who have lost life this turn (damage or loss,
// b18LifeLostThisTurn) are read as the ability resolves; each of them
// chooses among their creatures with the greatest power (ties are theirs
// to break) and the choices are sacrificed together (CR 101.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "334e83cd-b513-4bd4-89a2-86875667e6c4",
		Name:         "Papalymo Totolymo",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Papalymo Totolymo — 1 damage to each opponent, you gain 1 life",
				damageEachOpponentThenGainLife(1)),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}, Sacrifice Papalymo Totolymo: Each opponent who lost life this turn sacrifices a creature with the greatest power among creatures they control.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    Plus(ManaCost("{4}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				var losers []uuid.UUID
				for _, opp := range ctx.Opponents() {
					if b18LifeLostThisTurn(g, opp) > 0 {
						losers = append(losers, opp)
					}
				}
				if len(losers) == 0 {
					return nil
				}
				return g.PlayersSacrificeThenForEffect(item.SourceCardID, losers,
					sacrificeSpec("a creature with the greatest power among creatures you control", GreatestPowerYouControl()),
					"Papalymo Totolymo — sacrifice a creature with the greatest power",
					func(*game.Game, game.PromptedSacrifices) error { return nil })
			},
		}},
	})
}
